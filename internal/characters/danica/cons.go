package danica

const c1ICDKey = "danica-c1-icd"

func (c *char) c1OnSkillEntry() {
	if c.Base.Cons < 1 {
		return
	}

	if c.StatusIsActive(c1ICDKey) {
		return
	}

	c.AddStatus(c1ICDKey, 14*60, true)
	c.AddEnergy("danica-c1", 20)
}

func (c *char) c1BonusAtkBuff() float64 {
	if c.Base.Cons < 1 {
		return 0
	}

	return 0.8
}
