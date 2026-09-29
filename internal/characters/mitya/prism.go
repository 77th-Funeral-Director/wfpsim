package mitya

import (
	"fmt"

	"github.com/genshinsim/gcsim/pkg/core/attacks"
	"github.com/genshinsim/gcsim/pkg/core/attributes"
	"github.com/genshinsim/gcsim/pkg/core/combat"
	"github.com/genshinsim/gcsim/pkg/core/event"
	"github.com/genshinsim/gcsim/pkg/core/glog"
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
		if c.Core.Combat.Debug {
			c.Core.Log.NewEvent(fmt.Sprintf("mitya gained prism beyond max (%d) during tap skill", maxPrismStacks+c.c1MaxPrisms()), glog.LogCharacterEvent, c.Index())
		}
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
	} else {
		c.prisms = min(c.prisms, maxPrismStacks+c.c1MaxPrisms())
		if c.Core.Combat.Debug {
			c.Core.Log.NewEvent(fmt.Sprintf("mitya gained prism (%d)", c.prisms), glog.LogCharacterEvent, c.Index())
		}
	}

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
		if c.Core.Combat.Debug {
			c.Core.Log.NewEvent(fmt.Sprintf("mitya consumed prism (%d)", c.prisms), glog.LogCharacterEvent, c.Index())
		}
		c.c1OnPrismConsume()
		c.c6OnPrismConsume()
	}
	if c.prisms < 0 {
		c.prisms = 0
	}
}
