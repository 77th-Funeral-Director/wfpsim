package valeriy

import (
	"github.com/genshinsim/gcsim/internal/frames"
	"github.com/genshinsim/gcsim/pkg/core/action"
	"github.com/genshinsim/gcsim/pkg/core/attacks"
	"github.com/genshinsim/gcsim/pkg/core/attributes"
	"github.com/genshinsim/gcsim/pkg/core/combat"
	"github.com/genshinsim/gcsim/pkg/core/info"
)

var (
	burstFrames []int
	burstTicks  = [][3]float64{
		{0.46, 0.51, 0.2},
		{0.66, 0.46, 0.2},
		{0.48, 0.69, 0.2},
		{0.61, 0.61, 0.2},
		{0.36, 0.91, 0.2},
		{0.71, 0.61, 0.2},
	}
	burstTickDelay = 0.15
)

const (
	burstKey     = "valeriy-burst"
	burstDur     = 12 * 60
	burstHitmark = 43
)

func init() {
	burstFrames = frames.InitAbilSlice(60)
}

func (c *char) Burst(p map[string]int) (action.Info, error) {
	c.QueueCharTask(func() {
		ai := info.AttackInfo{
			ActorIndex: c.Index(),
			Abil:       "Burst",
			AttackTag:  attacks.AttackTagElementalBurst,
			ICDTag:     attacks.ICDTagNone,
			ICDGroup:   attacks.ICDGroupDefault,
			Element:    attributes.Electro,
			Durability: 25,
			Mult:       burst[c.TalentLvlBurst()],
		}
		ap := combat.NewCircleHitOnTarget(c.Core.Combat.Player(), nil, 5)
		c.Core.QueueAttack(ai, ap, 0, 0)

		c.AddStatus(burstKey, burstDur, true)

		src := c.Core.F
		c.burstSrc = src
		c.burstTicker(src)

		c.a1OnBurst()
	}, burstHitmark)

	c.SetCD(action.ActionBurst, 15*60)
	c.ConsumeEnergy(7)

	return action.Info{
		Frames:          frames.NewAbilFunc(burstFrames),
		AnimationLength: burstFrames[action.InvalidAction],
		CanQueueAfter:   burstFrames[action.ActionSwap], // earliest cancel
		State:           action.BurstState,
	}, nil
}

func (c *char) burstTicker(src int) {
	if c.burstSrc != src {
		return
	}

	if !c.StatusIsActive(burstKey) {
		return
	}

	delays := burstTicks[c.Core.Rand.Int()%len(burstTicks)]

	for _, d := range []float64{0.0, delays[0], delays[0] + delays[1]} {
		c.Core.Tasks.Add(func() {
			if c.burstSrc != src {
				return
			}

			if !c.StatusIsActive(burstKey) {
				return
			}

			ai := info.AttackInfo{
				ActorIndex:     c.Index(),
				Abil:           "Burst (Tick)",
				AdditionalTags: []attacks.AttackTag{attacks.AttackTagValeriySpecial},
				AttackTag:      attacks.AttackTagElementalBurst,
				ICDTag:         attacks.ICDTagElementalBurst,
				ICDGroup:       attacks.ICDGroupValeriyBurst,
				Element:        attributes.Electro,
				Durability:     25,
				Mult:           burstDot[c.TalentLvlBurst()],
			}

			c.Core.QueueAttack(ai, combat.NewCircleHitOnTarget(c.Core.Combat.Player(), nil, 5), 0, 0, c.c6CB)
		}, int((d+burstTickDelay)*60+0.5))
	}

	c.Core.Tasks.Add(func() { c.burstTicker(src) }, int((delays[0]+delays[1]+delays[2])*60+0.5))
}
