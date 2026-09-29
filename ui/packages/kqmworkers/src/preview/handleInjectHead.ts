import { Env } from "../bindings";
const title = "KQM Sim — Shared simulation";
const description =
	"Open this Genshin Impact simulation and explore team damage, rotations, and results.";
const imageAlt =
	"Simulation characters, weapons, artifact sets, and damage charts.";

export async function handleInjectHead(
	request,
	_context: ExecutionContext,
	env: Env,
): Promise<Response> {
	const res = await env.ASSETS.fetch(request);
	const image = `https://sim.kqm.gg/api/preview/${encodeURIComponent(request.params.key)}.png?v=2`;
	const content = (value: string) => ({
		element: (element: Element) => {
			element.setAttribute("content", value);
		},
	});
	return new HTMLRewriter()
		.on("title", {
			element: (element) => {
				element.setInnerContent(title);
			},
		})
		.on('meta[property="og:title"], meta[name="twitter:title"]', content(title))
		.on(
			'meta[name="description"], meta[property="og:description"], meta[name="twitter:description"]',
			content(description),
		)
		.on('meta[property="og:image"], meta[name="twitter:image"]', content(image))
		.on('meta[property="og:image:width"]', content("1080"))
		.on('meta[property="og:image:height"]', content("500"))
		.on(
			'meta[property="og:image:alt"], meta[name="twitter:image:alt"]',
			content(imageAlt),
		)
		.transform(res);
}

export const handleInjectHeadDB = handleInjectHead;
