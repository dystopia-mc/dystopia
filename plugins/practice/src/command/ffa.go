package command

import (
	"github.com/df-mc/dragonfly/server/cmd"
	"github.com/df-mc/dragonfly/server/world"
	"github.com/k4ties/dystopia/plugins/practice/src/ffa"
	"github.com/k4ties/dystopia/plugins/practice/src/instance/lobby"
)

type FFA struct {
	onlyPlayer
}

func (c FFA) Run(s cmd.Source, o *cmd.Output, tx *world.Tx) {
	if !lobby.Instance().Active(inPl(s).UUID()) || dead(s) {
		o.Errorf("Can only teleport to FFA in lobby")
		return
	}

	p(s).SendForm(ffa.NewForm())
}
