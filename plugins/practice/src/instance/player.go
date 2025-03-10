package instance

import (
	"context"
	"errors"
	"fmt"
	atomic2 "github.com/df-mc/atomic"
	"github.com/df-mc/dragonfly/server/player"
	"github.com/df-mc/dragonfly/server/session"
	"github.com/df-mc/dragonfly/server/world"
	plugin "github.com/k4ties/df-plugin/df-plugin"
	"github.com/k4ties/dystopia/internal"
	"github.com/k4ties/dystopia/plugins/practice/src/bot"
	"github.com/k4ties/dystopia/plugins/practice/src/kit"
	hud2 "github.com/k4ties/dystopia/plugins/practice/src/user/hud"
	"github.com/k4ties/dystopia/plugins/practice/src/user/locker"
	"github.com/sandertv/gophertunnel/minecraft/protocol"
	"github.com/sandertv/gophertunnel/minecraft/protocol/packet"
	"github.com/sandertv/gophertunnel/minecraft/text"
	"github.com/sasha-s/go-deadlock"
	"image/color"
	"strings"
	"sync/atomic"
	"time"
)

type Player struct {
	*player.Player

	instance   Instance
	instanceMu deadlock.Mutex

	def Instance
	c   session.Conn

	safeTxActive atomic.Bool
	safeTxCtx    atomic2.Value[*CustomContext]
	safeTxNum    atomic.Int64

	transferring atomic.Bool
	frozen       atomic.Bool

	w   *world.World
	wMu deadlock.Mutex

	execSafeMu deadlock.Mutex
}

func (pl *Player) world() *world.World {
	return pl.w
}

func (pl *Player) setWorld(w *world.World) {
	pl.wMu.Lock()
	defer pl.wMu.Unlock()
	pl.w = w
}

func (pl *Player) Transferring() bool {
	return pl.transferring.Load()
}

func (pl *Player) setTransferring(t bool) {
	pl.transferring.Store(t)
}

func (pl *Player) setInstance(i Instance) {
	if i == nil {
		i = Nop
	}
	if i == Nop {
		pl.setWorld(plugin.M().Srv().World())
	} else {
		i.World(func(tx *world.Tx) {
			pl.setWorld(tx.World())
		})
	}

	pl.instanceMu.Lock()
	defer pl.instanceMu.Unlock()
	pl.instance = i
}

func (pl *Player) Instance() Instance {
	pl.instanceMu.Lock()
	defer pl.instanceMu.Unlock()

	if pl.instance == nil {
		return Nop
	}

	return pl.instance
}

func GetTypedInstance[T any](pl *Player) (T, bool) {
	for _, i := range AllInstances() {
		if i.Active(pl.UUID()) {
			v, ok := i.(T)
			return v, ok
		}
	}

	return *new(T), false
}

type WithinTransaction func(p *player.Player)

var causeSuccess = "successfully done"

func (pl *Player) ctx() (*CustomContext, bool) {
	if !pl.safeTxActive.Load() {
		return nil, false
	}
	ctx := pl.safeTxCtx.Load()
	if ctx == nil {
		return nil, false
	}
	return ctx, true
}

func (pl *Player) CloseTransactions() {
	if ctx, ok := pl.ctx(); ok {
		ctx.Cancel(causeSuccess)
	}
}

var count int

func (pl *Player) ExecSafe(f func(*player.Player, *world.Tx), w ...WithinTransaction) {
	go func() {
		if pl.safeTxActive.Load() {
			if ctx := pl.safeTxCtx.Load(); ctx != nil && ctx.Err() == nil {
				<-ctx.Done()
			}
		}

		count++
		ctx := NewCustomContext(context.Background(), time.Second)

		pl.safeTxActive.Store(true)
		pl.safeTxCtx.Store(ctx)

		pl.H().ExecWorld(func(tx *world.Tx, e world.Entity) {
			defer ctx.Cancel(causeSuccess)

			p := e.(*player.Player)
			if ctx.Err() != nil {
				return
			}
			f(p, tx)

			for _, w := range w {
				if ctx.Err() != nil {
					return
				}
				go w(p)
			}
		})

		select {
		case <-ctx.Done():
			pl.safeTxActive.Store(false)
			pl.safeTxCtx.Store(nil)

			if ctx.Err().Error() != causeSuccess {
				pl.Messagef(text.Red + "Your transaction is abnormal. If you can, send our support your nickname.")

				bot.Log(fmt.Sprintf("unexpected transaction finish cause: %s (%s)\n\ninstance: '%s'\ntransferring: %t\n\n%s", ctx.Err().Error(), pl.Name(), pl.Instance().Name(), pl.Transferring(), strings.Join(internal.GetCallers(6), "\n")))
				time.AfterFunc(time.Second, func() {
					bot.Log(fmt.Sprintf("(1s timeout) user status: %s\n\ninstance: '%s'\ntransferring: %t", pl.Name(), pl.Instance().Name(), pl.Transferring()))
				})
			}
		}
	}()
}

func (pl *Player) SendKit(k kit.Playered, tx *world.Tx, inTransaction ...func(*player.Player, *world.Tx)) {
	pl.Reset(tx, func(*player.Player) {
		pl.ExecSafe(func(p *player.Player, tx *world.Tx) {
			kit.Send(k(p), p)
			for _, f := range inTransaction {
				f(p, tx)
			}
		})
	})
}

func (pl *Player) Conn() (session.Conn, bool) {
	return pl.c, pl.c != nil
}

func (pl *Player) MustConn(f func(session.Conn)) {
	c, ok := pl.Conn()
	if !ok {
		Kick(pl.Player, ErrorSponge)
		return
	}

	f(c)
}

func (pl *Player) limitChunks(v int) {
	pl.MustConn(func(conn session.Conn) {
		_ = conn.WritePacket(&packet.ChunkRadiusUpdated{ChunkRadius: int32(v)})
	})
}

func (pl *Player) HideElements(e ...hud2.Element) {
	pl.MustConn(func(conn session.Conn) {
		hud2.Hide(conn, e...)
	})
}

func resetFunctions(p *player.Player) {
	p.Inventory().Clear()
	p.Armour().Clear()

	for _, e := range p.Effects() {
		p.RemoveEffect(e.Type())
	}

	p.SetGameMode(world.GameModeAdventure)

	p.CloseForm()
	p.CloseDialogue()

	p.EnableInstantRespawn()
	p.SetMobile()

	p.SetVisible()
	p.SetScale(1.0)

	//p.ShowCoordinates()

	p.SetExperienceProgress(0)
	p.SetExperienceLevel(0)

	p.SetScoreTag("")
}

func (pl *Player) Pause(pauseScreen bool) {
	pl.MustConn(func(conn session.Conn) {
		_ = conn.WritePacket(&packet.LevelEvent{
			EventType: func() int32 {
				if !pauseScreen {
					return 3006
				}

				return 3005
			}(),
			EventData: 1,
		})
	})
}

func (pl *Player) UnPause() {
	pl.MustConn(func(conn session.Conn) {
		_ = conn.WritePacket(&packet.LevelEvent{
			EventType: 3005,
			EventData: 0,
		})
	})
}

func (pl *Player) Frozen() bool {
	return pl.frozen.Load()
}

func (pl *Player) Freeze(by string) {
	if pl.Frozen() {
		// already frozen
		return
	}

	pl.LockElements(true, true)
	pl.frozen.Store(true)
	pl.Messagef(text.Yellow + "You've been frozen")

	bot.Log(fmt.Sprintf("%s was frozen by %s", pl.Name(), by))
}

func (pl *Player) UnFreeze(by string) {
	if !pl.Frozen() {
		// not frozen
		return
	}

	pl.UnlockElements()
	pl.frozen.Store(false)
	pl.Messagef(text.Yellow + "You've been unfrozen")

	bot.Log(fmt.Sprintf("%s was unfrozen by %s", pl.Name(), by))
}

func (pl *Player) LockElements(camera, movement bool) {
	if camera && movement {
		locker.LockCameraAndMovement(pl)
		return
	}
	if camera {
		locker.LockCamera(pl)
	}
	if movement {
		locker.LockMovement(pl)
	}
}

func (pl *Player) ShowCredits() {
	pl.MustConn(func(conn session.Conn) {
		_ = conn.WritePacket(&packet.ShowCredits{
			PlayerRuntimeID: 1,
		})
	})
}

func (pl *Player) UnlockElements() {
	locker.ResetLocks(pl)
}

func (pl *Player) Reset(tx *world.Tx, after ...func(p *player.Player)) {
	selfReset := func(pl *Player) {
		pl.ExecSafe(func(p *player.Player, tx *world.Tx) {
			resetFunctions(p)

			for _, f := range after {
				f(p)
			}
		})
	}

	if tx == nil {
		selfReset(pl)
		return
	}

	e, ok := pl.H().Entity(tx)
	if !ok {
		selfReset(pl)
		return
	}

	p := e.(*player.Player)
	resetFunctions(p)

	for _, f := range after {
		f(p)
	}
}

func (pl *Player) ResetElements(e ...hud2.Element) {
	pl.MustConn(func(conn session.Conn) {
		hud2.Reset(conn, e...)
	})
}

func (pl *Player) syncConn(m *plugin.Manager) error {
	c, ok := m.Conn(pl.Name())
	if !ok {
		return errors.New("could not find connection")
	}

	pl.c = c
	return nil
}

func (pl *Player) enableChunkCache() {
	if c, ok := pl.Conn(); ok {
		_ = c.WritePacket(&packet.ClientCacheStatus{
			Enabled: true,
		})
	}
}

func (pl *Player) Crash() {
	pl.MustConn(func(conn session.Conn) {
		_ = conn.WritePacket(&packet.RemoveActor{
			EntityUniqueID: 1,
		})
	})
}

func FadeInCamera(c session.Conn, dur float32, fadeIn bool) {
	var duration, fadeInDuration float32

	duration = dur / 2

	if fadeIn {
		fadeInDuration = dur / 3
		duration = dur / 3
	}

	_ = c.WritePacket(&packet.CameraInstruction{
		Fade: protocol.Option(protocol.CameraInstructionFade{
			TimeData: protocol.Option(protocol.CameraFadeTimeData{
				FadeInDuration:  fadeInDuration,
				WaitDuration:    duration,
				FadeOutDuration: duration,
			}),
			Colour: protocol.Option(color.RGBA{R: 10}),
		}),
	})
}
