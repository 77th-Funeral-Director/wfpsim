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
	burstFrames  []int
	burstHitmark = 105
	burstLoveKey = "anastasya-burst-love"

	burstTriumphKey = "anastasya-burst-triumph"

	burstHonorFrames  []int
	burstHonorSeal    = 100
	burstHonorHitmark = 116
	burstHonorKey     = "anastasya-burst-honor"

	skillTriumphFrames []int

	skillDoomFrames   [3][]int
	skillDoomHitmarks = [3][]int{{5, 10, 15, 25, 30, 35}, {19}, {74}}
)

func init() {
	burstFrames = frames.InitAbilSlice(109)
	burstHonorFrames = frames.InitAbilSlice(108)
	burstHonorFrames[action.ActionSwap] = 106

	skillTriumphFrames = frames.InitAbilSlice(26)

	skillDoomFrames[0] = frames.InitAbilSlice(34)
	skillDoomFrames[1] = frames.InitAbilSlice(45)
	skillDoomFrames[2] = frames.InitAbilSlice(100)
}

func (c *char) Burst(p map[string]int) (action.Info, error) {
	switch c.getBlessing() {
	case love:
		return c.burstLove(p)
	case honor:
		return c.burstHonor(p)
	default:
		return action.Info{}, fmt.Errorf("unknown blessing type: %v", c.getBlessing())
	}
}

func (c *char) burstLove(_ map[string]int) (action.Info, error) {
	c.QueueCharTask(func() {
		ai := info.AttackInfo{
			ActorIndex: c.Index(),
			Abil:       "To the New World",
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
		// TODO: the timer freeze mechanic so it's just extending the existing duration
		c.ExtendStatus(skillTapKey, 20*60)

		for _, char := range c.Core.Player.Chars() {
			char.Sealed = false
		}
		c.burstLoveToSeal = true
	}, burstHitmark)

	c.burstSrc = c.Core.F
	c.ConsumeEnergy(7)
	c.SetCD(action.ActionBurst, 18*60)

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

		if c.burstLoveToSeal {
			c.burstLoveToSeal = false
			c.Sealed = true
			c.unsealPoints = 100
			c.Core.Tasks.Add(func() {
				c.burstUnsealTicker(c.burstSrc)
			}, 12)
			return
		}

		for _, char := range c.Core.Player.Chars() {
			if char.Index() == c.Index() {
				continue
			}

			if c.unsealReady[char.Index()] {
				char.Sealed = false
			}
		}
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
	c.Core.Log.NewEvent(fmt.Sprintf("unseal points (%v)", c.unsealPoints), glog.LogCharacterEvent, c.Index()).Write("amt", amt)
	if c.unsealPoints <= 0 {
		c.unsealPoints = 0
		for _, char := range c.Core.Player.Chars() {
			if !char.Sealed {
				continue
			}

			if c.Index() == char.Index() {
				c.Core.Log.NewEvent("swap unsealed", glog.LogCooldownEvent, char.Index())
				c.AddStatus(burstTriumphKey, 20*60, false)
				char.Sealed = false
				c.a1OnUnseal(char)
			} else {
				// this char is ready to be unsealed using the burst skill
				c.unsealReady[char.Index()] = true
			}
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

func (c *char) burstHonor(_ map[string]int) (action.Info, error) {
	c.QueueCharTask(func() {
		c.AddStatus(burstHonorKey, 14*60, true)

		// TODO: the timer freeze mechanic so it's just extending the existing duration
		c.ExtendStatus(skillHoldKey, 14*60)

		// up to 3 chars, but we max at 3 non anastasya chars, so we can just do all of them
		for _, char := range c.Core.Player.Chars() {
			if char.Index() == c.Index() {
				continue
			}
			char.Sealed = true
		}
		c.unsealPoints = 120
		c.Core.Tasks.Add(func() {
			c.burstUnsealTicker(c.burstSrc)
		}, 12)
	}, burstHonorSeal)

	c.QueueCharTask(func() {
		ai := info.AttackInfo{
			ActorIndex: c.Index(),
			Abil:       "The Evernight Yields to Dawn",
			Element:    attributes.Cryo,
		}
		ssc := burstHonorSSC[c.TalentLvlBurst()] * (1 + c.c2MultBonus())
		ssw := burstHonorSSw[c.TalentLvlBurst()] * (1 + c.c2MultBonus())
		c.toStellarWithDefault(&ai, ssc, ssw)

		ap := combat.NewBoxHitOnTarget(
			c.Core.Combat.Player(),
			info.Point{Y: -5},
			14,
			12,
		)
		c.Core.QueueAttack(ai, ap, 0, 0)
	}, burstHonorHitmark)

	c.burstSrc = c.Core.F
	c.ConsumeEnergy(7)
	c.SetCD(action.ActionBurst, 18*60)

	return action.Info{
		Frames:          frames.NewAbilFunc(burstHonorFrames),
		AnimationLength: burstHonorFrames[action.InvalidAction],
		CanQueueAfter:   burstHonorFrames[action.ActionSkill], // earliest cancel
		State:           action.BurstState,
	}, nil
}

func (c *char) burstHonorOnAttack(counter int) {
	switch counter {
	case 0:
		c.burstUnsealPoints(10)
	case 1, 2:
		c.burstUnsealPoints(20)
	case 3, 4:
		c.burstUnsealPoints(30)
	default:
		panic("invalid attack counter")
	}
}

func (c *char) skillDoom(unseal int) (action.Info, error) {
	if unseal > len(c.Core.Player.Chars()) {
		return action.Info{}, fmt.Errorf("invalid unseal index: %v", unseal)
	}

	unseal -= 1 // convert to 0-indexed

	char := c.Core.Player.ByIndex(unseal)
	if !char.Sealed {
		return action.Info{}, fmt.Errorf("character %v is not sealed", char.Base.Key.String())
	}

	if c.getBlessing() != honor {
		return action.Info{}, fmt.Errorf("cannot use unsealing skill in love blessing")
	}

	counter := 3
	for _, char := range c.Core.Player.Chars() {
		if char.Index() == c.Index() {
			continue
		}
		if char.Sealed {
			counter--
		}
	}
	if counter < 0 {
		return action.Info{}, fmt.Errorf("invalid unseal counter: %v", counter)
	}

	ap := combat.NewCircleHitOnTarget(c.Core.Combat.Player(), nil, 7)

	if counter == 0 {
		ai := info.AttackInfo{
			ActorIndex:       c.Index(),
			Abil:             "Doom 1",
			AttackTag:        attacks.AttackTagElementalArt,
			Durability:       25,
			IgnoreDefPercent: 1,
			Element:          attributes.Cryo,
			Mult:             burstHonorSkill1[c.TalentLvlBurst()] * (1 + c.c2MultBonus()),
		}
		for i := range 5 {
			c.Core.QueueAttack(ai, ap, skillDoomHitmarks[counter][i], skillDoomHitmarks[counter][i])
		}
	}

	ai := info.AttackInfo{
		ActorIndex: c.Index(),
		Abil:       fmt.Sprintf("Doom %v", counter+1),
		Element:    attributes.Cryo,
	}
	ssc := burstHonorSkillSSC[counter][c.TalentLvlBurst()] * (1 + c.c2MultBonus())
	ssw := burstHonorSkillSSw[counter][c.TalentLvlBurst()] * (1 + c.c2MultBonus())
	c.toStellarWithDefault(&ai, ssc, ssw)
	c.Core.QueueAttack(
		ai,
		ap,
		skillDoomHitmarks[counter][len(skillDoomHitmarks[counter])-1],
		skillDoomHitmarks[counter][len(skillDoomHitmarks[counter])-1],
	)
	c.unsealReady[unseal] = false
	char.Sealed = false
	c.Core.Log.NewEvent("swap unsealed", glog.LogCooldownEvent, unseal)
	c.a1OnUnseal(char)
	c.c6OnUnseal()
	return action.Info{
		Frames:          frames.NewAbilFunc(skillDoomFrames[counter]),
		AnimationLength: skillDoomFrames[counter][action.InvalidAction],
		CanQueueAfter:   skillDoomFrames[counter][action.ActionDash], // earliest cancel
		State:           action.SkillState,
	}, nil
}
