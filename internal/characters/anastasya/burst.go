package anastasya

import (
	"fmt"

	"github.com/genshinsim/gcsim/internal/frames"
	"github.com/genshinsim/gcsim/pkg/core/action"
	"github.com/genshinsim/gcsim/pkg/core/attacks"
	"github.com/genshinsim/gcsim/pkg/core/attributes"
	"github.com/genshinsim/gcsim/pkg/core/combat"
	"github.com/genshinsim/gcsim/pkg/core/event"
	"github.com/genshinsim/gcsim/pkg/core/glog"
	"github.com/genshinsim/gcsim/pkg/core/info"
	"github.com/genshinsim/gcsim/pkg/core/player/character"
	"github.com/genshinsim/gcsim/pkg/enemy"
	"github.com/genshinsim/gcsim/pkg/modifier"
)

var (
	burstFrames     []int
	burstHitmark    = 105
	burstLoveKey    = "anastasya-burst-love"
	burstTriumphKey = "anastasya-burst-triumph"

	skillTriumphFrames []int
)

func init() {
	burstFrames = frames.InitAbilSlice(109)

	skillTriumphFrames = frames.InitAbilSlice(26)
}

func (c *char) Burst(p map[string]int) (action.Info, error) {
	if c.getBlessing() == love {
		return c.burstLove(p)
	}
	return action.Info{}, fmt.Errorf("burst not implemented for honor blessing")
}

func (c *char) burstLove(_ map[string]int) (action.Info, error) {
	c.QueueCharTask(func() {
		ai := info.AttackInfo{
			ActorIndex: c.Index(),
			Abil:       "Burst",
			AttackTag:  attacks.AttackTagElementalBurst,
			ICDTag:     attacks.ICDTagElementalBurst,
			ICDGroup:   attacks.ICDGroupDefault,
			StrikeType: attacks.StrikeTypeDefault,
			Element:    attributes.Cryo,
			Durability: 25,
			Mult:       burstLove[c.TalentLvlBurst()],
		}

		ap := combat.NewBoxHitOnTarget(
			c.Core.Combat.Player(),
			info.Point{Y: -5},
			14,
			12,
		)

		c.Core.QueueAttack(ai, ap, 0, 0)
		c.AddStatus(burstLoveKey, 20*60, true)

		c.burstLoveToSeal = true
	}, burstHitmark)

	c.burstSrc = c.Core.F
	c.ConsumeEnergy(7)
	c.SetCD(action.ActionBurst, 15*60)

	return action.Info{
		Frames:          frames.NewAbilFunc(burstFrames),
		AnimationLength: burstFrames[action.InvalidAction],
		CanQueueAfter:   burstFrames[action.ActionSkill], // earliest cancel
		State:           action.BurstState,
	}, nil
}

func (c *char) burstLoveOnSkillTick(ap info.AttackPattern) {
	if !c.StatusIsActive(burstLoveKey) {
		return
	}
	ai := info.AttackInfo{
		ActorIndex:       c.Index(),
		Abil:             "Burst (DoT)",
		AttackTag:        attacks.AttackTagDirectStellarConduct,
		IgnoreDefPercent: 1,
		Element:          attributes.Cryo,
		Mult:             burstLoveCrownSSC[c.TalentLvlBurst()],
	}
	c.toStellarWithDefault(&ai, burstLoveCrownSSC[c.TalentLvlBurst()], burstLoveCrownSSw[c.TalentLvlBurst()])
	c.Core.QueueAttack(ai, ap, 4, 4)
}

func (c *char) burstInit() {
	for _, char := range c.Core.Player.Chars() {
		char.AddReactBonusMod(character.ReactBonusMod{
			Base: modifier.NewBase(burstLoveKey+"-buff", -1),
			Amount: func(ai info.AttackInfo) float64 {
				if !ai.AttackTag.IsStellar() {
					return 0
				}
				if !c.StatusIsActive(burstLoveKey) {
					return 0
				}
				return burstLoveStellar[c.TalentLvlBurst()]
			},
		})
	}

	c.Core.Events.Subscribe(event.OnCharacterSwap, func(args ...interface{}) {
		prev := args[0].(int)
		if prev != c.Index() {
			return
		}
		c.burstLoveToSeal = false
		c.Sealed = true
		c.unsealPoints = 100
		c.Core.Tasks.Add(func() {
			c.burstUnsealTicker(c.burstSrc)
		}, 12)
	}, "anastasya-burst-seal")

	c.Core.Events.Subscribe(event.OnCharacterSwap, func(args ...interface{}) {
		next := args[1].(int)
		if next != c.Index() {
			return
		}
		if !c.StatusIsActive(burstTriumphKey) {
			return
		}
		c.Core.Player.Exec(action.ActionSkill, c.Base.Key, map[string]int{})
	}, "anastasya-burst-triumph")

	c.Core.Events.Subscribe(event.OnEnemyHit, func(args ...interface{}) {
		_, ok := args[0].(*enemy.Enemy)
		if !ok {
			return
		}
		atk, ok := args[1].(*info.AttackEvent)
		if !ok {
			return
		}
		if !atk.Info.AttackTag.IsStellar() {
			return
		}

		if !c.Sealed {
			return
		}

		c.burstUnsealPoints(2)
	}, "anastasya-burst")
}

func (c *char) burstUnsealTicker(src int) {
	if c.burstSrc != src {
		return
	}

	if c.unsealPoints <= 0 {
		return
	}

	// 1 point every 12f = 5 per second
	c.burstUnsealPoints(1)

	c.Core.Tasks.Add(func() {
		c.burstUnsealTicker(src)
	}, 12)
}

func (c *char) burstUnsealPoints(amt float64) {
	if c.unsealPoints <= 0 {
		return
	}
	c.unsealPoints -= amt
	if c.unsealPoints <= 0 {
		c.unsealPoints = 0
		for _, char := range c.Core.Player.Chars() {
			if !char.Sealed {
				continue
			}
			c.Core.Log.NewEvent("swap unsealed", glog.LogCooldownEvent, char.Index())
			if c.Index() == char.Index() {
				c.AddStatus(burstTriumphKey, 20*60, false)
			}
			char.Sealed = false
			c.a1OnUnseal(char)
		}
	}
}

func (c *char) skillTriumph() (action.Info, error) {
	ai := info.AttackInfo{
		ActorIndex:       c.Index(),
		Abil:             "Triumph",
		AttackTag:        attacks.AttackTagDirectStellarConduct,
		IgnoreDefPercent: 1,
		Element:          attributes.Cryo,
		Mult:             burstLoveTriumphSSC[c.TalentLvlBurst()],
	}
	c.toStellarWithDefault(&ai, burstLoveTriumphSSC[c.TalentLvlBurst()], burstLoveTriumphSSw[c.TalentLvlBurst()])
	c.Core.QueueAttack(
		ai,
		combat.NewCircleHitOnTarget(c.Core.Combat.PrimaryTarget(), nil, 5),
		28,
		28,
	)
	c.DeleteStatus(burstTriumphKey)
	return action.Info{
		Frames:          frames.NewAbilFunc(skillTriumphFrames),
		AnimationLength: skillTriumphFrames[action.InvalidAction],
		CanQueueAfter:   skillTriumphFrames[action.ActionDash], // earliest cancel
		State:           action.SkillState,
	}, nil
}
