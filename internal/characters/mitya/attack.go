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
	attackHitmarks = []int{15, 21, 22}
)

const normalHitNum = 3

func init() {
	attackFrames = make([][]int, normalHitNum)

	attackFrames[0] = frames.InitNormalCancelSlice(attackHitmarks[0], 500)
	attackFrames[0][action.ActionAttack] = 26
	attackFrames[0][action.ActionCharge] = 26

	attackFrames[1] = frames.InitNormalCancelSlice(attackHitmarks[1], 500)
	attackFrames[1][action.ActionAttack] = 33
	attackFrames[1][action.ActionCharge] = 33

	attackFrames[2] = frames.InitNormalCancelSlice(attackHitmarks[2], 500)
	attackFrames[2][action.ActionAttack] = 50
	attackFrames[2][action.ActionCharge] = 50
}

// Standard attack damage function
// Has "travel" parameter, used to set the number of frames that the projectile is in the air (default = 10)
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
	// TODO: Assume that this is not dynamic (snapshot on projectile release)
	c.Core.QueueAttack(
		ai,
		ap,
		attackHitmarks[c.NormalCounter],
		attackHitmarks[c.NormalCounter],
	)

	defer c.AdvanceNormalIndex()

	return action.Info{
		Frames:          frames.NewAttackFunc(c.Character, attackFrames),
		AnimationLength: attackFrames[c.NormalCounter][action.InvalidAction],
		CanQueueAfter:   attackHitmarks[c.NormalCounter],
		State:           action.NormalAttackState,
	}, nil
}
