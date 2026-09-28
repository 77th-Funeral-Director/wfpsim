package mitya

import (
	"github.com/genshinsim/gcsim/pkg/core/attacks"
	"github.com/genshinsim/gcsim/pkg/core/attributes"
	"github.com/genshinsim/gcsim/pkg/core/combat"
	"github.com/genshinsim/gcsim/pkg/core/event"
	"github.com/genshinsim/gcsim/pkg/core/info"
)

const prismGainDur = 3*60 - 1

func (c *char) prismInit() {
	c.prismSrc = -1
	c.Core.Events.Subscribe(event.OnStellarConduct, func(args ...any) {
		if c.StatusIsActive(prismGainKey) {
			c.AddStatus(prismGainKey, prismGainDur+c.a1PrismGainDur(), false)
			return
		}
		c.AddStatus(prismGainKey, prismGainDur+c.a1PrismGainDur(), false)
		if c.prismSrc < 0 {
			src := c.Core.F
			c.prismSrc = src
			// TODO: does he gain a stack immediately or after a delay?
			// c.Core.Tasks.Add(func() { c.prismTicker(src) }, 1.5*60)
			c.prismTicker(src)
		}
	}, "mitya-prism")
}

func (c *char) prismTicker(src int) {
	if c.prismSrc != src {
		return
	}

	if !c.StatusIsActive(prismGainKey) {
		c.prismSrc = -1
		return
	}

	c.addPrism()

	c.Core.Tasks.Add(func() { c.prismTicker(src) }, 1.5*60)
}

func (c *char) addPrism() {
	c.prisms++
	if c.getSkillState() == skillStatePress && c.prisms > maxPrismStacks+c.c1MaxPrisms() {
		c.consumePrism(1)
		ai := info.AttackInfo{
			ActorIndex:       c.Index(),
			Abil:             "Prism Surge" + stellarConductText,
			AttackTag:        attacks.AttackTagDirectStellarConduct,
			Element:          attributes.Electro,
			Mult:             skillPrismSurge[c.TalentLvlSkill()] * (1 + c.c6PrismMult()),
			UseEM:            true,
			IgnoreDefPercent: 1,
		}

		ap := combat.NewCircleHitOnTarget(c.Core.Combat.Player(), info.Point{Y: 2}, 5)
		c.Core.QueueAttack(ai, ap, 0, 0)
	} else if c.getSkillState() == skillStatePress && c.skillChargeFinalAnim {
		// preempt the final CA hitmark if we go from 1 prism to 2
		// trigger the wave immediately
	}

	c.prisms = min(c.prisms, maxPrismStacks+c.c1MaxPrisms())

	c.a4OnPrismGain()
	c.c4OnPrismGain()

	src := c.Core.F
	c.prismRemoveSrc = src
	c.Core.Tasks.Add(func() {
		if c.prismRemoveSrc != src {
			return
		}
		c.prisms = 0
	}, 6*60+c.a4PrismDur())
}

func (c *char) consumePrism(count int) {
	for range count {
		c.prisms--
		c.c1OnPrismConsume()
		c.c6OnPrismConsume()
	}
	if c.prisms < 0 {
		c.prisms = 0
	}
}
