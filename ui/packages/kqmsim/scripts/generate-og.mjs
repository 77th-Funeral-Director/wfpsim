import { readFile, mkdir, writeFile } from "node:fs/promises";
import { createRequire } from "node:module";
import satori from "satori";
import { Resvg } from "@resvg/resvg-js";

const require = createRequire(import.meta.url);
const publicDir = new URL("../public/", import.meta.url);
const logo = await readFile(new URL("kqm-logo.png", publicDir));
const font = await readFile(
	require.resolve("@fontsource/inter/files/inter-latin-600-normal.woff"),
);
const box = (style, children) => ({
	type: "div",
	props: { style: { display: "flex", ...style }, children },
});

// Add a card here to reuse the same layout, logo, and export settings.
const cards = [
	{
		name: "site",
		label: "GENSHIN IMPACT COMBAT SIMULATOR",
		title: "Build. Simulate.\nUnderstand.",
		description: "Explore team damage, rotations, and results.",
	},
	{
		name: "share",
		label: "SHARED SIMULATION",
		title: "Every rotation.\nEvery result.",
		description: "Open this simulation and explore the details.",
	},
];
await mkdir(new URL("og/", publicDir), { recursive: true });
for (const card of cards) {
	const svg = await satori(
		box(
			{
				width: "100%",
				height: "100%",
				background: "#232024",
				color: "#ffffff",
				fontFamily: "Inter",
				padding: "54px 64px",
				flexDirection: "column",
				position: "relative",
			},
			[
				box(
					{
						position: "absolute",
						right: -100,
						top: -160,
						width: 650,
						height: 900,
						border: "1px solid #66576e",
						borderRadius: 320,
						transform: "rotate(28deg)",
						background: "linear-gradient(135deg, #423745, #232024)",
					},
					[],
				),
				box({ alignItems: "center", gap: 22 }, [
					{
						type: "img",
						props: {
							src: `data:image/png;base64,${logo.toString("base64")}`,
							width: 76,
							height: 76,
							style: { objectFit: "contain" },
						},
					},
					box({ fontSize: 38 }, "KQM Sim"),
				]),
				box(
					{ marginTop: 48, fontSize: 17, letterSpacing: 4, color: "#dab2f9" },
					card.label,
				),
				box(
					{
						marginTop: 20,
						fontSize: 76,
						lineHeight: 1.05,
						whiteSpace: "pre",
						letterSpacing: -3,
					},
					card.title,
				),
				box(
					{ marginTop: 24, fontSize: 24, color: "#d1c6d8" },
					card.description,
				),
				box(
					{
						marginTop: "auto",
						justifyContent: "space-between",
						alignItems: "center",
						borderTop: "1px solid #66576e",
						paddingTop: 22,
						fontSize: 18,
					},
					[
						box({ color: "#dab2f9" }, "THEORYCRAFTING IN ACTION"),
						box({}, "sim.kqm.gg"),
					],
				),
			],
		),
		{
			width: 1200,
			height: 630,
			fonts: [{ name: "Inter", data: font, weight: 600, style: "normal" }],
		},
	);
	await writeFile(
		new URL(`og/${card.name}.png`, publicDir),
		new Resvg(svg).render().asPng(),
	);
}
console.log("Generated KQM Sim social images (1200 × 630).");
