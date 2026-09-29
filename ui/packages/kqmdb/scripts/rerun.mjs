// Resumable upgrade runner. Results stay local unless --publish is supplied.
import { parseArgs } from "node:util";
import { createHash } from "node:crypto";
import {
	existsSync,
	mkdirSync,
	readFileSync,
	writeFileSync,
	renameSync,
	rmSync,
} from "node:fs";
import { resolve, join } from "node:path";
import { spawnSync } from "node:child_process";
import { resultSummary } from "../worker/submissions.mjs";

const { values } = parseArgs({
	options: {
		origin: { type: "string", default: "http://localhost:8787" },
		out: { type: "string" },
		sim: { type: "string" },
		publish: { type: "boolean", default: false },
		limit: { type: "string" },
		timeout: { type: "string", default: "600" },
	},
});
if (!values.out || !values.sim)
	throw new Error(
		"Supply --out for saved results and --sim for the simulator binary.",
	);
const origin = new URL(values.origin);
if (
	!["https://db.kqm.gg", "http://localhost:8787"].includes(origin.origin) ||
	origin.username ||
	origin.password
)
	throw new Error("Use the KQM database or localhost origin.");
const token = process.env.SYNC_TOKEN;
if (!token)
	throw new Error(
		"Set SYNC_TOKEN, or load the existing .dev.vars file with node --env-file.",
	);
const timeout = Number(values.timeout) * 1000;
const limit = values.limit ? Number(values.limit) : Infinity;
if (
	!Number.isSafeInteger(timeout) ||
	timeout < 1000 ||
	!(limit === Infinity || (Number.isSafeInteger(limit) && limit > 0))
)
	throw new Error("Invalid timeout or limit.");
const out = resolve(values.out),
	sim = resolve(values.sim);
mkdirSync(out, { recursive: true, mode: 0o700 });
const sha = (text) => createHash("sha256").update(text).digest("hex");
const binaryHash = sha(readFileSync(sim));
const manifestPath = join(out, "manifest.json");
async function api(path, body) {
	const response = await fetch(new URL(path, origin), {
		method: body ? "POST" : "GET",
		redirect: "error",
		headers: {
			Authorization: `Bearer ${token}`,
			...(body ? { "Content-Type": "application/json" } : {}),
		},
		body: body ? JSON.stringify(body) : undefined,
		signal: AbortSignal.timeout(60000),
	});
	const result = await response.json();
	if (!response.ok)
		throw new Error(
			`HTTP ${response.status}: ${result.error ?? "Request failed"}`,
		);
	return result;
}
let manifest;
if (existsSync(manifestPath)) {
	manifest = JSON.parse(readFileSync(manifestPath, "utf8"));
	if (manifest.origin !== origin.origin || manifest.binaryHash !== binaryHash)
		throw new Error(
			"Use a new output folder when the server or simulator build changes.",
		);
} else {
	const rows = [];
	for (const table of ["simulations", "submissions"]) {
		let after = "";
		while (true) {
			const page = await api(
				`/api/admin/export?table=${table}&after=${encodeURIComponent(after)}`,
			);
			rows.push(...page.rows);
			if (!page.cursor) break;
			after = page.cursor;
		}
	}
	manifest = {
		origin: origin.origin,
		binaryHash,
		createdAt: new Date().toISOString(),
		rows,
	};
	writeFileSync(manifestPath, JSON.stringify(manifest), { mode: 0o600 });
}
const report = {
	total: manifest.rows.length,
	completed: 0,
	published: 0,
	failed: [],
};
for (const row of manifest.rows.slice(0, limit)) {
	const key = `${row.table}-${row.id}`;
	if (!/^[a-zA-Z0-9_-]+$/.test(key)) throw new Error("Invalid exported ID");
	const resultPath = join(out, `${key}.json`);
	const temporaryResult = `${resultPath}.partial`;
	const receipt = join(out, `${key}.published.json`);
	try {
		if (!existsSync(resultPath)) {
			// Config imports can read local files. Database jobs must be self-contained.
			if (/^\s*#?\s*import\b/im.test(row.entry.config))
				throw new Error(
					"Config imports are not allowed in database upgrade jobs.",
				);
			const configPath = join(out, `${key}.txt`);
			writeFileSync(configPath, row.entry.config, { mode: 0o600 });
			const run = spawnSync(sim, ["-c", configPath, "-out", temporaryResult], {
				timeout,
				encoding: "utf8",
				maxBuffer: 8_000_000,
			});
			if (run.error || run.status !== 0)
				throw new Error(
					run.error?.message ??
						run.stderr?.slice(-1500) ??
						`Simulator exited ${run.status}`,
				);
		}
		const prepared = existsSync(resultPath) ? resultPath : temporaryResult;
		const result = JSON.parse(readFileSync(prepared, "utf8"));
		resultSummary(result);
		if (
			result.config_file?.replace(/\r\n/g, "\n").trimEnd() !==
			row.entry.config.replace(/\r\n/g, "\n").trimEnd()
		)
			throw new Error("Result config does not match the exported config.");
		if (prepared === temporaryResult) renameSync(temporaryResult, resultPath);
		report.completed++;
		if (values.publish) {
			if (!existsSync(receipt)) {
				const saved = await api("/api/admin/rerun", {
					...row,
					engine: result.sim_version,
					result,
				});
				writeFileSync(receipt, JSON.stringify(saved), { mode: 0o600 });
			}
			report.published++;
		}
	} catch (error) {
		rmSync(temporaryResult, { force: true });
		report.failed.push({ table: row.table, id: row.id, error: error.message });
	}
	writeFileSync(join(out, "report.json"), JSON.stringify(report, null, 2), {
		mode: 0o600,
	});
	console.log(
		`${report.completed} complete, ${report.published} published, ${report.failed.length} failed`,
	);
}
if (report.failed.length) process.exitCode = 1;
