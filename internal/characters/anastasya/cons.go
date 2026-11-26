package anastasya

import (
	"fmt"

	"github.com/genshinsim/gcsim/internal/frames"
	"github.com/genshinsim/gcsim/pkg/core/action"
	"github.com/genshinsim/gcsim/pkg/core/attacks"
	"github.com/genshinsim/gcsim/pkg/core/attributes"
	"github.com/genshinsim/gcsim/pkg/core/combat"
	"github.com/genshinsim/gcsim/pkg/core/event"
	"github.com/genshinsim/gcsim/pkg/core/info"
	"github.com/genshinsim/gcsim/pkg/core/player/character"
	"github.com/genshinsim/gcsim/pkg/modifier"
)

const (
	c2Key      = "anastasya-c2"
	c2LoveKey  = c2Key + "-love"
	c2HonorKey = c2Key + "-honor"
	c6Key      = "anastasya-c6"
	c6ICDKey   = c6Key + "-icd"
)

// Must be called after a1Init. Replaces the buff from a1
func (c *char) c1Init() {
	if c.Base.Ascension < 1 {
		return
	}
	if c.Base.Cons < 1 {
		return
	}
	m := make([]float64, attributes.EndStatType)
	m2 := make([]float64, attributes.EndStatType)
	for _, char := range c.Core.Player.Chars() {
		char.AddStatMod(character.StatMod{
			Base: modifier.NewBase(a1Key+"-buff", -1),
			Amount: func() []float64 {
				if !c.StatusIsActive(a1Key) {
					return nil
				}

				if c.c6a1Override && char.Index() == c.Index() {
					return c.c6a1OverrideBuffs(m, m2)
				}

				c.c1BaseBuff(char, m)
				if char.Sealed {
					return m
				}

				if char.StatusIsActive(a1SealedKey) {
					return m
				}

				unsealed := len(c.Core.Player.Chars()) - c.a1SealedCount()
				for _, ch := range c.Core.Player.Chars() {
					if char.Index() == ch.Index() {
						continue
					}
					if !ch.Sealed && !ch.StatusIsActive(a1SealedKey) {
						continue
					}
					mult := 1.0 / float64(unsealed)
					if ch.Index() == c.Index() {
						mult *= 3.0
					}
					c.c1BaseBuff(ch, m2)
					for _, i := range []attributes.Stat{attributes.CD, attributes.ATKP, attributes.EM} {
						m[i] += m2[i] * mult
					}
				}

				return m
			},
		})
	}
}

func (c *char) c1BaseBuff(char *character.CharWrapper, buff []float64) {
	mult := 1 + c.c4a1Bonus()
	buff[attributes.CD] = 0.12 * mult
	buff[attributes.ATKP] = 0
	buff[attributes.EM] = 0

	if char.Base.Element == attributes.Cryo {
		buff[attributes.ATKP] = 0.15 * mult
	} else {
		buff[attributes.EM] = 60 * mult
	}
}

func (c *char) c2Init() {
	if c.Base.Cons < 2 {
		return
	}

	m := make([]float64, attributes.EndStatType)
	m[attributes.CD] = 0.5
	for _, char := range c.Core.Player.Chars() {
		char.AddAttackMod(character.AttackMod{
			Base: modifier.NewBase(c2LoveKey+"-buff", -1),
			Amount: func(atk *info.AttackEvent, t info.Target) []float64 {
				if !atk.Info.AttackTag.IsStellar() {
					return nil
				}
				if !c.StatusIsActive(c2LoveKey) {
					return nil
				}
				return m
			},
		})
	}

	c.Core.Events.Subscribe(event.OnApplyAttack, func(args ...any) {
		atk := args[0].(*info.AttackEvent)

		// don't apply elevation to the reaction attack, since the subcomponent contributor attacks each got elevation applied already
		if !atk.Info.AttackTag.IsStellarDirect() {
			return
		}

		atk.Info.Elevation += 0.05
	}, c2Key)

	c.Core.Events.Subscribe(event.OnSpecialReactionAttack, func(args ...any) {
		atk := args[1].(*info.AttackEvent)

		if !atk.Info.AttackTag.IsStellarReact() {
			return
		}

		atk.Info.Elevation += 0.05
	}, c2Key)
}

func (c *char) c2OnSkillLove() {
	if c.Base.Cons < 2 {
		return
	}
	c.AddStatus(c2LoveKey, 30*60, true)
}

func (c *char) c2OnSkillHonor() {
	if c.Base.Cons < 2 {
		return
	}
	c.AddStatus(c2HonorKey, 15*60, true)
}

func (c *char) c2MultBonus() float64 {
	if c.Base.Cons < 2 {
		return 0
	}
	if c.StatusIsActive(c2HonorKey) {
		return 0.3
	}
	return 0
}

func (c *char) c4a1Bonus() float64 {
	if c.Base.Cons < 4 {
		return 0
	}
	return 0.3
}

func (c *char) c6Init() {
	if c.Base.Cons < 6 {
		return
	}

	c.Core.Events.Subscribe(event.OnCharacterSwap, func(args ...any) {
		if !c.c6LoveReady {
			return
		}
		ai := info.AttackInfo{
			ActorIndex:       c.Index(),
			Abil:             "Triumph (C6)",
			AttackTag:        attacks.AttackTagDirectStellarConduct,
			IgnoreDefPercent: 1,
			Element:          attributes.Cryo,
			Mult:             burstLoveTriumphSSC[c.TalentLvlBurst()] * 2,
		}
		c.toStellarWithDefault(&ai, burstLoveTriumphSSC[c.TalentLvlBurst()]*2, burstLoveTriumphSSw[c.TalentLvlBurst()]*2)
		c.Core.QueueAttack(
			ai,
			combat.NewCircleHitOnTarget(c.Core.Combat.PrimaryTarget(), nil, 5),
			28,
			28,
		)
	}, c6Key)

	c.Core.Events.Subscribe(event.OnApplyAttack, func(args ...any) {
		atk := args[0].(*info.AttackEvent)

		// don't apply elevation to the reaction attack, since the subcomponent contributor attacks each got elevation applied already
		if !atk.Info.AttackTag.IsStellarDirect() {
			return
		}

		self, other := c.c6Elevation()
		if atk.Info.ActorIndex == c.Index() {
			atk.Info.Elevation += self
		} else {
			atk.Info.Elevation += other
		}
	}, c6Key)

	c.Core.Events.Subscribe(event.OnSpecialReactionAttack, func(args ...any) {
		atk := args[1].(*info.AttackEvent)

		if !atk.Info.AttackTag.IsStellarReact() {
			return
		}

		self, other := c.c6Elevation()
		if atk.Info.ActorIndex == c.Index() {
			atk.Info.Elevation += self
		} else {
			atk.Info.Elevation += other
		}
	}, c6Key)
}

func (c *char) c6Elevation() (float64, float64) {
	if c.getBlessing() == honor {
		return 0.2, 0
	}
	return 0.1, 0.1
}

func (c *char) c6OnSkillLove() {
	if c.Base.Cons < 6 {
		return
	}

	if c.StatusIsActive(c6ICDKey) {
		return
	}
	c.AddStatus(c6ICDKey, 15*60, true)
	c.c6LoveReady = true
	c.c6HonorReady = false
}

func (c *char) c6OnSkillHonor() {
	if c.Base.Cons < 6 {
		return
	}
	if c.StatusIsActive(c6ICDKey) {
		return
	}
	c.AddStatus(c6ICDKey, 15*60, true)
	c.c6LoveReady = false
	c.c6HonorReady = true
}

func (c *char) c6OnUnseal() {
	if c.Base.Cons < 6 {
		return
	}
	c.c6HonorReady = false
}

func (c *char) c6a1OverrideBuffs(m, m2 []float64) []float64 {
	c.c1BaseBuff(c.CharWrapper, m)
	for _, ch := range c.Core.Player.Chars() {
		if c.Index() == ch.Index() {
			continue
		}
		c.c1BaseBuff(ch, m2)
		for _, i := range []attributes.Stat{attributes.CD, attributes.ATKP, attributes.EM} {
			m[i] += m2[i]
		}
	}

	return m
}

func (c *char) c6Skill() (action.Info, error) {
	ai := info.AttackInfo{
		ActorIndex: c.Index(),
		Abil:       fmt.Sprintf("Doom %v (C6)", 3),
		Element:    attributes.Cryo,
	}
	ap := combat.NewCircleHitOnTarget(c.Core.Combat.Player(), nil, 7)

	ssc := burstHonorSkillSSC[2][c.TalentLvlBurst()] * (1 + c.c2MultBonus()) * 2.5
	ssw := burstHonorSkillSSw[2][c.TalentLvlBurst()] * (1 + c.c2MultBonus()) * 2.5
	c.toStellarWithDefault(&ai, ssc, ssw)
	c.QueueCharTask(func() {
		c.c6a1Override = true
		c.Core.QueueAttack(ai, ap, 0, 0)
		c.c6a1Override = false
	}, skillDoomHitmarks[2][len(skillDoomHitmarks[2])-1])

	c.c6HonorReady = false
	return action.Info{
		Frames:          frames.NewAbilFunc(skillDoomFrames[2]),
		AnimationLength: skillDoomFrames[2][action.InvalidAction],
		CanQueueAfter:   skillDoomFrames[2][action.ActionDash], // earliest cancel
		State:           action.SkillState,
	}, nil
}
