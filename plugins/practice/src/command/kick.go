package command

import (
	"github.com/df-mc/dragonfly/server/cmd"
	"github.com/df-mc/dragonfly/server/world"
	"github.com/k4ties/dystopia/plugins/practice/src/punishment"
	"github.com/k4ties/dystopia/plugins/practice/src/user"
)

type KickEnum struct {
	onlyStaff
	Target PlayerEnum  `cmd:"target"`
	Reason cmd.Varargs `cmd:"reason"`
}

func (k KickEnum) Run(src cmd.Source, o *cmd.Output, tx *world.Tx) {
	u, ok := user.P().UserByName(string(k.Target))
	if !ok {
		o.Errorf("User not found: %s", k.Target)
		return
	}
	if !isConsole(src) {
		if self, ok := usr(src); ok {
			if u == self {
				o.Errorf("Cannot kick yourself.")
				return
			}
			if self.Data().Rank().Priority <= u.Data().Rank().Priority {
				o.Errorf("Can only kick players with lower rank.")
				return
			}
		}
	}

	if err := punishment.P().Kick(u, string(k.Reason), src.(cmd.NamedTarget).Name()); err != nil {
		o.Errorf("Unable to kick player %s: %s", k.Target, err)
		return
	}

	systemMessage(o, "Successfully kicked <grey>%s</grey> for reason <grey>%s</grey>", u.Data().Name(), k.Reason)
}

type Kick struct {
	onlyConsole
	Target string      `cmd:"target"`
	Reason cmd.Varargs `cmd:"reason"`
}

func (k Kick) Run(src cmd.Source, o *cmd.Output, tx *world.Tx) {
	KickEnum{Target: PlayerEnum(k.Target), Reason: k.Reason}.Run(src, o, tx)
}
