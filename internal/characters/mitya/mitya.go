package mitya

import (
	tmpl "github.com/genshinsim/gcsim/internal/template/character"
	"github.com/genshinsim/gcsim/pkg/core"
	"github.com/genshinsim/gcsim/pkg/core/info"
	"github.com/genshinsim/gcsim/pkg/core/player/character"
)

type char struct {
	*tmpl.Character
	skillSrc             int
	skillState           skillState
	skillChargeFinalAnim bool
	prisms               int
	prismSrc             int
	prismRemoveSrc       int
	a4Stacks             int
	a4Buff               []float64
	c2Src                int
	c4Stacks             int
}

func NewChar(s *core.Core, w *character.CharWrapper, p info.CharacterProfile) error {
	c := char{}
	c.Character = tmpl.NewWithWrapper(s, w)

	c.EnergyMax = 60
	c.BurstCon = 5
	c.SkillCon = 3
	c.NormalHitNum = normalHitNum

	w.Character = &c

	return nil
}

func (c *char) Init() error {
	c.prismInit()
	c.stellarInit()
	c.a1Init()
	c.a4Init()
	c.c1Init()
	c.c2Init()
	c.c6Init()
	return nil
}

func (c *char) AnimationStartDelay(k info.AnimationDelayKey) int {
	if k == info.AnimationXingqiuN0StartDelay {
		return 10
	}
	return c.Character.AnimationStartDelay(k)
}
