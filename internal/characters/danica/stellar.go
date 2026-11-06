package danica

import (
	"github.com/genshinsim/gcsim/pkg/core/attacks"
	"github.com/genshinsim/gcsim/pkg/core/event"
	"github.com/genshinsim/gcsim/pkg/core/glog"
	"github.com/genshinsim/gcsim/pkg/core/info"
	"github.com/genshinsim/gcsim/pkg/reactable"
)

type radianceState int

const (
	radianceNone radianceState = iota
	radianceStellarConduct
)

const (
	stellarConductText = " (Stellar-Conduct)"
)

const stellarBonusKey = "danica-stellar-bonus"

func (c *char) getRadiance() radianceState {
	if c.StatusIsActive(reactable.PolestarFieldKey) {
		return radianceStellarConduct
	}

	return radianceNone
}

func (c *char) toStellar(ai *info.AttackInfo, sscMult float64) {
	if c.getRadiance() == radianceStellarConduct {
		ai.Abil += stellarConductText
		ai.AttackTag = attacks.AttackTagDirectStellarConduct
		ai.Durability = 0
		ai.Mult = sscMult
		ai.IgnoreDefPercent = 1
	}
}

func (c *char) stellarInit() {
	c.Core.Flags.Custom[reactable.StellarConductEnableKey] = 1

	c.Core.Events.Subscribe(event.OnEnemyHit, func(args ...any) {
		atk := args[1].(*info.AttackEvent)

		if atk.Info.AttackTag != attacks.AttackTagDirectStellarConduct {
			return
		}

		bonus := min(c.TotalAtk()/100.0*0.007, 0.14)

		if c.Core.Flags.LogDebug {
			c.Core.Log.NewEvent("danica adding stellar base damage", glog.LogCharacterEvent, c.Index()).Write("bonus", bonus)
		}

		atk.Info.BaseDmgBonus += bonus
	}, stellarBonusKey)
}
