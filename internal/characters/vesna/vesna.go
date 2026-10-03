package vesna

import (
	tmpl "github.com/genshinsim/gcsim/internal/template/character"
	"github.com/genshinsim/gcsim/pkg/core"
	"github.com/genshinsim/gcsim/pkg/core/action"
	"github.com/genshinsim/gcsim/pkg/core/info"
	"github.com/genshinsim/gcsim/pkg/core/player/character"
	"github.com/genshinsim/gcsim/pkg/reactable"
)

type char struct {
	*tmpl.Character
	skillStacks      int
	skillNACount     int
	skillLvl         int
	skillsMaxLvlUsed int
	a1Stacks         RingQueue[int]
}

const stellarSwirlText = " (Stellar Swirl)"

func NewChar(s *core.Core, w *character.CharWrapper, _ info.CharacterProfile) error {
	c := char{}
	c.Character = tmpl.NewWithWrapper(s, w)

	c.EnergyMax = 60
	c.BurstCon = 5
	c.SkillCon = 3
	c.NormalHitNum = normalHitNum

	w.Character = &c

	c.Moonsign = 1

	return nil
}

func (c *char) Init() error {
	c.skillInit()
	c.a1Init()
	c.a4Init()
	c.stellarInit()
	c.c1Init()
	c.c2Init()
	c.c6Init()
	return nil
}

func (c *char) AnimationStartDelay(k info.AnimationDelayKey) int {
	if k == info.AnimationXingqiuN0StartDelay {
		return 9
	}
	return c.Character.AnimationStartDelay(k)
}

func (c *char) ActionReady(a action.Action, p map[string]int) (bool, action.Failure) {
	if a == action.ActionSkill && c.StatusIsActive(c6Key) {
		return true, action.NoFailure
	}

	// check if it is possible to use next skill
	if c.StatusIsActive(skillKey) && a == action.ActionSkill {
		if c.skillStacks <= 0 {
			return false, action.SkillCD
		}
		return true, action.NoFailure
	}

	return c.Character.ActionReady(a, p)
}

func (c *char) Condition(fields []string) (any, error) {
	switch fields[0] {
	case "sword-essence":
		if c.StatusIsActive(skillKey) {
			return c.skillStacks, nil
		}
		return 0, nil
	default:
		return c.Character.Condition(fields)
	}
}

func (c *char) isRadianceSSw() bool {
	return c.StatusIsActive(reactable.SswKey)
}
