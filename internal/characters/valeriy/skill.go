package valeriy

import (
	"github.com/genshinsim/gcsim/internal/frames"
	"github.com/genshinsim/gcsim/pkg/core/action"
	"github.com/genshinsim/gcsim/pkg/core/attacks"
	"github.com/genshinsim/gcsim/pkg/core/attributes"
	"github.com/genshinsim/gcsim/pkg/core/combat"
	"github.com/genshinsim/gcsim/pkg/core/info"
)

// copied from nicole

var skillFrames []int

const (
	skillHitmark   = 26
	particleICDKey = "valeriy-particle-icd"
	skillDur       = 18 * 60
)

func init() {
	skillFrames = frames.InitAbilSlice(34)
}

func (c *char) Skill(p map[string]int) (action.Info, error) {
	// TODO: figure out order between shield and hitmark
	c.QueueCharTask(func() {
		c.addMomentum(40 * (1 + c.c1SkillMomentumBonus()))
		c.addShield()
	}, skillHitmark-1)

	c.QueueCharTask(func() {
		ai := info.AttackInfo{
			ActorIndex: c.Index(),
			Abil:       "Skill",
			AttackTag:  attacks.AttackTagElementalArt,
			ICDTag:     attacks.ICDTagNone,
			ICDGroup:   attacks.ICDGroupDefault,
			StrikeType: attacks.StrikeTypeDefault,
			Element:    attributes.Electro,
			Durability: 25,
			Mult:       skill[c.TalentLvlSkill()],
		}
		ap := combat.NewCircleHitOnTarget(c.Core.Combat.Player(), nil, 5)

		c.Core.QueueAttack(ai, ap, 0, 0, c.particleCB)

		c.c2OnSkill()
	}, skillHitmark)

	c.SetCD(action.ActionSkill, 10*60)
	return action.Info{
		Frames:          frames.NewAbilFunc(skillFrames),
		AnimationLength: skillFrames[action.InvalidAction],
		CanQueueAfter:   skillFrames[action.ActionSwap],
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
	c.AddStatus(particleICDKey, 0.3*60, true)
	c.Core.QueueParticle(c.Base.Key.String(), 3, attributes.Electro, c.ParticleDelay)
}
