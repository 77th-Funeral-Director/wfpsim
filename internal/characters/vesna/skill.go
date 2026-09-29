package vesna

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
)

var (
	skillFrames      []int
	skillSkillFrames [][]int

	skillWindborneHitmarks = []int{0, 9, 17}

	skillSpirit3DotHitmarks = []int{24, 24 + 9, 24 + 9 + 8, 24 + 9 + 8 + 9}

	skillSkillCanQueueAfter = []int{5000, skillWindborneHitmarks[1], skillWindborneHitmarks[2], skillSpirit3DotHitmarks[3]}
)

const (
	skillHitmark             = 29
	skillMaxLvl              = 3
	particleICDKey           = "vesna-particle-icd"
	skillKey                 = "vesna-skill"
	skillSpirit2Hitmark      = 17 + 30
	skillSpirit3FinalHitmark = 24 + 9 + 8 + 9 + 20
)

func init() {
	skillFrames = frames.InitAbilSlice(32)
	skillFrames[action.ActionAttack] = 26
	skillFrames[action.ActionSkill] = 27
	skillFrames[action.ActionBurst] = 29
	skillFrames[action.ActionDash] = 28
	skillFrames[action.ActionJump] = 27
	skillFrames[action.ActionWalk] = 32
	skillFrames[action.ActionSwap] = 26

	skillSkillFrames = make([][]int, 4)
	skillSkillFrames[0] = frames.InitAbilSlice(5000)
	skillSkillFrames[1] = frames.InitAbilSlice(27)
	skillSkillFrames[1][action.ActionAttack] = 27
	skillSkillFrames[1][action.ActionSkill] = 25
	skillSkillFrames[1][action.ActionBurst] = 25
	skillSkillFrames[1][action.ActionDash] = 25
	skillSkillFrames[1][action.ActionJump] = 24
	skillSkillFrames[1][action.ActionWalk] = 24
	skillSkillFrames[1][action.ActionSwap] = skillWindborneHitmarks[1]

	skillSkillFrames[2] = frames.InitAbilSlice(49)
	skillSkillFrames[2][action.ActionAttack] = 49
	skillSkillFrames[2][action.ActionSkill] = 39
	skillSkillFrames[2][action.ActionBurst] = 40
	skillSkillFrames[2][action.ActionDash] = 40
	skillSkillFrames[2][action.ActionJump] = 42
	skillSkillFrames[2][action.ActionWalk] = 41
	skillSkillFrames[2][action.ActionSwap] = skillWindborneHitmarks[2]

	skillSkillFrames[3] = frames.InitAbilSlice(76)
	skillSkillFrames[3][action.ActionAttack] = 70
	skillSkillFrames[3][action.ActionSkill] = 69
	skillSkillFrames[3][action.ActionBurst] = skillSpirit3DotHitmarks[3]
	skillSkillFrames[3][action.ActionDash] = skillSpirit3DotHitmarks[3]
	skillSkillFrames[3][action.ActionJump] = skillSpirit3DotHitmarks[3]
	skillSkillFrames[3][action.ActionWalk] = skillSpirit3DotHitmarks[3]
	skillSkillFrames[3][action.ActionSwap] = skillSpirit3DotHitmarks[3]
}

func (c *char) skillInit() {
	c.Core.Events.Subscribe(event.OnCharacterSwap, func(args ...any) {
		// do nothing if previous char wasn't vesna
		prev := args[0].(int)
		if prev != c.Index() {
			return
		}
		if !c.StatusIsActive(skillKey) {
			return
		}

		c.DeleteStatus(skillKey)
	}, "vesna-skill-exit")
}

func (c *char) Skill(p map[string]int) (action.Info, error) {
	if c.StatusIsActive(c6Key) {
		return c.c6Attack()
	}

	if c.StatusIsActive(skillKey) {
		return c.skillSpecial()
	}
	ai := info.AttackInfo{
		ActorIndex: c.Index(),
		Abil:       "The Art of Victory",
		AttackTag:  attacks.AttackTagElementalArt,
		ICDTag:     attacks.ICDTagNone,
		ICDGroup:   attacks.ICDGroupDefault,
		StrikeType: attacks.StrikeTypeDefault,
		Element:    attributes.Anemo,
		Durability: 25,
		Mult:       skill[c.TalentLvlSkill()],
	}
	c.Core.QueueAttack(ai, combat.NewCircleHitOnTarget(c.Core.Combat.Player(), nil, 3), skillHitmark, skillHitmark, c.particleCB)

	c.AddStatus(skillKey, 15*60+skillHitmark, true)

	c.SetCDWithDelay(action.ActionSkill, 18*60, skillHitmark)

	c.skillsMaxLvlUsed = 0
	c.skillLvl = 1
	c.skillStacks = 0
	c.addSkillStacks(2)

	c.a1OnSkill()
	c.c2OnSkill()

	return action.Info{
		Frames:          func(next action.Action) int { return skillFrames[next] },
		AnimationLength: skillFrames[action.InvalidAction],
		CanQueueAfter:   skillFrames[action.ActionSwap], // earliest cancel
		State:           action.SkillState,
	}, nil
}

func (c *char) skillSpecial() (action.Info, error) {
	lvl := min(c.skillLvl, skillMaxLvl)
	ai := info.AttackInfo{
		ActorIndex: c.Index(),
		Abil:       fmt.Sprintf("Windborne Blade Lv. %d", lvl),
		AttackTag:  attacks.AttackTagElementalArt,
		ICDTag:     attacks.ICDTagElementalArt,
		ICDGroup:   attacks.ICDGroupVesnaSkill,
		StrikeType: attacks.StrikeTypeDefault,
		Element:    attributes.Anemo,
		Durability: 25,
		Mult:       skillSpecial[lvl][c.TalentLvlSkill()],
	}

	ap := combat.NewCircleHitOnTargetFanAngle(c.Core.Combat.Player(), nil, 5, 60)

	// this buff is added before the hits land
	c.a1OnSpecialSkillOrBurst()

	defer func() {
		if c.c1UseSkillStack() {
			c.skillStacks -= 1
		}

		if lvl == skillMaxLvl {
			c.skillsMaxLvlUsed += 1
		} else {
			c.skillLvl += 1
		}

		if c.skillsMaxLvlUsed >= 3+c.c1ExtraMaxLvlSkills() {
			c.DeleteStatus(skillKey)
		}
	}()

	switch lvl {
	case 1:
		c.Core.QueueAttack(ai, ap, skillWindborneHitmarks[lvl], skillWindborneHitmarks[lvl], c.particleCB)

		return action.Info{
			Frames:          func(next action.Action) int { return skillSkillFrames[lvl][next] },
			AnimationLength: skillSkillFrames[lvl][action.InvalidAction],
			CanQueueAfter:   skillSkillCanQueueAfter[lvl], // earliest cancel
			State:           action.SkillState,
		}, nil
	case 2:
		c.Core.QueueAttack(ai, ap, skillWindborneHitmarks[lvl], skillWindborneHitmarks[lvl], c.particleCB)
		// level 2
		ai := info.AttackInfo{
			ActorIndex: c.Index(),
			Abil:       fmt.Sprintf("Windborne Blade Lv. %d Spirit Blade", lvl),
			AttackTag:  attacks.AttackTagElementalArt,
			ICDTag:     attacks.ICDTagElementalArt,
			ICDGroup:   attacks.ICDGroupVesnaSkill,
			StrikeType: attacks.StrikeTypeDefault,
			Element:    attributes.Anemo,
			Durability: 25,
			Mult:       skillSpirit2[c.TalentLvlSkill()] * c.a1Mult(),
		}
		if c.isRadianceSSw() {
			ai.Abil += stellarSwirlText
			ai.AttackTag = attacks.AttackTagDirectStellarSwirl
			ai.IgnoreDefPercent = 1
			ai.Durability = 0
		}

		ap := combat.NewCircleHitOnTargetFanAngle(c.Core.Combat.Player(), nil, 5, 60)
		c.Core.QueueAttack(ai, ap, skillSpirit2Hitmark, skillSpirit2Hitmark, c.particleCB)

		return action.Info{
			Frames:          func(next action.Action) int { return skillSkillFrames[lvl][next] },
			AnimationLength: skillSkillFrames[lvl][action.InvalidAction],
			CanQueueAfter:   skillSkillCanQueueAfter[lvl], // earliest cancel
			State:           action.SkillState,
		}, nil
	default:
		// level 3
		aiDoT := info.AttackInfo{
			ActorIndex: c.Index(),
			Abil:       fmt.Sprintf("Windborne Blade Lv. %d Spirit Blade", lvl),
			AttackTag:  attacks.AttackTagElementalArt,
			ICDTag:     attacks.ICDTagElementalArt,
			ICDGroup:   attacks.ICDGroupVesnaSkill,
			StrikeType: attacks.StrikeTypeDefault,
			Element:    attributes.Anemo,
			Durability: 25,
			Mult:       skillSpirit3DoT[c.TalentLvlSkill()] * c.a1Mult(),
		}

		aiFinale := info.AttackInfo{
			ActorIndex: c.Index(),
			Abil:       fmt.Sprintf("Windborne Blade Lv. %d Spirit Blade Final", lvl),
			AttackTag:  attacks.AttackTagElementalArt,
			ICDTag:     attacks.ICDTagElementalArt,
			ICDGroup:   attacks.ICDGroupVesnaSkill,
			StrikeType: attacks.StrikeTypeDefault,
			Element:    attributes.Anemo,
			Durability: 25,
			Mult:       skillSpirit3Final[c.TalentLvlSkill()] * c.a1Mult(),
		}

		if c.isRadianceSSw() {
			aiDoT.Abil += stellarSwirlText
			aiDoT.AttackTag = attacks.AttackTagDirectStellarSwirl
			aiDoT.IgnoreDefPercent = 1
			aiDoT.Durability = 0

			aiFinale.Abil += stellarSwirlText
			aiFinale.AttackTag = attacks.AttackTagDirectStellarSwirl
			aiFinale.IgnoreDefPercent = 1
			aiFinale.Durability = 0
		}

		apDoT := combat.NewCircleHitOnTargetFanAngle(c.Core.Combat.Player(), nil, 5, 60)
		for _, hitmark := range skillSpirit3DotHitmarks {
			c.Core.QueueAttack(aiDoT, apDoT, hitmark, hitmark, c.particleCB)
		}
		c.Core.QueueAttack(aiFinale, apDoT, skillSpirit3FinalHitmark, skillSpirit3FinalHitmark, c.particleCB)

		c.c6OnMaxLvlSkill()

		return action.Info{
			Frames:          func(next action.Action) int { return skillSkillFrames[lvl][next] },
			AnimationLength: skillSkillFrames[lvl][action.InvalidAction],
			CanQueueAfter:   skillSkillCanQueueAfter[lvl], // earliest cancel
			State:           action.SkillState,
		}, nil
	}
}

func (c *char) pinionAttack(delay int) {
	ai := info.AttackInfo{
		ActorIndex: c.Index(),
		Abil:       "Wind Pinion",
		AttackTag:  attacks.AttackTagElementalArt,
		ICDTag:     attacks.ICDTagElementalArt,
		ICDGroup:   attacks.ICDGroupVesnaSkill,
		StrikeType: attacks.StrikeTypeDefault,
		Element:    attributes.Anemo,
		Durability: 25,
		Mult:       skillPinion[c.TalentLvlSkill()],
	}

	ap := combat.NewCircleHitOnTarget(c.Core.Combat.PrimaryTarget(), nil, 2)

	c.Core.QueueAttack(ai, ap, delay, delay, c.particleCB)
}

func (c *char) particleCB(ac info.AttackCB) {
	if ac.Target.Type() != info.TargettableEnemy {
		return
	}

	if c.StatusIsActive(particleICDKey) {
		return
	}

	c.AddStatus(particleICDKey, 2.5*60, true)
	c.Core.QueueParticle(c.Base.Key.String(), 1, attributes.Anemo, c.ParticleDelay)
}

func (c *char) skillStacksCB(stacks int) info.AttackCBFunc {
	done := false
	return func(ac info.AttackCB) {
		if ac.Target.Type() != info.TargettableEnemy {
			return
		}
		if done {
			return
		}
		c.skillNACount += stacks
		if c.skillNACount >= 6 {
			c.skillNACount -= 6
			c.addSkillStacks(1)
		}
	}
}

func (c *char) addSkillStacks(count int) {
	c.skillStacks += count
	c.Core.Log.NewEvent(fmt.Sprintf("Gained Sword Essence (current: %d)", c.skillStacks), glog.LogCharacterEvent, c.Index())
}
