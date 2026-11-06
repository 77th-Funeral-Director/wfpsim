package anastasya

import (
	"github.com/genshinsim/gcsim/pkg/core/attributes"
	"github.com/genshinsim/gcsim/pkg/core/player/character"
	"github.com/genshinsim/gcsim/pkg/modifier"
	"github.com/genshinsim/gcsim/pkg/reactable"
)

const (
	a1Key       = "anastasya-a1"
	a1SealedKey = "anastasya-a1-sealed"
)

func (c *char) a1Init() {
	if c.Base.Ascension < 1 {
		return
	}
	m := make([]float64, attributes.EndStatType)
	baseBuff := 0.12
	for _, char := range c.Core.Player.Chars() {
		char.AddStatMod(character.StatMod{
			Base: modifier.NewBase(a1Key+"-buff", -1),
			Amount: func() []float64 {
				if char.Sealed {
					return nil
				}
				if !c.StatusIsActive(a1Key) {
					return nil
				}
				if char.StatusIsActive(a1SealedKey) {
					return nil
				}

				sealed := c.a1SealedCount()
				unsealed := len(c.Core.Player.Chars()) - sealed

				m[attributes.CD] = baseBuff + baseBuff*float64(sealed)/float64(unsealed)

				return m
			},
		})
	}
}

func (c *char) a1OnUnseal(char *character.CharWrapper) {
	if c.Base.Ascension < 1 {
		return
	}
	char.AddStatus(a1SealedKey, 8*60, true)
}

func (c *char) a1SealedCount() int {
	count := 0
	for _, char := range c.Core.Player.Chars() {
		if char.Sealed || char.StatusIsActive(a1SealedKey) {
			count++
		}
	}
	return count
}

func (c *char) a1OnSkill() {
	if c.Base.Ascension < 1 {
		return
	}

	for _, char := range c.Core.Player.Chars() {
		char.AddStatus(a1Key, 20*60, true)
	}
}

func (c *char) a4Init() {
	if c.Base.Ascension < 4 {
		return
	}

	c.Core.Flags.Custom[reactable.StellarSwirlBonusMultKey] = 0.15
	c.Core.Flags.Custom[reactable.PolestarFieldBonusStacksKey] = 2
}
