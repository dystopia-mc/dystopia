package command

import (
	"github.com/df-mc/dragonfly/server/block/cube"
	"github.com/df-mc/dragonfly/server/cmd"
	"github.com/df-mc/dragonfly/server/world"
	"github.com/go-gl/mathgl/mgl64"
	"github.com/k4ties/dystopia/plugins/practice/src/user"
)

type TeleportToPos struct {
	onlyAdmin
	Position mgl64.Vec3 `cmd:"destination"`
}

func (c TeleportToPos) Run(src cmd.Source, o *cmd.Output, tx *world.Tx) {
	p(src).Teleport(c.Position)
	systemMessage(o, "You've been teleported to <grey>%s</grey>", cube.PosFromVec3(c.Position).String())
}

type TeleportToTarget struct {
	onlyAdmin
	Target PlayerEnum `cmd:"destination"`
}

func (c TeleportToTarget) Run(src cmd.Source, o *cmd.Output, tx *world.Tx) {
	usr, ok := user.P().UserByName(string(c.Target))
	if !ok {
		o.Errorf("Cannot find player: %s", c.Target)
		return
	}
	if _, ok := usr.Player(); !ok || !usr.Online() {
		o.Errorf("Cannot find player: %s", c.Target)
		return
	}

	pl, _ := usr.Player()
	handle := tx.RemoveEntity(p(src))

	go pl.H().ExecWorld(func(tx *world.Tx, e world.Entity) {
		tx.AddEntity(handle)
	})

	p(src).Teleport(pl.Position())
	systemMessage(o, "You've been teleported to <grey>%s</grey>", c.Target)
}
