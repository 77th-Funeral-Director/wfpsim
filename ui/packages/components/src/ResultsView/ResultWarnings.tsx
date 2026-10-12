import { dynamicKey } from "@gcsim/localization";
import { Alert, AlertDescription, AlertTitle, Button } from "@gcsim/primitives";
import type { model } from "@gcsim/types";
import { X } from "lucide-react";
import { useState } from "react";
import { Trans, useTranslation } from "react-i18next";

const actionWarnings = [
	["energy", "insufficient_energy"],
	["burst_cd", "burst_cd"],
	["skill", "skill_cd"],
	["stamina", "insufficient_stamina"],
	["swap", "swap_cd"],
	["dash", "dash_cd"],
] as const;

export function ResultWarnings({ data }: { data: model.SimulationResult }) {
	const { t, i18n } = useTranslation();
	const [dismissal, setDismissal] = useState<{
		result: model.SimulationResult;
		keys: string[];
	} | null>(null);
	const dismissed = dismissal?.result === data ? dismissal.keys : [];

	const warnings: {
		key: string;
		active: boolean;
		stat?: keyof model.FailedActions;
	}[] = [
		{
			key: "incomplete_char",
			active: Boolean(data.incomplete_characters?.length),
		},
		{
			key: "position_overlap",
			active: data.statistics?.warnings?.target_overlap ?? false,
		},
		...actionWarnings.map(([key, stat]) => ({
			key,
			stat,
			active: data.statistics?.warnings?.[stat] ?? false,
		})),
		{
			key: "ignore_burst_energy",
			active: data.simulator_settings?.ignore_burst_energy ?? false,
		},
	];

	function seconds(value?: number) {
		return value == null
			? "—"
			: value.toLocaleString(i18n.language, { maximumFractionDigits: 2 }) +
					t("result.seconds_short");
	}

	return (
		<div className="flex w-full flex-col gap-2 mb-4 empty:mb-0">
			{warnings
				.filter(({ key, active }) => active && !dismissed.includes(key))
				.map(({ key, stat }) => {
					const title = t(dynamicKey(`warnings.${key}_title`));
					const delays =
						data.statistics?.failed_actions?.flatMap((actions, index) => {
							const stats = stat ? actions[stat] : undefined;
							const name = data.character_details?.[index]?.name;
							return stats && (stats.max ?? 0) > 0 && name
								? [{ name, stats }]
								: [];
						}) ?? [];
					return (
						<Alert
							key={key}
							aria-label={title}
							variant={
								key === "ignore_burst_energy" ? "destructive" : "warning"
							}
						>
							<div className="col-start-2 flex min-w-0 items-start justify-between gap-3">
								<AlertTitle className="line-clamp-none">{title}</AlertTitle>
								<Button
									variant="ghost"
									size="icon-sm"
									className="shrink-0"
									aria-label={`${t("viewer.close")}: ${title}`}
									onClick={() =>
										setDismissal({ result: data, keys: [...dismissed, key] })
									}
								>
									<X />
								</Button>
							</div>
							<AlertDescription className="block text-current">
								<p>
									{key === "incomplete_char" ? (
										<Trans i18nKey="warnings.incomplete_char_body">
											<a
												href="https://discord.gg/m7jvjdxx7q"
												target="_blank"
												rel="noreferrer"
											>
												gcsim Discord
											</a>
										</Trans>
									) : (
										t(dynamicKey(`warnings.${key}_body`))
									)}
								</p>
								{key === "incomplete_char" && (
									<div className="pt-2 font-g-mono text-g-xs">
										<strong>{t("warnings.incomplete_char_data_header")}</strong>
										<ul className="list-disc pl-4">
											{data.incomplete_characters?.map((name) => (
												<li key={name}>
													{t(dynamicKey(`game:character_names.${name}`))}
												</li>
											))}
										</ul>
									</div>
								)}
								{delays.length > 0 && (
									<div className="pt-2 font-g-mono text-g-xs">
										<strong>
											{t(dynamicKey(`warnings.${key}_data_header`))}
										</strong>
										<ul className="list-disc pl-4">
											{delays.map(({ name, stats }) => (
												<li key={name}>
													{t(dynamicKey(`game:character_names.${name}`))}{" "}
													<span className="inline-block">
														mean: {seconds(stats.mean)} | min:{" "}
														{seconds(stats.min)} | max: {seconds(stats.max)} |
														std: {seconds(stats.sd)}
													</span>
												</li>
											))}
										</ul>
									</div>
								)}
							</AlertDescription>
						</Alert>
					);
				})}
		</div>
	);
}
