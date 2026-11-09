package danica

import (
	"github.com/genshinsim/gcsim/internal/frames"
	"github.com/genshinsim/gcsim/pkg/core/action"
	"github.com/genshinsim/gcsim/pkg/core/attacks"
	"github.com/genshinsim/gcsim/pkg/core/attributes"
	"github.com/genshinsim/gcsim/pkg/core/combat"
	"github.com/genshinsim/gcsim/pkg/core/event"
	"github.com/genshinsim/gcsim/pkg/core/info"
	"github.com/genshinsim/gcsim/pkg/enemy"
)

var (
	skillFrames      []int
	skillEntryFrames []int
)

const (
	skillHitmark      = 22
	skillEntryHitmark = 42
	skillKey          = "danica-skill"
	skillICDKey       = "danica-skill-icd"
	skillExciteICDKey = "danica-skill-excite-icd"
	particleICDKey    = "danica-particle-icd"

	maidDefendKey = "danica-maid-defend"
	maidCombatKey = "danica-maid-combat"

	entrySkillCDKey = "danica-entry-skill-cd"

	exciteMax = 100
)

func init() {
	skillFrames = frames.InitAbilSlice(44)
	skillFrames[action.ActionBurst] = 34
	skillFrames[action.ActionDash] = 34
	skillFrames[action.ActionJump] = 34

	skillEntryFrames = frames.InitAbilSlice(60)
}

func (c *char) Skill(p map[string]int) (action.Info, error) {
	if c.doEntrySkill {
		return c.entrySKill()
	}

	c.QueueCharTask(func() {
		c.skillSrc = c.Core.F
		ai := info.AttackInfo{
			ActorIndex:         c.Index(),
			Abil:               "Skill",
			AttackTag:          attacks.AttackTagElementalArt,
			ICDTag:             attacks.ICDTagElementalArt,
			ICDGroup:           attacks.ICDGroupDefault,
			StrikeType:         attacks.StrikeTypeBlunt,
			PoiseDMG:           100,
			Element:            attributes.Electro,
			Durability:         25,
			HitlagFactor:       0.01,
			HitlagHaltFrames:   0.04 * 60,
			Mult:               skill[c.TalentLvlSkill()],
			CanBeDefenseHalted: true,
		}

		ap := combat.NewCircleHitOnTarget(c.Core.Combat.Player(), nil, 2)
		c.Core.QueueAttack(ai, ap, 0, 0, c.particleCB)
		c.AddStatus(skillKey, 25*60, true)

		c.a1OnSkill()
	}, skillHitmark)

	c.SetCDWithDelay(action.ActionSkill, 18*60, 20)

	return action.Info{
		Frames:          frames.NewAbilFunc(skillFrames),
		AnimationLength: skillFrames[action.InvalidAction],
		CanQueueAfter:   skillFrames[action.ActionDash], // earliest cancel
		State:           action.SkillState,
	}, nil
}

func (c *char) entrySKill() (action.Info, error) {
	c.excite = 0
	c.AddStatus(maidDefendKey, 25*60, true)
	c.AddStatus(skillKey, 25*60, true)
	c.newShield()
	c.doEntrySkill = false
	c.AddStatus(entrySkillCDKey, 14*60, false)
	c.a4OnSkillEntry()
	c.c1OnSkillEntry()
	c.QueueCharTask(func() {
		ai := info.AttackInfo{
			ActorIndex:       c.Index(),
			Abil:             "Skill Entry",
			AttackTag:        attacks.AttackTagDirectStellarConduct,
			StrikeType:       attacks.StrikeTypeDefault,
			Element:          attributes.Electro,
			Mult:             skillEntrySSC[c.TalentLvlSkill()],
			IgnoreDefPercent: 1,
		}

		ap := combat.NewCircleHitOnTarget(c.Core.Combat.Player(), nil, 2)
		c.Core.QueueAttack(ai, ap, 0, 0, c.particleCB)
	}, skillEntryHitmark)

	return action.Info{
		Frames:          frames.NewAbilFunc(skillEntryFrames),
		AnimationLength: skillEntryFrames[action.InvalidAction],
		CanQueueAfter:   skillEntryFrames[action.ActionDash], // earliest cancel
		State:           action.SkillState,
	}, nil
}

func (c *char) skillInit() {
	c.Core.Events.Subscribe(event.OnEnemyDamage, func(args ...interface{}) {
		if c.Core.Player.Active() == c.Index() {
			return
		}

		e, ok := args[0].(*enemy.Enemy)
		if !ok {
			return
		}

		atk, ok := args[1].(*info.AttackEvent)
		if !ok {
			return
		}

		if c.Core.Player.Active() != atk.Info.ActorIndex {
			return
		}

		switch atk.Info.Element {
		case attributes.Electro, attributes.Cryo:
		default:
			return
		}

		if !c.StatusIsActive(skillKey) {
			return
		}

		if !c.StatusIsActive(skillExciteICDKey) {
			c.addExcite(50)
			c.AddStatus(skillExciteICDKey, 0.1*60, true)
		}

		if c.StatusIsActive(skillICDKey) {
			return
		}

		c.AddStatus(skillICDKey, 1.5*60, true)

		ai := info.AttackInfo{
			ActorIndex: c.Index(),
			Abil:       "Skill (Coordinated)",
			AttackTag:  attacks.AttackTagElementalArt,
			ICDTag:     attacks.ICDTagNone,
			ICDGroup:   attacks.ICDGroupDefault,
			StrikeType: attacks.StrikeTypeDefault,
			Element:    attributes.Electro,
			Durability: 25,
			Mult:       skillCoord[c.TalentLvlSkill()],
		}

		ap := combat.NewCircleHitOnTarget(e, nil, 2)
		c.Core.QueueAttack(ai, ap, 5, 5)
	}, skillKey)

	c.Core.Events.Subscribe(event.OnCharacterSwap, func(args ...interface{}) {
		next := args[1].(int)
		if next != c.Index() {
			return
		}

		if c.excite < exciteMax {
			return
		}

		if !c.StatusIsActive(skillKey) {
			return
		}

		if c.StatusIsActive(entrySkillCDKey) {
			return
		}

		if c.getMaid() == maidCombat {
			return
		}

		c.doEntrySkill = true
		c.Core.Player.Exec(action.ActionSkill, c.Base.Key, map[string]int{})
	}, skillKey)
}

func (c *char) addExcite(amt float64) {
	c.excite = min(c.excite+amt, exciteMax)
}

func (c *char) particleCB(a info.AttackCB) {
	if a.Target.Type() != info.TargettableEnemy {
		return
	}
	if c.StatusIsActive(particleICDKey) {
		return
	}
	c.AddStatus(particleICDKey, 14*60, false)
	c.Core.QueueParticle(c.Base.Key.String(), 5, attributes.Electro, c.ParticleDelay)
}

type maidState int

const (
	maidNone maidState = iota
	maidDefend
	maidCombat
)

func (c *char) getMaid() maidState {
	if c.StatusIsActive(maidDefendKey) {
		return maidDefend
	}
	if c.StatusIsActive(maidCombatKey) {
		return maidCombat
	}
	return maidNone
}
