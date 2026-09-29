package mitya

import (
	"github.com/genshinsim/gcsim/pkg/core/attacks"
	"github.com/genshinsim/gcsim/pkg/core/attributes"
	"github.com/genshinsim/gcsim/pkg/core/event"
	"github.com/genshinsim/gcsim/pkg/core/glog"
	"github.com/genshinsim/gcsim/pkg/core/info"
	"github.com/genshinsim/gcsim/pkg/reactable"
)

const (
	stellarBonusKey    = "mitya-stellar-bonus"
	stellarConductText = " (Stellar-Conduct)"
	prismGainKey       = "mitya-prism"
)

// Seems like he doesn't need radiance to do anything?
// type radianceState int

// const (
// 	radianceNone radianceState = iota
// 	radianceStellarConduct
// )

// func (c *char) isRadianceSSC() bool {
// 	return c.StatusIsActive(reactable.PolestarFieldKey)
// }

// ignore the carrying around a polestar field stuff
func (c *char) stellarInit() {
	c.Core.Flags.Custom[reactable.StellarConductEnableKey] = 1

	c.Core.Events.Subscribe(event.OnEnemyHit, func(args ...any) {
		atk := args[1].(*info.AttackEvent)

		if atk.Info.AttackTag != attacks.AttackTagDirectStellarConduct {
			return
		}

		bonus := min(c.Stat(attributes.EM)*0.028e-2, 0.14)

		if c.Core.Flags.LogDebug {
			c.Core.Log.NewEvent("mitya adding stellar base damage", glog.LogCharacterEvent, c.Index()).Write("bonus", bonus)
		}

		atk.Info.BaseDmgBonus += bonus
	}, stellarBonusKey)
}
