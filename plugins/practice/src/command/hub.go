package command

import (
	"github.com/df-mc/dragonfly/server/cmd"
	"github.com/df-mc/dragonfly/server/player"
	"github.com/df-mc/dragonfly/server/world"
	"github.com/k4ties/dystopia/plugins/practice/src/ffa"
	"github.com/k4ties/dystopia/plugins/practice/src/instance/lobby"
)

type Hub struct {
	onlyPlayer
}

func (Hub) Run(s cmd.Source, o *cmd.Output, _ *world.Tx) {
	if dead(s) {
		o.Errorf("Cannot teleport while you're dead")
		return
	}

	if pl, in := ffa.LookupPlayer(s.(*player.Player).UUID()); pl != nil && in != nil {
		if c, ok := in.Combat(pl.UUID()); ok {
			if c.Active() {
				o.Errorf("Cannot teleport while you're in combat")
				return
			}
		}
	}

	lobby.TransferWithRoutineSimple(p(s))
	systemMessage(o, "You've been teleported to the Lobby.")
}
