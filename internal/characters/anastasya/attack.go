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

var (
	attackFrames   [][]int
	attackHitmarks = [][]int{{17}, {25}, {11, 26}, {15, 24, 35}, {17}}

	attackHonorFrames      [][]int
	attackHonorHitmarks    = [][]int{{8}, {6, 6 + 14}, {14}, {12, 12 + 8, 12 + 8 + 10}, {32, 32, 32}}
	attackHonorHitHaltTime = []float64{0.02, 0.03, 0.03, 0.03, 0.03}
)

const normalHitNum = 5

func init() {
	attackFrames = make([][]int, normalHitNum)

	attackFrames[0] = frames.InitNormalCancelSlice(attackHitmarks[0][0], 35)
	attackFrames[1] = frames.InitNormalCancelSlice(attackHitmarks[1][0], 26)
	attackFrames[2] = frames.InitNormalCancelSlice(attackHitmarks[2][1], 39)
	attackFrames[3] = frames.InitNormalCancelSlice(attackHitmarks[3][2], 44)
	attackFrames[4] = frames.InitNormalCancelSlice(attackHitmarks[4][0], 52)

	attackHonorFrames = make([][]int, normalHitNum)

	attackHonorFrames[0] = frames.InitNormalCancelSlice(attackHonorHitmarks[0][0], 14)
	attackHonorFrames[1] = frames.InitNormalCancelSlice(attackHonorHitmarks[1][1], 28)
	attackHonorFrames[2] = frames.InitNormalCancelSlice(attackHonorHitmarks[2][0], 26)
	attackHonorFrames[3] = frames.InitNormalCancelSlice(attackHonorHitmarks[3][2], 40)
	attackHonorFrames[4] = frames.InitNormalCancelSlice(attackHonorHitmarks[4][0], 72)
}

func (c *char) Attack(p map[string]int) (action.Info, error) {
	if c.getBlessing() == honor {
		return c.attackHonor(p)
	}
	counter := c.NormalCounter
	travel, ok := p["travel"]
	if !ok {
		travel = 10
	}

	ai := info.AttackInfo{
		ActorIndex: c.Index(),
		Abil:       fmt.Sprintf("Normal %v", counter),
		AttackTag:  attacks.AttackTagNormal,
		ICDTag:     attacks.ICDTagNormalAttack,
		ICDGroup:   attacks.ICDGroupDefault,
		StrikeType: attacks.StrikeTypePierce,
		Element:    attributes.Physical,
		Mult:       attack[counter][c.TalentLvlAttack()],
		Durability: 25,
	}

	for _, delay := range attackHitmarks[counter] {
		c.Core.QueueAttack(
			ai,
			combat.NewBoxHit(
				c.Core.Combat.Player(),
				c.Core.Combat.PrimaryTarget(),
				info.Point{Y: -0.5},
				0.1,
				1,
			),
			delay,
			delay+travel,
		)
	}

	defer c.AdvanceNormalIndex()

	return action.Info{
		Frames:          frames.NewAttackFunc(c.Character, attackFrames),
		AnimationLength: attackFrames[counter][action.InvalidAction],
		CanQueueAfter:   attackHitmarks[counter][len(attackHitmarks[counter])-1],
		State:           action.NormalAttackState,
	}, nil
}

func (c *char) attackHonor(_ map[string]int) (action.Info, error) {
	counter := c.NormalCounter
	ai := info.AttackInfo{
		ActorIndex:       c.Index(),
		Abil:             fmt.Sprintf("Normal %v", counter),
		AttackTag:        attacks.AttackTagNormal,
		ICDTag:           attacks.ICDTagAnastasyaHonorNA,
		ICDGroup:         attacks.ICDGroupAnastasyaHonorNA,
		StrikeType:       attacks.StrikeTypePierce,
		Element:          attributes.Cryo,
		Mult:             skillAttack[counter][c.TalentLvlSkill()] * (1 + c.c2MultBonus()),
		HitlagFactor:     0.01,
		HitlagHaltFrames: attackHonorHitHaltTime[counter] * 60,
		Durability:       25,
	}

	if counter == 4 {
		ai.IsDeployable = true
	}

	for i, delay := range attackHonorHitmarks[counter] {
		if i == len(attackHonorHitmarks[counter])-1 {
			sscMult := skillAttackSSC[counter][c.TalentLvlSkill()] * (1 + c.c2MultBonus())
			sswMult := skillAttackSSw[counter][c.TalentLvlSkill()] * (1 + c.c2MultBonus())
			c.toStellar(&ai, sscMult, sswMult)
		}
		ap := combat.NewCircleHitOnTargetFanAngle(
			c.Core.Combat.Player(),
			info.Point{Y: -0.5},
			5,
			180,
		)
		c.Core.QueueAttack(
			ai,
			ap,
			delay,
			delay,
		)
	}

	c.burstHonorOnAttack(counter)

	defer c.AdvanceNormalIndex()

	return action.Info{
		Frames:          frames.NewAttackFunc(c.Character, attackHonorFrames),
		AnimationLength: attackHonorFrames[counter][action.InvalidAction],
		CanQueueAfter:   attackHonorHitmarks[counter][len(attackHonorHitmarks[counter])-1],
		State:           action.NormalAttackState,
	}, nil
}
