package command

import (
	"github.com/df-mc/dragonfly/server/cmd"
	"github.com/df-mc/dragonfly/server/player"
	"github.com/df-mc/dragonfly/server/world"
	"github.com/k4ties/dystopia/plugins/practice/src/instance"
	"github.com/k4ties/dystopia/plugins/practice/src/user"
)

type Instance string

func (Instance) Type() string {
	return "ffa"
}

func (Instance) Options(cmd.Source) []string {
	return instance.AllInstancesNames()
}

type TransferEnum struct {
	onlyAdmin
	Target   PlayerEnum `cmd:"target"`
	Instance Instance   `cmd:"instance"`
}

func (f TransferEnum) Run(_ cmd.Source, o *cmd.Output, _ *world.Tx) {
	usr, ok := user.P().UserByName(string(f.Target))
	if !ok {
		o.Errorf("User not found: %s", f.Target)
		return
	}

	if _, ok := usr.Player(); !ok || !usr.Online() {
		o.Errorf("User is not online: %s", f.Target)
		return
	}

	p, _ := usr.Player()

	pl, in := instance.LookupPlayer(p.UUID())
	if pl == nil || in == nil {
		o.Errorf("Player not found: %s", f.Target)
		return
	}

	in, ok = instance.ByName(string(f.Instance))
	if !ok {
		o.Errorf("Instance not found: %s", f.Instance)
		return
	}

	pl.ExecSafe(func(p *player.Player, tx *world.Tx) {
		in.Transfer(pl, tx)
	})

	systemMessage(o, "You've successfully transferred <grey>%s</grey> to %s", f.Target, f.Instance)
}

type Transfer struct {
	onlyConsole
	Target   string   `cmd:"target"`
	Instance Instance `cmd:"instance"`
}

func (f Transfer) Run(src cmd.Source, o *cmd.Output, tx *world.Tx) {
	TransferEnum{Target: PlayerEnum(f.Target), Instance: f.Instance}.Run(src, o, tx)
}
