import { dynamicKey } from "@gcsim/localization";
import { Button, Spinner } from "@gcsim/primitives";
import type { model } from "@gcsim/types";
import { ResultsView } from "@gcsim/components";
import { useEffect, useState } from "react";
import { useTranslation } from "react-i18next";
import { Link } from "wouter";

export default function Results({
	id,
	submission = false,
}: {
	id: string;
	submission?: boolean;
}) {
	const { t } = useTranslation();
	const [data, setData] = useState<model.SimulationResult | null>(null);
	const [error, setError] = useState<string | null>(null);
	const [attempt, setAttempt] = useState(0);
	const [copied, setCopied] = useState(false);
	// biome-ignore lint/correctness/useExhaustiveDependencies: attempt explicitly retries the request.
	useEffect(() => {
		const controller = new AbortController();
		setData(null);
		setError(null);
		document.title = "Simulation results — KQM Sim Database";
		fetch(
			submission
				? `/api/submissions/${encodeURIComponent(id)}/result`
				: `/api/share/db/${encodeURIComponent(id)}`,
			{
				signal: controller.signal,
			},
		)
			.then(async (response) => {
				if (!response.ok)
					throw new Error(
						response.status === 404
							? "This simulation was not found."
							: "The results could not load. Please try again.",
					);
				return response.json() as Promise<model.SimulationResult>;
			})
			.then((result) => {
				if (!controller.signal.aborted) setData(result);
			})
			.catch((reason) => {
				if (!controller.signal.aborted) setError(reason.message);
			});
		return () => controller.abort();
	}, [id, attempt, submission]);
	const names = data?.character_details?.map((c) =>
		t(dynamicKey(`game:character_names.${c.name}`)),
	);
	return (
		<main className="mx-auto max-w-[1440px] px-4 py-8 sm:px-8">
			<div className="mb-6 flex flex-wrap items-center justify-between gap-4">
				<Link
					href={submission ? `/submission/${id}` : "/database"}
					className="text-g-accent"
				>
					{submission ? "← Back to submission" : "← Back to database"}
				</Link>
				<h1 className="font-g-display text-g-h2 font-bold">
					Simulation results
				</h1>
			</div>
			{error ? (
				<div role="alert" className="rounded-g-lg border border-g-danger p-6">
					<p>{error}</p>
					<Button className="mt-4" onClick={() => setAttempt((n) => n + 1)}>
						Try again
					</Button>
				</div>
			) : !data ? (
				<div className="flex min-h-64 items-center justify-center">
					<Spinner className="size-12" />
				</div>
			) : (
				<>
					<ResultsView model={data} names={names} />
					<section className="mt-8 rounded-g-lg border border-g-line-soft bg-g-surface p-5">
						<div className="mb-4 flex flex-wrap items-center justify-between gap-3">
							<h2 className="text-g-h2 font-bold">Simulation configuration</h2>
							<Button
								onClick={async () => {
									try {
										await navigator.clipboard.writeText(data.config_file ?? "");
										setCopied(true);
									} catch {
										setCopied(false);
									}
								}}
							>
								{copied ? "Copied" : "Copy config"}
							</Button>
						</div>
						<pre className="overflow-x-auto whitespace-pre-wrap break-words font-g-mono text-g-sm">
							{data.config_file}
						</pre>
						<a
							className="mt-4 inline-block text-g-accent"
							href="https://sim.kqm.gg/simulator"
						>
							Open KQM Sim
						</a>
					</section>
				</>
			)}
		</main>
	);
}
