import { ID_PATTERN } from "./storage.mjs";
import { resultSummary, SubmissionError } from "./submissions.mjs";

const json = (value, status = 200) =>
	Response.json(value, {
		status,
		headers: {
			"Cache-Control": "no-store",
			"X-Content-Type-Options": "nosniff",
		},
	});
export async function digest(text) {
	return Array.from(
		new Uint8Array(
			await crypto.subtle.digest("SHA-256", new TextEncoder().encode(text)),
		),
		(b) => b.toString(16).padStart(2, "0"),
	).join("");
}
function tableName(value) {
	if (!["simulations", "submissions"].includes(value))
		throw new SubmissionError("Invalid record type.");
	return value;
}

export async function applyUpgrade(input, env) {
	const table = tableName(input.table);
	if (
		!ID_PATTERN.test(input.id ?? "") ||
		!/^[a-f0-9]{64}$/.test(input.previous ?? "") ||
		!/^[a-f0-9]{40}$/.test(input.engine ?? "")
	)
		throw new SubmissionError("Invalid upgrade request.");
	const row = await env.DB.prepare(`SELECT * FROM ${table} WHERE id=?`)
		.bind(input.id)
		.first();
	if (!row) throw new SubmissionError("Record not found.", 404);
	const original = JSON.parse(row.document);
	if ((await digest(row.document)) !== input.previous)
		throw new SubmissionError(
			"The record changed. Export it again before retrying.",
			409,
		);
	const result = input.result;
	if (
		!result ||
		typeof result.config_file !== "string" ||
		result.config_file.replace(/\r\n/g, "\n").trimEnd() !==
			original.config.replace(/\r\n/g, "\n").trimEnd() ||
		result.sim_version !== input.engine ||
		result.modified
	)
		throw new SubmissionError(
			"The result must use this record's config and the selected clean simulator build.",
			422,
		);
	const summary = resultSummary(result);
	const text = JSON.stringify(result);
	const shareKey = `kqm-${await digest(text)}`;
	const now = Date.now();
	const updated = {
		...original,
		summary,
		hash: input.engine,
		share_key: shareKey,
		last_update: Math.floor(now / 1000),
	};
	const document = JSON.stringify(updated);
	const historyId = await digest(
		`${table}:${input.id}:${input.previous}:${shareKey}`,
	);
	await env.FILES.put(`results/${shareKey}.json`, text, {
		httpMetadata: { contentType: "application/json" },
	});
	// D1 batch is transactional. The document predicate prevents an in-flight review
	// or another upgrade from being overwritten. Approval and tags are unchanged.
	const history =
		env.DB.prepare(`INSERT INTO upgrade_history(id,target_table,target_id,engine_hash,previous_row,new_document,created_at)
    SELECT ?,?,?,?,?,?,? FROM ${table} WHERE id=? AND document=?`).bind(
			historyId,
			table,
			input.id,
			input.engine,
			JSON.stringify(row),
			document,
			now,
			input.id,
			row.document,
		);
	const update =
		table === "simulations"
			? env.DB.prepare(
					"UPDATE simulations SET document=?,dps=?,duration=?,imported_at=?,source='local' WHERE id=? AND document=?",
				).bind(
					document,
					summary.mean_dps_per_target,
					summary.sim_duration.mean,
					now,
					input.id,
					row.document,
				)
			: env.DB.prepare(
					"UPDATE submissions SET document=? WHERE id=? AND document=?",
				).bind(document, input.id, row.document);
	const results = await env.DB.batch([history, update]);
	if (!results[1].meta.changes)
		throw new SubmissionError(
			"The record changed. Export it again before retrying.",
			409,
		);
	return { id: input.id, table, engine: input.engine, shareKey };
}

export async function handleUpgrades(request, env) {
	const url = new URL(request.url);
	try {
		if (url.pathname === "/api/admin/export" && request.method === "GET") {
			const table = tableName(url.searchParams.get("table"));
			const after = url.searchParams.get("after") ?? "";
			if (after && !ID_PATTERN.test(after))
				throw new SubmissionError("Invalid cursor.");
			const { results } = await env.DB.prepare(
				`SELECT id,document FROM ${table} WHERE id > ? ORDER BY id LIMIT 50`,
			)
				.bind(after)
				.all();
			const rows = await Promise.all(
				results.map(async (row) => ({
					id: row.id,
					table,
					previous: await digest(row.document),
					entry: JSON.parse(row.document),
				})),
			);
			return json({ rows, cursor: rows.at(-1)?.id ?? null });
		}
		if (url.pathname === "/api/admin/rerun" && request.method === "POST") {
			if (!request.headers.get("Content-Type")?.startsWith("application/json"))
				throw new SubmissionError("Use JSON.");
			const reader = request.body?.getReader();
			if (!reader) throw new SubmissionError("Missing result.");
			const chunks = [];
			let size = 0;
			while (true) {
				const { value, done } = await reader.read();
				if (done) break;
				size += value.byteLength;
				if (size > 8_000_000) {
					await reader.cancel();
					throw new SubmissionError("Result too large.", 413);
				}
				chunks.push(value);
			}
			let input;
			try {
				input = JSON.parse(await new Blob(chunks).text());
			} catch {
				throw new SubmissionError("Invalid JSON.");
			}
			return json(await applyUpgrade(input, env));
		}
		return json({ error: "Not found" }, 404);
	} catch (error) {
		if (error instanceof SubmissionError)
			return json({ error: error.message }, error.status);
		console.error("Upgrade failed", error.message);
		return json(
			{
				error: "The upgrade could not finish. The current record is retained.",
			},
			503,
		);
	}
}
