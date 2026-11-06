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
)

const normalHitNum = 5

func init() {
	attackFrames = make([][]int, normalHitNum)

	attackFrames[0] = frames.InitNormalCancelSlice(attackHitmarks[0][0], 35)
	attackFrames[1] = frames.InitNormalCancelSlice(attackHitmarks[1][0], 26)
	attackFrames[2] = frames.InitNormalCancelSlice(attackHitmarks[2][1], 39)
	attackFrames[3] = frames.InitNormalCancelSlice(attackHitmarks[3][2], 44)
	attackFrames[4] = frames.InitNormalCancelSlice(attackHitmarks[4][0], 52)
}

func (c *char) Attack(p map[string]int) (action.Info, error) {
	travel, ok := p["travel"]
	if !ok {
		travel = 10
	}

	ai := info.AttackInfo{
		ActorIndex: c.Index(),
		Abil:       fmt.Sprintf("Normal %v", c.NormalCounter),
		AttackTag:  attacks.AttackTagNormal,
		ICDTag:     attacks.ICDTagNone,
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
