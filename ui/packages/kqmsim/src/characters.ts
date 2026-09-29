import {
	characters,
	type TeamComposerCharacterSource,
} from "@gcsim/components";
import type { model } from "@gcsim/types";

export const kqmCharacterNames = {
	alyosha: "Alyosha",
	linnea: "Linnea",
	lohen: "Lohen",
	sandrone: "Sandrone",
	vesna: "Vesna",
	vodyanitsa: "Vodyanitsa",
	zibai: "Zibai",
};

for (const key of Object.keys(kqmCharacterNames)) {
	if (!characters.includes(key)) characters.push(key);
}

function createCharacter(name: string): model.Character {
	return {
		name,
		level: 80,
		max_level: 90,
		cons: 0,
		weapon: { name: "dullblade", refine: 1, level: 1, max_level: 20 },
		talents: { attack: 6, skill: 6, burst: 6 },
		stats: Array(22).fill(0),
		snapshot: Array(22).fill(0),
		sets: {},
	};
}

export const teamCharacters: TeamComposerCharacterSource = {
	createCharacter,
};
