package mitya

import (
	"github.com/genshinsim/gcsim/pkg/core/attributes"
	"github.com/genshinsim/gcsim/pkg/core/player/character"
	"github.com/genshinsim/gcsim/pkg/modifier"
)

const (
	a1Key = "mitya-a1"
	a4Key = "mitya-a4"
)

func (c *char) a1Init() {
	if c.Base.Ascension < 1 {
		return
	}

	m := make([]float64, attributes.EndStatType)

	for _, char := range c.Core.Player.Chars() {
		char.AddStatMod(character.StatMod{
			Base:         modifier.NewBase(a1Key, -1),
			AffectedStat: attributes.EM,
			Amount: func() []float64 {
				if c.getSkillState() != skillStatePress {
					return nil
				}
				// this probably only updates every 1s or something?
				m[attributes.EM] = 30 * float64(c.prisms)
				return m
			},
		})
	}
}

func (c *char) a1PrismGainDur() int {
	if c.Base.Ascension < 1 {
		return 0
	}
	if c.getSkillState() != skillStateHold {
		return 0
	}
	return 3 * 60
}

func (c *char) a4Init() {
	if c.Base.Ascension < 4 {
		return
	}

	c.a4Buff = make([]float64, attributes.EndStatType)
}

func (c *char) a4OnPrismGain() {
	if c.Base.Ascension < 4 {
		return
	}

	if dur := c.StatusDuration(a4Key); dur > 0 {
		c.a4Stacks = min(c.a4Stacks+1, 4)
		c.ExtendStatus(a4Key, 15*60-dur)
		return
	}

	c.a4Stacks = 1
	c.AddStatMod(character.StatMod{
		Base:         modifier.NewBaseWithHitlag(a4Key, 15*60),
		AffectedStat: attributes.CR,
		Amount: func() []float64 {
			c.a4Buff[attributes.CR] = 0.05 * float64(c.a4Stacks)
			return c.a4Buff
		},
	})
}

func (c *char) a4PrismDur() int {
	if c.Base.Ascension < 4 {
		return 0
	}
	return 4 * 60
}
