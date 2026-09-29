import { Button } from "@gcsim/primitives";
import { useEffect, useState } from "react";
import { Link } from "wouter";
import {
	type Submission as SubmissionData,
	submissionApi,
} from "../lib/submissions";
import {
	DecisionSummary,
	FlowError,
	SubmissionDetails,
} from "../SharedComponents/SubmissionDetails";

export default function Submission({ id }: { id: string }) {
	const [data, setData] = useState<SubmissionData | null>(null);
	const [error, setError] = useState("");
	const [attempt, setAttempt] = useState(0);
	// biome-ignore lint/correctness/useExhaustiveDependencies: attempt refreshes the saved status.
	useEffect(() => {
		const controller = new AbortController();
		setError("");
		submissionApi<SubmissionData>(
			`/api/submissions/${encodeURIComponent(id)}`,
			{ signal: controller.signal },
		)
			.then(setData)
			.catch((reason) => {
				if (!controller.signal.aborted) setError(reason.message);
			});
		return () => controller.abort();
	}, [id, attempt]);
	return (
		<main className="submission-flow">
			<Link href="/submit" className="flow-text-link">
				← Submit a simulation
			</Link>
			<header className="flow-heading">
				<h1>Submission status</h1>
				<p>Keep this link to check the review decision.</p>
			</header>
			{error && (
				<FlowError message={error} retry={() => setAttempt((n) => n + 1)} />
			)}
			{!data && !error && <p role="status">Loading submission…</p>}
			{data && (
				<div className="flex flex-col gap-6">
					{data.status === "pending" ? (
						<section className="flow-panel">
							<h2 className="font-g-display text-g-h3 font-semibold">
								Submitted for review
							</h2>
							<p className="mt-2 text-g-ink-dim">
								The result is saved. A reviewer will check the team and rotation
								before it appears in the database.
							</p>
							<Button
								className="mt-4"
								variant="secondary"
								onClick={() => setAttempt((n) => n + 1)}
							>
								Refresh status
							</Button>
						</section>
					) : (
						<DecisionSummary data={data} />
					)}
					<SubmissionDetails data={data} />
					<p className="text-g-sm text-g-ink-dim break-all">
						Submission ID: {data.id}
					</p>
				</div>
			)}
		</main>
	);
}
