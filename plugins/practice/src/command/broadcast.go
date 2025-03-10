package command

import (
	"github.com/df-mc/dragonfly/server/cmd"
	"github.com/df-mc/dragonfly/server/world"
	"github.com/k4ties/dystopia/plugins/practice/src/user/chatter"
)

type Broadcast struct {
	onlyAdminAndConsole
	Text cmd.Varargs `cmd:"text"`
}

func (b Broadcast) Run(s cmd.Source, o *cmd.Output, _ *world.Tx) {
	name := s.(cmd.NamedTarget).Name()
	chatter.Messagef("<purple>[%s]: %s</purple>", name, string(b.Text))
	systemMessage(o, "You've successfully sent message.")
}
