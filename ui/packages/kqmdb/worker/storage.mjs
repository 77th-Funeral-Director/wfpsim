export const ID_PATTERN = /^[a-zA-Z0-9_-]{1,128}$/;
export const UPSERT = `INSERT INTO simulations
  (id, document, create_date, dps, duration, imported_at, seen_run, visible)
  SELECT json_extract(value, '$._id'), value,
    CAST(json_extract(value, '$.create_date') AS INTEGER),
    COALESCE(json_extract(value, '$.summary.mean_dps_per_target'), 0),
    COALESCE(json_extract(value, '$.summary.sim_duration.mean'), 0),
    ?2, ?3, 1
  FROM json_each(?1) WHERE 1
  ON CONFLICT(id) DO UPDATE SET document=excluded.document,
    create_date=excluded.create_date, dps=excluded.dps,
    duration=excluded.duration, imported_at=excluded.imported_at,
    seen_run=excluded.seen_run, visible=1 WHERE simulations.source='upstream'`;

export function validateEntries(entries) {
	if (!Array.isArray(entries) || entries.length > 100)
		throw new Error("Invalid source page");
	for (const entry of entries) {
		if (
			!entry ||
			!ID_PATTERN.test(entry._id) ||
			!ID_PATTERN.test(entry.share_key) ||
			!Number.isFinite(Number(entry.create_date)) ||
			!entry.is_db_valid ||
			typeof entry.config !== "string" ||
			!Array.isArray(entry.summary?.char_names) ||
			!Array.isArray(entry.summary?.team) ||
			!Number.isFinite(entry.summary?.mean_dps_per_target)
		) {
			throw new Error("Invalid simulation in source page");
		}
	}
	return entries;
}

export async function syncPublicDatabase(env, options = {}) {
	// A scheduled import and an admin import can overlap. Only one may move
	// the cursor. A short lease also permits recovery if a Worker is stopped.
	const lease = JSON.stringify({
		owner: crypto.randomUUID(),
		expires: Date.now() + 180000,
	});
	const acquired = await env.DB.prepare(
		"INSERT INTO sync_state(key,value) VALUES ('lock',?1) ON CONFLICT(key) DO UPDATE SET value=excluded.value WHERE json_extract(sync_state.value,'$.expires') < ?2",
	)
		.bind(lease, Date.now())
		.run();
	if (!acquired.meta.changes) return { busy: true };
	try {
		return await importPages(env, options);
	} finally {
		await env.DB.prepare("DELETE FROM sync_state WHERE key='lock' AND value=?")
			.bind(lease)
			.run();
	}
}

async function importPages(env, { force = false, pages = 5 } = {}) {
	const saved = await env.DB.prepare(
		"SELECT value FROM sync_state WHERE key='progress'",
	).first();
	let state = saved ? JSON.parse(saved.value) : {};
	if (!state.run && !force && Date.now() - (state.completedAt ?? 0) < 86400000)
		return state;
	state = {
		...state,
		run: state.run ?? crypto.randomUUID(),
		offset: state.offset ?? 0,
	};
	for (let page = 0; page < pages; page++) {
		const query = JSON.stringify({
			query: {},
			limit: 100,
			skip: state.offset,
			sort: { _id: 1 },
		});
		const response = await fetch(
			`https://simpact.app/api/db?q=${encodeURIComponent(query)}`,
			{
				headers: { Accept: "application/json" },
				signal: AbortSignal.timeout(25000),
			},
		);
		if (!response.ok)
			throw new Error(`Public database returned ${response.status}`);
		const payload = await response.json();
		if (!payload || typeof payload !== "object" || Array.isArray(payload))
			throw new Error("Invalid source response");
		if (!("data" in payload) && Object.keys(payload).length > 0)
			throw new Error("Source response contains no simulation data");
		const entries = validateEntries("data" in payload ? payload.data : []);
		if (state.offset === 0 && entries.length === 0)
			throw new Error("Source is empty; retain the current database");
		const count = entries.length;
		const finished = count < 100;
		const next = finished
			? { offset: 0, completedAt: Date.now(), count: state.offset + count }
			: { ...state, offset: state.offset + count };
		const statements = [
			env.DB.prepare(UPSERT).bind(
				JSON.stringify(entries),
				Date.now(),
				state.run,
			),
		];
		if (finished) {
			// Retain withdrawn records for recovery, but exclude them from public searches.
			statements.push(
				env.DB.prepare(
					"UPDATE simulations SET visible=0 WHERE source='upstream' AND seen_run != ?",
				).bind(state.run),
			);
		}
		statements.push(
			env.DB.prepare(
				"INSERT INTO sync_state(key,value) VALUES ('progress',?) ON CONFLICT(key) DO UPDATE SET value=excluded.value",
			).bind(JSON.stringify(next)),
		);
		await env.DB.batch(statements);
		state = next;
		if (finished) break;
	}
	return state;
}

export async function loadResult(entry, env) {
	// share_key addresses the exact immutable run referenced by the DB entry.
	const key = `results/${entry.share_key}.json`;
	const saved = await env.FILES.get(key);
	if (saved) return saved.body;
	const response = await fetch(
		`https://gcsim.app/api/share/${encodeURIComponent(entry.share_key)}`,
		{
			headers: { Accept: "application/json" },
			signal: AbortSignal.timeout(25000),
		},
	);
	if (!response.ok)
		throw new Error(`Public result returned ${response.status}`);
	const text = await response.text();
	if (text.length > 20000000) throw new Error("Simulation result is too large");
	const result = JSON.parse(text);
	if (
		!Array.isArray(result.character_details) ||
		!result.statistics ||
		typeof result.config_file !== "string"
	)
		throw new Error("Invalid simulation result");
	await env.FILES.put(key, text, {
		httpMetadata: { contentType: "application/json" },
	});
	return text;
}
