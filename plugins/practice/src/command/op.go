package command

import (
	"github.com/df-mc/dragonfly/server/cmd"
	"github.com/df-mc/dragonfly/server/world"
	rank2 "github.com/k4ties/dystopia/plugins/practice/src/rank"
	"github.com/k4ties/dystopia/plugins/practice/src/user"
	"github.com/sandertv/gophertunnel/minecraft/text"
)

type Op struct {
	onlyOwnerAndConsole
	Target string
}

func (c Op) Run(_ cmd.Source, o *cmd.Output, _ *world.Tx) {
	u, ok := user.P().UserByName(c.Target)
	if !ok {
		o.Errorf("User not found: %s", c.Target)
		return
	}

	role := rank2.Owner

	if u.Data().Rank().Priority >= rank2.PriorityOwner {
		role = rank2.Player
		role.Name = "Player"
		role.Format = text.Grey
	}

	u.UpdateRank(role)
	o.Printf("You've successfully changed rank of %s to %s", u.Name(), role.Format+role.Name+text.Reset)
}
