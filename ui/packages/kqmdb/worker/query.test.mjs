import assert from "node:assert/strict";
import { readFileSync } from "node:fs";
import { DatabaseSync } from "node:sqlite";
import test from "node:test";
import worker from "./index.mjs";
import { compileQuery, QueryError } from "./query.mjs";
import {
	loadResult,
	syncPublicDatabase,
	UPSERT,
	validateEntries,
} from "./storage.mjs";

const db = new DatabaseSync(":memory:");
const schema =
	readFileSync(
		new URL("../migrations/0001_database.sql", import.meta.url),
		"utf8",
	) +
	readFileSync(
		new URL("../migrations/0002_submissions.sql", import.meta.url),
		"utf8",
	);
db.exec(schema);
const entry = (id, names, tags, dps, description = "Team") => ({
	_id: id,
	share_key: `share${id}`,
	create_date: "1700000000",
	is_db_valid: true,
	config: "options iteration=1;",
	description,
	submitter: "Tester",
	accepted_tags: tags,
	summary: {
		char_names: names,
		team: names.map((name) => ({ name })),
		mean_dps_per_target: dps,
		sim_duration: { mean: 90 },
	},
});
const rows = [
	entry("a", ["nahida", "furina"], [1], 90000, "Nahida's team [test]"),
	entry("b", ["hutao", "furina"], [1, 9], 100000),
	entry("c", ["aetheranemo"], [8], 50000),
];
validateEntries(rows);
db.prepare(UPSERT).run({ 1: JSON.stringify(rows), 2: Date.now(), 3: "run" });
function ids(query, extra = {}) {
	const { sql, params } = compileQuery({ query, ...extra });
	return db
		.prepare(sql)
		.all(...params)
		.map((r) => JSON.parse(r.document)._id);
}
test("character, tag, traveller, and exclusion filters select the right records", () => {
	assert.deepEqual(ids({ "summary.mode": 0 }), ["a", "b", "c"]);
	assert.deepEqual(ids({ "summary.char_names": "nahida" }), ["a"]);
	assert.deepEqual(
		ids({
			$and: [
				{ "summary.char_names": "furina" },
				{ accepted_tags: { $nin: [9] } },
			],
		}),
		["a"],
	);
	assert.deepEqual(ids({ "summary.char_names": { $ne: "furina" } }), ["c"]);
	assert.deepEqual(
		ids({
			$or: [
				{ "summary.char_names": "aetheranemo" },
				{ "summary.char_names": "lumineanemo" },
			],
		}),
		["c"],
	);
});
test("text search preserves punctuation, matches case-insensitively, and cannot inject SQL", () => {
	assert.deepEqual(
		ids({ description: { $regex: "NAHIDA's team \\[test\\]", $options: "i" } }),
		["a"],
	);
	assert.deepEqual(ids({ description: "' OR 1=1 --" }), []);
	assert.throws(() => ids({ "description') OR 1=1 --": "x" }), QueryError);
	assert.throws(() => ids({ description: { $regex: ".*" } }), QueryError);
});
test("DPS sorting and page offsets are stable", () => {
	assert.deepEqual(
		ids({}, { sort: { "summary.mean_dps_per_target": -1 }, limit: 1, skip: 1 }),
		["a"],
	);
	assert.throws(() => ids({}, { limit: 1000 }), QueryError);
});
test("the homepage sample filter selects rows before sorting and limiting", () => {
	const sampled = new DatabaseSync(":memory:");
	try {
		sampled.exec(schema);
		sampled.prepare(UPSERT).run({
			1: JSON.stringify(rows),
			2: Date.now(),
			3: "sample-test",
		});
		let roll = 0;
		sampled.function("random", () => roll);
		const select = (rate, extra = {}) => {
			const { sql, params } = compileQuery({
				query: { $sampleRate: rate },
				limit: 3,
				skip: 0,
				sort: { create_date: -1 },
				...extra,
			});
			return sampled
				.prepare(sql)
				.all(...params)
				.map((r) => JSON.parse(r.document)._id);
		};
		assert.deepEqual(select(0), []);
		assert.deepEqual(select(0.02), ["a", "b", "c"]);
		roll = 2147483647;
		assert.deepEqual(select(0.02), []);
		assert.deepEqual(
			select(1, { sort: { "summary.mean_dps_per_target": -1 }, limit: 2 }),
			["b", "a"],
		);
		sampled.prepare("UPDATE simulations SET visible=0 WHERE id='c'").run();
		assert.deepEqual(select(1), ["a", "b"]);
		assert.deepEqual(
			select(1, { query: { $sampleRate: 1, "summary.char_names": "nahida" } }),
			["a"],
		);
		for (const rate of [-1, 1.1, "0.02", null, {}, Number.NaN, Infinity])
			assert.throws(() => select(rate), QueryError);
	} finally {
		sampled.close();
	}
});
test("upserts replace a simulation without duplicates, and hidden entries are not public", () => {
	db.prepare(UPSERT).run({
		1: JSON.stringify([entry("a", ["nahida"], [1], 120000)]),
		2: Date.now(),
		3: "next",
	});
	assert.equal(db.prepare("SELECT count(*) AS n FROM simulations").get().n, 3);
	assert.deepEqual(ids({}, { sort: { "summary.mean_dps_per_target": -1 } }), [
		"a",
		"b",
		"c",
	]);
	db.prepare("UPDATE simulations SET visible=0 WHERE id='c'").run();
	assert.deepEqual(ids({ "summary.char_names": "aetheranemo" }), []);
	assert.equal(
		db.prepare("SELECT count(*) AS n FROM simulations WHERE id='c'").get().n,
		1,
	);
});
test("an invalid import is rejected and public clients cannot start an import", async () => {
	assert.throws(() => validateEntries([{ _id: "../other" }]));
	const response = await worker.fetch(
		new Request("https://db.kqm.gg/api/admin/sync", { method: "POST" }),
		{},
		{},
	);
	assert.equal(response.status, 401);
});

test("external sites can read card data, but cannot write or access admin routes", async () => {
	const env = {
		DB: {
			prepare: (sql) => ({
				bind: (...params) => ({
					all: () => ({ results: db.prepare(sql).all(...params) }),
				}),
			}),
		},
	};
	const request = (path, options = {}) =>
		worker.fetch(
			new Request(`https://db.kqm.gg${path}`, {
				headers: { Origin: "https://gcsim.app" },
				...options,
			}),
			env,
			{},
		);
	const query = encodeURIComponent(
		JSON.stringify({ query: { $sampleRate: 1 }, limit: 3 }),
	);
	const response = await request(`/api/db?q=${query}`);
	assert.equal(response.status, 200);
	assert.equal(response.headers.get("Access-Control-Allow-Origin"), "*");
	assert.equal(response.headers.get("Access-Control-Allow-Credentials"), null);
	const { data } = await response.json();
	assert.deepEqual(
		data.map((e) => e._id),
		["a", "b"],
	);
	assert.equal(data[0].summary.mean_dps_per_target, 120000);
	assert.equal(data[0].summary.team[0].name, "nahida");
	assert.equal(data[0].submitter, "Tester");

	const preflight = await request("/api/db/", { method: "OPTIONS" });
	assert.equal(preflight.status, 204);
	assert.equal(preflight.headers.get("Access-Control-Allow-Origin"), "*");
	assert.equal(
		preflight.headers.get("Access-Control-Allow-Methods"),
		"GET, HEAD, OPTIONS",
	);
	assert.equal((await request("/api/db", { method: "POST" })).status, 405);
	const invalid = await request("/api/db?q=invalid");
	assert.equal(invalid.status, 400);
	assert.equal(invalid.headers.get("Access-Control-Allow-Origin"), "*");
	const admin = await request("/api/admin/sync", { method: "OPTIONS" });
	assert.equal(admin.status, 401);
	assert.equal(admin.headers.get("Access-Control-Allow-Origin"), null);
});

test("failed imports preserve records and their cursor, release the lease, and resume", async (t) => {
	const storage = new DatabaseSync(":memory:");
	storage.exec(schema);
	const prepare = (sql) => {
		let params = [];
		return {
			bind(...values) {
				params = values;
				return this;
			},
			first() {
				return storage.prepare(sql).get(...params) ?? null;
			},
			run() {
				const stmt = storage.prepare(sql);
				const values = sql.includes("?1")
					? [Object.fromEntries(params.map((value, i) => [i + 1, value]))]
					: params;
				return { meta: stmt.run(...values) };
			},
		};
	};
	const env = {
		DB: {
			prepare,
			batch: async (statements) => statements.map((s) => s.run()),
		},
	};
	const page = Array.from({ length: 100 }, (_, i) =>
		entry(`sim${i}`, ["nahida"], [1], 50000),
	);
	const source = t.mock.method(globalThis, "fetch", async () =>
		Response.json({ data: page }),
	);
	const progress = await syncPublicDatabase(env, { pages: 1 });
	assert.equal(progress.offset, 100);
	source.mock.mockImplementation(async () =>
		Response.json({ error: "Temporarily unavailable" }),
	);
	await assert.rejects(syncPublicDatabase(env), /no simulation data/);
	assert.equal(
		storage
			.prepare("SELECT COUNT(*) AS n FROM simulations WHERE visible=1")
			.get().n,
		100,
	);
	assert.equal(
		JSON.parse(
			storage.prepare("SELECT value FROM sync_state WHERE key='progress'").get()
				.value,
		).offset,
		100,
	);
	assert.equal(
		storage.prepare("SELECT value FROM sync_state WHERE key='lock'").get(),
		undefined,
	);
	source.mock.mockImplementation(async () => Response.json({}));
	const completed = await syncPublicDatabase(env);
	assert.equal(completed.count, 100);
	assert.equal(completed.run, undefined);
	source.mock.mockImplementation(async () => {
		throw new Error("Should not fetch again today");
	});
	assert.equal(
		(await syncPublicDatabase(env)).completedAt,
		completed.completedAt,
	);
	storage.close();
});

test("result files use the exact share key and remain available from R2 without the source", async (t) => {
	const saved = new Map();
	const env = {
		FILES: {
			get: async (key) => (saved.has(key) ? { body: saved.get(key) } : null),
			put: async (key, body) => saved.set(key, body),
		},
	};
	const result = {
		character_details: [],
		statistics: {},
		config_file: "options iteration=1;",
	};
	const source = t.mock.method(globalThis, "fetch", async (url) => {
		assert.equal(url, "https://gcsim.app/api/share/sharea");
		return Response.json(result);
	});
	assert.deepEqual(JSON.parse(await loadResult(rows[0], env)), result);
	source.mock.mockImplementation(async () => {
		throw new Error("Source unavailable");
	});
	assert.deepEqual(JSON.parse(await loadResult(rows[0], env)), result);
	assert.equal(source.mock.callCount(), 1);
});
