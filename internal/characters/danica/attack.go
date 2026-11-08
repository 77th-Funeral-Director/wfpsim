package danica

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
	attackFrames          [][]int
	attackHitmarks        = []int{27, 24, 19, 41}
	attackPoiseDMG        = []float64{105.8, 98.1, 115.34, 151.68}
	attackHitlagHaltFrame = []float64{0.10, 0.10, 0.09, 0.15}
	attackHitboxes        = [][]float64{{2}, {2}, {2}, {2, 3}}
	attackOffsets         = []float64{1, 1, 1, -1}
)

const normalHitNum = 4

func init() {
	attackFrames = make([][]int, normalHitNum)

	attackFrames[0] = frames.InitNormalCancelSlice(attackHitmarks[0], 54)
	attackFrames[0][action.ActionAttack] = 38
	attackFrames[0][action.ActionCharge] = 41

	attackFrames[1] = frames.InitNormalCancelSlice(attackHitmarks[1], 66)
	attackFrames[1][action.ActionAttack] = 46
	attackFrames[1][action.ActionCharge] = 48

	attackFrames[2] = frames.InitNormalCancelSlice(attackHitmarks[2], 72)
	attackFrames[2][action.ActionAttack] = 31
	attackFrames[2][action.ActionCharge] = 40

	attackFrames[3] = frames.InitNormalCancelSlice(attackHitmarks[3], 112)
	attackFrames[3][action.ActionAttack] = 106
	attackFrames[3][action.ActionCharge] = 109
}

func (c *char) Attack(p map[string]int) (action.Info, error) {
	if c.StatusIsActive(skillKey) {
		return action.Info{}, fmt.Errorf("danica attack during skill not implemented")
	}

	ai := info.AttackInfo{
		ActorIndex:         c.Index(),
		Abil:               fmt.Sprintf("Normal %v", c.NormalCounter),
		AttackTag:          attacks.AttackTagNormal,
		ICDTag:             attacks.ICDTagNormalAttack,
		ICDGroup:           attacks.ICDGroupDefault,
		StrikeType:         attacks.StrikeTypeBlunt,
		Element:            attributes.Physical,
		Durability:         25,
		Mult:               attack[c.NormalCounter][c.TalentLvlAttack()],
		HitlagFactor:       0.01,
		HitlagHaltFrames:   attackHitlagHaltFrame[c.NormalCounter] * 60,
		CanBeDefenseHalted: true,
	}

	counter := c.NormalCounter
	// need char queue because of potential hitlag from C4
	c.QueueCharTask(func() {
		ai.PoiseDMG = attackPoiseDMG[counter]

		ap := combat.NewCircleHitOnTarget(
			c.Core.Combat.Player(),
			info.Point{Y: attackOffsets[counter]},
			attackHitboxes[counter][0],
		)
		if counter == 3 {
			ap = combat.NewBoxHitOnTarget(
				c.Core.Combat.Player(),
				info.Point{Y: attackOffsets[counter]},
				attackHitboxes[counter][0],
				attackHitboxes[counter][1],
			)
		}

		c.Core.QueueAttack(ai, ap, 0, 0)
	}, attackHitmarks[c.NormalCounter])

	defer c.AdvanceNormalIndex()

	return action.Info{
		Frames:          frames.NewAttackFunc(c.Character, attackFrames),
		AnimationLength: attackFrames[c.NormalCounter][action.InvalidAction],
		CanQueueAfter:   attackHitmarks[c.NormalCounter],
		State:           action.NormalAttackState,
	}, nil
}
