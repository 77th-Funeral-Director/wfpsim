package danica

import (
	"github.com/genshinsim/gcsim/pkg/core/attributes"
	"github.com/genshinsim/gcsim/pkg/core/combat"
	"github.com/genshinsim/gcsim/pkg/core/info"
	"github.com/genshinsim/gcsim/pkg/core/player/character"
	"github.com/genshinsim/gcsim/pkg/enemy"
	"github.com/genshinsim/gcsim/pkg/modifier"
)

const (
	a1Key = "danica-a1"
	a4Key = "danica-a4"
)

func (c *char) a1OnSkill() {
	if c.Base.Ascension < 1 {
		return
	}

	c.a1Ticker(c.skillSrc)
}

func (c *char) a1Ticker(src int) {
	if c.skillSrc != src {
		return
	}

	if !c.StatusIsActive(skillKey) {
		return
	}

	ap := combat.NewCircleHitOnTarget(c.Core.Combat.Player(), nil, 9)
	for _, e := range c.Core.Combat.EnemiesWithinArea(ap, nil) {
		e, ok := e.(*enemy.Enemy)
		if !ok {
			continue
		}

		val := 0.15

		if c.getRadiance() == radianceStellarConduct {
			val = 0.3
			e.AddResistMod(info.ResistMod{
				Base:  modifier.NewBaseWithHitlag(a1Key+"-"+attributes.Cryo.String(), 2*60),
				Ele:   attributes.Cryo,
				Value: -val,
			})
		}
		e.AddResistMod(info.ResistMod{
			Base:  modifier.NewBaseWithHitlag(a1Key+"-"+attributes.Electro.String(), 2*60),
			Ele:   attributes.Electro,
			Value: -val,
		})
	}
}

func (c *char) a4Init() {
	if c.Base.Ascension < 4 {
		return
	}
	c.a4Buff = make([]float64, attributes.EndStatType)
	c.a4Buff[attributes.ATKP] = 0.2
}

func (c *char) a4OnSkillEntry() {
	if c.Base.Ascension < 4 {
		return
	}
	c.AddStatMod(character.StatMod{
		Base:         modifier.NewBaseWithHitlag(a4Key, 25*60),
		AffectedStat: attributes.ATKP,
		Amount: func() []float64 {
			return c.a4Buff
		},
	})
}
