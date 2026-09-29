import { Button, Input } from "@gcsim/primitives";
import { type FormEvent, useState } from "react";
import { Link, useLocation } from "wouter";
import { ApiError, type Submission, submissionApi } from "../lib/submissions";
import { FlowError } from "../SharedComponents/SubmissionDetails";

export default function Submit() {
	const [, navigate] = useLocation();
	const [requestId] = useState(() => crypto.randomUUID());
	const [error, setError] = useState<ApiError | null>(null);
	const [busy, setBusy] = useState(false);
	async function submit(event: FormEvent<HTMLFormElement>) {
		event.preventDefault();
		const form = event.currentTarget;
		const fields = new FormData(form);
		setError(null);
		setBusy(true);
		try {
			const data = await submissionApi<Submission>("/api/submissions", {
				body: {
					requestId,
					link: fields.get("link"),
					author: fields.get("author"),
					description: fields.get("description"),
				},
			});
			navigate(`/submission/${data.id}`);
		} catch (reason) {
			const failure =
				reason instanceof ApiError
					? reason
					: new ApiError(
							"The submission could not be sent. Check your connection and try again.",
							0,
						);
			setError(failure);
			if (failure.field)
				requestAnimationFrame(() =>
					form
						.querySelector<HTMLInputElement>(`[name="${failure.field}"]`)
						?.focus(),
				);
		} finally {
			setBusy(false);
		}
	}
	return (
		<main className="submission-flow">
			<Link href="/database" className="flow-text-link">
				← Browse simulations
			</Link>
			<header className="flow-heading">
				<p className="text-g-sm text-g-ink-dim">Community submissions</p>
				<h1>Submit a simulation</h1>
				<p>
					Share a completed result for review. Approved simulations appear in
					the KQM database.
				</p>
			</header>
			<div className="submission-columns">
				<form
					className="flow-panel flex flex-col gap-6"
					onSubmit={submit}
					noValidate
				>
					<div>
						<label htmlFor="sim-link">Simulation link</label>
						<Input
							id="sim-link"
							name="link"
							defaultValue={
								new URLSearchParams(window.location.search).get("link") ?? ""
							}
							type="url"
							autoComplete="url"
							placeholder="https://sim.kqm.gg/sh/…"
							required
							maxLength={500}
							aria-invalid={error?.field === "link"}
							aria-describedby="link-help link-error"
						/>
						<p id="link-help" className="field-help">
							Use a KQM or gcsim result link. Run the simulation before you
							share it.
						</p>
						<p id="link-error" className="field-error">
							{error?.field === "link" ? error.message : ""}
						</p>
					</div>
					<div>
						<label htmlFor="sim-author">Display name</label>
						<Input
							id="sim-author"
							name="author"
							autoComplete="nickname"
							maxLength={80}
							required
							placeholder="Name to show with the simulation"
							aria-invalid={error?.field === "author"}
							aria-describedby="author-help author-error"
						/>
						<p id="author-help" className="field-help">
							This name will be public if the simulation is approved.
						</p>
						<p id="author-error" className="field-error">
							{error?.field === "author" ? error.message : ""}
						</p>
					</div>
					<div>
						<label htmlFor="sim-description">Simulation description</label>
						<textarea
							id="sim-description"
							name="description"
							rows={6}
							minLength={10}
							maxLength={3000}
							required
							placeholder="Describe the team, rotation, and any assumptions."
							aria-invalid={error?.field === "description"}
							aria-describedby="description-help description-error"
						/>
						<p id="description-help" className="field-help">
							Include rotation length, key actions, and any limits that affect
							the result.
						</p>
						<p id="description-error" className="field-error">
							{error?.field === "description" ? error.message : ""}
						</p>
					</div>
					{error && <FlowError message={error.message} />}
					<div className="flex flex-wrap items-center gap-4">
						<Button type="submit" disabled={busy}>
							{busy ? "Submitting…" : "Submit for review"}
						</Button>
						<span role="status" className="text-g-sm text-g-ink-dim">
							{busy ? "Loading and saving the shared result." : ""}
						</span>
					</div>
				</form>
				<aside className="flow-panel self-start">
					<h2 className="font-g-display text-g-h3 font-semibold">
						Before you submit
					</h2>
					<ol className="mt-5 flex list-decimal flex-col gap-5 ps-5 text-g-ink-dim">
						<li>
							Check that the shared result has the correct characters, gear, and
							rotation.
						</li>
						<li>Explain assumptions that a reviewer needs to understand.</li>
						<li>
							Keep the submission link to check the decision and review note.
						</li>
					</ol>
					<p className="mt-6 text-g-sm text-g-ink-dim">
						Reviewers use the result saved with your link. This site does not
						run the simulation again.
					</p>
				</aside>
			</div>
		</main>
	);
}
