import assert from "node:assert/strict";
import { readFileSync } from "node:fs";
import { DatabaseSync } from "node:sqlite";
import test from "node:test";
import worker from "./index.mjs";
import { applyUpgrade, digest } from "./upgrades.mjs";

const fixture = JSON.parse(
	readFileSync(new URL("../test-fixtures/razor-result.json", import.meta.url)),
);
function setup(t) {
	const sqlite = new DatabaseSync(":memory:");
	for (const name of [
		"0001_database.sql",
		"0002_submissions.sql",
		"0003_upgrades.sql",
	])
		sqlite.exec(
			readFileSync(new URL(`../migrations/${name}`, import.meta.url), "utf8"),
		);
	t.after(() => sqlite.close());
	const files = new Map();
	const env = {
		SYNC_TOKEN: "test-only",
		FILES: { put: async (key, value) => files.set(key, value) },
		DB: {
			prepare(sql) {
				let params = [];
				return {
					bind(...values) {
						params = values;
						return this;
					},
					first() {
						return sqlite.prepare(sql).get(...params);
					},
					all() {
						return { results: sqlite.prepare(sql).all(...params) };
					},
					run() {
						return { meta: sqlite.prepare(sql).run(...params) };
					},
				};
			},
			async batch(statements) {
				sqlite.exec("BEGIN");
				try {
					const result = statements.map((s) => s.run());
					sqlite.exec("COMMIT");
					return result;
				} catch (error) {
					sqlite.exec("ROLLBACK");
					throw error;
				}
			},
		},
	};
	return { sqlite, env, files };
}
function entry() {
	return {
		_id: "entry-1",
		config: fixture.config_file,
		share_key: "old-result",
		accepted_tags: [1, 5],
		rejected_tags: [6],
		is_db_valid: true,
		description: "Keep this description",
		submitter: "Keep this author",
		create_date: 1700000000,
		summary: { mean_dps_per_target: 1, sim_duration: { mean: 90 } },
	};
}

test("upgrade keeps review data, records recovery history, and rejects stale writes", async (t) => {
	const { sqlite, env, files } = setup(t);
	const original = JSON.stringify(entry());
	sqlite
		.prepare(
			"INSERT INTO simulations VALUES (?,?,1700000000,1,90,0,'before',1,'archive')",
		)
		.run("entry-1", original);
	const result = { ...fixture, sim_version: "a".repeat(40), modified: false };
	const input = {
		id: "entry-1",
		table: "simulations",
		previous: await digest(original),
		engine: result.sim_version,
		result,
	};
	await applyUpgrade(input, env);
	const row = sqlite.prepare("SELECT * FROM simulations").get(),
		updated = JSON.parse(row.document);
	for (const key of [
		"accepted_tags",
		"rejected_tags",
		"is_db_valid",
		"description",
		"submitter",
		"create_date",
		"config",
	])
		assert.deepEqual(updated[key], entry()[key]);
	assert.equal(row.source, "local");
	assert.equal(files.size, 1);
	assert.equal(
		JSON.parse(
			sqlite.prepare("SELECT previous_row FROM upgrade_history").get()
				.previous_row,
		).document,
		original,
	);
	await assert.rejects(
		applyUpgrade(input, env),
		(error) => error.status === 409,
	);
	assert.equal(
		sqlite.prepare("SELECT COUNT(*) n FROM upgrade_history").get().n,
		1,
	);
});

test("rerunning an unapproved record does not approve or publish it", async (t) => {
	const { sqlite, env } = setup(t);
	const original = JSON.stringify({
		...entry(),
		accepted_tags: [],
		is_db_valid: false,
	});
	sqlite
		.prepare(
			"INSERT INTO submissions(id,request_hash,source_url,document,submitted_at) VALUES ('entry-1','hash','source',?,0)",
		)
		.run(original);
	const result = { ...fixture, sim_version: "a".repeat(40), modified: false };
	const input = {
		id: "entry-1",
		table: "submissions",
		previous: await digest(original),
		engine: result.sim_version,
		result,
	};
	await assert.rejects(
		applyUpgrade(
			{ ...input, result: { ...result, config_file: "wrong" } },
			env,
		),
		(error) => error.status === 422,
	);
	await assert.rejects(
		applyUpgrade({ ...input, result: { ...result, modified: true } }, env),
		(error) => error.status === 422,
	);
	await applyUpgrade(input, env);
	const saved = sqlite.prepare("SELECT * FROM submissions").get();
	assert.equal(saved.status, "pending");
	assert.equal(JSON.parse(saved.document).is_db_valid, false);
	assert.equal(sqlite.prepare("SELECT COUNT(*) n FROM simulations").get().n, 0);
});

test("upgrade export requires admin credentials and validates the table", async (t) => {
	const { env } = setup(t);
	const request = (path, auth) =>
		worker.fetch(
			new Request(`https://db.kqm.gg${path}`, {
				headers: auth ? { Authorization: "Bearer test-only" } : {},
			}),
			env,
			{},
		);
	assert.equal(
		(await request("/api/admin/export?table=submissions", false)).status,
		401,
	);
	assert.equal(
		(await request("/api/admin/export?table=users", true)).status,
		400,
	);
	const response = await request("/api/admin/export?table=submissions", true);
	assert.equal(response.status, 200);
	assert.equal(response.headers.get("Cache-Control"), "no-store");
	assert.deepEqual(await response.json(), { rows: [], cursor: null });
});
