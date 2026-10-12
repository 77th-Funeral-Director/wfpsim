import type { model } from "@gcsim/types";
import { render, screen, within } from "@testing-library/react";
import userEvent from "@testing-library/user-event";
import { createInstance } from "i18next";
import { I18nextProvider } from "react-i18next";
import { beforeAll, describe, expect, it } from "vitest";
import english from "../../../localization/src/locales/English.json";
import { ResultWarnings } from "./ResultWarnings";

const i18n = createInstance();
beforeAll(async () => {
	await i18n.init({
		lng: "en",
		resources: {
			en: {
				translation: english,
				game: { character_names: { furina: "Furina", skirk: "Skirk" } },
			},
		},
	});
});

const energyTitle = english.warnings.energy_title;
const energyResult: model.SimulationResult = {
	character_details: [{ name: "skirk" }, { name: "furina" }],
	statistics: {
		warnings: { insufficient_energy: true },
		failed_actions: [
			{ insufficient_energy: { max: 0 } },
			{ insufficient_energy: { mean: 2.07, min: 0.95, max: 8.68, sd: 1.29 } },
		],
	},
};

function view(data: model.SimulationResult) {
	return (
		<I18nextProvider i18n={i18n}>
			<ResultWarnings data={data} />
		</I18nextProvider>
	);
}

describe("ResultWarnings", () => {
	it("shows the energy warning and the affected character's delay in seconds", () => {
		render(view(energyResult));
		const alert = screen.getByRole("alert", { name: energyTitle });
		expect(within(alert).getByText(english.warnings.energy_body)).toBeVisible();
		expect(within(alert).getByText("Furina")).toBeVisible();
		expect(within(alert).queryByText("Skirk")).not.toBeInTheDocument();
		expect(
			within(alert).getByText(
				"mean: 2.07s | min: 0.95s | max: 8.68s | std: 1.29s",
			),
		).toBeVisible();
	});

	it("dismisses only the selected warning and restores it for a new result", async () => {
		const data = {
			...energyResult,
			statistics: {
				...energyResult.statistics,
				warnings: { insufficient_energy: true, target_overlap: true },
			},
		};
		const { rerender } = render(view(data));
		await userEvent.click(
			within(screen.getByRole("alert", { name: energyTitle })).getByRole(
				"button",
			),
		);
		expect(
			screen.queryByRole("alert", { name: energyTitle }),
		).not.toBeInTheDocument();
		expect(
			screen.getByRole("alert", {
				name: english.warnings.position_overlap_title,
			}),
		).toBeVisible();
		rerender(view(data));
		expect(
			screen.queryByRole("alert", { name: energyTitle }),
		).not.toBeInTheDocument();
		rerender(view({ ...data }));
		expect(screen.getByRole("alert", { name: energyTitle })).toBeVisible();
	});

	it("shows no warnings for results with missing or false warning flags", () => {
		const { rerender } = render(view({}));
		expect(screen.queryByRole("alert")).not.toBeInTheDocument();
		rerender(
			view({ statistics: { warnings: { insufficient_energy: false } } }),
		);
		expect(screen.queryByRole("alert")).not.toBeInTheDocument();
	});

	it("keeps the warning readable when old results have no delay details", () => {
		render(view({ statistics: { warnings: { insufficient_energy: true } } }));
		expect(screen.getByRole("alert", { name: energyTitle })).toBeVisible();
		expect(screen.queryByText(/undefined/)).not.toBeInTheDocument();
		expect(screen.queryByRole("list")).not.toBeInTheDocument();
	});

	it.each([
		["burst_cd", "burst_cd"],
		["skill_cd", "skill"],
		["insufficient_stamina", "stamina"],
		["swap_cd", "swap"],
		["dash_cd", "dash"],
	] as const)(
		"shows the %s warning with its own delay details",
		(flag, key) => {
			render(
				view({
					character_details: [{ name: "furina" }],
					statistics: {
						warnings: { [flag]: true },
						failed_actions: [{ [flag]: { mean: 1, min: 0, max: 2, sd: 0.5 } }],
					},
				}),
			);
			const alert = screen.getByRole("alert", {
				name: english.warnings[`${key}_title`],
			});
			expect(
				within(alert).getByText(english.warnings[`${key}_data_header`]),
			).toBeVisible();
			expect(
				within(alert).getByText("mean: 1s | min: 0s | max: 2s | std: 0.5s"),
			).toBeVisible();
		},
	);

	it("shows incomplete characters and the ignore-energy mode warning", () => {
		render(
			view({
				incomplete_characters: ["skirk"],
				simulator_settings: { ignore_burst_energy: true },
			}),
		);
		const incomplete = screen.getByRole("alert", {
			name: english.warnings.incomplete_char_title,
		});
		expect(within(incomplete).getByText("Skirk")).toBeVisible();
		expect(
			within(incomplete).getByRole("link", { name: "gcsim Discord" }),
		).toHaveAttribute("href", "https://discord.gg/m7jvjdxx7q");
		expect(
			screen.getByRole("alert", {
				name: english.warnings.ignore_burst_energy_title,
			}),
		).toHaveAttribute("data-variant", "destructive");
	});
});
