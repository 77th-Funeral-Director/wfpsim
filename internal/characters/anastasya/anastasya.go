package anastasya

import (
	tmpl "github.com/genshinsim/gcsim/internal/template/character"
	"github.com/genshinsim/gcsim/pkg/core"
	"github.com/genshinsim/gcsim/pkg/core/action"
	"github.com/genshinsim/gcsim/pkg/core/info"
	"github.com/genshinsim/gcsim/pkg/core/player/character"
)

type char struct {
	*tmpl.Character

	skillSrc        int
	burstSrc        int
	burstLoveToSeal bool
	unsealPoints    float64
	unsealReady     [4]bool
	c6LoveReady     bool
	c6HonorReady    bool
	c6a1Override    bool
}

func NewChar(s *core.Core, w *character.CharWrapper, _ info.CharacterProfile) error {
	c := char{}
	c.Character = tmpl.NewWithWrapper(s, w)

	c.EnergyMax = 50
	c.BurstCon = 3
	c.SkillCon = 5
	c.NormalHitNum = normalHitNum

	w.Character = &c

	return nil
}

func (c *char) Init() error {
	c.burstInit()
	c.a1Init()
	c.a4Init()
	c.stellarInit()

	c.c1Init()
	c.c2Init()
	c.c6Init()
	return nil
}

func (c *char) ActionReady(a action.Action, p map[string]int) (bool, action.Failure) {
	if a == action.ActionSkill {
		unseal, ok := p["unseal"]
		if ok {
			if c.getBlessing() == honor && 0 < unseal && unseal <= len(c.Core.Player.Chars()) {
				return c.unsealReady[unseal-1], action.NoFailure
			}
			// let this pass through to the skillDoom function to return an error
			return true, action.NoFailure
		}

		if c.c6HonorReady {
			return true, action.NoFailure
		}
	}
	return c.Character.ActionReady(a, p)
}

func (c *char) AnimationStartDelay(k info.AnimationDelayKey) int {
	if k == info.AnimationXingqiuN0StartDelay {
		return 17
	}
	return c.Character.AnimationStartDelay(k)
}
