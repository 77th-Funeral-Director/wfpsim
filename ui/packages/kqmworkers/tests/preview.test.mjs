import assert from "node:assert/strict";
import { readFile } from "node:fs/promises";
import test from "node:test";
import ts from "typescript";

const source = await readFile(
	new URL("../src/preview/handlePreview.ts", import.meta.url),
	"utf8",
);
const { outputText } = ts.transpileModule(source, {
	compilerOptions: {
		module: ts.ModuleKind.ESNext,
		target: ts.ScriptTarget.ES2022,
	},
});
const { handlePreview } = await import(
	`data:text/javascript;base64,${Buffer.from(outputText).toString("base64")}`
);

test("simulation preview capture and cache behavior", async () => {
	let captures = 0;
	let writes = 0;
	let exists = true;
	let stored = null;
	let cached;
	let fail = false;
	const png = new Uint8Array([137, 80, 78, 71]);
	globalThis.caches = {
		default: {
			match: async (request) => {
				assert.equal(
					request.url,
					"https://sim.kqm.gg/api/preview/team-123.png?v=2",
				);
				return cached;
			},
			put: async () => {},
		},
	};
	const context = { waitUntil() {} };
	const env = {
		kqmsim_kv: { get: async () => (exists ? new ArrayBuffer(1) : null) },
		kqmsim_r2: {
			get: async (key) => {
				assert.equal(key, "previews/v2/team-123.png");
				return stored;
			},
			put: async (key) => {
				assert.equal(key, "previews/v2/team-123.png");
				writes++;
			},
		},
		BROWSER: {
			quickAction: async (action, options) => {
				captures++;
				assert.equal(action, "screenshot");
				assert.equal(
					options.url,
					"https://sim.kqm.gg/embed/index.html?key=team-123",
				);
				assert.deepEqual(options.viewport, {
					width: 540,
					height: 250,
					deviceScaleFactor: 2,
				});
				assert.equal(
					options.waitForSelector.selector,
					'[data-preview-ready="true"]',
				);
				return fail
					? new Response("failed", { status: 429 })
					: new Response(png, { headers: { "Content-Type": "image/png" } });
			},
		},
	};
	const request = (key = "team-123.png") => ({ params: { key } });
	assert.equal(
		(await handlePreview(request("../bad"), context, env)).status,
		400,
	);
	exists = false;
	assert.equal((await handlePreview(request(), context, env)).status, 404);
	assert.equal(captures, 0);
	exists = true;
	const response = await handlePreview(request(), context, env);
	assert.equal(response.headers.get("Content-Type"), "image/png");
	assert.deepEqual(new Uint8Array(await response.arrayBuffer()), png);
	assert.equal(writes, 1);
	stored = { body: png };
	assert.equal((await handlePreview(request(), context, env)).status, 200);
	assert.equal(captures, 1);
	cached = new Response(png);
	assert.equal(await handlePreview(request(), context, env), cached);
	cached = undefined;
	stored = null;
	fail = true;
	const failure = await handlePreview(request(), context, env);
	assert.equal(failure.status, 503);
	assert.equal(failure.headers.get("Cache-Control"), null);
	assert.equal(writes, 1);
});
