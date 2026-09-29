import assert from "node:assert/strict";
import { readFileSync } from "node:fs";
import { DatabaseSync } from "node:sqlite";
import test from "node:test";
import worker from "./index.mjs";
import { syncPublicDatabase, UPSERT } from "./storage.mjs";
import { resultSummary, SubmissionError, sourceLink } from "./submissions.mjs";

const result = JSON.parse(
	readFileSync(
		new URL("../test-fixtures/razor-result.json", import.meta.url),
		"utf8",
	),
);
const schema = ["0001_database.sql", "0002_submissions.sql"]
	.map((name) =>
		readFileSync(new URL(`../migrations/${name}`, import.meta.url), "utf8"),
	)
	.join("\n");
const basic = `Basic ${btoa("reviewer:test-password-only")}`;

function setup(t) {
	const db = new DatabaseSync(":memory:");
	db.exec(schema);
	t.after(() => db.close());
	const files = new Map();
	const env = {
		REVIEW_USER: "reviewer",
		REVIEW_PASSWORD: "test-password-only",
		SUBMISSION_LIMITER: { limit: async () => ({ success: true }) },
		DB: {
			prepare(sql) {
				let params = [];
				const args = () =>
					/\?[1-9]/.test(sql)
						? [Object.fromEntries(params.map((value, i) => [i + 1, value]))]
						: params;
				return {
					bind(...values) {
						params = values;
						return this;
					},
					first() {
						return db.prepare(sql).get(...args()) ?? null;
					},
					all() {
						return { results: db.prepare(sql).all(...args()) };
					},
					run() {
						return { meta: db.prepare(sql).run(...args()) };
					},
				};
			},
			async batch(statements) {
				db.exec("BEGIN");
				try {
					const responses = statements.map((statement) => statement.run());
					db.exec("COMMIT");
					return responses;
				} catch (error) {
					db.exec("ROLLBACK");
					throw error;
				}
			},
		},
		FILES: {
			put: async (key, body) => files.set(key, body),
			get: async (key) => (files.has(key) ? { body: files.get(key) } : null),
			head: async (key) =>
				files.has(key) ? { size: files.get(key).length } : null,
		},
	};
	const request = (path, body, auth = false, extra = {}) =>
		worker.fetch(
			new Request(`https://db.kqm.gg${path}`, {
				method: body ? "POST" : "GET",
				headers: {
					Origin: "https://db.kqm.gg",
					...(body ? { "Content-Type": "application/json" } : {}),
					...(auth ? { Authorization: basic } : {}),
					...extra,
				},
				body: body ? JSON.stringify(body) : undefined,
			}),
			env,
			{},
		);
	const input = (overrides = {}) => ({
		requestId: crypto.randomUUID(),
		link: "https://gcsim.app/sh/example",
		author: "KQM contributor",
		description: "Razor team with a complete saved rotation.",
		...overrides,
	});
	const fetcher = t.mock.method(globalThis, "fetch", async (url, options) => {
		assert.equal(url, "https://gcsim.app/api/share/example");
		assert.equal(options.redirect, "manual");
		return Response.json(result);
	});
	const create = async (overrides) => {
		const response = await request("/api/submissions", input(overrides));
		assert.equal(
			response.status,
			201,
			JSON.stringify(await response.clone().json()),
		);
		return response.json();
	};
	return { db, env, files, request, input, fetcher, create };
}

test("public submissions save an immutable result once and remain outside the database until approval", async (t) => {
	const { db, files, request, input, fetcher } = setup(t);
	const body = input();
	const created = await (await request("/api/submissions", body)).json();
	assert.equal(created.status, "pending");
	assert.equal(created.entry.submitter, body.author);
	assert.equal(
		created.entry.summary.mean_dps_per_target,
		resultSummary(result).mean_dps_per_target,
	);
	assert.equal(files.size, 1);
	assert.deepEqual((await (await request("/api/db")).json()).data, []);
	assert.equal(
		(await (await request(`/api/submissions/${created.id}`)).json()).status,
		"pending",
	);
	assert.deepEqual(
		await (await request(`/api/submissions/${created.id}/result`)).json(),
		result,
	);
	assert.equal(
		(await (await request("/api/submissions", body)).json()).id,
		created.id,
	);
	assert.equal(fetcher.mock.callCount(), 1);
	assert.equal(db.prepare("SELECT count(*) AS n FROM submissions").get().n, 1);
	assert.equal(
		(await request("/api/submissions", { ...body, author: "Different author" }))
			.status,
		409,
	);
});

test("review APIs require Basic Auth and reject cross-origin writes", async (t) => {
	const { request, create } = setup(t);
	const created = await create();
	for (const path of [
		"/api/review/session",
		"/api/review/submissions",
		`/api/review/submissions/${created.id}`,
	]) {
		const response = await request(path);
		assert.equal(response.status, 401);
		assert.equal(response.headers.get("Access-Control-Allow-Origin"), null);
		assert.equal(response.headers.get("Cache-Control"), "no-store");
	}
	assert.equal(
		(
			await request(`/api/review/submissions/${created.id}/decision`, {
				action: "approve",
			})
		).status,
		401,
	);
	assert.equal(
		(
			await request(
				`/api/review/submissions/${created.id}/decision`,
				{ action: "approve" },
				true,
				{ Origin: "https://other.example" },
			)
		).status,
		403,
	);
	assert.equal(
		(
			await request(
				`/api/review/submissions/${created.id}/decision`,
				{ action: "approve" },
				true,
				{ "Content-Type": "text/plain" },
			)
		).status,
		415,
	);
	assert.equal(
		(await request("/api/review/session", undefined, true)).status,
		200,
	);
	assert.equal(
		(await (await request("/api/review/submissions", undefined, true)).json())
			.total,
		1,
	);
});

test("approval publishes the saved result and survives a completed upstream import", async (t) => {
	const { env, request, create, fetcher } = setup(t);
	const created = await create();
	const response = await request(
		`/api/review/submissions/${created.id}/decision`,
		{ action: "approve", reason: "Checked the rotation." },
		true,
	);
	assert.equal(response.status, 200);
	const approved = await response.json();
	assert.equal(approved.status, "approved");
	assert.equal(approved.reviewedBy, "reviewer");
	assert.equal(approved.publishedId, created.id);
	assert.deepEqual(
		(
			await (
				await request(`/api/review/submissions/${created.id}`, undefined, true)
			).json()
		).duplicates,
		[],
	);
	assert.equal(
		(await (await request("/api/db")).json()).data[0]._id,
		created.id,
	);
	assert.deepEqual(
		await (await request(`/api/share/db/${created.id}`)).json(),
		result,
	);
	assert.equal(
		(
			await request(
				`/api/review/submissions/${created.id}/decision`,
				{ action: "reject", reason: "Second decision" },
				true,
			)
		).status,
		409,
	);
	const upstream = {
		...created.entry,
		_id: "upstream",
		share_key: "source",
		is_db_valid: true,
	};
	fetcher.mock.mockImplementation(async () =>
		Response.json({ data: [upstream] }),
	);
	await syncPublicDatabase(env, { force: true });
	assert.equal((await (await request("/api/db")).json()).data.length, 2);
});

test("rejection requires a note, records the reviewer, and does not publish", async (t) => {
	const { request, create } = setup(t);
	const created = await create();
	assert.equal(
		(
			await request(
				`/api/review/submissions/${created.id}/decision`,
				{ action: "reject", reason: "" },
				true,
			)
		).status,
		400,
	);
	const response = await request(
		`/api/review/submissions/${created.id}/decision`,
		{ action: "reject", reason: "Add the missing rotation assumptions." },
		true,
	);
	assert.equal(response.status, 200);
	const receipt = await (
		await request(`/api/submissions/${created.id}`)
	).json();
	assert.equal(receipt.status, "rejected");
	assert.equal(receipt.reason, "Add the missing rotation assumptions.");
	assert.equal(receipt.reviewedBy, undefined);
	assert.deepEqual((await (await request("/api/db")).json()).data, []);
});

test("replacement keeps the viewer ID and old entry, and source sync cannot overwrite it", async (t) => {
	const { db, env, request, create } = setup(t);
	const created = await create();
	const old = {
		...created.entry,
		_id: "existing",
		description: "Old published rotation",
		share_key: "oldshare",
		accepted_tags: [1, 2],
		is_db_valid: true,
	};
	db.prepare(UPSERT).run({
		1: JSON.stringify([old]),
		2: Date.now(),
		3: "upstream",
	});
	const detail = await (
		await request(`/api/review/submissions/${created.id}`, undefined, true)
	).json();
	assert.equal(detail.duplicates[0]._id, "existing");
	const response = await request(
		`/api/review/submissions/${created.id}/decision`,
		{ action: "replace", replaceId: "existing", reason: "Updated rotation." },
		true,
	);
	assert.equal(response.status, 200);
	assert.equal((await response.json()).publishedId, "existing");
	assert.deepEqual(
		(await (await request("/api/db/id/existing")).json()).accepted_tags,
		old.accepted_tags,
	);
	assert.deepEqual(
		(
			await (
				await request(`/api/review/submissions/${created.id}`, undefined, true)
			).json()
		).duplicates,
		[],
	);
	assert.equal(
		JSON.parse(
			db
				.prepare("SELECT previous_document FROM submissions WHERE id=?")
				.get(created.id).previous_document,
		).description,
		old.description,
	);
	db.prepare(UPSERT).run({
		1: JSON.stringify([old]),
		2: Date.now(),
		3: "next-import",
	});
	assert.equal(
		(await (await request("/api/db/id/existing")).json()).description,
		created.entry.description,
	);
	assert.equal(
		db.prepare("SELECT source FROM simulations WHERE id='existing'").get()
			.source,
		"local",
	);
	assert.ok(env.FILES);
});

test("two competing review decisions produce one winner", async (t) => {
	const { request, create } = setup(t);
	const created = await create();
	const responses = await Promise.all([
		request(
			`/api/review/submissions/${created.id}/decision`,
			{ action: "approve" },
			true,
		),
		request(
			`/api/review/submissions/${created.id}/decision`,
			{ action: "reject", reason: "Needs revision." },
			true,
		),
	]);
	assert.deepEqual(responses.map((r) => r.status).sort(), [200, 409]);
	const receipt = await (
		await request(`/api/submissions/${created.id}`)
	).json();
	assert.equal(
		(await (await request("/api/db")).json()).data.length,
		receipt.status === "approved" ? 1 : 0,
	);
});

test("importing a tag helper link keeps its author and description without sending a Discord command", async (t) => {
	const { request, input, fetcher } = setup(t);
	const calls = [];
	fetcher.mock.mockImplementation(async (url) => {
		calls.push(url);
		if (url === "https://taghelper.simpact.app/api/db/id/LNJ6bpFTFDCw")
			return Response.json({
				share_key: "savedrun",
				submitter: "original-author",
				description: "Original complete rotation description.",
			});
		assert.equal(url, "https://gcsim.app/api/share/savedrun");
		return Response.json(result);
	});
	const response = await request(
		"/api/review/import",
		input({ link: "https://taghelper.simpact.app/id/LNJ6bpFTFDCw" }),
		true,
	);
	assert.equal(response.status, 201);
	const data = await response.json();
	assert.equal(data.entry.submitter, "original-author");
	assert.equal(
		data.entry.description,
		"Original complete rotation description.",
	);
	assert.equal(calls.length, 2);
});

test("submission validation rejects unsupported URLs, unfinished results, and excess requests", async (t) => {
	const { env, request, input, fetcher, db } = setup(t);
	for (const url of [
		"http://gcsim.app/sh/example",
		"https://evil.example/sh/example",
		"https://gcsim.app.evil.example/sh/example",
		"https://name:secret@gcsim.app/sh/example",
		"https://gcsim.app:8080/sh/example",
		"https://gcsim.app/sh/../../private",
	])
		assert.throws(() => sourceLink(url), SubmissionError);
	assert.equal(
		sourceLink("https://sim.kqm.gg/sh/slug-123").result,
		"https://sim.kqm.gg/api/share/slug-123",
	);
	fetcher.mock.mockImplementation(async () =>
		Response.json({ config_file: "No results" }),
	);
	assert.equal((await request("/api/submissions", input())).status, 422);
	assert.equal(db.prepare("SELECT count(*) AS n FROM submissions").get().n, 0);
	env.SUBMISSION_LIMITER.limit = async () => ({ success: false });
	assert.equal((await request("/api/submissions", input())).status, 429);
});
