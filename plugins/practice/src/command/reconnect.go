package command

import (
	"github.com/df-mc/dragonfly/server/cmd"
	"github.com/df-mc/dragonfly/server/world"
)

type ReConnect struct {
	onlyPlayer
}

func (c ReConnect) Run(s cmd.Source, o *cmd.Output, tx *world.Tx) {
	if err := p(s).Transfer(p(s).Data().Session.ClientData().ServerAddress); err != nil {
		o.Errorf("Failed to reconnect: %v", err)
	}
}
