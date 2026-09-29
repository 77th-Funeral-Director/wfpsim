package vodyanitsa

import (
	tmpl "github.com/genshinsim/gcsim/internal/template/character"
	"github.com/genshinsim/gcsim/pkg/core"
	"github.com/genshinsim/gcsim/pkg/core/event"
	"github.com/genshinsim/gcsim/pkg/core/info"
	"github.com/genshinsim/gcsim/pkg/core/keys"
	"github.com/genshinsim/gcsim/pkg/core/player/character"
)

func init() {
	core.RegisterCharFunc(keys.Vodyanitsa, NewChar)
}

type char struct {
	*tmpl.Character

	skillSrc int

	lastVortexDetonateExp int
	leadVocal             int
	chorus                int

	c1Buff   []float64
	c2Buff   []float64
	c4Buff   []float64
	c6Buff   []float64
	c4Stacks RingQueue[int]
}

func NewChar(s *core.Core, w *character.CharWrapper, _ info.CharacterProfile) error {
	c := char{}
	c.Character = tmpl.NewWithWrapper(s, w)

	c.EnergyMax = 60
	c.NormalHitNum = normalHitNum
	c.BurstCon = 5
	c.SkillCon = 3

	w.Character = &c

	return nil
}

func (c *char) Init() error {
	c.stellarSwirlInit()

	c.a1Init()
	c.a4Init()

	c.c1Init()
	c.c2Init()
	c.c4Init()
	c.c6Init()
	return nil
}

func (c *char) AnimationStartDelay(k info.AnimationDelayKey) int {
	if k == info.AnimationXingqiuN0StartDelay {
		return 11
	}
	return c.Character.AnimationStartDelay(k)
}

func (c *char) stellarSwirlInit() {
	c.lastVortexDetonateExp = -1
	c.Core.Events.Subscribe(event.OnStellarVortexDetonate, func(args ...any) {
		c.lastVortexDetonateExp = c.Core.F + 5*60
	}, "vodyanitsa-ssw")
}

func (c *char) recentSSW() bool {
	return c.Core.F < c.lastVortexDetonateExp || c.hasStellarVortex()
}

func (c *char) hasStellarVortex() bool {
	for _, g := range c.Core.Combat.Gadgets() {
		if g != nil && g.GadgetTyp() == info.GadgetTypStellarVortex {
			return true
		}
	}
	return false
}
