package mitya

import (
	"fmt"

	"github.com/genshinsim/gcsim/internal/frames"
	"github.com/genshinsim/gcsim/pkg/core/action"
	"github.com/genshinsim/gcsim/pkg/core/attacks"
	"github.com/genshinsim/gcsim/pkg/core/attributes"
	"github.com/genshinsim/gcsim/pkg/core/combat"
	"github.com/genshinsim/gcsim/pkg/core/info"
)

var (
	skillFrames       []int
	skillHoldFrames   []int
	skillRecastFrames []int
)

const (
	skillHitmark        = 28
	skillCDStart        = 22
	skillHoldHitmark    = 44
	skillHoldCDStart    = 36
	skillFirstTickDelay = 72 + 45 - 27
	skillInterval       = 92
	maxPrismStacks      = 4
	skillKey            = "mitya-skill"
	particleICDKey      = "mitya-particle-icd"
	skillRecastKey      = "mitya-skill-recast"
	skillRecastCDKey    = "mitya-skill-recast-cd"
)

func init() {
	skillFrames = frames.InitAbilSlice(24)
	skillHoldFrames = frames.InitAbilSlice(46)
	skillRecastFrames = frames.InitAbilSlice(8)
}

func (c *char) Skill(p map[string]int) (action.Info, error) {
	if p["recast"] != 0 {
		return c.skillRecast()
	}

	if p["hold"] != 0 {
		return c.skillHold(), nil
	}
	return c.skillPress(), nil
}

func (c *char) skillPress() action.Info {
	c.QueueCharTask(func() {
		ai := info.AttackInfo{
			ActorIndex: c.Index(),
			Abil:       "Skill (Press)",
			AttackTag:  attacks.AttackTagElementalArt,
			ICDTag:     attacks.ICDTagNone,
			ICDGroup:   attacks.ICDGroupDefault,
			StrikeType: attacks.StrikeTypeDefault,
			Element:    attributes.Electro,
			Durability: 25,
			Mult:       skillPress[c.TalentLvlSkill()],
		}

		c.Core.QueueAttack(
			ai,
			combat.NewBoxHitOnTarget(c.Core.Combat.Player(), info.Point{Y: -0.4}, 2.5, 10),
			skillHitmark-skillCDStart,
			skillHitmark-skillCDStart,
			c.particleCB,
		)

		c.addSkillState(skillStatePress)
	}, skillCDStart)

	c.SetCDWithDelay(action.ActionSkill, 18*60, skillCDStart)

	return action.Info{
		Frames:          frames.NewAbilFunc(skillFrames),
		AnimationLength: skillFrames[action.InvalidAction],
		CanQueueAfter:   skillFrames[action.ActionWalk], // earliest cancel
		State:           action.SkillState,
	}
}

func (c *char) skillHold() action.Info {
	c.QueueCharTask(func() {
		ai := info.AttackInfo{
			ActorIndex: c.Index(),
			Abil:       "Skill (Hold)",
			AttackTag:  attacks.AttackTagElementalArt,
			ICDTag:     attacks.ICDTagNone,
			ICDGroup:   attacks.ICDGroupDefault,
			StrikeType: attacks.StrikeTypeDefault,
			Element:    attributes.Electro,
			Durability: 25,
			Mult:       skillHold[c.TalentLvlSkill()],
		}

		c.Core.QueueAttack(
			ai,
			combat.NewSingleTargetHit(c.Core.Combat.PrimaryTarget().Key()),
			skillHoldHitmark-skillHoldCDStart,
			skillHoldHitmark-skillHoldCDStart,
			c.particleCB,
		)
		c.addSkillState(skillStateHold)
		c.addPrism()
	}, skillHoldCDStart)

	c.SetCDWithDelay(action.ActionSkill, 18*60, skillHoldCDStart)

	return action.Info{
		Frames:          frames.NewAbilFunc(skillHoldFrames),
		AnimationLength: skillHoldFrames[action.InvalidAction],
		CanQueueAfter:   skillHoldFrames[action.ActionDash], // earliest cancel
		State:           action.SkillState,
	}
}

func (c *char) skillRecast() (action.Info, error) {
	if !c.StatusIsActive(skillRecastKey) {
		return action.Info{}, fmt.Errorf("cannot recast skill, not in recast period")
	}
	c.AddStatus(skillRecastCDKey, 1.5*60, false)
	if c.getSkillState() == skillStatePress {
		c.skillDetonate()
		c.skillState = skillStateHold
		c.skillSrc = -1
	} else {
		c.skillState = skillStatePress
		src := c.Core.F
		c.skillSrc = src
		c.Core.Tasks.Add(func() { c.skillTicker(src) }, skillFirstTickDelay)
		c.Core.Tasks.Add(func() {
			if c.skillSrc != src {
				return
			}
			c.skillDetonate()
		}, c.StatusDuration(skillKey))
	}

	return action.Info{
		Frames:          frames.NewAbilFunc(skillRecastFrames),
		AnimationLength: skillRecastFrames[action.InvalidAction],
		CanQueueAfter:   skillRecastFrames[action.ActionDash], // earliest cancel
		State:           action.SkillState,
	}, nil
}

func (c *char) particleCB(a info.AttackCB) {
	if a.Target.Type() != info.TargettableEnemy {
		return
	}
	if c.StatusIsActive(particleICDKey) {
		return
	}
	c.AddStatus(particleICDKey, 6*60, true)
	c.Core.QueueParticle(c.Base.Key.String(), 5, attributes.Electro, c.ParticleDelay)
}

func (c *char) addSkillState(newState skillState) {
	if c.getSkillState() == skillStatePress {
		c.skillDetonate()
	}

	c.skillState = newState
	c.AddStatus(skillRecastKey, 6*60, false)
	c.AddStatus(skillRecastCDKey, 1.5*60, false)
	c.AddStatus(skillKey, 25*60, false)
	c.c1OnSkill()
	c.c2OnSkill()

	if newState == skillStatePress {
		src := c.Core.F
		c.skillSrc = src
		c.Core.Tasks.Add(func() { c.skillTicker(src) }, skillFirstTickDelay)
		c.Core.Tasks.Add(func() {
			if c.skillSrc != src {
				return
			}
			c.skillDetonate()
		}, 25*60)
	} else {
		c.skillSrc = -1
	}
}

func (c *char) getSkillState() skillState {
	if !c.StatusIsActive(skillKey) {
		return skillStateNone
	}
	return c.skillState
}

func (c *char) skillDetonate() {
	// detonate existing prisms
	// ignoring what happens when he consumes all prisms... should it remove the polestar field?
	ai := info.AttackInfo{
		ActorIndex:       c.Index(),
		Abil:             "Prism Shatter" + stellarConductText,
		AttackTag:        attacks.AttackTagDirectStellarConduct,
		Element:          attributes.Electro,
		Mult:             skillPrismShardPerPrism[c.TalentLvlSkill()] * float64(c.prisms) * (1 + c.c6PrismMult()),
		UseEM:            true,
		IgnoreDefPercent: 1,
	}

	ap := combat.NewCircleHitOnTarget(c.Core.Combat.Player(), info.Point{Y: 2}, 5)
	c.Core.QueueAttack(ai, ap, 0, 0)
	c.prisms = 0
}

func (c *char) skillTicker(src int) {
	if c.skillSrc != src {
		return
	}

	if !c.StatusIsActive(skillKey) {
		return
	}

	ai := info.AttackInfo{
		ActorIndex: c.Index(),
		Abil:       "Skill Pulse",
		AttackTag:  attacks.AttackTagElementalArt,
		ICDTag:     attacks.ICDTagElementalArt,
		ICDGroup:   attacks.ICDGroupDefault,
		StrikeType: attacks.StrikeTypeDefault,
		Element:    attributes.Electro,
		Durability: 25,
		Mult:       skillPulse[c.TalentLvlSkill()] * (1 + c.c6PrismMult()),
	}

	ap := combat.NewCircleHitOnTarget(c.Core.Combat.PrimaryTarget(), nil, 3)

	c.Core.QueueAttack(ai, ap, 0, 0)

	c.Core.Tasks.Add(func() { c.skillTicker(src) }, skillInterval)
}

type skillState int

const (
	skillStateNone skillState = iota
	skillStatePress
	skillStateHold
)
