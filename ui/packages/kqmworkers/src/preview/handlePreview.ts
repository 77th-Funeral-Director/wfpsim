import { IRequest } from "itty-router";
import { Env } from "../bindings";

export async function handlePreview(
	request: IRequest,
	context: ExecutionContext,
	env: Env,
): Promise<Response> {
	const key = request.params?.key?.replace(/\.png$/, "") ?? "";
	if (!/^[a-zA-Z0-9_-]{1,150}$/.test(key)) {
		return new Response("Invalid simulation key", { status: 400 });
	}
	const cacheKey = new Request(`https://sim.kqm.gg/api/preview/${key}.png?v=2`);
	const cached = await caches.default.match(cacheKey);
	if (cached) return cached;

	try {
		if ((await env.kqmsim_kv.get(key, "arrayBuffer")) === null) {
			return new Response("Simulation not found", { status: 404 });
		}
		const storedKey = `previews/v2/${key}.png`;
		const stored = await env.kqmsim_r2.get(storedKey);
		let body: BodyInit;
		if (stored) {
			body = stored.body;
		} else {
			const capture = await env.BROWSER.quickAction("screenshot", {
				url: `https://sim.kqm.gg/embed/index.html?key=${key}`,
				viewport: { width: 540, height: 250, deviceScaleFactor: 2 },
				screenshotOptions: { type: "png" },
				gotoOptions: { waitUntil: "networkidle0", timeout: 30000 },
				waitForSelector: {
					selector: '[data-preview-ready="true"]',
					timeout: 30000,
				},
			});
			if (
				!capture.ok ||
				!capture.headers.get("Content-Type")?.includes("image/png")
			) {
				console.error("Preview capture failed", capture.status);
				return new Response("Preview is temporarily unavailable", {
					status: 503,
				});
			}
			body = await capture.arrayBuffer();
			await env.kqmsim_r2.put(storedKey, body, {
				httpMetadata: { contentType: "image/png" },
			});
		}
		const response = new Response(body, {
			headers: {
				"Content-Type": "image/png",
				"Cache-Control": "public, max-age=86400",
			},
		});
		context.waitUntil(caches.default.put(cacheKey, response.clone()));
		return response;
	} catch (error) {
		console.error("Preview generation failed", error);
		return new Response("Preview is temporarily unavailable", { status: 503 });
	}
}
