package valeriy

import (
	"fmt"

	"github.com/genshinsim/gcsim/internal/frames"
	"github.com/genshinsim/gcsim/pkg/core/action"
	"github.com/genshinsim/gcsim/pkg/core/attacks"
	"github.com/genshinsim/gcsim/pkg/core/attributes"
	"github.com/genshinsim/gcsim/pkg/core/combat"
	"github.com/genshinsim/gcsim/pkg/core/event"
	"github.com/genshinsim/gcsim/pkg/core/glog"
	"github.com/genshinsim/gcsim/pkg/core/info"
)

var (
	chargeFrames        []int
	specialChargeFrames []int
)

const (
	maxMomentum = 100
	reqMomentum = 40

	chargeHitmark        = 39
	specialChargeHitmark = 43

	caBuffKey = "valeriy-ca-buff"
)

func init() {
	chargeFrames = frames.InitAbilSlice(58)
	specialChargeFrames = frames.InitAbilSlice(67)
}

func (c *char) ChargeAttack(p map[string]int) (action.Info, error) {
	if c.momentum >= reqMomentum {
		return c.specialChargeAttack(p)
	}
	ai := info.AttackInfo{
		Abil:       "Charge",
		ActorIndex: c.Index(),
		AttackTag:  attacks.AttackTagExtra,
		ICDTag:     attacks.ICDTagExtraAttack,
		ICDGroup:   attacks.ICDGroupDefault,
		StrikeType: attacks.StrikeTypeSlash,
		Element:    attributes.Physical,
		Durability: 25,
		Mult:       charge[c.TalentLvlAttack()],
	}

	ap := combat.NewCircleHitOnTarget(
		c.Core.Combat.Player(),
		info.Point{Y: 1},
		2.2,
	)

	c.Core.QueueAttack(ai, ap, chargeHitmark, chargeHitmark)

	return action.Info{
		Frames:          frames.NewAbilFunc(chargeFrames),
		AnimationLength: chargeFrames[action.InvalidAction],
		CanQueueAfter:   chargeFrames[action.ActionJump], // earliest cancel
		State:           action.ChargeAttackState,
	}, nil
}

func (c *char) specialChargeAttack(_ map[string]int) (action.Info, error) {
	c.QueueCharTask(func() {
		ai := info.AttackInfo{
			Abil:           "Special Charge",
			ActorIndex:     c.Index(),
			AdditionalTags: []attacks.AttackTag{attacks.AttackTagValeriySpecial},
			AttackTag:      attacks.AttackTagExtra,
			ICDTag:         attacks.ICDTagNone,
			ICDGroup:       attacks.ICDGroupDefault,
			StrikeType:     attacks.StrikeTypePierce,
			Element:        attributes.Electro,
			Durability:     25,
			Mult:           specialCharge[c.TalentLvlAttack()],
		}

		ap := combat.NewBoxHitOnTarget(
			c.Core.Combat.Player(),
			info.Point{Y: -0.1},
			5,
			8,
		)

		c.Core.QueueAttack(ai, ap, 0, 0, c.c6CB)

		c.caStacks = 15
		c.AddStatus(caBuffKey, 18*60, true)
		c.a4OnMomentumConsume(c.momentum)
		c.c1OnMomentumConsume(c.momentum)

		if c.Core.Flags.LogDebug {
			c.Core.Log.NewEvent(fmt.Sprintf("Valeriy consumed momentum (%v)", c.momentum), glog.LogCharacterEvent, c.Index())
		}
		c.momentum = 0
	}, specialChargeHitmark)

	return action.Info{
		Frames:          frames.NewAbilFunc(specialChargeFrames),
		AnimationLength: specialChargeFrames[action.InvalidAction],
		CanQueueAfter:   specialChargeFrames[action.ActionJump], // earliest cancel
		State:           action.ChargeAttackState,
	}, nil
}

func (c *char) addMomentum(amount float64) float64 {
	startingMomentum := c.momentum

	amount *= 1 + c.c4BonusMomentum()

	c.momentum = min(c.momentum+amount, maxMomentum+c.c4BonusMaxMomentum())
	if c.Core.Flags.LogDebug {
		c.Core.Log.NewEvent(fmt.Sprintf("Valeriy add momentum (%v)", c.momentum), glog.LogCharacterEvent, c.Index()).
			Write("starting", startingMomentum).
			Write("amount", amount)
	}
	return c.momentum - startingMomentum
}

func (c *char) chargeInit() {
	c.Core.Events.Subscribe(event.OnEnemyHit, func(args ...any) {
		atk := args[1].(*info.AttackEvent)
		if atk.Info.Element != attributes.Electro {
			return
		}

		if c.caStacks <= 0 {
			return
		}

		if atk.Info.ActorIndex != c.Core.Player.Active() {
			return
		}

		if !c.StatusIsActive(caBuffKey) {
			return
		}

		var scaling float64
		if c.isRadianceSSC() {
			if atk.Info.AttackTag != attacks.AttackTagDirectStellarConduct {
				return
			}
			scaling = ssc[c.TalentLvlAttack()] + c.a4BonusSSCScaling()
		} else {
			switch atk.Info.AttackTag {
			case attacks.AttackTagElementalBurst:
			case attacks.AttackTagElementalArt:
			case attacks.AttackTagElementalArtHold:
			case attacks.AttackTagNormal:
			case attacks.AttackTagExtra:
			case attacks.AttackTagPlunge:
			default:
				return
			}

			scaling = electro[c.TalentLvlAttack()] + c.a4BonusElectroScaling()
		}

		amt := scaling * c.TotalAtk()
		if c.Core.Flags.LogDebug {
			c.Core.Log.NewEvent("Valeriy CA buff proc dmg add", glog.LogPreDamageMod, atk.Info.ActorIndex).
				Write("before", atk.Info.FlatDmg).
				Write("scaling", scaling).
				Write("addition", amt).
				Write("ca_stacks", c.caStacks)
		}

		atk.Info.FlatDmg += amt
		c.caStacks--
	}, "valeriy-charge")
}
