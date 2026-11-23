package silverdawn

import (
	"github.com/genshinsim/gcsim/pkg/core"
	"github.com/genshinsim/gcsim/pkg/core/attributes"
	"github.com/genshinsim/gcsim/pkg/core/event"
	"github.com/genshinsim/gcsim/pkg/core/info"
	"github.com/genshinsim/gcsim/pkg/core/player/character"
	"github.com/genshinsim/gcsim/pkg/enemy"
	"github.com/genshinsim/gcsim/pkg/modifier"
)

const (
	atkpKey      = "silverdawn-atkp"
	cdKey        = "silverdawn-cd"
	energyIcdKey = "silverdawn-energy-icd"
	extensionKey = "silverdawn-extension"
)

type Weapon struct {
	Index          int
	extensionsLeft int
}

func (w *Weapon) SetIndex(idx int) { w.Index = idx }
func (w *Weapon) Init() error      { return nil }

func NewWeapon(c *core.Core, char *character.CharWrapper, p info.WeaponProfile) (info.Weapon, error) {
	w := &Weapon{}
	r := p.Refine

	mAtk := make([]float64, attributes.EndStatType)
	mAtk[attributes.ATKP] = (0.28 + float64(r)*0.12)

	cd := (0.28 + float64(r)*0.12)
	mCD := make([]float64, attributes.EndStatType)
	mCD[attributes.CD] = cd

	energy := 9.5 + float64(r)*0.5

	onSkillBurst := func(args ...any) {
		if c.Player.Active() != char.Index() {
			return
		}

		char.AddStatMod(character.StatMod{
			Base:         modifier.NewBaseWithHitlag(atkpKey, 6*60),
			AffectedStat: attributes.ATKP,
			Amount: func() []float64 {
				return mAtk
			},
		})

		char.AddAttackMod(character.AttackMod{
			Base: modifier.NewBaseWithHitlag(cdKey, 6*60),
			Amount: func(atk *info.AttackEvent, t info.Target) []float64 {
				if !atk.Info.AttackTag.IsStellar() {
					return nil
				}
				return mCD
			},
		})
		w.extensionsLeft = 8

		if char.StatusIsActive(energyIcdKey) {
			return
		}
		char.AddStatus(energyIcdKey, 18*60, true)
		char.AddEnergy("silverdawn-energy", energy)
	}
	c.Events.Subscribe(event.OnSkill, onSkillBurst, "silverdawn-"+char.Base.Key.String())
	c.Events.Subscribe(event.OnBurst, onSkillBurst, "silverdawn-"+char.Base.Key.String())

	c.Events.Subscribe(event.OnSpecialReactionAttack, func(args ...any) {
		ae, ok := args[1].(*info.AttackEvent)
		if !ok {
			return
		}

		if !ae.Info.AttackTag.IsStellar() {
			return
		}

		if ae.Info.ActorIndex != char.Index() {
			return
		}

		if !char.StatusIsActive(cdKey) {
			return
		}

		ae.Snapshot.Stats[attributes.CR] += cd
	}, "silverdawn-"+char.Base.Key.String())

	c.Events.Subscribe(event.OnEnemyDamage, func(args ...any) {
		if w.extensionsLeft == 0 {
			return
		}

		_, ok := args[0].(*enemy.Enemy)
		if !ok {
			return
		}

		atk, ok := args[1].(*info.AttackEvent)
		if !ok {
			return
		}

		if atk.Info.ActorIndex != char.Index() {
			return
		}

		if !atk.Info.AttackTag.IsStellar() {
			return
		}

		char.ExtendStatus(atkpKey, 3*60)
		char.ExtendStatus(cdKey, 3*60)
		w.extensionsLeft -= 1
	}, "silverdawn-"+char.Base.Key.String())

	onStellarReact := func(args ...any) {
		if w.extensionsLeft == 0 {
			return
		}

		_, ok := args[0].(*enemy.Enemy)
		if !ok {
			return
		}

		atk, ok := args[1].(*info.AttackEvent)
		if !ok {
			return
		}

		if atk.Info.ActorIndex != char.Index() {
			return
		}

		char.ExtendStatus(atkpKey, 3*60)
		char.ExtendStatus(cdKey, 3*60)
		w.extensionsLeft -= 1
	}

	c.Events.Subscribe(event.OnStellarConduct, onStellarReact, "silverdawn-"+char.Base.Key.String())
	c.Events.Subscribe(event.OnStellarSwirl, onStellarReact, "silverdawn-"+char.Base.Key.String())
	return w, nil
}
