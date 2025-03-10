package command

import (
	"github.com/df-mc/dragonfly/server/cmd"
	"github.com/df-mc/dragonfly/server/world"
	"github.com/k4ties/dystopia/internal/restarter"
)

type Restart struct {
	onlyOwnerAndConsole
}

func (r Restart) Run(_ cmd.Source, o *cmd.Output, _ *world.Tx) {
	if err := restarter.Restart(); err != nil {
		o.Error(err)
		return
	}

	systemMessage(o, "You've successfully restarted server.")
}
