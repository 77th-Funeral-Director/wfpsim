import { existsSync, mkdirSync, readFileSync, writeFileSync } from "node:fs";
import { setTimeout } from "node:timers/promises";

const origin = new URL(process.argv[2] ?? "http://localhost:8787");
if (
	![
		"http://localhost:8787",
		"https://db.kqm.gg",
		"https://kqm-sim-database.kqm.workers.dev",
	].includes(origin.origin)
)
	throw new Error("Use the configured KQM database origin.");
const token =
	process.env.SYNC_TOKEN ??
	readFileSync(new URL("../.dev.vars", import.meta.url), "utf8").match(
		/^SYNC_TOKEN=(.+)$/m,
	)?.[1];
if (!token) throw new Error("Set SYNC_TOKEN before importing.");
async function post(path) {
	for (let attempt = 0; ; attempt++) {
		try {
			const response = await fetch(new URL(path, origin), {
				method: "POST",
				headers: { Authorization: `Bearer ${token}` },
				signal: AbortSignal.timeout(150000),
			});
			if (!response.ok)
				throw new Error(`Import returned HTTP ${response.status}`);
			return await response.json();
		} catch (error) {
			if (attempt === 2) throw error;
			await setTimeout(3000 * (attempt + 1));
		}
	}
}
if (process.argv.includes("--warm")) {
	mkdirSync(".wrangler", { recursive: true });
	const checkpoint = `.wrangler/warm-${origin.hostname}.json`;
	let state = existsSync(checkpoint)
		? JSON.parse(readFileSync(checkpoint, "utf8"))
		: { cursor: "", total: 0 };
	while (true) {
		const next = await post(
			`/api/admin/warm?after=${encodeURIComponent(state.cursor)}`,
		);
		if (!next.count) break;
		state = { cursor: next.cursor, total: state.total + next.count };
		writeFileSync(checkpoint, JSON.stringify(state));
		if (state.total % 100 === 0)
			console.log(`Saved ${state.total} result files.`);
	}
	console.log(`Result copy complete: ${state.total} files.`);
} else {
	for (let batch = 0; batch < 200; batch++) {
		const state = await post("/api/admin/sync");
		if (state.busy) {
			await setTimeout(3000);
			continue;
		}
		console.log(
			state.run
				? `Imported ${state.offset} simulation records.`
				: `Import complete: ${state.count} simulations.`,
		);
		if (!state.run) break;
		if (batch === 199)
			throw new Error("Import exceeded the expected database size.");
	}
}
