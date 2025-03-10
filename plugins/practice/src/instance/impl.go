package instance

import (
	"github.com/df-mc/atomic"
	"github.com/df-mc/dragonfly/server/block/cube"
	"github.com/df-mc/dragonfly/server/player"
	"github.com/df-mc/dragonfly/server/world"
	"github.com/go-gl/mathgl/mgl64"
	"github.com/google/uuid"
	plugin "github.com/k4ties/df-plugin/df-plugin"
	"github.com/k4ties/dystopia/plugins/practice/src/bot"
	"github.com/k4ties/dystopia/plugins/practice/src/user/hud"
	"github.com/sandertv/gophertunnel/minecraft/text"
	"github.com/sasha-s/go-deadlock"
	"iter"
	"log/slog"
	"math"
	"slices"
)

type Impl struct {
	errorLog *slog.Logger

	players   map[uuid.UUID]*Player
	playersMu deadlock.RWMutex

	name string

	world    *world.World
	gameMode world.GameMode

	hidden     []hud.Element
	defaultRot cube.Rotation

	heightThresholdStatus atomic.Bool
	heightThreshold       int
	heightThresholdMode   OnIntersectThreshold

	onExit []func(*Player, *player.Player, *world.Tx)
}

func (i *Impl) Name() string {
	return i.name
}

func (i *Impl) Player(u uuid.UUID) (*Player, bool) {
	i.playersMu.RLock()
	defer i.playersMu.RUnlock()

	p, ok := i.players[u]
	return p, ok
}

func (i *Impl) Messagef(s string, args ...any) {
	msg := text.Colourf(s, args...)
	bot.Log(msg)

	for p := range i.Players() {
		if plugin.M().Online(p.UUID()) {
			p.Message(msg)
		}
	}
}

func (i *Impl) HeightThresholdMode() OnIntersectThreshold {
	return i.heightThresholdMode
}

func (i *Impl) HeightThresholdEnabled() bool {
	return i.heightThresholdStatus.Load()
}

func (i *Impl) ToggleHeightThreshold() {
	i.heightThresholdStatus.Store(!i.heightThresholdStatus.Load())
}

func (i *Impl) HeightThreshold() int {
	return i.heightThreshold
}

func (i *Impl) Transfer(pl *Player, tx *world.Tx) {
	if i.Active(pl.UUID()) || pl.Transferring() {
		return
	}

	pl.setTransferring(true)

	if tx == nil || !i.inWorld(tx) {
		pl.ExecSafe(func(p *player.Player, tx *world.Tx) {
			i.transfer(pl, tx)
		})
		return
	}

	i.transfer(pl, tx)
}

func (i *Impl) transfer(pl *Player, tx *world.Tx) {
	FadeInCamera(pl.c, 0.895, false)

	if pl.Instance() != Nop {
		if imp, ok := pl.Instance().(interface {
			Hidden() []hud.Element
		}); ok {
			for _, elem := range imp.Hidden() {
				if !i.isHidden(elem) {
					pl.ResetElements(elem)
				} else {
					pl.HideElements(elem)
				}
			}
		}

		pl.Instance().removeFromList(pl)
	}

	if !i.inWorld(tx) {
		h := tx.RemoveEntity(pl.Player)

		i.World(func(tx *world.Tx) {
			tx.AddEntity(h)
		})
	}

	i.Rotate(pl)
	pl.ExecSafe(func(p *player.Player, tx *world.Tx) {
		p.SetGameMode(i.gameMode)
		p.Teleport(i.WorldSpawn())
	})

	i.addToList(pl)
	pl.HideElements(i.hidden...)

	pl.setTransferring(false)
}

type WorldSpawn interface {
	WorldSpawn() mgl64.Vec3
}

func (i *Impl) WorldSpawn() mgl64.Vec3 {
	return i.world.Spawn().Vec3Centre()
}

func (i *Impl) Rotate(p *Player) {
	p.ExecSafe(func(p *player.Player, tx *world.Tx) {
		currentRot := p.Rotation()
		maxRot := p.Rotation()

		yawDiff := findAngleDifference(currentRot.Yaw(), maxRot.Yaw()) - currentRot.Yaw()
		pitchDiff := findPitchDifference(currentRot.Pitch(), maxRot.Pitch()) - currentRot.Pitch()

		p.Move(mgl64.Vec3{}, yawDiff, pitchDiff)
	})
}

func findPitchDifference(currentPitch float64, targetPitch float64) float64 {
	diff := targetPitch - currentPitch
	if math.Abs(diff) > 180 {
		if diff < 0 {
			diff += 360
		} else {
			diff -= 360
		}
	}
	return diff
}

func findAngleDifference(yaw float64, expectedYaw float64) float64 {
	diff := math.Mod(expectedYaw-yaw+1080, 360) - 180
	return diff
}

func (i *Impl) Hidden() []hud.Element {
	return i.hidden
}

func (i *Impl) addToList(p *Player) {
	if p == nil {
		return
	}
	if i.Active(p.UUID()) {
		panic("player is already in instance")
	}

	i.playersMu.Lock()
	i.players[p.UUID()] = p
	i.playersMu.Unlock()

	p.setInstance(i)
}

func (i *Impl) removeFromList(p *Player) {
	if !i.Active(p.UUID()) {
		panic("cannot remove from instance player that is not in instance")
	}
	for _, f := range i.onExit {
		p.ExecSafe(func(safeP *player.Player, tx *world.Tx) {
			f(p, safeP, tx)
		})
	}

	i.playersMu.Lock()
	delete(i.players, p.UUID())
	i.playersMu.Unlock()

	p.setInstance(nil)
}

func (i *Impl) inWorld(tx *world.Tx) (found bool) {
	if tx == nil {
		return false
	}
	if tx.World() == i.world {
		found = true
	}

	return
}

func (i *Impl) isHidden(e hud.Element) bool {
	return slices.Contains(i.hidden, e)
}

func (i *Impl) DefaultRotation() cube.Rotation {
	return i.defaultRot
}

func (i *Impl) World(f func(*world.Tx)) {
	i.world.Exec(func(tx *world.Tx) {
		f(tx)
	})
}

func (i *Impl) GameMode() world.GameMode {
	return i.gameMode
}

func (i *Impl) Players() iter.Seq[*Player] {
	return func(yield func(*Player) bool) {
		i.playersMu.RLock()
		defer i.playersMu.RUnlock()

		for _, p := range i.players {
			if !yield(p) {
				return
			}
		}
	}
}

func (i *Impl) Active(u uuid.UUID) bool {
	i.playersMu.RLock()
	defer i.playersMu.RUnlock()

	_, ok := i.players[u]
	return ok
}

func (i *Impl) ErrorLog() *slog.Logger {
	return i.errorLog
}

func (i *Impl) NewPlayer(p *player.Player) *Player {
	pl := &Player{Player: p, instance: i}

	if err := pl.syncConn(plugin.M()); err != nil {
		Kick(p, ErrorSponge)
		return nil
	}

	pl.enableChunkCache()
	pl.limitChunks(16)
	return pl
}

type OnIntersectThreshold int

const (
	EventDeath OnIntersectThreshold = iota
	EventTeleportToSpawn
)

type HeightThresholdConfig struct {
	Enabled   bool
	Threshold int
	OnDeath   OnIntersectThreshold
}

func New(name string, w *world.World, g world.GameMode, errorLogger *slog.Logger, defaultRot cube.Rotation, htc HeightThresholdConfig, hidden ...hud.Element) Instance {
	i := &Impl{name: name, players: make(map[uuid.UUID]*Player), world: w, gameMode: g, errorLog: errorLogger, defaultRot: defaultRot, hidden: hidden}
	if htc.Enabled {
		i.ToggleHeightThreshold()
		i.heightThreshold = htc.Threshold
		i.heightThresholdMode = htc.OnDeath
	}
	return i
}

func (i *Impl) WithOnExitFuncs(f ...func(*Player, *player.Player, *world.Tx)) *Impl {
	i.onExit = append(i.onExit, f...)
	return i
}
