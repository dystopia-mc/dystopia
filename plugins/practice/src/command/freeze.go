package command

import (
	"github.com/df-mc/dragonfly/server/cmd"
	"github.com/df-mc/dragonfly/server/world"
	"github.com/k4ties/dystopia/plugins/practice/src/instance"
	"github.com/k4ties/dystopia/plugins/practice/src/user"
)

type FreezeEnum struct {
	onlyStaff
	Target PlayerEnum `cmd:"target"`
}

func (f FreezeEnum) Run(src cmd.Source, o *cmd.Output, tx *world.Tx) {
	u, ok := user.P().UserByName(string(f.Target))
	if !ok {
		o.Errorf("User not found: %s", f.Target)
		return
	}
	if !isConsole(src) {
		if self, ok := usr(src); ok {
			if u == self {
				o.Errorf("Cannot freeze yourself.")
				return
			}
			if self.Data().Rank().Priority <= u.Data().Rank().Priority {
				o.Errorf("Can only freeze players with lower rank.")
				return
			}
		}
	}
	p, ok := u.Player()
	if !ok || !u.Online() {
		o.Errorf("User %s is not online", f.Target)
		return
	}
	pl, in := instance.LookupPlayer(p.UUID())
	if pl == nil || in == nil {
		o.Errorf("Player %s not found. Please contact our support.", f.Target)
		return
	}

	pl.Freeze(src.(cmd.NamedTarget).Name())
	systemMessage(o, "You've successfully frozen %s", f.Target)
}

type Freeze struct {
	onlyConsole
	Target string `cmd:"target"`
}

func (f Freeze) Run(src cmd.Source, o *cmd.Output, tx *world.Tx) {
	FreezeEnum{Target: PlayerEnum(f.Target)}.Run(src, o, tx)
}

type UnFreezeEnum struct {
	onlyStaff
	Target PlayerEnum `cmd:"target"`
}

func (c UnFreezeEnum) Run(src cmd.Source, o *cmd.Output, _ *world.Tx) {
	u, ok := user.P().UserByName(string(c.Target))
	if !ok {
		o.Errorf("User not found: %s", c.Target)
		return
	}
	if !isConsole(src) {
		if self, ok := usr(src); ok {
			if u == self {
				o.Errorf("Cannot freeze yourself.")
				return
			}
			if self.Data().Rank().Priority <= u.Data().Rank().Priority {
				o.Errorf("Can only freeze players with lower rank.")
				return
			}
		}
	}
	p, ok := u.Player()
	if !ok || !u.Online() {
		o.Errorf("User %s is not online", c.Target)
		return
	}
	pl, in := instance.LookupPlayer(p.UUID())
	if pl == nil || in == nil {
		o.Errorf("Player %s not found. Please contact our support.", c.Target)
		return
	}

	pl.UnFreeze(src.(cmd.NamedTarget).Name())
	systemMessage(o, "You've successfully unfrozen %s", c.Target)
}

type UnFreeze struct {
	onlyConsole
	Target string `cmd:"target"`
}

func (u UnFreeze) Run(src cmd.Source, o *cmd.Output, tx *world.Tx) {
	UnFreezeEnum{Target: PlayerEnum(u.Target)}.Run(src, o, tx)
}
