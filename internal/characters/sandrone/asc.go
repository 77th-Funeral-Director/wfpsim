package sandrone

import (
	"math"

	"github.com/genshinsim/gcsim/pkg/core/attributes"
	"github.com/genshinsim/gcsim/pkg/core/player/character"
	"github.com/genshinsim/gcsim/pkg/modifier"
)

const (
	a1MaxStacks = 10
	a4Scale     = 0.08
	a4Max       = 160
	a4Key       = "sandrone-a4"
)

func (c *char) a1OnSkill() float64 {
	if c.Base.Ascension < 1 {
		return 1.0
	}

	if c.decode > 50 {
		return 4.0
	}

	return 1.0
}

func (c *char) a1OnDecreaseDecode(amt float64) {
	if c.Base.Ascension < 1 {
		return
	}

	c.a1DecreasedDecode += amt

	// TODO: RefinedTactics lasts 60s
	c.a1RefinedTactics += int(c.a1DecreasedDecode / 10)
	if c.a1RefinedTactics > a1MaxStacks {
		c.a1RefinedTactics = a1MaxStacks
	}
	c.a1DecreasedDecode = math.Mod(c.a1DecreasedDecode, 10)
}

func (c *char) a1OnBurstRayStellar() float64 {
	if c.Base.Ascension < 1 {
		return 1.0
	}

	return 1.0 + float64(c.a1RefinedTactics)*0.1
}

func (c *char) a4Init() {
	if c.Base.Ascension < 4 {
		return
	}

	m := make([]float64, attributes.EndStatType)
	c.AddStatMod(character.StatMod{
		Base:         modifier.NewBase(a4Key, -1),
		Extra:        true,
		AffectedStat: attributes.EM,
		Amount: func() []float64 {
			stats := c.SelectStat(true, attributes.BaseATK, attributes.ATKP, attributes.ATK)
			m[attributes.EM] = min(stats.TotalATK()*a4Scale, a4Max)
			return m
		},
	})
}
