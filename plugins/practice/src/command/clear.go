package command

import (
	"github.com/df-mc/dragonfly/server/cmd"
	"github.com/df-mc/dragonfly/server/player"
	"github.com/df-mc/dragonfly/server/world"
	"github.com/k4ties/dystopia/plugins/practice/src/user"
)

type ClearEnum struct {
	onlyManager
	Target cmd.Optional[PlayerEnum] `cmd:"target"`
}

func (c ClearEnum) Run(src cmd.Source, o *cmd.Output, _ *world.Tx) {
	target, hasTarget := c.Target.Load()

	if !hasTarget {
		pl := p(src)

		pl.Inventory().Clear()
		pl.Armour().Clear()

		systemMessage(o, "You've successfully cleared your inventory.")
		return
	}

	usr, ok := user.P().UserByName(string(target))
	if !ok {
		o.Errorf("Can't find user %s", target)
		return
	}

	if _, ok := usr.Player(); !ok || !usr.Online() {
		o.Errorf("Can't find user %s", target)
		return
	}

	pl, _ := usr.Player()
	go pl.H().ExecWorld(func(tx *world.Tx, e world.Entity) {
		p := e.(*player.Player)
		p.Inventory().Clear()
		p.Armour().Clear()
	})

	systemMessage(o, "You've successfully cleared %s inventory.", target)
}

type Clear struct {
	onlyConsole
	Player string `cmd:"target"`
}

func (c Clear) Run(_ cmd.Source, o *cmd.Output, _ *world.Tx) {
	usr, ok := user.P().UserByName(c.Player)
	if !ok {
		o.Errorf("Can't find user %s", c.Player)
		return
	}

	if _, ok := usr.Player(); !ok || !usr.Online() {
		o.Errorf("Can't find user %s", c.Player)
		return
	}

	pl, _ := usr.Player()
	go pl.H().ExecWorld(func(tx *world.Tx, e world.Entity) {
		p := e.(*player.Player)
		p.Inventory().Clear()
		p.Armour().Clear()
	})

	systemMessage(o, "You've successfully cleared %s inventory.", c.Player)
}
