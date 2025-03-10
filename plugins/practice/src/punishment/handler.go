package punishment

import (
	"github.com/df-mc/dragonfly/server/event"
	"github.com/df-mc/dragonfly/server/player"
	"github.com/df-mc/dragonfly/server/session"
	"github.com/google/uuid"
	plugin "github.com/k4ties/df-plugin/df-plugin"
	"github.com/k4ties/dystopia/internal"
	user2 "github.com/k4ties/dystopia/plugins/practice/src/user"
	"github.com/sandertv/gophertunnel/minecraft/protocol/packet"
	"github.com/sandertv/gophertunnel/minecraft/text"
	"slices"
	"time"
)

type Handler struct {
	plugin.PlayerHandler
	p *Pool
}

func NewHandler(relay plugin.PlayerHandler) *Handler {
	return &Handler{PlayerHandler: relay, p: P()}
}

func (h *Handler) HandleLogin(ctx *event.Context[session.Conn]) {
	h.PlayerHandler.HandleLogin(ctx)

	if !ctx.Cancelled() {
		conn := ctx.Val()
		identity := conn.IdentityData()
		client := conn.ClientData()

		name := identity.DisplayName
		xuid := identity.XUID
		id, err := uuid.Parse(identity.Identity)
		if err != nil {
			return
		}
		if _, err := uuid.Parse(client.DeviceID); err != nil {
			if u, ok := user2.P().User(id); ok {
				_ = P().Ban(u, NilTime, "DID Spoof", "Anti Cheat")
			}
			return
		}

		dids := []string{client.DeviceID}
		ips := []string{internal.Ip(conn.RemoteAddr())}

		if u, ok := user2.P().User(id); ok {
			if ban, banned := h.p.Banned(u); banned {
				if !slices.Contains(ban.Punished().IPs, ips[0]) {
					ban.punished.IPs = append(ban.punished.IPs, ips[0])
				}
				if !slices.Contains(ban.Punished().DeviceIDs, dids[0]) {
					ban.punished.DeviceIDs = append(ban.punished.DeviceIDs, dids[0])
				}

				_ = h.p.update(ban)
				_ = conn.WritePacket(&packet.Disconnect{
					Message: BanReason(ban),
				})
				ctx.Cancel()
				return
			}
		}

		m, ok := h.p.modelExists(name, xuid, id, ips, dids, BanType)
		if ok {
			if ban, err := BanFromModel(m.Model); err == nil {
				if !slices.Contains(ban.Punished().IPs, ips[0]) {
					ban.punished.IPs = append(ban.punished.IPs, ips[0])
				}
				if !slices.Contains(ban.Punished().DeviceIDs, dids[0]) {
					ban.punished.DeviceIDs = append(ban.punished.DeviceIDs, dids[0])
				}

				_ = h.p.update(ban)
				_ = conn.WritePacket(&packet.Disconnect{
					Message: BanReason(ban),
				})
				ctx.Cancel()
				return
			}
		}
	}
}

func (h *Handler) HandleChat(ctx *player.Context, msg *string) {
	// on chat event user must be initialized by any way, so we will check for his instance
	cancelled := false
	user2.P().MustUser(ctx.Val(), func(u *user2.User) {
		if mute, muted := h.p.Muted(u); muted {
			cancelled = true
			ctx.Val().Messagef(text.Colourf(MuteNotifyFormat, internal.DurationString(time.Until(mute.Expiration())), mute.Reason()))
			ctx.Cancel()
			return
		}
	})
	if !cancelled {
		h.PlayerHandler.HandleChat(ctx, msg)
	}
}
