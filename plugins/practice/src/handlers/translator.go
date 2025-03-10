package handlers

import (
	"github.com/df-mc/dragonfly/server/block/cube"
	"github.com/df-mc/dragonfly/server/cmd"
	"github.com/df-mc/dragonfly/server/item"
	"github.com/df-mc/dragonfly/server/player"
	"github.com/df-mc/dragonfly/server/player/skin"
	"github.com/df-mc/dragonfly/server/session"
	"github.com/df-mc/dragonfly/server/world"
	"github.com/go-gl/mathgl/mgl64"
	plugin "github.com/k4ties/df-plugin/df-plugin"
	"github.com/k4ties/dystopia/plugins/practice/src/instance"
	"github.com/k4ties/dystopia/plugins/practice/src/rank"
	"github.com/k4ties/dystopia/plugins/practice/src/user"
	"github.com/k4ties/dystopia/plugins/practice/src/user/chatter"
	"github.com/sandertv/gophertunnel/minecraft/protocol/packet"
	"github.com/sandertv/gophertunnel/minecraft/text"
	"net"
	"time"
)

// Translator is made to relay any handler to a specific user. If you want to know why we need this,
// you can use default practice handler instead of Translator and see what happens. It is an issue with my
// library "df-plugin", I will fix it in the future, but for now I have only this solution.
type Translator struct {
	Decliner
	plugin.NopPlayerHandler

	p *user.Pool
	i instance.Instance
}

func NewTranslator(p *user.Pool, i instance.Instance) *Translator {
	if p == nil {
		panic("nil pool")
	}
	if i == nil {
		panic("nil instance")
	}
	return &Translator{p: p, i: i}
}

func (t *Translator) HandleJump(p *player.Player) {
	MustPractice(p.UUID(), p, t.i).HandleJump(p)
}

func (t *Translator) HandleTeleport(ctx *player.Context, pos mgl64.Vec3) {
	MustPractice(ctx.Val().UUID(), ctx.Val(), t.i).HandleTeleport(ctx, pos)
}

func (t *Translator) HandleChangeWorld(p *player.Player, before, after *world.World) {
	MustPractice(p.UUID(), p, t.i).HandleChangeWorld(p, before, after)
}

func (t *Translator) HandleToggleSprint(ctx *player.Context, after bool) {
	MustPractice(ctx.Val().UUID(), ctx.Val(), t.i).HandleToggleSprint(ctx, after)
}

func (t *Translator) HandleToggleSneak(ctx *player.Context, after bool) {
	MustPractice(ctx.Val().UUID(), ctx.Val(), t.i).HandleToggleSneak(ctx, after)
}

func (t *Translator) HandleRespawn(p *player.Player, pos *mgl64.Vec3, w **world.World) {
	MustPractice(p.UUID(), p, t.i).HandleRespawn(p, pos, w)
}

func (t *Translator) HandleBlockPick(ctx *player.Context, pos cube.Pos, b world.Block) {
	MustPractice(ctx.Val().UUID(), ctx.Val(), t.i).HandleBlockPick(ctx, pos, b)
}

func (t *Translator) HandleItemUseOnEntity(ctx *player.Context, e world.Entity) {
	MustPractice(ctx.Val().UUID(), ctx.Val(), t.i).HandleItemUseOnEntity(ctx, e)
}

func (t *Translator) HandleItemRelease(ctx *player.Context, item item.Stack, dur time.Duration) {
	MustPractice(ctx.Val().UUID(), ctx.Val(), t.i).HandleItemRelease(ctx, item, dur)
}

func (t *Translator) HandleItemConsume(ctx *player.Context, item item.Stack) {
	MustPractice(ctx.Val().UUID(), ctx.Val(), t.i).HandleItemConsume(ctx, item)
}

func (t *Translator) HandleExperienceGain(ctx *player.Context, amount *int) {
	MustPractice(ctx.Val().UUID(), ctx.Val(), t.i).HandleExperienceGain(ctx, amount)
}

func (t *Translator) HandleLecternPageTurn(ctx *player.Context, pos cube.Pos, oldPage int, newPage *int) {
	MustPractice(ctx.Val().UUID(), ctx.Val(), t.i).HandleLecternPageTurn(ctx, pos, oldPage, newPage)
}

func (t *Translator) HandleTransfer(ctx *player.Context, addr *net.UDPAddr) {
	MustPractice(ctx.Val().UUID(), ctx.Val(), t.i).HandleTransfer(ctx, addr)
}

func (t *Translator) HandleDiagnostics(p *player.Player, d session.Diagnostics) {
	MustPractice(p.UUID(), p, t.i).HandleDiagnostics(p, d)
}

func (t *Translator) RelayedHandleSpawn(pl *player.Player) {
	MustPractice(pl.UUID(), pl, t.i).HandleSpawn(pl)
}

func (t *Translator) HandlePunchAir(ctx *player.Context) {
	MustPractice(ctx.Val().UUID(), ctx.Val(), t.i).HandlePunchAir(ctx)
}

func (t *Translator) RelayedHandleQuit(pl *player.Player) {
	MustPractice(pl.UUID(), pl, t.i).HandleQuit(pl)
	RemovePractice(pl.UUID())
}

func (t *Translator) RelayedHandleDeath(deadPlayer *player.Player, src world.DamageSource, keepInv *bool) {
	MustPractice(deadPlayer.UUID(), deadPlayer, t.i).HandleDeath(deadPlayer, src, keepInv)
}

func (t *Translator) HandleSkinChange(ctx *player.Context, s *skin.Skin) {
	MustPractice(ctx.Val().UUID(), ctx.Val(), t.i).HandleSkinChange(ctx, s)
}

func (t *Translator) HandleItemUseOnBlock(ctx *player.Context, p cube.Pos, f cube.Face, p2 mgl64.Vec3) {
	MustPractice(ctx.Val().UUID(), ctx.Val(), t.i).HandleItemUseOnBlock(ctx, p, f, p2)
}

func (t *Translator) HandleItemUse(ctx *player.Context) {
	MustPractice(ctx.Val().UUID(), ctx.Val(), t.i).HandleItemUse(ctx)
}

func (t *Translator) HandleHeal(ctx *player.Context, health *float64, src world.HealingSource) {
	MustPractice(ctx.Val().UUID(), ctx.Val(), t.i).HandleHeal(ctx, health, src)
}

func (t *Translator) HandleClientPacket(ctx *player.Context, pk packet.Packet) {
	MustPractice(ctx.Val().UUID(), ctx.Val(), t.i).HandleClientPacket(ctx, pk)
}

func (t *Translator) HandleServerPacket(ctx *player.Context, pk packet.Packet) {
	MustPractice(ctx.Val().UUID(), ctx.Val(), t.i).HandleServerPacket(ctx, pk)
}

func (t *Translator) HandleChat(ctx *player.Context, msg *string) {
	var u *user.User
	var found bool

	t.p.MustUser(ctx.Val(), func(usr *user.User) {
		u = usr
		found = true
	})

	if !found {
		return
	}

	if !chatter.ChatEnabled() {
		if u.Data().Rank().Priority < rank.PriorityAdmin {
			ctx.Val().Messagef(text.Yellow + "Chat is temporarily disabled.")
			ctx.Cancel()
			return
		}
	}

	MustPractice(ctx.Val().UUID(), ctx.Val(), t.i).HandleChat(ctx, msg)
}

func (t *Translator) HandleMove(ctx *player.Context, newPos mgl64.Vec3, rot cube.Rotation) {
	MustPractice(ctx.Val().UUID(), ctx.Val(), t.i).HandleMove(ctx, newPos, rot)
}

func (t *Translator) HandleCommandExecution(ctx *player.Context, cmd cmd.Command, args []string) {
	MustPractice(ctx.Val().UUID(), ctx.Val(), t.i).HandleCommandExecution(ctx, cmd, args)
}

func (t *Translator) HandleHurt(ctx *player.Context, damage *float64, immune bool, attackImmunity *time.Duration, src world.DamageSource) {
	MustPractice(ctx.Val().UUID(), ctx.Val(), t.i).HandleHurt(ctx, damage, immune, attackImmunity, src)
}

func (t *Translator) HandleAttackEntity(ctx *player.Context, attacked world.Entity, force *float64, height *float64, critical *bool) {
	MustPractice(ctx.Val().UUID(), ctx.Val(), t.i).HandleAttackEntity(ctx, attacked, force, height, critical)
}

func (t *Translator) HandleHeldSlotChange(ctx *player.Context, from int, to int) {
	MustPractice(ctx.Val().UUID(), ctx.Val(), t.i).HandleHeldSlotChange(ctx, from, to)
}

//func (t *Translator) HandleQuit(p *player.Player) {
//
//}

//func (t *Translator) HandleDeath(p *player.Player, src world.DamageSource, keepInv *bool) {
//
//}

//func (t *Translator) HandleLogin(ctx *event.Context[session.Conn]) {
//
//}

//func (t *Translator) HandleSpawn(p *player.Player) {
//
//}
