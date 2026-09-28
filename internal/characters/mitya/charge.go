package mitya

import (
	"github.com/genshinsim/gcsim/internal/frames"
	"github.com/genshinsim/gcsim/pkg/core/action"
	"github.com/genshinsim/gcsim/pkg/core/attacks"
	"github.com/genshinsim/gcsim/pkg/core/attributes"
	"github.com/genshinsim/gcsim/pkg/core/combat"
	"github.com/genshinsim/gcsim/pkg/core/info"
)

// copied from kokomi

var (
	chargeFrames          []int
	specialChargeEndlag   []int
	specialChargeHitmarks = []int{62, 62 + 17, 62 + 17 + 12, 62 + 17 + 12 + 14 + 19}
)

// hitmark frame, includes CA windup
const (
	chargeHitmark     = 57
	specialCAFinisher = 45
)

func init() {
	chargeFrames = frames.InitAbilSlice(76)
	chargeFrames[action.ActionAttack] = 70
	chargeFrames[action.ActionCharge] = 70
	chargeFrames[action.ActionSkill] = 70
	chargeFrames[action.ActionBurst] = 70
	chargeFrames[action.ActionDash] = chargeHitmark
	chargeFrames[action.ActionJump] = chargeHitmark
	chargeFrames[action.ActionSwap] = 70

	specialChargeEndlag = frames.InitAbilSlice(76)
}

// Standard charge attack
// CA has no travel time
func (c *char) ChargeAttack(p map[string]int) (action.Info, error) {
	ai := info.AttackInfo{
		ActorIndex: c.Index(),
		Abil:       "Charge",
		AttackTag:  attacks.AttackTagExtra,
		ICDTag:     attacks.ICDTagNone,
		ICDGroup:   attacks.ICDGroupDefault,
		StrikeType: attacks.StrikeTypeDefault,
		Element:    attributes.Electro,
		Durability: 25,
		Mult:       charge[c.TalentLvlAttack()],
	}

	// skip CA windup if we're in NA animation
	windup := 0
	if c.Core.Player.CurrentState() == action.NormalAttackState {
		windup = 14
	}

	radius := 3.5

	c.Core.QueueAttack(
		ai,
		combat.NewCircleHitOnTarget(c.Core.Combat.PrimaryTarget(), nil, radius),
		chargeHitmark-windup,
		chargeHitmark-windup,
	)

	return action.Info{
		Frames:          func(next action.Action) int { return chargeFrames[next] - windup },
		AnimationLength: chargeFrames[action.InvalidAction] - windup,
		CanQueueAfter:   chargeHitmark - windup,
		State:           action.ChargeAttackState,
	}, nil
}

func (c *char) specialChargeAttack() (action.Info, error) {
	ai := info.AttackInfo{
		ActorIndex: c.Index(),
		Abil:       "Special Charge",
		AttackTag:  attacks.AttackTagExtra,
		ICDTag:     attacks.ICDTagNone,
		ICDGroup:   attacks.ICDGroupDefault,
		StrikeType: attacks.StrikeTypeDefault,
		Element:    attributes.Electro,
		Durability: 25,
		Mult:       charge[c.TalentLvlAttack()],
	}

	// skip CA windup if we're in NA animation
	windup := 0
	if c.Core.Player.CurrentState() == action.NormalAttackState {
		windup = 14
	}

	radius := 3.5

	c.Core.QueueAttack(
		ai,
		combat.NewCircleHitOnTarget(c.Core.Combat.PrimaryTarget(), nil, radius),
		chargeHitmark-windup,
		chargeHitmark-windup,
	)

	return action.Info{
		Frames:          func(next action.Action) int { return chargeFrames[next] - windup },
		AnimationLength: chargeFrames[action.InvalidAction] - windup,
		CanQueueAfter:   chargeHitmark - windup,
		State:           action.ChargeAttackState,
	}, nil
}

func (c *char) specialChargeTicker() {
}
