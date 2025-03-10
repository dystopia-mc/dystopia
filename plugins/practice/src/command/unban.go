package command

import (
	"github.com/df-mc/dragonfly/server/cmd"
	"github.com/df-mc/dragonfly/server/world"
	"github.com/k4ties/dystopia/plugins/practice/src/punishment"
	"github.com/k4ties/dystopia/plugins/practice/src/user"
)

type UnbanEnum struct {
	onlyStaff
	Target PlayerEnum `cmd:"target"`
}

func (c UnbanEnum) Run(src cmd.Source, o *cmd.Output, _ *world.Tx) {
	u, ok := user.P().UserByName(string(c.Target))
	if !ok {
		o.Errorf("User not found: %s", c.Target)
		return
	}
	if !isConsole(src) {
		if self, ok := usr(src); ok {
			if u == self {
				o.Errorf("Cannot unban yourself.")
				return
			}
			if self.Data().Rank().Priority <= u.Data().Rank().Priority {
				o.Errorf("Can only unban players with lower rank.")
				return
			}
		}
	}

	pool := punishment.P()
	if err := pool.Unban(u, src.(cmd.NamedTarget).Name()); err != nil {
		o.Errorf("Couldn't unban: %s", err.Error())
		return
	}

	systemMessage(o, "Successfully unbanned %s", c.Target)
}

type Unban struct {
	onlyStaffAndConsole
	Target string `cmd:"target"`
}

func (u Unban) Run(src cmd.Source, o *cmd.Output, tx *world.Tx) {
	UnbanEnum{Target: PlayerEnum(u.Target)}.Run(src, o, tx)
}
