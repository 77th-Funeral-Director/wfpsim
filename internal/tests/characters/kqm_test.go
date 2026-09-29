package characters

import (
	"context"
	"fmt"
	"math"
	"strings"
	"testing"

	"github.com/genshinsim/gcsim/pkg/core/info"
	"github.com/genshinsim/gcsim/pkg/gcs/ast"
	"github.com/genshinsim/gcsim/pkg/simulator"
)

// These teams exercise the restored characters with the current reaction engine.
func TestKQMTeams(t *testing.T) {
	teams := [][][2]string{
		{{"vodyanitsa", "dullblade"}, {"sandrone", "wastergreatsword"}, {"vesna", "dullblade"}, {"kaeya", "dullblade"}},
		{{"zibai", "dullblade"}, {"linnea", "huntersbow"}, {"xingqiu", "dullblade"}, {"fischl", "huntersbow"}},
		{{"alyosha", "beginnersprotector"}, {"lohen", "beginnersprotector"}, {"diona", "huntersbow"}, {"klee", "apprenticesnotes"}},
		{{"albedo", "dullblade"}, {"venti", "huntersbow"}, {"klee", "apprenticesnotes"}, {"diona", "huntersbow"}},
	}
	for _, team := range teams {
		for _, cons := range []int{0, 6} {
			t.Run(fmt.Sprintf("%s-C%d", team[0][0], cons), func(t *testing.T) {
				var cfg strings.Builder
				cfg.WriteString("options iteration=2 workers=1 duration=45;\ntarget lvl=100 resist=0.1 radius=2 pos=0,2;\nenergy every interval=60,61 amount=100;\n")
				for _, char := range team {
					fmt.Fprintf(&cfg, "%s char lvl=90/90 cons=%d talent=9,9,9;\n%s add weapon=%q refine=1 lvl=90/90;\n%s add stats hp=30000 atk=1000 def=1000 er=3 em=300 cr=0.5 cd=1;\n", char[0], cons, char[0], char[1], char[0])
				}
				fmt.Fprintf(&cfg, "%s add stats stellar-swirl%%=0.25 lunarcrystallize%%=0.2;\nactive %s;\nfor let i=0; i<3; i=i+1 {\n", team[0][0], team[0][0])
				for _, char := range team {
					fmt.Fprintf(&cfg, "%s skill, burst, attack:5, dash;\n", char[0])
				}
				cfg.WriteString("}\n")
				file := ast.NewFile()
				parsed, program, err := simulator.Parse(file, cfg.String())
				if err != nil {
					t.Fatal(err)
				}
				if parsed.Characters[0].ReactBonus[info.ReactionTypeStellarSwirl] != 0.25 {
					t.Fatal("reaction bonus was not preserved")
				}
				result, err := simulator.RunWithSeededConfig(context.Background(), file, cfg.String(), parsed, program, []int64{1, 2})
				if err != nil {
					t.Fatal(err)
				}
				dps := result.GetStatistics().GetDPS().GetMean()
				if dps <= 0 || math.IsNaN(dps) || math.IsInf(dps, 0) {
					t.Fatalf("invalid DPS %v", dps)
				}
			})
		}
	}
}
