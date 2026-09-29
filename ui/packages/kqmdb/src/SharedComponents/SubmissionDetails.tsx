import { TeamTile } from "@gcsim/components";
import { dynamicKey } from "@gcsim/localization";
import { Button } from "@gcsim/primitives";
import type { db } from "@gcsim/types";
import { useTranslation } from "react-i18next";
import { Link } from "wouter";
import { team } from "../lib/entry";
import { fullDate, type Submission, statusLabels } from "../lib/submissions";

export function FlowError({
	message,
	retry,
}: {
	message: string;
	retry?: () => void;
}) {
	return (
		<div role="alert" className="flow-error">
			<p>{message}</p>
			{retry && (
				<Button variant="secondary" onClick={retry}>
					Try again
				</Button>
			)}
		</div>
	);
}

export function SubmissionBadge({ status }: { status: Submission["status"] }) {
	return (
		<span className={`submission-badge submission-badge-${status}`}>
			{statusLabels[status]}
		</span>
	);
}

export function TeamNames({ entry }: { entry: db.Entry }) {
	const { t } = useTranslation();
	return (
		<>
			{(entry.summary?.char_names ?? [])
				.map((name) => t(dynamicKey(`game:character_names.${name}`)))
				.join(" · ")}
		</>
	);
}

export function SubmissionDetails({ data }: { data: Submission }) {
	const { t } = useTranslation();
	const summary = data.entry.summary;
	return (
		<section
			className="flow-panel submission-summary"
			aria-label="Submitted simulation"
		>
			<div className="flex flex-wrap items-start justify-between gap-4">
				<div>
					<SubmissionBadge status={data.status} />
					<p className="mt-3 text-g-sm text-g-ink-dim">
						Submitted {fullDate(data.submittedAt)}
					</p>
				</div>
				<div className="submission-dps">
					<strong>
						{Math.round(summary?.mean_dps_per_target ?? 0).toLocaleString()}
					</strong>
					<span>DPS per target</span>
				</div>
			</div>
			<div className="mt-6 max-w-[520px]">
				<TeamTile chars={team(data.entry)} className="w-full" />
			</div>
			<h2 className="mt-4 font-g-display text-g-h3 font-semibold">
				<TeamNames entry={data.entry} />
			</h2>
			<div className="mt-2 flex flex-wrap gap-x-4 gap-y-2 text-g-sm text-g-ink-dim">
				<span>{summary?.mode === 2 ? "TTK" : "Duration"}</span>
				<span>{summary?.sim_duration?.mean?.toFixed(1)} seconds</span>
				<span>
					{summary?.target_count ?? 1}{" "}
					{(summary?.target_count ?? 1) === 1 ? "target" : "targets"}
				</span>
			</div>
			<p className="mt-6 whitespace-pre-wrap break-words leading-relaxed">
				{data.entry.description}
			</p>
			<p className="mt-3 text-g-sm text-g-ink-dim">
				Submitted by <bdi>{data.entry.submitter}</bdi>
			</p>
			<details className="mt-6 flow-disclosure">
				<summary>View character builds</summary>
				<ul className="mt-4 grid gap-4 sm:grid-cols-2">
					{summary?.team?.map((character) => (
						<li
							key={character.name}
							className="rounded-g-md bg-g-surface-2 p-4"
						>
							<h3 className="font-semibold">
								{t(dynamicKey(`game:character_names.${character.name}`))} · C
								{character.cons ?? 0}
							</h3>
							<p className="mt-1 text-g-sm text-g-ink-dim">
								Level {character.level} · Talents {character.talents?.attack}/
								{character.talents?.skill}/{character.talents?.burst}
							</p>
							<p className="mt-2">
								{t(dynamicKey(`game:weapon_names.${character.weapon?.name}`))} ·
								R{character.weapon?.refine}
							</p>
							<p className="mt-1 text-g-sm text-g-ink-dim">
								Weapon level {character.weapon?.level}
							</p>
							{Object.entries(character.sets ?? {}).map(([name, count]) => (
								<p key={name} className="mt-2 text-g-sm">
									{count} pieces ·{" "}
									{t(dynamicKey(`game:artifact_names.${name}`))}
								</p>
							))}
						</li>
					))}
				</ul>
			</details>
			<div className="mt-6 flex flex-wrap gap-3">
				<Button variant="secondary" asChild>
					<a
						href={`/submission/${data.id}/results`}
						target="_blank"
						rel="noreferrer"
					>
						Open result viewer <span className="sr-only">in a new tab</span>↗
					</a>
				</Button>
				<a
					className="flow-text-link"
					href={data.sourceUrl}
					target="_blank"
					rel="noreferrer"
				>
					Open original link <span className="sr-only">in a new tab</span>↗
				</a>
			</div>
		</section>
	);
}

export function DecisionSummary({ data }: { data: Submission }) {
	return (
		<section className="flow-panel" aria-labelledby="decision-heading">
			<h2
				id="decision-heading"
				tabIndex={-1}
				className="font-g-display text-g-h3 font-semibold"
			>
				{data.status === "approved" ? "Published to KQM" : "Review decision"}
			</h2>
			<p className="mt-2 text-g-ink-dim">
				{data.status === "approved"
					? "This simulation is now in the public database."
					: "This submission was not added to the database. Use the review note to prepare a new submission."}
			</p>
			{data.reason && (
				<div className="mt-4 rounded-g-md bg-g-surface-2 p-4">
					<h3 className="text-g-sm font-semibold">Review note</h3>
					<p className="mt-2 whitespace-pre-wrap break-words">{data.reason}</p>
				</div>
			)}
			{data.reviewedAt && (
				<p className="mt-4 text-g-sm text-g-ink-dim">
					Reviewed {fullDate(data.reviewedAt)}
					{data.reviewedBy ? ` by ${data.reviewedBy}` : ""}
				</p>
			)}
			<div className="mt-5">
				{data.publishedId ? (
					<Button asChild>
						<Link href={`/db/${data.publishedId}`}>
							View published simulation
						</Link>
					</Button>
				) : (
					<Button variant="secondary" asChild>
						<Link href="/submit">Submit a revised simulation</Link>
					</Button>
				)}
			</div>
		</section>
	);
}
