package user

import (
	"github.com/df-mc/dragonfly/server/block/cube"
	"github.com/df-mc/dragonfly/server/cmd"
	"github.com/df-mc/dragonfly/server/event"
	"github.com/df-mc/dragonfly/server/item"
	"github.com/df-mc/dragonfly/server/player"
	"github.com/df-mc/dragonfly/server/player/skin"
	"github.com/df-mc/dragonfly/server/session"
	"github.com/df-mc/dragonfly/server/world"
	"github.com/go-gl/mathgl/mgl64"
	"github.com/google/uuid"
	plugin "github.com/k4ties/df-plugin/df-plugin"
	"github.com/k4ties/dystopia/internal"
	"github.com/k4ties/dystopia/plugins/practice/src/instance"
	"github.com/k4ties/dystopia/plugins/practice/src/rank"
	"github.com/k4ties/dystopia/plugins/practice/src/user/chatter"
	"github.com/sandertv/gophertunnel/minecraft/protocol/packet"
	"github.com/sandertv/gophertunnel/minecraft/text"
	"net"
	"slices"
	"time"
)

// Handler is an implementation of plugin.PlayerHandler
type Handler struct {
	l     login
	relay plugin.PlayerHandler

	pr practiceHandler
}

var DisabledDiagnostics = session.Diagnostics{
	AverageFramesPerSecond:        -1,
	AverageServerSimTickTime:      -1,
	AverageClientSimTickTime:      -1,
	AverageBeginFrameTime:         -1,
	AverageInputTime:              -1,
	AverageRenderTime:             -1,
	AverageEndFrameTime:           -1,
	AverageRemainderTimePercent:   -1,
	AverageUnaccountedTimePercent: -1,
}

func NewHandler(loginHandler login, relay plugin.PlayerHandler) *Handler {
	return &Handler{relay: relay, l: loginHandler, pr: relay.(practiceHandler)}
}

func (h *Handler) HandleDeath(p *player.Player, src world.DamageSource, keepInv *bool) {
	h.pr.RelayedHandleDeath(p, src, keepInv)
}

func (h *Handler) HandleClientPacket(ctx *player.Context, pk packet.Packet) {
	h.relay.HandleClientPacket(ctx, pk)
}

func (h *Handler) HandleServerPacket(ctx *player.Context, pk packet.Packet) {
	h.relay.HandleServerPacket(ctx, pk)
}

func (h *Handler) HandleMove(ctx *player.Context, newPos mgl64.Vec3, newRot cube.Rotation) {
	h.relay.HandleMove(ctx, newPos, newRot)
}

func (h *Handler) HandleJump(p *player.Player) {
	h.relay.HandleJump(p)
}

func (h *Handler) HandleTeleport(ctx *player.Context, pos mgl64.Vec3) {
	h.relay.HandleTeleport(ctx, pos)
}

func (h *Handler) HandleChangeWorld(p *player.Player, before, after *world.World) {
	h.relay.HandleChangeWorld(p, before, after)
}

func (h *Handler) HandleToggleSprint(ctx *player.Context, after bool) {
	h.relay.HandleToggleSprint(ctx, after)
}

func (h *Handler) HandleToggleSneak(ctx *player.Context, after bool) {
	h.relay.HandleToggleSneak(ctx, after)
}

func (h *Handler) HandleChat(ctx *player.Context, message *string) {
	h.relay.HandleChat(ctx, message)
}

func (h *Handler) HandleFoodLoss(ctx *player.Context, from int, to *int) {
	h.relay.HandleFoodLoss(ctx, from, to)
}

func (h *Handler) HandleHeal(ctx *player.Context, health *float64, src world.HealingSource) {
	h.relay.HandleHeal(ctx, health, src)
}

func (h *Handler) HandleHurt(ctx *player.Context, damage *float64, immune bool, attackImmunity *time.Duration, src world.DamageSource) {
	h.relay.HandleHurt(ctx, damage, immune, attackImmunity, src)
}

func (h *Handler) HandleRespawn(p *player.Player, pos *mgl64.Vec3, w **world.World) {
	h.relay.HandleRespawn(p, pos, w)
}

func (h *Handler) HandleSkinChange(ctx *player.Context, skin *skin.Skin) {
	h.relay.HandleSkinChange(ctx, skin)
}

func (h *Handler) HandleFireExtinguish(ctx *player.Context, pos cube.Pos) {
	h.relay.HandleFireExtinguish(ctx, pos)
}

func (h *Handler) HandleStartBreak(ctx *player.Context, pos cube.Pos) {
	h.relay.HandleStartBreak(ctx, pos)
}

func (h *Handler) HandleBlockBreak(ctx *player.Context, pos cube.Pos, drops *[]item.Stack, xp *int) {
	h.relay.HandleBlockBreak(ctx, pos, drops, xp)
}

func (h *Handler) HandleBlockPlace(ctx *player.Context, pos cube.Pos, b world.Block) {
	h.relay.HandleBlockPlace(ctx, pos, b)
}

func (h *Handler) HandleBlockPick(ctx *player.Context, pos cube.Pos, b world.Block) {
	h.relay.HandleBlockPick(ctx, pos, b)
}

func (h *Handler) HandleItemUse(ctx *player.Context) {
	h.relay.HandleItemUse(ctx)
}

func (h *Handler) HandleItemUseOnBlock(ctx *player.Context, pos cube.Pos, face cube.Face, clickPos mgl64.Vec3) {
	h.relay.HandleItemUseOnBlock(ctx, pos, face, clickPos)
}

func (h *Handler) HandleItemUseOnEntity(ctx *player.Context, e world.Entity) {
	h.relay.HandleItemUseOnEntity(ctx, e)
}

func (h *Handler) HandleItemRelease(ctx *player.Context, item item.Stack, dur time.Duration) {
	h.relay.HandleItemRelease(ctx, item, dur)
}

func (h *Handler) HandleItemConsume(ctx *player.Context, item item.Stack) {
	h.relay.HandleItemConsume(ctx, item)
}

func sendActionBar(p interface{ Data() player.Config }, format string, args ...any) {
	p.Data().Session.SendActionBarMessage(text.Colourf(format, args...))
}

func (h *Handler) HandleAttackEntity(ctx *player.Context, e world.Entity, force, height *float64, critical *bool) {
	MustUsePool(func(pool *Pool) {
		pool.MustUser(ctx.Val(), func(u *User) {
			if _, ok := u.Player(); ok && u.Online() {
				u.online.cps.add()
				h.HandleInternal(u.online.cps, u.Data().FPS(), u, ctx.Val())
			}
		})
	})

	h.relay.HandleAttackEntity(ctx, e, force, height, critical)
}

func (h *Handler) HandleExperienceGain(ctx *player.Context, amount *int) {
	h.relay.HandleExperienceGain(ctx, amount)
}

const CPSLimit = 30

func (h *Handler) HandleInternal(c *cps, fps int, u *User, p *player.Player) {
	if c.Amount() >= CPSLimit {
		chatter.Messagef("<yellow>%s was kicked for exceeding cps limit.</yellow>", p.Name())
		instance.Kick(p, instance.ErrorPeedor)
		return
	}
	if fps == -1 && c.Amount() == 0 {
		return
	}

	var hasFps bool
	var hasCps bool

	if fps != -1 && u.Settings().Visual.ShowFPS {
		hasFps = true
	}
	if c.Amount() > 0 && u.Settings().Visual.ShowCPS {
		hasCps = true
	}

	const (
		onlyCPSFormat   = "<grey>CPS: <white>%d</white></grey>"
		onlyFPSFormat   = "<grey>FPS: <white>%d</white></grey>"
		CPSAndFPSFormat = "<grey>CPS: <white>%d</white> <dark-grey>|</dark-grey> FPS: <white>%d</white></grey>"
	)

	if hasCps && !hasFps {
		sendActionBar(p, onlyCPSFormat, c.Amount())
	}
	if hasFps && !hasCps {
		sendActionBar(p, onlyFPSFormat, fps)
	}
	if hasFps && hasCps {
		sendActionBar(p, CPSAndFPSFormat, c.Amount(), fps)
	}
}

func (h *Handler) HandlePunchAir(ctx *player.Context) {
	MustUsePool(func(pool *Pool) {
		pool.MustUser(ctx.Val(), func(u *User) {
			if _, ok := u.Player(); ok && u.Online() {
				u.online.cps.add()
				h.HandleInternal(u.online.cps, u.Data().FPS(), u, ctx.Val())
			}
		})
	})

	h.relay.HandlePunchAir(ctx)
}

func (h *Handler) HandleSignEdit(ctx *player.Context, pos cube.Pos, frontSide bool, oldText, newText string) {
	h.relay.HandleSignEdit(ctx, pos, frontSide, oldText, newText)
}

func (h *Handler) HandleLecternPageTurn(ctx *player.Context, pos cube.Pos, oldPage int, newPage *int) {
	h.relay.HandleLecternPageTurn(ctx, pos, oldPage, newPage)
}

func (h *Handler) HandleItemDamage(ctx *player.Context, i item.Stack, damage int) {
	h.relay.HandleItemDamage(ctx, i, damage)
}

func (h *Handler) HandleItemPickup(ctx *player.Context, i *item.Stack) {
	h.relay.HandleItemPickup(ctx, i)
}

func (h *Handler) HandleHeldSlotChange(ctx *player.Context, from, to int) {
	h.relay.HandleHeldSlotChange(ctx, from, to)
}

func (h *Handler) HandleItemDrop(ctx *player.Context, s item.Stack) {
	h.relay.HandleItemDrop(ctx, s)
}

func (h *Handler) HandleTransfer(ctx *player.Context, addr *net.UDPAddr) {
	h.relay.HandleTransfer(ctx, addr)
}

func (h *Handler) HandleCommandExecution(ctx *player.Context, command cmd.Command, args []string) {
	h.relay.HandleCommandExecution(ctx, command, args)
}

func (h *Handler) HandleDiagnostics(p *player.Player, d session.Diagnostics) {
	MustUsePool(func(pool *Pool) {
		pool.MustUser(p, func(u *User) {
			u.Data().setFPS(int(d.AverageFramesPerSecond))
			h.HandleInternal(u.online.cps, u.Data().FPS(), u, p)
		})
	})

	h.relay.HandleDiagnostics(p, d)
}

func (h *Handler) HandleLogin(ctx *event.Context[session.Conn]) {
	MustUsePool(func(p *Pool) {
		conn := ctx.Val()

		id, err := uuid.Parse(ctx.Val().IdentityData().Identity)
		if err != nil {
			_ = conn.WritePacket(&packet.Disconnect{
				Message: instance.KickMessage(instance.ErrorAngus),
			})
			ctx.Cancel()
			return
		}

		if u, ok := p.User(id); !ok {
			// user haven't existed in database, so we're creating new one
			offline := newBlankUser(conn, id)
			usr := FromOffline(offline)

			p.NewUser(usr)
			p.DB().NewAccount(offline)
		} else {
			// user have existed in database
			addDidIfNotExists(u, conn)
			addIpIfNotExists(u, conn)
		}
	})

	h.l.HandleLogin(ctx)
}

func addDidIfNotExists(u *User, c session.Conn) {
	if !slices.Contains(u.Data().DIDs(), c.ClientData().DeviceID) {
		u.Data().addDID(c.ClientData().DeviceID)
	}
}

func addIpIfNotExists(u *User, c session.Conn) {
	if !slices.Contains(u.Data().IPs(), internal.Ip(c.RemoteAddr())) {
		u.Data().addIp(internal.Ip(c.RemoteAddr()))
	}
}

func (h *Handler) HandleSpawn(pl *player.Player) {
	pl.UpdateDiagnostics(DisabledDiagnostics)

	MustUsePool(func(p *Pool) {
		p.MustUser(pl, func(u *User) {
			c, ok := plugin.M().Conn(pl.Name())
			if !ok {
				instance.Kick(pl, instance.ErrorValagalishe)
				return
			}

			u.SetOnline(pl, c)
		})
	})

	h.pr.RelayedHandleSpawn(pl)
}

func (h *Handler) HandleQuit(pl *player.Player) {
	MustUsePool(func(p *Pool) {
		p.MustUser(pl, func(u *User) {
			u.SetOffline()

			if err := p.DB().Update(Offline(u)); err != nil {
				l := plugin.M().Logger()
				l.Error("failed to update player data to database", "player", pl.Name(), "error", err.Error())
			}
		})
	})

	h.pr.RelayedHandleQuit(pl)
}

func newBlankUser(conn session.Conn, id uuid.UUID) OfflineUser {
	i := conn.IdentityData()
	c := conn.ClientData()

	return OfflineUser{
		Name:      i.DisplayName,
		XUID:      i.XUID,
		IPs:       []string{internal.Ip(conn.RemoteAddr())},
		DIDs:      []string{c.DeviceID},
		UUID:      id,
		Rank:      rank.Player,
		FirstJoin: time.Now(),
		Settings:  DefaultSettings(),
		KillStreak: struct{ Max, Current int }{
			Max:     0,
			Current: 0,
		},
	}
}

type login interface {
	HandleLogin(ctx *event.Context[session.Conn])
}

type practiceHandler interface {
	RelayedHandleSpawn(pl *player.Player)
	RelayedHandleQuit(p *player.Player)
	RelayedHandleDeath(deadPlayer *player.Player, src world.DamageSource, keepInv *bool)

	HandleSkinChange(ctx *player.Context, sk *skin.Skin)
	HandleMove(ctx *player.Context, newPos mgl64.Vec3, newRot cube.Rotation)
	HandleItemUseOnBlock(ctx *player.Context, blockPos cube.Pos, blockFace cube.Face, clickPos mgl64.Vec3)
	HandleItemUse(ctx *player.Context)
	HandleHeal(ctx *player.Context, health *float64, src world.HealingSource)
	HandleClientPacket(ctx *player.Context, pk packet.Packet)
	HandleChat(ctx *player.Context, msg *string)
	HandleCommandExecution(ctx *player.Context, cmd cmd.Command, args []string)
	HandleHurt(ctx *player.Context, damage *float64, immune bool, attackImmunity *time.Duration, src world.DamageSource)
	HandleAttackEntity(ctx *player.Context, attacked world.Entity, force *float64, height *float64, critical *bool)
	HandleHeldSlotChange(ctx *player.Context, from int, to int)
}
