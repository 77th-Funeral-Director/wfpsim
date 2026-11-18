package mitya

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
	attackHitmarks = []int{19, 20, 32}
)

const normalHitNum = 3

func init() {
	attackFrames = make([][]int, normalHitNum)

	attackFrames[0] = frames.InitNormalCancelSlice(attackHitmarks[0], 42)
	attackFrames[0][action.ActionAttack] = 28
	attackFrames[0][action.ActionCharge] = 40

	attackFrames[1] = frames.InitNormalCancelSlice(attackHitmarks[1], 52)
	attackFrames[1][action.ActionAttack] = 33
	attackFrames[1][action.ActionCharge] = 45

	attackFrames[2] = frames.InitNormalCancelSlice(attackHitmarks[2], 63)
	attackFrames[2][action.ActionAttack] = 51
	attackFrames[2][action.ActionWalk] = 59
}

func (c *char) Attack(_ map[string]int) (action.Info, error) {
	ai := info.AttackInfo{
		ActorIndex: c.Index(),
		Abil:       fmt.Sprintf("Normal %v", c.NormalCounter),
		AttackTag:  attacks.AttackTagNormal,
		ICDTag:     attacks.ICDTagNormalAttack,
		ICDGroup:   attacks.ICDGroupDefault,
		StrikeType: attacks.StrikeTypeDefault,
		Element:    attributes.Electro,
		Durability: 25,
		Mult:       attack[c.NormalCounter][c.TalentLvlAttack()],
	}

	ap := combat.NewCircleHit(c.Core.Combat.Player(), c.Core.Combat.PrimaryTarget(), nil, 1.2)
	c.Core.QueueAttack(
		ai,
		ap,
		attackHitmarks[c.NormalCounter],
		attackHitmarks[c.NormalCounter],
	)

	defer c.AdvanceNormalIndex()
	n := c.NormalCounter
	framesFunc := frames.NewAttackFunc(c.Character, attackFrames)
	return action.Info{
		Frames: func(next action.Action) int {
			if c.getSkillState() == skillStateHold {
				return attackHitmarks[n]
			}
			return framesFunc(next)
		},
		AnimationLength: attackFrames[c.NormalCounter][action.InvalidAction],
		CanQueueAfter:   attackHitmarks[c.NormalCounter],
		State:           action.NormalAttackState,
	}, nil
}
