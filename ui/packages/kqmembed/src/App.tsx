import { Preview } from "./Preview";
import "@gcsim/components/src/index.css";
import type { model } from "@gcsim/types";

import React from "react";

const App = ({ id, src }: { id: string; src: string }) => {
	const [err, setError] = React.useState<string>("");
	const [data, setData] = React.useState<model.SimulationResult | undefined>(
		undefined,
	);
	const [completed, setCompleted] = React.useState(false);
	React.useEffect(() => {
		//https://gcsim.app/api/share/db/nFLhjtD9dfFN
		fetch("/api/share/" + src + "/" + encodeURIComponent(id))
			.then(async (res) => {
				if (!res.ok) throw new Error("Result not found");
				return res.json();
			})
			.then((data) => {
				if (data) {
					setData(data);
				} else {
					setError("unexpected no data");
				}
			})
			.catch((e) => {
				setError(JSON.stringify(e));
			});
	}, []);
	React.useEffect(() => {
		if (!data) return;
		let cancelled = false;
		// Preload SVG gear images and the portrait background as well as img elements.
		const urls = new Set([
			...Array.from(document.querySelectorAll("img")).map((img) => img.src),
			...Array.from(document.querySelectorAll("svg image")).map(
				(img) => img.getAttribute("href") ?? "",
			),
			"/api/assets/misc/overlay.jpg",
		]);
		Promise.all(
			[...urls].filter(Boolean).map(
				(url) =>
					new Promise<void>((resolve, reject) => {
						const image = new Image();
						image.onload = () => resolve();
						image.onerror = () =>
							reject(new Error("Could not load preview image"));
						image.src = url;
					}),
			),
		)
			.then(() => document.fonts.ready)
			.then(() => {
				requestAnimationFrame(() =>
					requestAnimationFrame(() => {
						if (!cancelled) setCompleted(true);
					}),
				);
			})
			.catch((error) => {
				if (!cancelled) setError(error.message);
			});
		return () => {
			cancelled = true;
		};
	}, [data]);

	if (err !== "") {
		return (
			<>
				<input id="has-error" disabled hidden value={err} />
				<div>{err}</div>
			</>
		);
	}

	if (data === undefined) {
		return (
			<div id="status" className="disabled">
				no data
			</div>
		);
	}

	return (
		<div className="kqm-preview">
			{completed ? (
				<span
					className="hidden absolute top-0 left-0"
					id="images_loaded"
					data-preview-ready="true"
				></span>
			) : null}
			<Preview data={data} />
		</div>
	);
};

const Routes = () => {
	const key = new URLSearchParams(window.location.search).get("key");
	const match = window.location.pathname.match(/^\/(db|sh)\/([A-Za-z0-9_-]+)$/);
	if (key) return <App id={key} src="sh" />;
	return match ? (
		<App id={match[2]} src={match[1]} />
	) : (
		<p>Preview not found</p>
	);
};
export default Routes;
