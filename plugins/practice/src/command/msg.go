package command

import (
	"github.com/df-mc/dragonfly/server/cmd"
	"github.com/df-mc/dragonfly/server/world"
	"github.com/k4ties/dystopia/plugins/practice/src/punishment"
	"github.com/k4ties/dystopia/plugins/practice/src/user"
	"github.com/sandertv/gophertunnel/minecraft/text"
)

type Msg struct {
	onlyPlayer
	Player PlayerEnum  `cmd:"target"`
	Text   cmd.Varargs `cmd:"text"`
}

const WhisperFormat = "<red>%s whispers to you:</red> %s"

func (m Msg) Run(s cmd.Source, o *cmd.Output, _ *world.Tx) {
	u, ok := user.P().UserByName(string(m.Player))
	if !ok {
		o.Errorf("User not found: %s", m.Player)
		return
	}
	if !u.Settings().Global.DirectMessages {
		o.Errorf("Player %s disabled the messages in the settings.", m.Player)
		return
	}
	if !u.Online() {
		o.Errorf("User is not online: %s", m.Player)
		return
	}
	pl, ok := u.Player()
	if !ok {
		o.Errorf("Cannot get user: %s. Please contact our support", u.Name())
		return
	}
	if pl.UUID() == p(s).UUID() {
		o.Errorf("Cannot text to yourself")
		return
	}
	if usr, ok := usr(s); ok {
		if _, muted := punishment.P().Muted(usr); muted {
			o.Errorf("Cannot text while you're muted.")
			return
		}
	}

	pl.Message(text.Colourf(WhisperFormat, p(s).Name(), m.Text))
	systemMessage(o, "Successfully sent message to <grey>%s</grey>", m.Player)
}
