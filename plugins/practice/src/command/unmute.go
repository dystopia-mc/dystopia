package command

import (
	"github.com/df-mc/dragonfly/server/cmd"
	"github.com/df-mc/dragonfly/server/world"
	"github.com/k4ties/dystopia/plugins/practice/src/punishment"
	"github.com/k4ties/dystopia/plugins/practice/src/user"
)

type UnmuteEnum struct {
	onlyStaff
	Target PlayerEnum `cmd:"target"`
}

func (c UnmuteEnum) Run(src cmd.Source, o *cmd.Output, _ *world.Tx) {
	u, ok := user.P().UserByName(string(c.Target))
	if !ok {
		o.Errorf("User not found: %s", c.Target)
		return
	}
	if !isConsole(src) {
		if self, ok := usr(src); ok {
			if u == self {
				o.Errorf("Cannot unmute yourself.")
				return
			}
			if self.Data().Rank().Priority <= u.Data().Rank().Priority {
				o.Errorf("Can only unmute players with lower rank.")
				return
			}
		}
	}

	pool := punishment.P()
	if err := pool.Unmute(u, src.(cmd.NamedTarget).Name()); err != nil {
		o.Errorf("Couldn't unmute: %s", err.Error())
		return
	}

	systemMessage(o, "Successfully unmuted %s", c.Target)
}

type Unmute struct {
	onlyStaffAndConsole
	Target string `cmd:"target"`
}

func (u Unmute) Run(src cmd.Source, o *cmd.Output, tx *world.Tx) {
	UnmuteEnum{Target: PlayerEnum(u.Target)}.Run(src, o, tx)
}
