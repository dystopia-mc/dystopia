package command

import (
	"github.com/df-mc/dragonfly/server/cmd"
	"github.com/df-mc/dragonfly/server/world"
	"github.com/k4ties/dystopia/plugins/practice/src/user/chatter"
	"github.com/sandertv/gophertunnel/minecraft/text"
)

type Chat struct {
	onlyAdminAndConsole
	Toggle cmd.SubCommand `cmd:"toggle"`
}

func (c Chat) Run(s cmd.Source, o *cmd.Output, _ *world.Tx) {
	chatter.ToggleChat()
	now := chatter.ChatEnabled()
	name := s.(cmd.NamedTarget).Name()

	chatter.Messagef("<red><b>>></b></red> Player <grey>%s</grey> %s the chat", name, formatBool(now))
	o.Printf(text.Green + "You've successfully toggled chat")
}

func formatBool(b bool) string {
	switch b {
	case true:
		return "opened"
	default:
		return "closed"
	}
}
