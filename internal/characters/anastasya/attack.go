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
	attackHonorHitmarks    = [][]int{{8}, {6, 6 + 14}, {14}, {14, 14 + 11, 14 + 11 + 10}, {28, 29, 30}}
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

	attackHonorFrames[0] = frames.InitNormalCancelSlice(attackHonorHitmarks[0][0], 21)
	attackHonorFrames[1] = frames.InitNormalCancelSlice(attackHonorHitmarks[1][1], 31)
	attackHonorFrames[2] = frames.InitNormalCancelSlice(attackHonorHitmarks[2][0], 33)
	attackHonorFrames[3] = frames.InitNormalCancelSlice(attackHonorHitmarks[3][2], 49)
	attackHonorFrames[4] = frames.InitNormalCancelSlice(attackHonorHitmarks[4][0], 57)
}

func (c *char) Attack(p map[string]int) (action.Info, error) {
	if c.getBlessing() == honor {
		return c.attackHonor(p)
	}

	travel, ok := p["travel"]
	if !ok {
		travel = 10
	}

	ai := info.AttackInfo{
		ActorIndex: c.Index(),
		Abil:       fmt.Sprintf("Normal %v", c.NormalCounter),
		AttackTag:  attacks.AttackTagNormal,
		ICDTag:     attacks.ICDTagNormalAttack,
		ICDGroup:   attacks.ICDGroupDefault,
		StrikeType: attacks.StrikeTypePierce,
		Element:    attributes.Physical,
		Mult:       attack[c.NormalCounter][c.TalentLvlAttack()],
		Durability: 25,
	}

	for _, delay := range attackHitmarks[c.NormalCounter] {
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
		AnimationLength: attackFrames[c.NormalCounter][action.InvalidAction],
		CanQueueAfter:   attackHitmarks[c.NormalCounter][len(attackHitmarks[c.NormalCounter])-1],
		State:           action.NormalAttackState,
	}, nil
}

func (c *char) attackHonor(_ map[string]int) (action.Info, error) {
	ai := info.AttackInfo{
		ActorIndex:       c.Index(),
		Abil:             fmt.Sprintf("Normal %v", c.NormalCounter),
		AttackTag:        attacks.AttackTagNormal,
		ICDTag:           attacks.ICDTagAnastasyaHonorNA,
		ICDGroup:         attacks.ICDGroupAnastasyaHonorNA,
		StrikeType:       attacks.StrikeTypePierce,
		Element:          attributes.Cryo,
		Mult:             skillAttack[c.NormalCounter][c.TalentLvlSkill()],
		HitlagFactor:     0.01,
		HitlagHaltFrames: attackHonorHitHaltTime[c.NormalCounter] * 60,
		Durability:       25,
	}

	for i, delay := range attackHonorHitmarks[c.NormalCounter] {
		if i == len(attackHonorHitmarks[c.NormalCounter])-1 {
			sscMult := skillAttackSSC[c.NormalCounter][c.TalentLvlSkill()]
			sswMult := skillAttackSSw[c.NormalCounter][c.TalentLvlSkill()]
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

	c.burstHonorOnAttack(c.NormalCounter)

	defer c.AdvanceNormalIndex()

	return action.Info{
		Frames:          frames.NewAttackFunc(c.Character, attackHonorFrames),
		AnimationLength: attackHonorFrames[c.NormalCounter][action.InvalidAction],
		CanQueueAfter:   attackHonorHitmarks[c.NormalCounter][len(attackHonorHitmarks[c.NormalCounter])-1],
		State:           action.NormalAttackState,
	}, nil
}
