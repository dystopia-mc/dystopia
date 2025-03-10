package lcn

import (
	"github.com/df-mc/dragonfly/server/cmd"
	"github.com/df-mc/dragonfly/server/player"
	"github.com/df-mc/dragonfly/server/world"
	user2 "github.com/k4ties/dystopia/plugins/practice/src/user"
	"github.com/sandertv/gophertunnel/minecraft/text"
	"slices"
	"strings"
)

var members = []string{
	"n0ni0n",
	"Sq1resEngine",
	"Widee1095",
	"kqnepphvh",
	"vanechek2012",
	"XxSanifeXx",
	"mo6sv3t",
	"eatingaciid",
	"rekitstyle",
	"zenon3071",
	"xCookyta",
	"XrenReki",
	"omgcxtnz",
	"SmokyFox11",
	"kxrsqe",
	"K0taG0q1",
	"RepaidFLEXER",
}

func init() {
	// lowercase names
	for i, m := range members {
		members[i] = strings.ToLower(m)
	}
}

type command struct{}

func (command) Run(s cmd.Source, o *cmd.Output, _ *world.Tx) {
	if !slices.Contains(members, strings.ToLower(p(s).Name())) {
		o.Errorf("You're not a clan member.")
		return
	}

	user2.MustUsePool(func(pl *user2.Pool) {
		pl.MustUser(p(s), func(u *user2.User) {
			d := u.Data()
			if !d.HasTag() || d.HasTag() && d.Tag() != "LCN" {
				u.Data().SetTag(text.Colourf("<orange><b>LCN</b></orange>"))
				o.Printf(text.Green + "You're successfully updated tag")
				return
			}

			if d.HasTag() && d.Tag() == "LCN" {
				u.Data().SetTag("")
				o.Printf(text.Green + "You're successfully disabled tag")
			}
		})
	})
}

func (command) Allow(s cmd.Source) bool {
	_, ok := s.(*player.Player)
	return ok
}

func p(s cmd.Source) *player.Player {
	return s.(*player.Player)
}
