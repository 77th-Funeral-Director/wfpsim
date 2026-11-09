package danica

import (
	"github.com/genshinsim/gcsim/internal/frames"
	"github.com/genshinsim/gcsim/pkg/core/action"
	"github.com/genshinsim/gcsim/pkg/core/attacks"
	"github.com/genshinsim/gcsim/pkg/core/attributes"
	"github.com/genshinsim/gcsim/pkg/core/combat"
	"github.com/genshinsim/gcsim/pkg/core/info"
	"github.com/genshinsim/gcsim/pkg/core/player/character"
	"github.com/genshinsim/gcsim/pkg/modifier"
)

var (
	burstHitmarks    = []int{102, 102 + 11, 102 + 11 + 14, 102 + 11 + 14 + 15, 102 + 11 + 14 + 15 + 26}
	burstDura        = []info.Durability{25, 0, 25, 0, 25}
	burstHitHaltTime = []float64{0.03, 0.02, 0.02, 0.03, 0.06}
	burstFrames      []int

	buffDelay = 106
)

const (
	burstKey     = "danica-burst"
	burstBuffKey = burstKey + "-buff"
)

func init() {
	// Mavuika
	burstFrames = frames.InitAbilSlice(100) // Q -> Walk
	burstFrames[action.ActionSwap] = 100
	burstFrames[action.ActionAttack] = 100
	burstFrames[action.ActionCharge] = 100
	burstFrames[action.ActionSkill] = 100
	burstFrames[action.ActionDash] = 100
	burstFrames[action.ActionJump] = 100
}

func (c *char) Burst(p map[string]int) (action.Info, error) {
	for i, delay := range burstHitmarks {
		c.QueueCharTask(func() {
			ai := info.AttackInfo{
				ActorIndex:       c.Index(),
				Abil:             "Burst",
				AttackTag:        attacks.AttackTagElementalBurst,
				ICDTag:           attacks.ICDTagNone,
				ICDGroup:         attacks.ICDGroupDefault,
				StrikeType:       attacks.StrikeTypeDefault,
				Element:          attributes.Electro,
				Durability:       burstDura[i],
				Mult:             burst[i][c.TalentLvlBurst()],
				HitlagFactor:     0.01,
				HitlagHaltFrames: burstHitHaltTime[i] * 60,
				IsDeployable:     true,
			}
			switch i {
			case 0:
				ai.StrikeType = attacks.StrikeTypeBlunt
				ai.PoiseDMG = 60
			case 4:
				ai.StrikeType = attacks.StrikeTypeBlunt
				ai.PoiseDMG = 90
			}

			c.Core.QueueAttack(ai, combat.NewCircleHitOnTarget(c.Core.Combat.Player(), nil, 2), 0, 0)

			if c.getRadiance() == radianceNone {
				return
			}

			switch i {
			case 0, 4:
				c.toStellar(&ai, burst1SSC[c.TalentLvlBurst()])
				c.Core.QueueAttack(ai, combat.NewCircleHitOnTarget(c.Core.Combat.Player(), nil, 2), 1, 1)
			}
		}, delay)
	}

	c.QueueCharTask(func() {
		if c.getRadiance() == radianceStellarConduct {
			c.AddStatus(burstKey, 20*60, true)
		}
	}, buffDelay)

	c.SetCD(action.ActionBurst, 15*60)
	c.ConsumeEnergy(8)

	return action.Info{
		Frames:          frames.NewAbilFunc(burstFrames),
		AnimationLength: burstFrames[action.InvalidAction],
		CanQueueAfter:   burstFrames[action.ActionDash], // earliest cancel
		State:           action.BurstState,
	}, nil
}

func (c *char) burstInit() {
	m := make([]float64, attributes.EndStatType)
	scaling := burstATKBonus[c.TalentLvlBurst()]
	maxBuff := burstMaxATKBonus[c.TalentLvlBurst()]
	for _, char := range c.Core.Player.Chars() {
		char.AddStatMod(character.StatMod{
			Base:         modifier.NewBase(burstBuffKey, -1),
			Extra:        true,
			AffectedStat: attributes.ATK,
			Amount: func() []float64 {
				if !c.StatusIsActive(burstKey) {
					return nil
				}
				totalNonExtraAtk := c.SelectStat(true, attributes.BaseATK, attributes.ATKP, attributes.ATK).TotalATK()
				m[attributes.ATK] = min(totalNonExtraAtk*scaling, maxBuff) * (1 + c.c1BonusAtkBuff())
				return m
			},
		})
	}
}
