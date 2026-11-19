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

// hitmark frame, excludes CA windup
const (
	chargeHitmark = 42

	specialCAStart    = 62
	specialCATick     = 18
	specialCALastTick = 21
	specialCAFinisher = 30
)

func init() {
	chargeFrames = frames.InitAbilSlice(63)
	chargeFrames[action.ActionAttack] = 53
	chargeFrames[action.ActionCharge] = 52
	chargeFrames[action.ActionSkill] = 53
	chargeFrames[action.ActionBurst] = 54
	chargeFrames[action.ActionDash] = chargeHitmark
	chargeFrames[action.ActionJump] = chargeHitmark
	chargeFrames[action.ActionWalk] = 63
	chargeFrames[action.ActionSwap] = 41

	specialChargeEndlag = frames.InitAbilSlice(14)
	specialChargeEndlag[action.ActionAttack] = 3
	specialChargeEndlag[action.ActionSkill] = 1
	specialChargeEndlag[action.ActionBurst] = 3
	specialChargeEndlag[action.ActionDash] = 0
	specialChargeEndlag[action.ActionJump] = 0
	specialChargeEndlag[action.ActionWalk] = 1
	specialChargeEndlag[action.ActionSwap] = 0
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
	switch c.Core.Player.CurrentState() {
	case action.Idle, action.ChargeAttackState, action.SkillState:
		windup = 13
	case action.BurstState:
		windup = 11
	}

	radius := 3.5

	c.Core.QueueAttack(
		ai,
		combat.NewCircleHitOnTarget(c.Core.Combat.PrimaryTarget(), nil, radius),
		chargeHitmark+windup,
		chargeHitmark+windup,
	)

	return action.Info{
		Frames:          func(next action.Action) int { return chargeFrames[next] + windup },
		AnimationLength: chargeFrames[action.InvalidAction] + windup,
		CanQueueAfter:   chargeFrames[action.ActionSwap] + windup,
		State:           action.ChargeAttackState,
	}, nil
}

func (c *char) specialChargeAttack() (action.Info, error) {
	// skip CA windup if we're in NA animation
	windup := 0
	if c.Core.Player.CurrentState() == action.Idle {
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
	}, specialCAStart+windup)

	c.specialChargeDur = specialCAStart + windup

	return action.Info{
		Frames: func(next action.Action) int {
			return c.specialChargeDur + specialChargeEndlag[next]
		},
		AnimationLength: 1200,
		CanQueueAfter:   specialCAStart + windup + specialChargeEndlag[action.ActionSwap],
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
			c.specialChargeSSC(0)
			c.specialChargeTicker(src)
		}, specialCATick)
	case c.prisms == 1:
		c.specialChargeDur += specialCALastTick + specialCAFinisher
		c.QueueCharTask(func() {
			c.specialChargeSSC(specialCAFinisher)
		}, specialCALastTick)
	}
}

func (c *char) specialChargeSSC(delay int) {
	c.consumePrism(1)
	ai := info.AttackInfo{
		ActorIndex:       c.Index(),
		Abil:             "Special Charge" + stellarConductText,
		AttackTag:        attacks.AttackTagDirectStellarConduct,
		Element:          attributes.Electro,
		Mult:             chargeSSC[c.TalentLvlAttack()],
		UseEM:            true,
		IgnoreDefPercent: 1,
	}

	ap := combat.NewCircleHitOnTarget(c.Core.Combat.PrimaryTarget(), nil, 5)
	c.Core.QueueAttack(ai, ap, delay, delay)
}
