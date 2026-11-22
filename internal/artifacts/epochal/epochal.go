package epochal

import (
	"github.com/genshinsim/gcsim/pkg/core"
	"github.com/genshinsim/gcsim/pkg/core/attributes"
	"github.com/genshinsim/gcsim/pkg/core/event"
	"github.com/genshinsim/gcsim/pkg/core/info"
	"github.com/genshinsim/gcsim/pkg/core/player/character"
	"github.com/genshinsim/gcsim/pkg/modifier"
)

type Set struct {
	Index     int
	Count     int
	char      *character.CharWrapper
	core      *core.Core
	starSpark int
	src       int
}

func (s *Set) SetIndex(idx int) { s.Index = idx }
func (s *Set) GetCount() int    { return s.Count }
func (s *Set) Init() error      { return nil }

func NewSet(c *core.Core, char *character.CharWrapper, count int, param map[string]int) (info.Set, error) {
	s := Set{
		Count: count,
		char:  char,
		core:  c,
	}

	if count < 2 {
		return &s, nil
	}
	m := make([]float64, attributes.EndStatType)
	m[attributes.ATKP] = 0.18
	char.AddStatMod(character.StatMod{
		Base:         modifier.NewBase("epochal-2pc", -1),
		AffectedStat: attributes.ATKP,
		Amount: func() []float64 {
			return m
		},
	})

	if s.Count < 4 {
		return &s, nil
	}

	cryoCR := 0.2
	stellarCryoCR := 0.4
	m2 := make([]float64, attributes.EndStatType)
	char.AddAttackMod(character.AttackMod{
		Base: modifier.NewBase("epochal-4pc", -1),
		Amount: func(atk *info.AttackEvent, t info.Target) []float64 {
			if atk.Info.Element != attributes.Cryo {
				return nil
			}

			if s.starSpark == 0 {
				return nil
			}

			if !atk.Info.AttackTag.IsStellar() {
				m2[attributes.CR] = cryoCR
				return m2
			}

			m2[attributes.CR] = stellarCryoCR
			return m2
		},
	})

	gainStarSpark := func(args ...any) {
		if c.Player.Active() != char.Index() {
			return
		}
		s.starSpark = min(s.starSpark+1, 2)
		src := c.F
		s.src = src
		char.QueueCharTask(func() { s.starSparkReduceTicker(src) }, 3*60)
	}
	c.Events.Subscribe(event.OnSkill, gainStarSpark, "epochal-4pc-"+char.Base.Key.String())
	c.Events.Subscribe(event.OnBurst, gainStarSpark, "epochal-4pc-"+char.Base.Key.String())
	c.Events.Subscribe(event.OnAttack, gainStarSpark, "epochal-4pc-"+char.Base.Key.String())

	c.Events.Subscribe(event.OnCharacterSwap, func(args ...any) {
		next := args[1].(int)

		if next != s.char.Index() {
			return
		}

		if s.starSpark == 0 {
			return
		}
		src := c.F
		s.src = src
		char.QueueCharTask(func() { s.starSparkReduceTicker(src) }, 3*60)
	}, "epochal-4pc-"+char.Base.Key.String())

	return &s, nil
}

func (s *Set) starSparkReduceTicker(src int) {
	if s.core.Player.Active() != s.char.Index() {
		return
	}

	if s.src != src {
		return
	}

	s.starSpark -= 1
	if s.starSpark > 0 {
		s.char.QueueCharTask(func() { s.starSparkReduceTicker(src) }, 3*60)
	}
}
