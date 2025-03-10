package trademc

import (
	"github.com/df-mc/dragonfly/server/cmd"
	"github.com/df-mc/dragonfly/server/world"
	"github.com/go-gl/mathgl/mgl64"
	"github.com/gorcon/rcon/rcontest"
	"github.com/k4ties/dystopia/plugins/practice/src/console"
	"github.com/sandertv/gophertunnel/minecraft/text"
)

func (r *RCON) handleMessage(ctx *rcontest.Context) {
	msg := ctx.Request().Body()
	r.requestConsoleCommand(msg, ctx)
}

func (r *RCON) requestConsoleCommand(text string, c *rcontest.Context) {
	console.ProceedCommand(text, &source{
		r: r,
		c: c,
	})
}

type source struct {
	r *RCON
	c *rcontest.Context
}

func (c *source) IsConsole() {}

func (c *source) Position() mgl64.Vec3 {
	return mgl64.Vec3{}
}

func (c *source) World() *world.World {
	return nil
}

func (c *source) SendCommandOutput(o *cmd.Output) {
	for _, m := range o.Messages() {
		c.r.responseMessage(text.Clean(m.String()))(c.c)
	}
	for _, e := range o.Errors() {
		c.r.responseMessage(text.Clean(e.Error()))(c.c)
	}
}

func (c *source) Name() string {
	return "RCON"
}
