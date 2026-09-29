package mitya

import (
	"github.com/genshinsim/gcsim/pkg/core/attacks"
	"github.com/genshinsim/gcsim/pkg/core/attributes"
	"github.com/genshinsim/gcsim/pkg/core/combat"
	"github.com/genshinsim/gcsim/pkg/core/event"
	"github.com/genshinsim/gcsim/pkg/core/info"
	"github.com/genshinsim/gcsim/pkg/core/player/character"
	"github.com/genshinsim/gcsim/pkg/enemy"
	"github.com/genshinsim/gcsim/pkg/modifier"
)

const (
	c1Key    = "mitya-c1"
	c1ICDKey = "mitya-c1-icd"
	c2Key    = "mitya-c2"
	c4Key    = "mitya-c4"
	c6Key    = "mitya-c6"
)

func (c *char) c1MaxPrisms() int {
	if c.Base.Cons < 1 {
		return 0
	}
	return 1
}

func (c *char) c1OnSkill() {
	if c.Base.Cons < 1 {
		return
	}
	c.addPrism()
}

func (c *char) c1OnPrismConsume() {
	if c.Base.Cons < 1 {
		return
	}
	if c.StatusIsActive(c1ICDKey) {
		return
	}
	c.AddStatus(c1ICDKey, 8*60, true)
	c.AddEnergy(c1Key, 10)
}

// While the Energy Device is present, the active character's CRIT DMG is increased by 50%;
func (c *char) c1Init() {
	if c.Base.Cons < 1 {
		return
	}
	m := make([]float64, attributes.EndStatType)
	m[attributes.CD] = 0.5
	for _, char := range c.Core.Player.Chars() {
		char.AddStatMod(character.StatMod{
			Base:         modifier.NewBase(c1Key, -1),
			AffectedStat: attributes.CD,
			Amount: func() []float64 {
				if c.Core.Player.Active() != char.Index() {
					return nil
				}

				if !c.StatusIsActive(skillKey) {
					return nil
				}

				return m
			},
		})
	}
}

func (c *char) c2Init() {
	if c.Base.Cons < 2 {
		return
	}

	m := make([]float64, attributes.EndStatType)
	for _, char := range c.Core.Player.Chars() {
		char.AddStatMod(character.StatMod{
			Base:         modifier.NewBase(c2Key, -1),
			AffectedStat: attributes.EM,
			Amount: func() []float64 {
				switch c.getSkillState() {
				case skillStatePress:
					m[attributes.EM] = 100
				case skillStateHold:
					if c.Core.Player.Active() != char.Index() {
						return nil
					}
					m[attributes.EM] = 200
				default:
					return nil
				}
				return m
			},
		})
	}
}

func (c *char) c2OnSkill() {
	if c.Base.Cons < 2 {
		return
	}

	c.c2Src = c.Core.F
	c.c2Ticker(c.c2Src)
}

func (c *char) c2Ticker(src int) {
	if !c.StatusIsActive(skillKey) {
		return
	}

	if c.c2Src != src {
		return
	}

	c.QueueCharTask(func() { c.c2Ticker(src) }, 1*60)

	ap := combat.NewCircleHitOnTarget(c.Core.Combat.Player(), nil, 10)
	for _, e := range c.Core.Combat.EnemiesWithinArea(ap, nil) {
		e, ok := e.(*enemy.Enemy)
		if !ok {
			continue
		}
		e.AddResistMod(info.ResistMod{
			Base:  modifier.NewBaseWithHitlag(c2Key+"-"+attributes.Cryo.String(), 2*60),
			Ele:   attributes.Cryo,
			Value: -0.2,
		})

		e.AddResistMod(info.ResistMod{
			Base:  modifier.NewBaseWithHitlag(c2Key+"-"+attributes.Electro.String(), 2*60),
			Ele:   attributes.Electro,
			Value: -0.2,
		})
	}
}

func (c *char) c4OnPrismGain() {
	if c.Base.Cons < 4 {
		return
	}

	c.c4Stacks++

	if c.c4Stacks < 5 {
		return
	}

	c.c4Stacks = 0
	ai := info.AttackInfo{
		ActorIndex:       c.Index(),
		Abil:             "Mitya C4",
		AttackTag:        attacks.AttackTagDirectStellarConduct,
		Element:          attributes.Electro,
		Mult:             4,
		UseEM:            true,
		IgnoreDefPercent: 1,
	}

	ap := combat.NewCircleHitOnTarget(c.Core.Combat.Player(), info.Point{Y: 2}, 5)
	c.Core.QueueAttack(ai, ap, 0, 0)
}

func (c *char) c6Init() {
	if c.Base.Cons < 6 {
		return
	}

	c.Core.Events.Subscribe(event.OnApplyAttack, func(args ...any) {
		atk := args[0].(*info.AttackEvent)

		if atk.Info.AttackTag != attacks.AttackTagDirectStellarConduct {
			return
		}

		atk.Info.Elevation += 0.2
	}, c6Key)
}

func (c *char) c6PrismMult() float64 {
	if c.Base.Cons < 6 {
		return 0
	}

	if c.getSkillState() != skillStatePress {
		return 0
	}

	return 0.15
}

func (c *char) c6OnPrismConsume() {
	if c.Base.Cons < 6 {
		return
	}

	if c.getSkillState() != skillStateHold {
		return
	}

	if c.Core.Combat.Rand.Float64() < 0.3 {
		c.addPrism()
	}
}
