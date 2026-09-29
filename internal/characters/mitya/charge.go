package mitya

import (
	"github.com/genshinsim/gcsim/internal/frames"
	"github.com/genshinsim/gcsim/pkg/core/action"
	"github.com/genshinsim/gcsim/pkg/core/attacks"
	"github.com/genshinsim/gcsim/pkg/core/attributes"
	"github.com/genshinsim/gcsim/pkg/core/combat"
	"github.com/genshinsim/gcsim/pkg/core/glog"
	"github.com/genshinsim/gcsim/pkg/core/info"
)

// copied from kokomi

var (
	chargeFrames        []int
	specialChargeEndlag []int
)

// hitmark frame, includes CA windup
const (
	chargeHitmark     = 57
	specialCAFinisher = 45
	specialCAStart    = 62
	specialCATick     = 15
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

	specialChargeEndlag = frames.InitAbilSlice(32)
}

// Standard charge attack
// CA has no travel time
func (c *char) ChargeAttack(p map[string]int) (action.Info, error) {
	if c.getSkillState() == skillStateHold {
		return c.specialChargeAttack()
	}

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
	// skip CA windup if we're in NA animation
	windup := 0
	if c.Core.Player.CurrentState() == action.NormalAttackState {
		windup = 14
	}

	c.QueueCharTask(func() {
		ai := info.AttackInfo{
			ActorIndex: c.Index(),
			Abil:       "Special Charge",
			AttackTag:  attacks.AttackTagExtra,
			ICDTag:     attacks.ICDTagNone,
			ICDGroup:   attacks.ICDGroupDefault,
			StrikeType: attacks.StrikeTypeDefault,
			Element:    attributes.Electro,
			Durability: 25,
			Mult:       chargeSpecial[c.TalentLvlAttack()],
		}

		c.Core.QueueAttack(
			ai,
			combat.NewCircleHitOnTarget(c.Core.Combat.PrimaryTarget(), nil, 5),
			0,
			0,
		)

		c.specialChargeTickSrc = c.Core.F
		c.specialChargeTicker(c.specialChargeTickSrc)
	}, specialCAStart-windup)

	c.specialChargeDur = specialCAStart - windup

	return action.Info{
		Frames: func(next action.Action) int {
			return c.specialChargeDur + specialChargeEndlag[next]
		},
		AnimationLength: 1200,
		CanQueueAfter:   specialCAStart - windup + specialChargeEndlag[action.ActionDash],
		State:           action.ChargeAttackState,
	}, nil
}

func (c *char) specialChargeTicker(src int) {
	if c.prisms == 0 {
		c.Core.Log.NewEvent("no prisms", glog.LogSimEvent, c.Index())
		return
	}

	if c.specialChargeTickSrc != src {
		c.Core.Log.NewEvent("special charge ticker src mismatch", glog.LogSimEvent, c.Index())
		return
	}

	switch {
	case c.prisms > 1:
		c.specialChargeDur += specialCATick
		c.QueueCharTask(func() {
			c.specialChargeSSC()
			c.specialChargeTicker(src)
		}, specialCATick)
	case c.prisms == 1:
		c.specialChargeDur += specialCAFinisher
		c.QueueCharTask(func() {
			c.specialChargeSSC()
		}, specialCAFinisher)
	}
}

func (c *char) specialChargeSSC() {
	c.consumePrism(1)
	ai := info.AttackInfo{
		ActorIndex:       c.Index(),
		Abil:             "Special Charge" + stellarConductText,
		AttackTag:        attacks.AttackTagDirectStellarConduct,
		Element:          attributes.Electro,
		Mult:             chargeSSC[c.TalentLvlSkill()],
		UseEM:            true,
		IgnoreDefPercent: 1,
	}

	ap := combat.NewCircleHitOnTarget(c.Core.Combat.PrimaryTarget(), nil, 5)
	c.Core.QueueAttack(ai, ap, 0, 0)
}
