package xilonen

import (
	"fmt"

	"github.com/genshinsim/gcsim/internal/frames"
	"github.com/genshinsim/gcsim/pkg/core/action"
	"github.com/genshinsim/gcsim/pkg/core/attacks"
	"github.com/genshinsim/gcsim/pkg/core/attributes"
	"github.com/genshinsim/gcsim/pkg/core/combat"
	"github.com/genshinsim/gcsim/pkg/core/geometry"
)

var (
	attackFrames           [][]int
	attackHitmarks         = [][]int{{10}, {9, 19}, {13}}
	attackHitlagHaltFrames = []float64{0.03, 0.03, 0.06}
	attackHitboxes         = [][][]float64{{{1.5}}, {{1.5}, {1.5}}, {{1.5}}}
	attackOffsets          = [][]float64{{0.8}, {0.6, 0.6}, {0.8}}

	rollerFrames           [][]int
	rollerHitmarks         = [][]int{{10}, {9}, {13}, {13}}
	rollerHitlagHaltFrames = []float64{0.03, 0.03, 0.06, 0.06}
	rollerHitboxes         = [][][]float64{{{1.5}}, {{1.5}}, {{1.5}}, {{1.5}}}
	rollerOffsets          = [][]float64{{0.8}, {0.6}, {0.8}, {0.8}}
)

const normalHitNum = 3
const rollerHitNum = 4

func init() {
	attackFrames = make([][]int, normalHitNum)

	attackFrames[0] = frames.InitNormalCancelSlice(attackHitmarks[0][0], 35)
	attackFrames[0][action.ActionAttack] = 18

	attackFrames[1] = frames.InitNormalCancelSlice(attackHitmarks[1][1], 29)
	attackFrames[1][action.ActionAttack] = 24

	attackFrames[2] = frames.InitNormalCancelSlice(attackHitmarks[2][0], 35)
	attackFrames[2][action.ActionCharge] = 500 //TODO: this action is illegal; need better way to handle it

	rollerFrames = make([][]int, rollerHitNum)

	rollerFrames[0] = frames.InitNormalCancelSlice(rollerHitmarks[0][0], 35)
	rollerFrames[0][action.ActionAttack] = 18

	rollerFrames[1] = frames.InitNormalCancelSlice(rollerHitmarks[1][0], 29)
	rollerFrames[1][action.ActionAttack] = 24

	rollerFrames[2] = frames.InitNormalCancelSlice(rollerHitmarks[2][0], 35)
	attackFrames[2][action.ActionAttack] = 29

	rollerFrames[3] = frames.InitNormalCancelSlice(rollerHitmarks[3][0], 35)
}

func (c *char) Attack(p map[string]int) (action.Info, error) {
	if c.nightsoulState.HasBlessing() {
		return c.nightsoulAttack(), nil
	}

	ai := combat.AttackInfo{
		ActorIndex:         c.Index,
		AttackTag:          attacks.AttackTagNormal,
		ICDTag:             attacks.ICDTagNormalAttack,
		ICDGroup:           attacks.ICDGroupDefault,
		StrikeType:         attacks.StrikeTypeSlash,
		Element:            attributes.Physical,
		Durability:         25,
		HitlagHaltFrames:   attackHitlagHaltFrames[c.NormalCounter] * 60,
		HitlagFactor:       0.01,
		CanBeDefenseHalted: true,
	}

	for i, mult := range attack[c.NormalCounter] {
		ax := ai
		ax.Abil = fmt.Sprintf("Normal %v", c.NormalCounter)
		ax.Mult = mult[c.TalentLvlAttack()]
		ap := combat.NewCircleHitOnTarget(
			c.Core.Combat.Player(),
			geometry.Point{Y: attackOffsets[c.NormalCounter][i]},
			attackHitboxes[c.NormalCounter][i][0],
		)
		c.QueueCharTask(func() {
			c.Core.QueueAttack(ax, ap, 0, 0)
		}, attackHitmarks[c.NormalCounter][i])
	}

	defer c.AdvanceNormalIndex()

	return action.Info{
		Frames:          frames.NewAttackFunc(c.Character, attackFrames),
		AnimationLength: attackFrames[c.NormalCounter][action.InvalidAction],
		CanQueueAfter:   attackHitmarks[c.NormalCounter][len(attackHitmarks[c.NormalCounter])-1],
		State:           action.NormalAttackState,
	}, nil
}

func (c *char) nightsoulAttack() action.Info {
	ai := combat.AttackInfo{
		ActorIndex:         c.Index,
		AttackTag:          attacks.AttackTagNormal,
		ICDTag:             attacks.ICDTagNormalAttack,
		ICDGroup:           attacks.ICDGroupDefault,
		StrikeType:         attacks.StrikeTypeSlash,
		Element:            attributes.Geo,
		Durability:         25,
		HitlagHaltFrames:   rollerHitlagHaltFrames[c.NormalCounter] * 60,
		HitlagFactor:       0.01,
		CanBeDefenseHalted: true,
		UseDef:             true,
		IgnoreInfusion:     true,
	}

	for i, mult := range attackE[c.NormalCounter] {
		ax := ai
		ax.Abil = fmt.Sprintf("Blade Roller %v", c.NormalCounter)
		ax.Mult = mult[c.TalentLvlAttack()]
		ap := combat.NewCircleHitOnTarget(
			c.Core.Combat.Player(),
			geometry.Point{Y: rollerOffsets[c.NormalCounter][i]},
			rollerHitboxes[c.NormalCounter][i][0],
		)
		c.QueueCharTask(func() {
			c.Core.QueueAttack(ax, ap, 0, 0, c.a1cb)
		}, rollerHitmarks[c.NormalCounter][i])
	}

	defer c.AdvanceNormalIndex()

	return action.Info{
		Frames:          frames.NewAttackFunc(c.Character, rollerFrames),
		AnimationLength: rollerFrames[c.NormalCounter][action.InvalidAction],
		CanQueueAfter:   rollerHitmarks[c.NormalCounter][len(rollerHitmarks[c.NormalCounter])-1],
		State:           action.NormalAttackState,
	}
}
