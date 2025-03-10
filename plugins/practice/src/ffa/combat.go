package ffa

import (
	"context"
	"github.com/bedrock-gophers/cooldown/cooldown"
	"github.com/df-mc/atomic"
	"github.com/df-mc/dragonfly/server/event"
	"github.com/df-mc/dragonfly/server/player/bossbar"
	"github.com/df-mc/dragonfly/server/world"
	"github.com/google/uuid"
	"github.com/k4ties/dystopia/internal"
	"github.com/k4ties/dystopia/plugins/practice/src/instance"
	"github.com/k4ties/dystopia/plugins/practice/src/user"
	"github.com/sandertv/gophertunnel/minecraft/text"
	"github.com/sasha-s/go-deadlock"
	"math"
	"strconv"
	"time"
)

type Combat struct {
	cd   *cooldown.CoolDown
	with atomic.Value[CombatUser]

	c   context.CancelFunc
	dur time.Duration

	t     *time.Timer
	owner CombatUser

	tk      *time.Ticker
	handler atomic.Value[CombatHandle]

	hiddenPlayers   map[uuid.UUID]*instance.Player
	hiddenPlayersMu deadlock.RWMutex

	ffa *Instance
}

type Context = event.Context[CombatUser]

type CombatHandle interface {
	HandleCombatStart(ctx *Context, with CombatUser)
	HandleRenewCooldown(ctx *Context, with CombatUser)
	HandleCombatStop(owner, with CombatUser)
}
type NopCombatHandle struct{}

func (n NopCombatHandle) HandleRenewCooldown(*Context, CombatUser) {}
func (n NopCombatHandle) HandleCombatStart(*Context, CombatUser)   {}
func (n NopCombatHandle) HandleCombatStop(CombatUser, CombatUser)  {}

type CombatUser interface {
	Name() string
	Messagef(string, ...any)

	UUID() uuid.UUID

	Settings() user.Settings

	InputMode() internal.InputMode
	OS() internal.OS

	GameMode() world.GameMode

	Latency() time.Duration

	SendBossBar(bossbar.BossBar)
	RemoveBossBar()

	SendTip(a ...any)
}

func NewCombat(cd time.Duration, owner CombatUser, ffa *Instance) *Combat {
	return &Combat{
		cd:    cooldown.NewCoolDown(),
		dur:   cd,
		owner: owner,
		ffa:   ffa,

		hiddenPlayers: make(map[uuid.UUID]*instance.Player),
	}
}

func (c *Combat) addHiddenPlayer(p *instance.Player) {
	c.hiddenPlayersMu.Lock()
	defer c.hiddenPlayersMu.Unlock()

	c.hiddenPlayers[p.UUID()] = p
}

func (c *Combat) removeHiddenPlayer(id uuid.UUID) {
	c.hiddenPlayersMu.Lock()
	defer c.hiddenPlayersMu.Unlock()

	delete(c.hiddenPlayers, id)
}

func (c *Combat) HiddenPlayers() []*instance.Player {
	c.hiddenPlayersMu.RLock()
	defer c.hiddenPlayersMu.RUnlock()

	var players []*instance.Player
	for _, p := range c.hiddenPlayers {
		players = append(players, p)
	}

	return players
}

func (c *Combat) WithHandler(h CombatHandle) *Combat {
	c.handler.Store(h)
	return c
}

func (c *Combat) Handler() CombatHandle {
	h := c.handler.Load()
	if h == nil {
		return NopCombatHandle{}
	}

	return h
}

func (c *Combat) Handle(h CombatHandle) {
	c.handler.Store(h)
}

func (c *Combat) InCombatWith(w CombatUser) bool {
	if !c.Active() {
		return false
	}

	return c.With().UUID() == w.UUID()
}

func (c *Combat) Stop() {
	c.CancelFunc()()
}

func (c *Combat) CancelFunc() context.CancelFunc {
	if c.c == nil || !c.Active() {
		return func() {}
	}

	return c.c
}

func (c *Combat) Active() bool {
	if c.cd != nil {
		return c.cd.Active()
	}

	return false
}

func (c *Combat) RenewCoolDown() {
	if c.Active() {
		ctx := event.C(c.owner)
		if c.Handler().HandleRenewCooldown(ctx, c.With()); !ctx.Cancelled() {
			c.cd.Set(c.dur)
			c.tk = time.NewTicker(time.Second / 10)

			if c.owner.Settings().FFA.ShowBossBar {
				c.owner.SendBossBar(c.boosBar(1.0))
			}

			if c.t != nil {
				c.t.Reset(c.dur)
			}
		}
	}
}

func (c *Combat) title() string {
	format := "%s: <grey>%s</grey> <dark-grey>(%s)</dark-grey>"
	return text.Colourf(format, c.With().Name(), c.With().OS().String(), strconv.Itoa(int(c.With().Latency().Milliseconds()))+"ms")
}

func (c *Combat) boosBar(healthPercentage float64) bossbar.BossBar {
	if healthPercentage < 0 {
		healthPercentage = 0
	}
	if healthPercentage > 1 {
		healthPercentage = 1.0
	}

	return bossbar.New(c.title()).WithColour(bossbar.Red()).WithHealthPercentage(healthPercentage)
}

type DisableCombatMessageFunc func()

func (c *Combat) transferring(ids ...uuid.UUID) bool {
	if len(ids) == 0 {
		return false
	}
	var players []*instance.Player
	for _, id := range ids {
		if pl, _ := LookupPlayer(id); pl != nil {
			players = append(players, pl)
		}
	}
	for _, pl := range players {
		if pl.Transferring() {
			return true
		}
	}

	return false
}

func (c *Combat) Start(w CombatUser) (context.CancelFunc, DisableCombatMessageFunc) {
	if c.transferring(w.UUID(), c.owner.UUID()) {
		return nil, nil
	}

	ctx := event.C(c.owner)
	var showMsg = true

	ctx2, cancel := context.WithCancel(context.Background())

	if c.Handler().HandleCombatStart(ctx, w); !ctx.Cancelled() {
		c.c = cancel
		c.with.Store(w)

		c.t = time.NewTimer(c.dur)
		c.cd.Set(c.dur)

		c.tk = time.NewTicker(time.Second / 10)

		go func() {
			defer func() {
				c.Handler().HandleCombatStop(c.owner, c.With())

				if showMsg {
					msg := text.Colourf("<red><b>>></b></red> Your combat is now expried.")
					c.owner.Messagef(msg)
				}

				c.owner.RemoveBossBar()
				c.with.Store(nil)
				c.cd.Set(0)
				c.t.Reset(0)
				c.t.Stop()
				c.tk.Stop()
				c.c = nil
				c.t = nil
				c.tk = nil
			}()

			if c.owner.Settings().FFA.ShowBossBar {
				c.owner.SendBossBar(c.boosBar(1.0))
			}

			for {
				select {
				case <-ctx2.Done():
					showMsg = false
					return
				case <-c.t.C:
					return
				case <-c.tk.C:
					if c.owner.Settings().FFA.ShowBossBar {
						z := 1.0
						o := math.Floor(c.dur.Seconds())
						v := c.cd.Remaining().Seconds()

						percentage := (z / o) * v
						c.owner.SendBossBar(c.boosBar(percentage))
					}
				}
			}
		}()
	}

	return cancel, func() {
		showMsg = false
	}
}

func (c *Combat) With() CombatUser {
	return c.with.Load()
}

func TypedWith[T any](c *Combat) (T, bool) {
	t, ok := c.with.Load().(T)
	if !ok {
		return *new(T), false
	}

	return t, true
}
