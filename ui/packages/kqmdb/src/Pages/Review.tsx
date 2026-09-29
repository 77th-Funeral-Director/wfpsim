import { Button, Input } from "@gcsim/primitives";
import type { db } from "@gcsim/types";
import { type FormEvent, useEffect, useRef, useState } from "react";
import { Link, useLocation } from "wouter";
import {
	ApiError,
	fullDate,
	type Submission,
	submissionApi,
} from "../lib/submissions";
import {
	DecisionSummary,
	FlowError,
	SubmissionBadge,
	SubmissionDetails,
	TeamNames,
} from "../SharedComponents/SubmissionDetails";

type Access = { authorization: string; reviewer: string };
type Decision = { action: "approve" | "reject" | "replace"; entry?: db.Entry };

export default function Review({ id }: { id?: string }) {
	const [access, setAccess] = useState<Access | null>(null);
	const [error, setError] = useState("");
	const [busy, setBusy] = useState(false);
	async function signIn(event: FormEvent<HTMLFormElement>) {
		event.preventDefault();
		const fields = new FormData(event.currentTarget);
		setBusy(true);
		setError("");
		try {
			const authorization = `Basic ${btoa(`${fields.get("username")}:${fields.get("password")}`)}`;
			const session = await submissionApi<{ reviewer: string }>(
				"/api/review/session",
				{ authorization },
			);
			setAccess({ authorization, reviewer: session.reviewer });
		} catch (reason) {
			setError(
				reason instanceof Error
					? reason.message
					: "Sign-in failed. Check your credentials and try again.",
			);
		} finally {
			setBusy(false);
		}
	}
	if (!access)
		return (
			<main className="submission-flow">
				<Link href="/database" className="flow-text-link">
					← Browse simulations
				</Link>
				<div className="mx-auto max-w-[480px] py-8 sm:py-16">
					<header className="flow-heading">
						<p className="text-g-sm text-g-ink-dim">Reviewer access</p>
						<h1>Sign in to review</h1>
						<p>Use the reviewer credentials to open the submission queue.</p>
					</header>
					<form className="flow-panel flex flex-col gap-6" onSubmit={signIn}>
						<div>
							<label htmlFor="review-username">Username</label>
							<Input
								id="review-username"
								name="username"
								autoComplete="username"
								required
								maxLength={80}
							/>
						</div>
						<div>
							<label htmlFor="review-password">Password</label>
							<Input
								id="review-password"
								name="password"
								type="password"
								autoComplete="current-password"
								required
								maxLength={200}
							/>
						</div>
						{error && <FlowError message={error} />}
						<Button disabled={busy} type="submit">
							{busy ? "Signing in…" : "Sign in"}
						</Button>
						<p role="status" className="sr-only">
							{busy ? "Checking reviewer access." : ""}
						</p>
					</form>
				</div>
			</main>
		);
	return (
		<main className="submission-flow">
			<div className="flex flex-wrap items-center justify-between gap-4">
				<Link href={id ? "/review" : "/database"} className="flow-text-link">
					{id ? "← Review queue" : "← Browse simulations"}
				</Link>
				<div className="flex flex-wrap items-center gap-3 text-g-sm text-g-ink-dim">
					<span>
						Signed in as <bdi>{access.reviewer}</bdi>
					</span>
					<Button variant="ghost" onClick={() => setAccess(null)}>
						Sign out
					</Button>
				</div>
			</div>
			{id ? (
				<ReviewDetail key={id} id={id} access={access} />
			) : (
				<ReviewQueue access={access} />
			)}
		</main>
	);
}

function ReviewQueue({ access }: { access: Access }) {
	const [status, setStatus] = useState<Submission["status"]>("pending");
	const [page, setPage] = useState(1);
	const [data, setData] = useState<{
		data: Submission[];
		total: number;
	} | null>(null);
	const [error, setError] = useState("");
	const [attempt, setAttempt] = useState(0);
	// biome-ignore lint/correctness/useExhaustiveDependencies: attempt retries the queue request.
	useEffect(() => {
		const controller = new AbortController();
		setData(null);
		setError("");
		submissionApi<{ data: Submission[]; total: number }>(
			`/api/review/submissions?status=${status}&page=${page}`,
			{ authorization: access.authorization, signal: controller.signal },
		)
			.then(setData)
			.catch((reason) => {
				if (!controller.signal.aborted) setError(reason.message);
			});
		return () => controller.abort();
	}, [status, page, access.authorization, attempt]);
	return (
		<>
			<header className="flow-heading">
				<p className="text-g-sm text-g-ink-dim">KQM submissions</p>
				<h1>Review queue</h1>
				<p>
					Check the saved results, compare the team, and decide what to publish.
				</p>
			</header>
			<div className="mb-6 flex flex-wrap items-end justify-between gap-4">
				<div>
					<label htmlFor="queue-status">Submission status</label>
					<select
						id="queue-status"
						value={status}
						onChange={(event) => {
							setStatus(event.target.value as Submission["status"]);
							setPage(1);
						}}
					>
						<option value="pending">Pending review</option>
						<option value="approved">Approved</option>
						<option value="rejected">Rejected</option>
					</select>
				</div>
				<p role="status" className="text-g-sm text-g-ink-dim">
					{data
						? `${data.total} ${data.total === 1 ? "submission" : "submissions"}`
						: error
							? "Queue not loaded."
							: "Loading queue…"}
				</p>
			</div>
			{error && (
				<FlowError message={error} retry={() => setAttempt((n) => n + 1)} />
			)}
			{data && data.data.length === 0 && (
				<section className="flow-panel py-12 text-center">
					<h2 className="font-g-display text-g-h2 font-semibold">
						{status === "pending"
							? "No submissions waiting"
							: `No ${status} submissions`}
					</h2>
					<p className="mx-auto mt-3 max-w-lg text-g-ink-dim">
						{status === "pending"
							? "New submissions will appear here. You can also import a pending entry with its tag helper link."
							: "Choose another status to return to the review queue."}
					</p>
					<Button
						variant="secondary"
						className="mt-5"
						onClick={() => {
							setStatus("pending");
							setPage(1);
							setAttempt((n) => n + 1);
						}}
					>
						{status === "pending"
							? "Refresh queue"
							: "View pending submissions"}
					</Button>
				</section>
			)}
			{data && data.data.length > 0 && (
				<div className="flex flex-col gap-4">
					{data.data.map((item) => (
						<article key={item.id} className="flow-panel queue-row">
							<div className="min-w-0 flex-1">
								<div className="flex flex-wrap items-center gap-3">
									<SubmissionBadge status={item.status} />
									<span className="text-g-sm text-g-ink-dim">
										{fullDate(item.submittedAt)}
									</span>
								</div>
								<h2 className="mt-4 font-g-display text-g-h3 font-semibold">
									<TeamNames entry={item.entry} />
								</h2>
								<p className="mt-2 line-clamp-3 whitespace-pre-wrap break-words leading-relaxed text-g-ink-dim">
									{item.entry.description}
								</p>
								<p className="mt-3 text-g-sm text-g-ink-dim">
									By <bdi>{item.entry.submitter}</bdi>
								</p>
							</div>
							<div className="queue-action">
								<strong className="font-g-mono text-g-h3 tabular-nums">
									{Math.round(
										item.entry.summary?.mean_dps_per_target ?? 0,
									).toLocaleString()}
								</strong>
								<span className="text-g-sm text-g-ink-dim">DPS per target</span>
								<Button variant="secondary" className="mt-3" asChild>
									<Link href={`/review/${item.id}`}>
										{item.status === "pending"
											? "Open review"
											: "View decision"}
										<span className="sr-only">
											{" "}
											for submission by {item.entry.submitter}
										</span>
									</Link>
								</Button>
							</div>
						</article>
					))}
				</div>
			)}
			{data && data.total > 20 && (
				<nav
					className="mt-6 flex flex-wrap items-center justify-between gap-3"
					aria-label="Queue pages"
				>
					<Button
						variant="secondary"
						disabled={page === 1}
						onClick={() => setPage((n) => n - 1)}
					>
						Previous page
					</Button>
					<span>
						Page {page} of {Math.ceil(data.total / 20)}
					</span>
					<Button
						variant="secondary"
						disabled={page * 20 >= data.total}
						onClick={() => setPage((n) => n + 1)}
					>
						Next page
					</Button>
				</nav>
			)}
			<ImportSubmission access={access} />
		</>
	);
}

function ImportSubmission({ access }: { access: Access }) {
	const [, navigate] = useLocation();
	const [requestId] = useState(() => crypto.randomUUID());
	const [busy, setBusy] = useState(false);
	const [error, setError] = useState("");
	async function submit(event: FormEvent<HTMLFormElement>) {
		event.preventDefault();
		const link = new FormData(event.currentTarget).get("link");
		setBusy(true);
		setError("");
		try {
			const saved = await submissionApi<Submission>("/api/review/import", {
				authorization: access.authorization,
				body: { requestId, link },
			});
			navigate(`/review/${saved.id}`);
		} catch (reason) {
			setError(
				reason instanceof Error
					? reason.message
					: "Import failed. Check the link and try again.",
			);
		} finally {
			setBusy(false);
		}
	}
	return (
		<details className="flow-panel flow-disclosure mt-8">
			<summary>Import an existing tag helper submission</summary>
			<p className="mt-4 max-w-2xl text-g-ink-dim">
				Copy a pending entry into the KQM queue. This keeps its author,
				description, and saved result. Decisions here apply to the KQM database.
			</p>
			<form onSubmit={submit} className="mt-5 flex flex-col gap-4">
				<div>
					<label htmlFor="import-link">Tag helper link</label>
					<Input
						id="import-link"
						name="link"
						type="url"
						required
						placeholder="https://taghelper.simpact.app/id/…"
					/>
				</div>
				{error && <FlowError message={error} />}
				<div>
					<Button variant="secondary" disabled={busy} type="submit">
						{busy ? "Importing…" : "Import submission"}
					</Button>
				</div>
				<p role="status" className="sr-only">
					{busy ? "Saving the existing submission." : ""}
				</p>
			</form>
		</details>
	);
}

function ReviewDetail({ id, access }: { id: string; access: Access }) {
	const [data, setData] = useState<Submission | null>(null);
	const [error, setError] = useState("");
	const [attempt, setAttempt] = useState(0);
	const [decision, setDecision] = useState<Decision | null>(null);
	// biome-ignore lint/correctness/useExhaustiveDependencies: attempt reloads a decision that another reviewer changed.
	useEffect(() => {
		const controller = new AbortController();
		setData(null);
		setError("");
		submissionApi<Submission>(
			`/api/review/submissions/${encodeURIComponent(id)}`,
			{ authorization: access.authorization, signal: controller.signal },
		)
			.then(setData)
			.catch((reason) => {
				if (!controller.signal.aborted) setError(reason.message);
			});
		return () => controller.abort();
	}, [id, access.authorization, attempt]);
	return (
		<>
			<header className="flow-heading">
				<h1>Review submission</h1>
				<p className="font-g-mono text-g-sm break-all">{id}</p>
			</header>
			{error && (
				<FlowError message={error} retry={() => setAttempt((n) => n + 1)} />
			)}
			{!data && !error && (
				<p role="status">Loading submission and matching teams…</p>
			)}
			{data && (
				<>
					<div className="submission-columns">
						<SubmissionDetails data={data} />
						{data.status === "pending" ? (
							<aside className="flow-panel self-start">
								<h2 className="font-g-display text-g-h3 font-semibold">
									Review decision
								</h2>
								<p className="mt-3 text-g-ink-dim">
									Check the result and compare existing entries before you
									publish.
								</p>
								<div className="mt-6 flex flex-col gap-3">
									<Button onClick={() => setDecision({ action: "approve" })}>
										Approve submission
									</Button>
									<Button
										variant="outline"
										className="reject-action"
										onClick={() => setDecision({ action: "reject" })}
									>
										Reject submission
									</Button>
								</div>
								<p className="mt-5 text-g-sm text-g-ink-dim">
									Rejection requires a note. The submission stays available in
									the review history.
								</p>
							</aside>
						) : (
							<DecisionSummary data={data} />
						)}
					</div>
					<section className="mt-10" aria-labelledby="matching-teams">
						<div className="mb-5 flex flex-wrap items-center justify-between gap-3">
							<h2
								id="matching-teams"
								className="font-g-display text-g-h2 font-semibold"
							>
								Existing simulations with this team
							</h2>
							<span className="text-g-sm text-g-ink-dim">
								{data.duplicates?.length ?? 0} matches
								{data.duplicates?.length === 100 ? " (first 100)" : ""}
							</span>
						</div>
						{!data.duplicates?.length ? (
							<div className="flow-panel">
								<p>No other published simulations use this team.</p>
								{data.status === "pending" && (
									<p className="mt-2 text-g-sm text-g-ink-dim">
										You can approve this submission as a new entry.
									</p>
								)}
							</div>
						) : (
							<div className="flex flex-col gap-4">
								{data.duplicates.map((entry) => (
									<article className="flow-panel" key={entry._id}>
										<div className="flex flex-wrap items-center justify-between gap-4">
											<h3 className="font-g-mono text-g-h3 font-semibold">
												{Math.round(
													entry.summary?.mean_dps_per_target ?? 0,
												).toLocaleString()}{" "}
												<span className="font-g-body text-g-sm font-normal text-g-ink-dim">
													DPS per target
												</span>
											</h3>
											<span className="text-g-sm text-g-ink-dim">
												{entry.summary?.sim_duration?.mean?.toFixed(1)} seconds
											</span>
										</div>
										<p className="mt-3 whitespace-pre-wrap break-words leading-relaxed">
											{entry.description}
										</p>
										<p className="mt-3 text-g-sm text-g-ink-dim">
											By <bdi>{entry.submitter}</bdi>
										</p>
										<div className="mt-5 flex flex-wrap gap-3">
											<Button variant="secondary" asChild>
												<a
													href={`/db/${entry._id}`}
													target="_blank"
													rel="noreferrer"
												>
													Compare result{" "}
													<span className="sr-only">
														{entry._id} in a new tab
													</span>
													↗
												</a>
											</Button>
											{data.status === "pending" && (
												<Button
													variant="outline"
													onClick={() =>
														setDecision({ action: "replace", entry })
													}
												>
													Replace this entry
													<span className="sr-only"> {entry._id}</span>
												</Button>
											)}
										</div>
									</article>
								))}
							</div>
						)}
					</section>
					{decision && (
						<DecisionDialog
							decision={decision}
							data={data}
							access={access}
							close={() => setDecision(null)}
							saved={(next) => {
								setData({
									...next,
									duplicates: data.duplicates?.filter(
										(entry) => entry._id !== next.publishedId,
									),
								});
								setDecision(null);
								requestAnimationFrame(() =>
									document.getElementById("decision-heading")?.focus(),
								);
							}}
						/>
					)}
				</>
			)}
		</>
	);
}

function DecisionDialog({
	decision,
	data,
	access,
	close,
	saved,
}: {
	decision: Decision;
	data: Submission;
	access: Access;
	close: () => void;
	saved: (next: Submission) => void;
}) {
	const dialog = useRef<HTMLDialogElement>(null);
	const trigger = useRef(document.activeElement as HTMLElement | null);
	const [busy, setBusy] = useState(false);
	const [error, setError] = useState("");
	const [noteError, setNoteError] = useState(false);
	const title =
		decision.action === "approve"
			? "Approve this submission?"
			: decision.action === "reject"
				? "Reject this submission?"
				: "Replace this published entry?";
	const action =
		decision.action === "approve"
			? "Approve and publish"
			: decision.action === "reject"
				? "Reject submission"
				: "Replace and publish";
	useEffect(() => {
		const element = dialog.current;
		element?.showModal();
		return () => {
			element?.close();
			trigger.current?.focus();
		};
	}, []);
	async function decide(event: FormEvent<HTMLFormElement>) {
		event.preventDefault();
		const form = event.currentTarget;
		const reason = String(new FormData(form).get("reason") ?? "").trim();
		if (decision.action !== "approve" && reason.length < 3) {
			setNoteError(true);
			form.querySelector("textarea")?.focus();
			return;
		}
		setBusy(true);
		setError("");
		setNoteError(false);
		try {
			const result = await submissionApi<Submission>(
				`/api/review/submissions/${data.id}/decision`,
				{
					authorization: access.authorization,
					body: {
						action: decision.action,
						reason,
						replaceId: decision.entry?._id,
					},
				},
			);
			saved(result);
		} catch (reason) {
			setError(
				reason instanceof ApiError
					? reason.message
					: "The decision could not be saved. Check your connection and try again.",
			);
		} finally {
			setBusy(false);
		}
	}
	return (
		<dialog
			ref={dialog}
			className="submission-flow decision-dialog"
			aria-labelledby="decision-title"
			aria-describedby="decision-description"
			onCancel={(event) => {
				if (busy) event.preventDefault();
				else close();
			}}
		>
			<form onSubmit={decide} noValidate>
				<h2
					id="decision-title"
					className="font-g-display text-g-h2 font-semibold"
				>
					{title}
				</h2>
				<p id="decision-description" className="mt-3 text-g-ink-dim">
					{decision.action === "approve"
						? "The saved result will appear in the public KQM database."
						: decision.action === "reject"
							? "The result will stay out of the public database. The submitter can read your note from their saved link."
							: "The existing viewer link will show this submission. The previous entry is saved with this review."}
				</p>
				{decision.entry && (
					<p className="mt-4 rounded-g-md bg-g-surface-2 p-4 text-g-sm break-words">
						Replacing: {decision.entry.description}
					</p>
				)}
				<div className="mt-6">
					<label htmlFor="decision-note">
						Review note{" "}
						{decision.action === "approve" ? "(optional)" : "(required)"}
					</label>
					<textarea
						id="decision-note"
						name="reason"
						rows={4}
						maxLength={2000}
						aria-invalid={noteError}
						aria-describedby="note-error"
					/>
					<p id="note-error" className="field-error">
						{noteError ? "Enter a review note with at least 3 characters." : ""}
					</p>
				</div>
				{error && <FlowError message={error} />}
				<div className="mt-6 flex flex-wrap justify-end gap-3">
					<Button
						variant="secondary"
						type="button"
						disabled={busy}
						onClick={close}
					>
						Cancel
					</Button>
					<Button
						variant={
							decision.action === "reject" || decision.action === "replace"
								? "destructive"
								: "default"
						}
						type="submit"
						disabled={busy}
					>
						{busy ? "Saving decision…" : action}
					</Button>
				</div>
				<p role="status" className="sr-only">
					{busy ? "Saving review decision." : ""}
				</p>
			</form>
		</dialog>
	);
}
