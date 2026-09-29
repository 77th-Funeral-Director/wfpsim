import { compileQuery } from "./query.mjs";
import { ID_PATTERN, loadResult } from "./storage.mjs";

const encoder = new TextEncoder();
const uuid =
	/^[0-9a-f]{8}-[0-9a-f]{4}-4[0-9a-f]{3}-[89ab][0-9a-f]{3}-[0-9a-f]{12}$/i;
const privateHeaders = {
	"Cache-Control": "no-store",
	"X-Content-Type-Options": "nosniff",
	"Referrer-Policy": "same-origin",
};
const json = (value, status = 200) =>
	Response.json(value, { status, headers: privateHeaders });

export class SubmissionError extends Error {
	constructor(message, status = 400, field) {
		super(message);
		this.status = status;
		this.field = field;
	}
}

async function hash(value) {
	const bytes = new Uint8Array(
		await crypto.subtle.digest("SHA-256", encoder.encode(value)),
	);
	return Array.from(bytes, (b) => b.toString(16).padStart(2, "0")).join("");
}

async function reviewer(request, env) {
	if (!env.REVIEW_USER || !env.REVIEW_PASSWORD)
		throw new SubmissionError(
			"Review access is not set up. Contact the site administrator.",
			503,
		);
	const actual = await hash(request.headers.get("Authorization") ?? "");
	const expected = await hash(
		`Basic ${btoa(`${env.REVIEW_USER}:${env.REVIEW_PASSWORD}`)}`,
	);
	let difference = 0;
	for (let i = 0; i < expected.length; i++)
		difference |= actual.charCodeAt(i) ^ expected.charCodeAt(i);
	if (difference !== 0) {
		const { success } = await env.SUBMISSION_LIMITER.limit({
			key: `sign-in:${request.headers.get("CF-Connecting-IP") ?? "local"}`,
		});
		throw new SubmissionError(
			success
				? "Enter the reviewer username and password, then try again."
				: "Too many sign-in attempts. Wait one minute, then try again.",
			success ? 401 : 429,
		);
	}
	return env.REVIEW_USER;
}

async function limitedText(body, maximum) {
	if (!body)
		throw new SubmissionError(
			"The response is empty. Check the simulation link.",
		);
	const reader = body.getReader();
	const chunks = [];
	let length = 0;
	try {
		while (true) {
			const { done, value } = await reader.read();
			if (done) break;
			length += value.byteLength;
			if (length > maximum)
				throw new SubmissionError(
					"The data is too large. Share a result without debug data.",
					413,
				);
			chunks.push(value);
		}
	} finally {
		await reader.cancel();
	}
	const bytes = new Uint8Array(length);
	let offset = 0;
	for (const chunk of chunks) {
		bytes.set(chunk, offset);
		offset += chunk.length;
	}
	return new TextDecoder().decode(bytes);
}

async function readBody(request) {
	if (request.headers.get("Origin") !== new URL(request.url).origin)
		throw new SubmissionError(
			"Open this form on the KQM website and try again.",
			403,
		);
	if (!request.headers.get("Content-Type")?.startsWith("application/json"))
		throw new SubmissionError("Send the form as JSON.", 415);
	try {
		const data = JSON.parse(await limitedText(request.body, 16000));
		if (!data || typeof data !== "object" || Array.isArray(data))
			throw new Error();
		return data;
	} catch (error) {
		if (error instanceof SubmissionError) throw error;
		throw new SubmissionError(
			"The form could not be read. Check the fields and try again.",
		);
	}
}

function textField(value, field, label, minimum, maximum) {
	if (
		typeof value !== "string" ||
		value.trim().length < minimum ||
		value.trim().length > maximum
	)
		throw new SubmissionError(
			`Enter ${label} with ${minimum}–${maximum} characters.`,
			400,
			field,
		);
	return value.trim();
}

export function sourceLink(value, importing = false) {
	let url;
	try {
		url = new URL(value);
	} catch {
		/* Handled below. */
	}
	if (url?.protocol !== "https:" || url.username || url.password || url.port)
		throw new SubmissionError(
			"Use a full HTTPS simulation link from KQM or gcsim.",
			400,
			"link",
		);
	const match = url.pathname.match(/^\/(sh|db|id)\/([a-zA-Z0-9_-]{1,128})\/?$/);
	if (!match)
		throw new SubmissionError(
			"Paste a shared result link, such as https://sim.kqm.gg/sh/…",
			400,
			"link",
		);
	const [, kind, id] = match;
	if (importing && url.hostname === "taghelper.simpact.app" && kind === "id")
		return {
			url: `https://taghelper.simpact.app/id/${id}`,
			metadata: `https://taghelper.simpact.app/api/db/id/${id}`,
		};
	const shared =
		["sim.kqm.gg", "gcsim.app", "simpact.app"].includes(url.hostname) &&
		kind === "sh";
	const database =
		["db.kqm.gg", "gcsim.app", "simpact.app"].includes(url.hostname) &&
		kind === "db";
	if (!shared && !database)
		throw new SubmissionError(
			importing
				? "Use a tag helper link or a KQM or gcsim result link."
				: "Use a shared result link from KQM or gcsim.",
			400,
			"link",
		);
	return {
		url: `${url.origin}/${kind}/${id}`,
		result: `${url.origin}/api/share/${database ? "db/" : ""}${id}`,
	};
}

async function fetchJson(url, maximum = 8000000) {
	let response;
	try {
		response = await fetch(url, {
			redirect: "manual",
			signal: AbortSignal.timeout(25000),
			headers: { Accept: "application/json" },
		});
	} catch (error) {
		console.error("Simulation source fetch failed", error.message);
		throw new SubmissionError(
			"The simulation source could not be reached. Try again shortly.",
			502,
			"link",
		);
	}
	if (!response.ok)
		throw new SubmissionError(
			response.status === 404
				? "This simulation was not found. Check the link and try again."
				: "The simulation could not load. Try again shortly.",
			422,
			"link",
		);
	try {
		return JSON.parse(await limitedText(response.body, maximum));
	} catch (error) {
		if (error instanceof SubmissionError) throw error;
		throw new SubmissionError(
			"This link has no readable result. Run the simulation and share it again.",
			422,
			"link",
		);
	}
}

export function resultSummary(result) {
	const stats = result?.statistics;
	const team = result?.character_details;
	const targets = result?.target_details;
	const validName = (name) =>
		typeof name === "string" && /^[a-z0-9_]{1,64}$/.test(name);
	if (
		!Array.isArray(team) ||
		team.length < 1 ||
		team.length > 4 ||
		team.some((c) => !validName(c?.name) || !validName(c?.weapon?.name)) ||
		new Set(team.map((c) => c.name)).size !== team.length ||
		!Array.isArray(targets) ||
		targets.length < 1 ||
		targets.length > 100 ||
		!Number.isFinite(stats?.duration?.mean) ||
		stats.duration.mean <= 0 ||
		!Number.isFinite(stats?.total_damage?.mean) ||
		stats.total_damage.mean <= 0 ||
		!Number.isFinite(stats?.dps?.mean) ||
		stats.dps.mean <= 0 ||
		!Number.isInteger(stats?.iterations) ||
		stats.iterations < 1 ||
		typeof result?.config_file !== "string" ||
		!result.config_file.trim() ||
		result.config_file.length > 100000
	)
		throw new SubmissionError(
			"This link needs a completed simulation with characters, gear, and damage results. Run it and share it again.",
			422,
			"link",
		);
	const summary = {
		mode: result.mode ?? 0,
		team,
		char_names: team.map((c) => c.name),
		sim_duration: {
			min: stats.duration.min,
			max: stats.duration.max,
			mean: stats.duration.mean,
			sd: stats.duration.sd,
		},
		total_damage: {
			min: stats.total_damage.min,
			max: stats.total_damage.max,
			mean: stats.total_damage.mean,
			sd: stats.total_damage.sd,
		},
		target_count: targets.length,
		mean_dps_per_target:
			stats.total_damage.mean / (targets.length * stats.duration.mean),
	};
	if (
		!Number.isFinite(summary.mean_dps_per_target) ||
		JSON.stringify(summary).length > 100000
	)
		throw new SubmissionError(
			"The result summary is invalid. Run the simulation and share it again.",
			422,
			"link",
		);
	return summary;
}

function submission(row, internal = false) {
	return {
		id: row.id,
		status: row.status,
		sourceUrl: row.source_url,
		submittedAt: row.submitted_at,
		reviewedAt: row.reviewed_at,
		reason: row.reason,
		publishedId: row.published_id,
		entry: JSON.parse(row.document),
		...(internal ? { reviewedBy: row.reviewed_by } : {}),
	};
}

async function getSubmission(id, env) {
	const row = await env.DB.prepare("SELECT * FROM submissions WHERE id=?")
		.bind(id)
		.first();
	if (!row)
		throw new SubmissionError(
			"This submission was not found. Check the saved link or submit it again.",
			404,
		);
	return row;
}

async function createSubmission(input, env, importing) {
	if (!uuid.test(input.requestId ?? ""))
		throw new SubmissionError("Reload the submission form and try again.");
	const id = `kqm-${input.requestId.toLowerCase()}`;
	const source = sourceLink(input.link, importing);
	if (importing && !source.metadata)
		throw new SubmissionError(
			"Use a tag helper link to import an existing submission.",
			400,
			"link",
		);
	const requestHash = await hash(
		JSON.stringify({
			url: source.url,
			author: input.author,
			description: input.description,
		}),
	);
	const existing = await env.DB.prepare("SELECT * FROM submissions WHERE id=?")
		.bind(id)
		.first();
	if (existing) {
		if (existing.request_hash !== requestHash)
			throw new SubmissionError(
				"This form was already submitted. Open a new submission form.",
				409,
			);
		return submission(existing);
	}
	let original;
	if (source.metadata) {
		original = await fetchJson(source.metadata, 500000);
		if (!ID_PATTERN.test(original.share_key ?? ""))
			throw new SubmissionError(
				"This submission has no result yet. Wait for its result and try again.",
				422,
				"link",
			);
		source.result = `https://gcsim.app/api/share/${original.share_key}`;
	}
	const author = textField(
		original?.submitter ?? input.author,
		"author",
		"a display name",
		1,
		80,
	);
	const description = textField(
		original?.description ?? input.description,
		"description",
		"a description",
		10,
		3000,
	);
	const result = await fetchJson(source.result);
	const summary = resultSummary(result);
	const resultText = JSON.stringify(result);
	const shareKey = `kqm-${await hash(resultText)}`;
	const now = Date.now();
	const entry = {
		_id: id,
		config: result.config_file,
		description,
		submitter: author,
		create_date: Math.floor(now / 1000),
		share_key: shareKey,
		last_update: Math.floor(now / 1000),
		accepted_tags: [],
		is_db_valid: false,
		summary,
	};
	await env.FILES.put(`results/${shareKey}.json`, resultText, {
		httpMetadata: { contentType: "application/json" },
	});
	await env.DB.prepare(
		"INSERT INTO submissions(id,request_hash,source_url,document,submitted_at) VALUES (?,?,?,?,?) ON CONFLICT(id) DO NOTHING",
	)
		.bind(id, requestHash, source.url, JSON.stringify(entry), now)
		.run();
	const saved = await getSubmission(id, env);
	if (saved.request_hash !== requestHash)
		throw new SubmissionError(
			"This form was already submitted. Open a new submission form.",
			409,
		);
	return submission(saved);
}

function normalizedTeam(entry) {
	return [...(entry.summary?.char_names ?? [])]
		.map((name) => name.replace(/^(aether|lumine)/, "traveler"))
		.sort()
		.join(",");
}

async function duplicates(entry, env, publishedId) {
	const query = {
		$and: entry.summary.char_names.map((name) => {
			const traveler = name.match(/^(aether|lumine)(.+)$/);
			return traveler
				? {
						$or: [
							{ "summary.char_names": `aether${traveler[2]}` },
							{ "summary.char_names": `lumine${traveler[2]}` },
						],
					}
				: { "summary.char_names": name };
		}),
	};
	const { sql, params } = compileQuery({
		query,
		limit: 100,
		sort: { create_date: -1 },
	});
	const { results } = await env.DB.prepare(sql)
		.bind(...params)
		.all();
	return results
		.map((r) => JSON.parse(r.document))
		.filter(
			(e) =>
				e._id !== publishedId && normalizedTeam(e) === normalizedTeam(entry),
		);
}

async function decide(id, input, user, env) {
	if (!["approve", "reject", "replace"].includes(input.action))
		throw new SubmissionError("Choose Approve or Reject.");
	const row = await getSubmission(id, env);
	if (row.status !== "pending")
		throw new SubmissionError(
			"This submission was already reviewed. Reload the page to see the decision.",
			409,
		);
	const reason = textField(
		input.reason ?? "",
		"reason",
		"a review note",
		input.action === "approve" ? 0 : 3,
		2000,
	);
	const entry = JSON.parse(row.document);
	let target = id;
	let previous = null;
	if (input.action === "replace") {
		if (!ID_PATTERN.test(input.replaceId ?? ""))
			throw new SubmissionError("Choose an existing simulation to replace.");
		const old = await env.DB.prepare(
			"SELECT document FROM simulations WHERE id=? AND visible=1",
		)
			.bind(input.replaceId)
			.first();
		if (
			!old ||
			normalizedTeam(JSON.parse(old.document)) !== normalizedTeam(entry)
		)
			throw new SubmissionError(
				"The existing simulation must use the same team. Reload the comparison list.",
				409,
			);
		target = input.replaceId;
		previous = old.document;
	}
	const status = input.action === "reject" ? "rejected" : "approved";
	const now = Date.now();
	const statements = [];
	if (status === "approved") {
		if (!(await env.FILES.head(`results/${entry.share_key}.json`))) {
			const body = await loadResult(entry, env);
			if (body instanceof ReadableStream) await body.cancel();
		}
		if (!(await env.FILES.head(`results/${entry.share_key}.json`)))
			throw new SubmissionError(
				"The saved result is unavailable. Try again before approving.",
				503,
			);
		const published = {
			...entry,
			_id: target,
			accepted_tags: previous
				? (JSON.parse(previous).accepted_tags ?? [1])
				: [1],
			is_db_valid: true,
		};
		statements.push(
			env.DB.prepare(`INSERT INTO simulations(id,document,create_date,dps,duration,imported_at,seen_run,visible,source)
			SELECT ?1,?2,?3,?4,?5,?6,'local',1,'local' FROM submissions WHERE id=?7 AND status='pending'
			AND (?8 IS NULL OR EXISTS (SELECT 1 FROM simulations WHERE id=?1 AND document=?8 AND visible=1))
			ON CONFLICT(id) DO UPDATE SET document=excluded.document,create_date=excluded.create_date,
			dps=excluded.dps,duration=excluded.duration,imported_at=excluded.imported_at,seen_run='local',visible=1,source='local'`).bind(
				target,
				JSON.stringify(published),
				published.create_date,
				published.summary.mean_dps_per_target,
				published.summary.sim_duration.mean,
				now,
				id,
				previous,
			),
		);
	}
	statements.push(
		env.DB.prepare(`UPDATE submissions SET status=?,reviewed_at=?,reviewed_by=?,reason=?,published_id=?,previous_document=?
		WHERE id=? AND status='pending' ${status === "approved" ? "AND changes()=1" : ""}`).bind(
			status,
			now,
			user,
			reason,
			status === "approved" ? target : null,
			previous,
			id,
		),
	);
	const results = await env.DB.batch(statements);
	if (!results.at(-1).meta.changes)
		throw new SubmissionError(
			"This entry changed during review. Reload the page before you decide.",
			409,
		);
	return submission(await getSubmission(id, env), true);
}

export async function handleSubmissions(request, env) {
	const url = new URL(request.url);
	try {
		const isReview = url.pathname.startsWith("/api/review/");
		const user = isReview ? await reviewer(request, env) : null;
		if (
			isReview &&
			url.pathname === "/api/review/session" &&
			request.method === "GET"
		)
			return json({ reviewer: user });
		if (request.method === "POST") {
			const input = await readBody(request);
			if (
				url.pathname === "/api/submissions" ||
				url.pathname === "/api/review/import"
			) {
				const { success } = await env.SUBMISSION_LIMITER.limit({
					key: `submit:${request.headers.get("CF-Connecting-IP") ?? "local"}`,
				});
				if (!success)
					throw new SubmissionError(
						"Too many submissions. Wait one minute, then try again.",
						429,
					);
				return json(await createSubmission(input, env, isReview), 201);
			}
			const decision = url.pathname.match(
				/^\/api\/review\/submissions\/([a-zA-Z0-9_-]{1,128})\/decision$/,
			);
			if (decision) return json(await decide(decision[1], input, user, env));
		}
		if (request.method === "GET") {
			if (url.pathname === "/api/review/submissions") {
				const status = url.searchParams.get("status") ?? "pending";
				const page = Number(url.searchParams.get("page") ?? 1);
				if (
					!["pending", "approved", "rejected"].includes(status) ||
					!Number.isInteger(page) ||
					page < 1 ||
					page > 10000
				)
					throw new SubmissionError("Choose a valid queue page.");
				const { results } = await env.DB.prepare(
					"SELECT * FROM submissions WHERE status=? ORDER BY submitted_at ASC,id LIMIT 20 OFFSET ?",
				)
					.bind(status, (page - 1) * 20)
					.all();
				const total = await env.DB.prepare(
					"SELECT COUNT(*) AS count FROM submissions WHERE status=?",
				)
					.bind(status)
					.first();
				return json({
					data: results.map((r) => submission(r, true)),
					total: total.count,
					page,
				});
			}
			const match = url.pathname.match(
				/^\/api\/(review\/)?submissions\/([a-zA-Z0-9_-]{1,128})(\/result)?$/,
			);
			if (match) {
				const row = await getSubmission(match[2], env);
				const entry = JSON.parse(row.document);
				if (match[3]) {
					return new Response(await loadResult(entry, env), {
						headers: { ...privateHeaders, "Content-Type": "application/json" },
					});
				}
				return json({
					...submission(row, isReview),
					...(isReview
						? { duplicates: await duplicates(entry, env, row.published_id) }
						: {}),
				});
			}
		}
		return json(
			{ error: "This action is not available. Reload the page and try again." },
			405,
		);
	} catch (error) {
		if (error instanceof SubmissionError)
			return json({ error: error.message, field: error.field }, error.status);
		console.error("Submission request failed", error.message);
		return json(
			{
				error:
					"The request could not finish. Reload the page to check its status before you try again.",
			},
			503,
		);
	}
}
