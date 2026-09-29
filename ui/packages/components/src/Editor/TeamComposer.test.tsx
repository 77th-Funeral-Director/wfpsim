import type { model } from "@gcsim/types";
import { render, screen } from "@testing-library/react";
import userEvent from "@testing-library/user-event";
import { beforeEach, describe, expect, it, vi } from "vitest";

type PickerItem = { key: string; source: string; text: string; label: string };

vi.mock("react-i18next", () => ({
	useTranslation: () => ({ t: (k: string) => k }),
}));

vi.mock("../Cards", () => ({
	TeamCard: ({
		team,
		handleRemove,
		handleAdd,
	}: {
		team: model.Character[];
		handleRemove: (index: number) => () => void;
		handleAdd?: () => void;
	}) => (
		<div data-testid="team-card">
			{team.map((c, index) => (
				<button
					key={c.name ?? index}
					type="button"
					onClick={handleRemove(index)}
				>
					delete-{c.name}
				</button>
			))}
			{handleAdd ? (
				<button type="button" onClick={handleAdd}>
					add
				</button>
			) : null}
		</div>
	),
}));

vi.mock("../common/gcsim", async (orig) => {
	const actual = await orig<typeof import("../common/gcsim")>();
	return {
		...actual,
		characters: ["klee", "amber", "bennett"],
		characterLabel: (k: string) => k,
		OmniSelect: ({
			isOpen,
			items,
			onSelect,
		}: {
			isOpen: boolean;
			items: PickerItem[];
			onSelect: (item: PickerItem) => void;
		}) =>
			isOpen ? (
				<div data-testid="picker">
					{items.map((it) => (
						<button
							type="button"
							key={`${it.source}-${it.key}`}
							onClick={() => onSelect(it)}
						>
							{`${it.source}:${it.key}`}
						</button>
					))}
				</div>
			) : null,
	};
});

import { TeamComposer } from "./TeamComposer";

function char(name: string): model.Character {
	return {
		name,
		level: 80,
		max_level: 90,
		element: "pyro",
		cons: 0,
		weapon: { name: "dullblade", refine: 1, level: 1, max_level: 20 },
		talents: { attack: 6, skill: 6, burst: 6 },
		sets: {},
		stats: new Array(22).fill(0),
		snapshot: new Array(22).fill(0),
	};
}

const source = {
	createCharacter: (key: string) => char(key),
	imported: [],
};

beforeEach(() => {
	vi.clearAllMocks();
});

describe("TeamComposer", () => {
	it("renders the team through TeamCard", () => {
		render(
			<TeamComposer
				parsedTeam={[char("amber"), char("bennett")]}
				error={null}
				config=""
				setConfig={() => {}}
			/>,
		);
		expect(screen.getByText("delete-amber")).toBeTruthy();
		expect(screen.getByText("delete-bennett")).toBeTruthy();
	});

	it("surfaces the validation error through a destructive alert", () => {
		render(
			<TeamComposer
				parsedTeam={[]}
				error="bad action list"
				config=""
				setConfig={() => {}}
			/>,
		);
		expect(screen.getByText("bad action list")).toBeTruthy();
	});

	it("removes only the selected character and preserves other character parameters", async () => {
		const setConfig = vi.fn();
		render(
			<TeamComposer
				parsedTeam={[char("amber"), char("bennett")]}
				error={null}
				config={
					"amber char lvl=1/1 cons=0 talent=1,1,1;\nbennett char lvl=80/90 cons=6 talent=9,9,9;\nbennett char params=[test=1];\ntarget lvl=100;"
				}
				setConfig={setConfig}
				characters={source}
			/>,
		);
		await userEvent.click(screen.getByText("delete-amber"));
		expect(setConfig).toHaveBeenCalledTimes(1);
		const written = setConfig.mock.calls[0][0] as string;
		expect(written).toContain("bennett char");
		expect(written).toContain("bennett char params=[test=1];");
		expect(written).toContain("target lvl=100;");
		expect(written).not.toContain("amber char");
	});

	it("adds a character without replacing existing config text", async () => {
		const setConfig = vi.fn();
		render(
			<TeamComposer
				parsedTeam={[char("amber")]}
				error={null}
				config={
					"# Keep this comment\namber char lvl=80/90 cons=6 talent=9,9,9;\namber char params=[test=1];\nactive amber;"
				}
				setConfig={setConfig}
				characters={source}
			/>,
		);
		await userEvent.click(screen.getByRole("button", { name: "add" }));
		await userEvent.click(screen.getByText("default:klee"));
		const written = setConfig.mock.calls[0][0] as string;
		expect(written).toContain("amber char");
		expect(written).toContain("klee char");
		expect(written).toContain(
			"# Keep this comment\namber char lvl=80/90 cons=6 talent=9,9,9;\namber char params=[test=1];\nactive amber;",
		);
	});

	it("removes character aliases without rewriting the other characters", async () => {
		const setConfig = vi.fn();
		render(
			<TeamComposer
				parsedTeam={[char("keqing"), char("bennett")]}
				error={null}
				config={
					"keq char lvl=90/90; keq add stats\n atk=1000;\n# Keep this comment\nbennett char lvl=80/90;\nbennett char params=[test=1];\nactive bennett;"
				}
				setConfig={setConfig}
				characters={source}
			/>,
		);
		await userEvent.click(screen.getByText("delete-keqing"));
		expect(setConfig).toHaveBeenCalledWith(
			"# Keep this comment\nbennett char lvl=80/90;\nbennett char params=[test=1];\nactive bennett;",
		);
	});

	it("keeps unparsed characters and excludes them from the picker while the config is incomplete", async () => {
		const setConfig = vi.fn();
		const config =
			'amber char lvl=80/90 cons=0 talent=6,6,6;\namber add weapon="dullblade" refine=1 lvl=1/20;\n';
		render(
			<TeamComposer
				parsedTeam={[]}
				error="Missing target and active character"
				config={config}
				setConfig={setConfig}
				characters={source}
			/>,
		);
		await userEvent.click(screen.getByRole("button", { name: "add" }));
		expect(screen.queryByText("default:amber")).toBeNull();
		await userEvent.click(screen.getByText("default:klee"));
		expect(setConfig.mock.calls[0][0]).toContain(config);
		expect(setConfig.mock.calls[0][0]).toContain("klee char");
	});

	it("hides the add affordance when no character source is injected", () => {
		render(
			<TeamComposer
				parsedTeam={[char("amber")]}
				error={null}
				config=""
				setConfig={() => {}}
			/>,
		);
		expect(screen.queryByRole("button", { name: "add" })).toBeNull();
	});
});
