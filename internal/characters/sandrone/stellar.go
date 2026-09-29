package sandrone

import (
	"github.com/genshinsim/gcsim/pkg/core/event"
	"github.com/genshinsim/gcsim/pkg/core/glog"
	"github.com/genshinsim/gcsim/pkg/core/info"
	"github.com/genshinsim/gcsim/pkg/reactable"
)

type radianceState int

const (
	radianceNone radianceState = iota
	radianceStellarConduct
	radianceStellarSwirl
)

const stellarBonusKey = "sandrone-stellar-bonus"

func (c *char) getRadiance() radianceState {
	if c.StatusIsActive(reactable.PolestarFieldKey) {
		return radianceStellarConduct
	}

	if c.StatusIsActive(reactable.SswKey) {
		return radianceStellarSwirl
	}

	return radianceNone
}

func (c *char) stellarInit() {
	c.Core.Flags.Custom[reactable.StellarConductEnableKey] = 1
	c.Core.Flags.Custom[reactable.StellarSwirlEnableKey] = 1

	c.Core.Events.Subscribe(event.OnEnemyHit, func(args ...any) {
		atk := args[1].(*info.AttackEvent)

		if !atk.Info.AttackTag.IsStellarDirect() {
			return
		}

		bonus := min(c.TotalAtk()/100.0*0.0035, 0.07)

		if c.Core.Flags.LogDebug {
			c.Core.Log.NewEvent("sandrone adding stellar base damage", glog.LogCharacterEvent, c.Index()).Write("bonus", bonus)
		}

		atk.Info.BaseDmgBonus += bonus
	}, stellarBonusKey)

	c.Core.Events.Subscribe(event.OnSpecialReactionAttack, func(args ...any) {
		atk := args[1].(*info.AttackEvent)

		if !atk.Info.AttackTag.IsStellarReact() {
			return
		}

		bonus := min(c.TotalAtk()/100.0*0.0035, 0.07)

		if c.Core.Flags.LogDebug {
			c.Core.Log.NewEvent("sandrone adding stellar base damage", glog.LogCharacterEvent, c.Index()).Write("bonus", bonus)
		}

		atk.Info.BaseDmgBonus += bonus
	}, stellarBonusKey+"-reaction")
}
