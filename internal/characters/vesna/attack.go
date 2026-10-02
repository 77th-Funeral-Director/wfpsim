package vesna

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
	attackHitmarks        = [][]int{{13}, {19}, {10, 10 + 11}, {21}, {20}, {29}}
	attackHitlagHaltFrame = [][]float64{{0.05}, {0.05}, {0.03, 0.03}, {0.02}, {0.03}, {0.08}}
	attackDefHalt         = [][]bool{{true}, {true}, {true, true}, {true}, {true}, {true}}
	attackHitboxes        = [][]float64{{1.7}, {2}, {1, 1.5}, {1.7}, {1.7}, {1.7}}
	attackOffsets         = []float64{1.8, 0.8, 0.5, 1.8, 1.8, 1.8}
	attackFanAngles       = []float64{360, 180, 360, 360, 360, 360}
	attackStacksPerHit    = []int{1, 1, 1, 1, 1, 3}
)

const normalHitNum = 6

func init() {
	// NA cancels
	attackFrames = make([][]int, normalHitNum)

	attackFrames[0] = frames.InitNormalCancelSlice(attackHitmarks[0][0], 25)
	attackFrames[0][action.ActionAttack] = 14
	attackFrames[0][action.ActionCharge] = 21

	attackFrames[1] = frames.InitNormalCancelSlice(attackHitmarks[1][0], 39)
	attackFrames[1][action.ActionAttack] = 33
	attackFrames[1][action.ActionCharge] = 33

	attackFrames[2] = frames.InitNormalCancelSlice(attackHitmarks[2][1], 30)
	attackFrames[2][action.ActionAttack] = 24
	attackFrames[2][action.ActionCharge] = 23

	attackFrames[3] = frames.InitNormalCancelSlice(attackHitmarks[3][0], 38)
	attackFrames[3][action.ActionAttack] = 31
	attackFrames[3][action.ActionCharge] = 30

	attackFrames[4] = frames.InitNormalCancelSlice(attackHitmarks[4][0], 42)
	attackFrames[4][action.ActionAttack] = 40
	attackFrames[4][action.ActionCharge] = 38

	attackFrames[5] = frames.InitNormalCancelSlice(attackHitmarks[5][0], 50)
	attackFrames[5][action.ActionWalk] = 49
}

func (c *char) Attack(p map[string]int) (action.Info, error) {
	if c.StatusIsActive(c6Key) {
		return c.c6Attack()
	}

	for i, hitmark := range attackHitmarks[c.NormalCounter] {
		ai := info.AttackInfo{
			ActorIndex:         c.Index(),
			Abil:               fmt.Sprintf("Normal %v", c.NormalCounter),
			AttackTag:          attacks.AttackTagNormal,
			ICDTag:             attacks.ICDTagNormalAttack,
			ICDGroup:           attacks.ICDGroupDefault,
			StrikeType:         attacks.StrikeTypeSlash,
			Element:            attributes.Physical,
			Durability:         25,
			Mult:               attack[c.NormalCounter][c.TalentLvlAttack()],
			HitlagFactor:       0.01,
			HitlagHaltFrames:   attackHitlagHaltFrame[c.NormalCounter][i] * 60,
			CanBeDefenseHalted: attackDefHalt[c.NormalCounter][i],
		}
		if c.NormalCounter == 1 || c.NormalCounter == 4 {
			ai.StrikeType = attacks.StrikeTypeSlash
		}

		var cb info.AttackCBFunc
		if c.StatusIsActive(skillKey) {
			ai.Element = attributes.Anemo
			ai.IgnoreInfusion = true
			cb = c.skillStacksCB(attackStacksPerHit[c.NormalCounter])
		}

		ap := combat.NewCircleHitOnTargetFanAngle(
			c.Core.Combat.Player(),
			info.Point{Y: attackOffsets[c.NormalCounter]},
			attackHitboxes[c.NormalCounter][0],
			attackFanAngles[c.NormalCounter],
		)
		if c.NormalCounter == 2 {
			ap = combat.NewBoxHitOnTarget(
				c.Core.Combat.Player(),
				info.Point{Y: attackOffsets[c.NormalCounter]},
				attackHitboxes[c.NormalCounter][0],
				attackHitboxes[c.NormalCounter][1],
			)
		}
		c.Core.QueueAttack(ai, ap, hitmark, hitmark, cb)
	}

	c.pinionAttack(30)

	defer c.AdvanceNormalIndex()

	return action.Info{
		Frames:          frames.NewAttackFunc(c.Character, attackFrames),
		AnimationLength: attackFrames[c.NormalCounter][action.InvalidAction],
		CanQueueAfter:   attackFrames[c.NormalCounter][action.ActionDash],
		State:           action.NormalAttackState,
	}, nil
}
