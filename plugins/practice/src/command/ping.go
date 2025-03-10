package command

import (
	"github.com/df-mc/dragonfly/server/cmd"
	"github.com/df-mc/dragonfly/server/world"
	"github.com/k4ties/dystopia/plugins/practice/src/user"
)

type PingEnum struct {
	onlyPlayer
	Player cmd.Optional[PlayerEnum] `cmd:"player"`
}

func (c PingEnum) Run(src cmd.Source, o *cmd.Output, _ *world.Tx) {
	if targetName, ok := c.Player.Load(); ok {
		if us, ok := user.P().UserByName(string(targetName)); ok {
			if p, ok := us.Player(); ok && us.Online() {
				systemMessage(o, "%s ping: <grey>%d</grey>", targetName, p.Latency().Milliseconds())
			}
		}
	} else {
		systemMessage(o, "Your ping: <grey>%d</grey>", p(src).Latency().Milliseconds())
	}
}

type Ping struct {
	onlyConsole
	Player string `cmd:"player"`
}

func (c Ping) Run(src cmd.Source, o *cmd.Output, tx *world.Tx) {
	us, ok := user.P().UserByName(c.Player)
	if !ok {
		o.Errorf("There's no user with name %s", c.Player)
		return
	}

	if _, ok := us.Player(); !ok || !us.Online() {
		o.Errorf("Player %s is not online", c.Player)
		return
	}

	p, _ := us.Player()
	systemMessage(o, "%s ping: <grey>%d</grey>", c.Player, p.Latency().Milliseconds())
}
