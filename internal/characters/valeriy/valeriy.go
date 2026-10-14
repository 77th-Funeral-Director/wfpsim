package valeriy

import (
	tmpl "github.com/genshinsim/gcsim/internal/template/character"
	"github.com/genshinsim/gcsim/pkg/core"
	"github.com/genshinsim/gcsim/pkg/core/action"
	"github.com/genshinsim/gcsim/pkg/core/info"
	"github.com/genshinsim/gcsim/pkg/core/keys"
	"github.com/genshinsim/gcsim/pkg/core/player/character"
)

type char struct {
	*tmpl.Character

	momentum float64
	caStacks int

	skillShield *shd
	burstSrc    int

	a1Momentum float64
	a4Momentum float64

	c1SkillUsed bool
	c2Buff      []float64
}

func NewChar(s *core.Core, w *character.CharWrapper, _ info.CharacterProfile) error {
	c := char{}
	c.Character = tmpl.NewWithWrapper(s, w)

	c.EnergyMax = 60
	c.NormalHitNum = normalHitNum
	c.NormalCon = 3
	c.SkillCon = 5

	w.Character = &c

	return nil
}

func (c *char) Init() error {
	c.chargeInit()
	c.a1Init()
	c.c2Init()
	c.c6Init()
	return nil
}

func (c *char) NextQueueItemIsValid(k keys.Char, a action.Action, p map[string]int) error {
	if a == action.ActionCharge && c.momentum >= reqMomentum {
		return nil
	}
	return c.Character.NextQueueItemIsValid(k, a, p)
}

func (c *char) AnimationStartDelay(k info.AnimationDelayKey) int {
	if k == info.AnimationXingqiuN0StartDelay {
		return 6
	}
	return c.Character.AnimationStartDelay(k)
}
