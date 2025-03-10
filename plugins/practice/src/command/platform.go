package command

import (
	"github.com/df-mc/dragonfly/server/cmd"
	"github.com/df-mc/dragonfly/server/world"
	user2 "github.com/k4ties/dystopia/plugins/practice/src/user"
)

type Platform struct {
	onlyPlayer
	Target cmd.Optional[PlayerEnum] `cmd:"target"`
}

func (c Platform) Run(src cmd.Source, o *cmd.Output, _ *world.Tx) {
	target, hasTarget := c.Target.Load()
	if !hasTarget {
		// get self input mode and os
		u(src, func(u *user2.User) {
			systemMessage(o, "Your platform: <grey>%s</grey> | <grey>%s</grey>", u.OS().String(), u.InputMode().String())
		})
		return
	}

	usr, ok := user2.P().UserByName(string(target))
	if !ok {
		o.Errorf("Couldn't find user: %s", target)
		return
	}

	systemMessage(o, "%s platform: <grey>%s</grey> | <grey>%s</grey>", usr.Data().Name(), usr.OS().String(), usr.InputMode().String())
}

type PlatformConsole struct {
	onlyConsole
	Target string `cmd:"target"`
}

func (p PlatformConsole) Run(_ cmd.Source, o *cmd.Output, _ *world.Tx) {
	usr, ok := user2.P().UserByName(p.Target)
	if !ok {
		o.Errorf("Couldn't find user: %s", p.Target)
		return
	}

	systemMessage(o, "%s platform: <grey>%s</grey> | <grey>%s</grey>", usr.Data().Name(), usr.OS().String(), usr.InputMode().String())
}
