package anastasya

import (
	"fmt"

	"github.com/genshinsim/gcsim/internal/frames"
	"github.com/genshinsim/gcsim/pkg/core/action"
	"github.com/genshinsim/gcsim/pkg/core/attacks"
	"github.com/genshinsim/gcsim/pkg/core/attributes"
	"github.com/genshinsim/gcsim/pkg/core/combat"
	"github.com/genshinsim/gcsim/pkg/core/info"
)

var skillFrames []int

const (
	skillTapKey         = "anastasya-skill-tap"
	skillHoldKey        = "anastasya-skill-hold"
	skillTapHitmark     = 23
	skillFirstTickDelay = 118
	skillParticleICDKey = "anastasya-particle-icd"
)

func init() {
	skillFrames = frames.InitAbilSlice(42) // E -> J
	skillFrames[action.ActionAttack] = 41
	skillFrames[action.ActionSkill] = 41
	skillFrames[action.ActionBurst] = 41
	skillFrames[action.ActionDash] = 40
	skillFrames[action.ActionWalk] = 41
	skillFrames[action.ActionSwap] = 40
}

func (c *char) Skill(p map[string]int) (action.Info, error) {
	if p["hold"] > 0 {
		return action.Info{}, fmt.Errorf("hold skill not implemented")
	}
	return c.skillTap(p)
}

func (c *char) skillTap(_ map[string]int) (action.Info, error) {
	c.QueueCharTask(func() {
		ai := info.AttackInfo{
			ActorIndex: c.Index(),
			Abil:       "Skill (Tap)",
			AttackTag:  attacks.AttackTagElementalArt,
			ICDTag:     attacks.ICDTagNone,
			ICDGroup:   attacks.ICDGroupDefault,
			StrikeType: attacks.StrikeTypeDefault,
			Element:    attributes.Cryo,
			Durability: 25,
			Mult:       skillTap[c.TalentLvlSkill()],
		}

		ap := combat.NewCircleHitOnTarget(c.Core.Combat.Player(), nil, 6)
		c.Core.QueueAttack(ai, ap, 0, 0, c.particleCB)

		src := c.Core.F
		c.skillSrc = src
		c.QueueCharTask(func() { c.skillTicker(src) }, skillFirstTickDelay)
	}, skillTapHitmark)

	c.AddStatus(skillTapKey, 20*60, true)

	c.SetCD(action.ActionSkill, 15*60)

	c.a1OnSkill()
	c.c2OnSkillLove()

	return action.Info{
		Frames:          frames.NewAbilFunc(skillFrames),
		AnimationLength: skillFrames[action.InvalidAction],
		CanQueueAfter:   skillFrames[action.ActionDash], // earliest cancel
		State:           action.SkillState,
	}, nil
}

func (c *char) skillTicker(src int) {
	if c.skillSrc != src {
		return
	}

	if !c.StatusIsActive(skillTapKey) {
		return
	}

	ai := info.AttackInfo{
		ActorIndex: c.Index(),
		Abil:       "Skill (DoT)",
		AttackTag:  attacks.AttackTagElementalArt,
		ICDTag:     attacks.ICDTagElementalArt,
		ICDGroup:   attacks.ICDGroupAnastasyaSkill,
		StrikeType: attacks.StrikeTypeDefault,
		Element:    attributes.Cryo,
		Durability: 25,
		Mult:       skillCrown[c.TalentLvlSkill()],
	}

	ap := combat.NewCircleHitOnTarget(c.Core.Combat.PrimaryTarget(), nil, 6)
	c.Core.QueueAttack(ai, ap, 0, 0)

	c.toStellar(&ai, skillCrownSSC[c.TalentLvlSkill()], skillCrownSSw[c.TalentLvlSkill()])
	c.Core.QueueAttack(ai, ap, 2, 2)

	c.burstLoveOnSkillTick(ap)
	c.QueueCharTask(func() {
		c.skillTicker(src)
	}, 118)
}

func (c *char) particleCB(a info.AttackCB) {
	if a.Target.Type() != info.TargettableEnemy {
		return
	}
	if c.StatusIsActive(skillParticleICDKey) {
		return
	}
	c.AddStatus(skillParticleICDKey, 15*60, true)
	c.Core.QueueParticle(c.Base.Key.String(), 5, attributes.Cryo, c.ParticleDelay)
}

type blessingState int

const (
	love blessingState = iota
	honor
)

func (c *char) getBlessing() blessingState {
	if c.StatusIsActive(skillHoldKey) {
		return honor
	}

	return love
}
