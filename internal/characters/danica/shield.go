package danica

import (
	"github.com/genshinsim/gcsim/pkg/core/attributes"
	"github.com/genshinsim/gcsim/pkg/core/player/shield"
)

type danicaShield struct{ *shield.Tmpl }

const shieldDur = 25 * 60

func (c *char) newShield() {
	n := &danicaShield{}
	n.Tmpl = &shield.Tmpl{
		ActorIndex: c.Index(),
		Target:     -1,
		Name:       "Danica Skill",
		Src:        c.Core.F,
		ShieldType: shield.DanicaSkill,
		Ele:        attributes.Electro,
		HP:         c.TotalAtk()*skillShieldAtk[c.TalentLvlSkill()] + skillShieldFlat[c.TalentLvlSkill()],
		Expires:    c.Core.F + shieldDur,
	}
	c.Core.Player.Shields.Add(n)
}
