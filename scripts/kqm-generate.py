"""Reapply the seven KQM character keys after the upstream data pipeline."""
from pathlib import Path
names = ["Alyosha", "Linnea", "Lohen", "Sandrone", "Vesna", "Vodyanitsa", "Zibai"]
p = Path("pkg/core/keys/character.dm.go")
s = p.read_text()
for name in names:
    if "// " + name.lower() + "\n" in s:
        continue
    s = s.replace("\tInvalidChar ", "\t" + name + " // " + name.lower() + "\n\tInvalidChar ", 1)
    s = s.replace('\t"invalidchar",', '\t"' + name.lower() + '",\n\t"invalidchar",', 1)
    s = s.replace("\tInvalidChar,", "\t" + name + ",\n\tInvalidChar,", 1)
p.write_text(s)

# Extra ICD definitions are kept separate from the upstream datamine.
import re
groups = {'AlyoshaBurst': ['114', '[]float64{1, 0, 0, 0, 0, 0, 0, 0}', '[]float64{1, 1, 1, 1, 1, 1, 1, 1}'], 'LohenSkillAttack': ['300', '[]float64{1, 0, 1, 0, 1, 0, 1, 0, 1, 0, 1, 0}', '[]float64{1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1}'], 'SandroneSweepingFire': ['85', '[]float64{1, 0}', '[]float64{1, 1}'], 'VesnaSkill': ['120', '[]float64{1, 0,}', '[]float64{1, 1}']}

p = Path("pkg/core/attacks/icd_groups.dm.go")
s = p.read_text()
for name, values in groups.items():
    key = "ICDGroup" + name
    if re.search(r"^\s*" + key + r"\s*$", s, re.M):
        continue
    s = s.replace("\n)", "\n\t" + key + "\n)", 1)
    for field, value in zip(["ResetTimer", "EleApplicationSequence", "DamageSequence"], values):
        value = value.replace("[]float64", "")
        start = s.index("var ICDGroup" + field)
        end = s.index("\n}", start)
        s = s[:end] + "\n\t" + key + ": " + value + "," + s[end:]
p.write_text(s)
p = Path("pkg/core/attacks/icd_tags.dm.go")
s = p.read_text()
for name in ["SandroneExtraAttackSweepingFire", "SandroneExtraAttackLaser"]:
    key = "ICDTag" + name
    if key not in s:
        s = s.replace("\n)", "\n\t" + key + "\n)", 1)
p.write_text(s)
