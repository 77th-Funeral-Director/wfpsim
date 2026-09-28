package mitya

import (
	"github.com/genshinsim/gcsim/internal/frames"
	"github.com/genshinsim/gcsim/pkg/core/action"
	"github.com/genshinsim/gcsim/pkg/core/attacks"
	"github.com/genshinsim/gcsim/pkg/core/attributes"
	"github.com/genshinsim/gcsim/pkg/core/combat"
	"github.com/genshinsim/gcsim/pkg/core/info"
)

var (
	burstFrames     []int
	burstPrismDelay = []int{119, 119, 119}
)

const (
	hitmark = 120
)

func init() {
	burstFrames = frames.InitAbilSlice(115)
}

func (c *char) Burst(p map[string]int) (action.Info, error) {
	c.QueueCharTask(func() {
		ai := info.AttackInfo{
			ActorIndex: c.Index(),
			Abil:       "Burst",
			AttackTag:  attacks.AttackTagElementalBurst,
			ICDTag:     attacks.ICDTagNone,
			ICDGroup:   attacks.ICDGroupDefault,
			StrikeType: attacks.StrikeTypeDefault,
			Element:    attributes.Electro,
			Durability: 25,
			Mult:       burst[c.TalentLvlBurst()],
			FlatDmg:    burstEM[c.TalentLvlBurst()] * c.Stat(attributes.EM),
		}
		ap := combat.NewCircleHitOnTarget(c.Core.Combat.Player(), nil, 6)
		c.Core.QueueAttack(ai, ap, 0, 0)
	}, hitmark)

	for i := range 3 {
		c.QueueCharTask(c.addPrism, burstPrismDelay[i])
	}

	c.SetCD(action.ActionBurst, 15*60)

	c.ConsumeEnergy(4)

	return action.Info{
		Frames:          frames.NewAbilFunc(burstFrames),
		AnimationLength: burstFrames[action.InvalidAction],
		CanQueueAfter:   burstFrames[action.ActionDash], // earliest cancel
		State:           action.BurstState,
	}, nil
}
