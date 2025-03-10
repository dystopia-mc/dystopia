package handlers

import (
	"github.com/df-mc/dragonfly/server/event"
	"github.com/df-mc/dragonfly/server/session"
	plugin "github.com/k4ties/df-plugin/df-plugin"
	"github.com/k4ties/dystopia/plugins/practice/src/handlers/whitelist"
	"github.com/k4ties/dystopia/plugins/practice/src/instance"
	"github.com/sandertv/gophertunnel/minecraft/protocol/packet"
	"strings"
)

type Login struct {
	plugin.NopPlayerHandler
}

func NewLoginHandler() *Login {
	l := &Login{}

	return l
}

func (l *Login) HandleLogin(ctx *event.Context[session.Conn]) {
	if ctx.Val().ClientData().CurrentInputMode == packet.InputModeTouch {
		_ = ctx.Val().WritePacket(&packet.Disconnect{
			Message: instance.KickMessage(instance.ErrorMobilizovan),
		})
		ctx.Cancel()
		return
	}

	if whitelist.Enabled() {
		name := strings.ToLower(ctx.Val().IdentityData().DisplayName)

		if !whitelist.Whitelisted(name) {
			_ = ctx.Val().WritePacket(&packet.Disconnect{
				Message: "Server is whitelisted.",
			})
			ctx.Cancel()
			return
		}
	}
}
