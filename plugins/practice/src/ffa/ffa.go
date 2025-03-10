package ffa

import (
	"context"
	"fmt"
	"github.com/bedrock-gophers/cooldown/cooldown"
	atomic2 "github.com/df-mc/atomic"
	"github.com/df-mc/dragonfly/server/entity/effect"
	"github.com/df-mc/dragonfly/server/player"
	"github.com/df-mc/dragonfly/server/world"
	"github.com/google/uuid"
	"github.com/k4ties/dystopia/plugins/practice/src/bot"
	instance2 "github.com/k4ties/dystopia/plugins/practice/src/instance"
	"github.com/k4ties/dystopia/plugins/practice/src/instance/lobby"
	"github.com/k4ties/dystopia/plugins/practice/src/kit"
	user2 "github.com/k4ties/dystopia/plugins/practice/src/user"
	"github.com/sandertv/gophertunnel/minecraft/text"
	"github.com/sasha-s/go-deadlock"
	"math"
	"slices"
	"strings"
	"sync/atomic"
	"time"
)

type PlayeredCombatHandle func(CombatUser) CombatHandle

type Instance struct {
	*instance2.Impl
	k kit.Playered

	c      Config
	closed atomic.Bool

	cooldowns   map[uuid.UUID]*cooldown.CoolDown
	cooldownsMu deadlock.RWMutex

	cooldownCancelFuncs   map[uuid.UUID]context.CancelFunc
	cooldownCancelFuncsMu deadlock.Mutex

	combats   map[uuid.UUID]*Combat
	combatsMu deadlock.RWMutex

	combatHandler atomic2.Value[PlayeredCombatHandle]
}

type Config struct {
	Name string
	Icon string

	PearlCooldown time.Duration
}

func New(i *instance2.Impl, k kit.Playered, c Config) *Instance {
	f := &Instance{k: k, Impl: i, c: c, cooldowns: make(map[uuid.UUID]*cooldown.CoolDown), cooldownCancelFuncs: make(map[uuid.UUID]context.CancelFunc), combats: make(map[uuid.UUID]*Combat)}
	f.Impl = f.Impl.WithOnExitFuncs(f.OnExit)
	f.combatHandler.Store(func(CombatUser) CombatHandle { return NopCombatHandle{} })

	if slices.Contains(Closed.Closed, strings.ToLower(c.Name)) {
		i.World(func(tx *world.Tx) {
			f.Close(tx)
		})
	}

	return f
}

func (i *Instance) CombatHandle(p CombatUser) CombatHandle {
	return i.combatHandler.Load()(p)
}

func (i *Instance) SetCombatHandle(p PlayeredCombatHandle) {
	i.combatHandler.Store(p)
}

func (i *Instance) HasPearCooldown() bool {
	return i.c.PearlCooldown > 0
}

func (i *Instance) PearlCooldown() time.Duration {
	return i.c.PearlCooldown
}

const (
	JoinFormat = "<red>%s</red> has joined to <red>%s</red>. <dark-grey>(%d)</dark-grey>"
	QuitFormat = "<red>%s</red> left from the <red>%s</red>. <dark-grey>(%v)</dark-grey>"
)

var combatCooldown = time.Second * 15

func (i *Instance) Transfer(pl *instance2.Player, tx *world.Tx) {
	if i.Closed() {
		pl.Messagef(text.Colourf("<red>Sorry, this game mode is closed.</red>"))
		return
	}

	if !i.Impl.Active(pl.UUID()) {
		msg := text.Colourf(JoinFormat, pl.Name(), i.c.Name, i.playerLen())

		i.Messagef(msg)
		pl.Messagef(msg)

		i.addToCooldownList(pl.UUID())

		canContinue := false

		user2.P().MustUser(pl.Player, func(u *user2.User) {
			i.addCombat(pl.UUID(), NewCombat(combatCooldown, u, i).WithHandler(i.CombatHandle(u)))
			canContinue = true
		})

		if !canContinue {
			return
		}
	}

	i.Impl.Transfer(pl, tx)

	pl.SendKit(i.k, tx)
	pl.Messagef(text.Colourf("<green>You've been teleported to the %s.</green>", i.c.Name))

	bot.Log(fmt.Sprintf("%s joined to the %s", pl.Name(), i.Name()))
}

func (i *Instance) ReKit(id uuid.UUID, tx *world.Tx) {
	if !i.Active(id) {
		return
	}
	pl, _ := i.Player(id)

	if c, ok := i.Combat(id); ok {
		if c.Active() {
			pl.Messagef(text.Red + "You cannot rekit in combat")
			return
		}
	}

	i.reKit(id, tx)
}

func (i *Instance) reKit(id uuid.UUID, tx *world.Tx) {
	if !i.Active(id) {
		return
	}

	pl, _ := i.Player(id)
	i.ResetPearlCooldown(pl)

	pl.Heal(pl.MaxHealth(), effect.InstantHealingSource{})

	if tx == nil {
		pl.ExecSafe(func(p *player.Player, tx *world.Tx) {
			pl.SendKit(i.Kit(), tx)
			p.SetGameMode(i.GameMode())
		})
		return
	}

	pl.SendKit(i.Kit(), tx, func(p *player.Player, tx *world.Tx) {
		p.SetGameMode(i.GameMode())
	})
}

func (i *Instance) AllPlayersInCombat() map[*instance2.Player]*Combat {
	m := make(map[*instance2.Player]*Combat)

	for pl := range i.Players() {
		if i.InCombat(pl.UUID()) {
			if c, ok := i.Combat(pl.UUID()); ok {
				m[pl] = c
			}
		}
	}

	return m
}

func (i *Instance) MustCombat(id uuid.UUID, f func(*Combat, *instance2.Player, *Instance)) {
	if c, ok := i.Combat(id); ok {
		if pl, ok := i.Player(id); ok {
			f(c, pl, i)
		}
	}
}

func (i *Instance) Combat(id uuid.UUID) (*Combat, bool) {
	i.combatsMu.RLock()
	defer i.combatsMu.RUnlock()

	c, ok := i.combats[id]
	return c, ok
}

func (i *Instance) StopCombat(id uuid.UUID) {
	if i.InCombat(id) {
		c, _ := i.Combat(id)
		c.Stop()
	}
}

func (i *Instance) InCombat(id uuid.UUID) bool {
	_, ok := i.Combat(id)
	return ok
}

func (i *Instance) addCombat(id uuid.UUID, c *Combat) {
	i.combatsMu.Lock()
	defer i.combatsMu.Unlock()

	i.combats[id] = c
}

func (i *Instance) removeCombat(id uuid.UUID) {
	i.combatsMu.Lock()
	defer i.combatsMu.Unlock()

	delete(i.combats, id)
}

func (i *Instance) StartPearlCoolDown(p *instance2.Player, c context.Context) context.CancelFunc {
	if !i.Active(p.UUID()) {
		return nil
	}
	if c == nil {
		c = context.Background()
	}

	ctx, cancel := context.WithCancel(c)

	i.MustCoolDown(p.UUID(), func(cd *cooldown.CoolDown) {
		if cd.Active() {
			// if cooldown is already active we are resetting the existing one
			i.ResetPearlCooldown(p)
		}

		cd.Set(i.PearlCooldown())
		i.addCooldownCancelFunc(p.UUID(), cancel)

		p.SetExperienceLevel(int(i.PearlCooldown().Seconds()))
		p.SetExperienceProgress(0.999) // if we will make 1.0 it will be new level

		go func() {
			var showMessage = true

			defer func() {
				i.ResetPearlCooldown(p)

				if showMessage {
					p.SendTip(text.Colourf("<green>Your cooldown has been expired</green>"))
				}
			}()

			ticker := time.NewTicker(time.Second / 10)
			timer := time.NewTimer(i.PearlCooldown())

			defer func() {
				ticker.Stop()
				timer.Stop()
				cd.Set(0)
			}()

			for {
				select {
				case <-ctx.Done():
					// if canceled, we must not show message
					showMessage = false
					return
				case <-timer.C:
					return
				case <-ticker.C:
					a := 1.0
					b := float64(int(i.PearlCooldown().Seconds()))
					c := cd.Remaining().Seconds()

					progress := (a / b) * c
					if progress > 1 {
						progress = 1
					}
					if progress < 0 {
						progress = 0
					}

					p.SetExperienceLevel(int(math.Floor(c)) + 1)
					p.SetExperienceProgress(progress)
				}
			}
		}()
	})

	return cancel
}

func (i *Instance) addCooldownCancelFunc(u uuid.UUID, c context.CancelFunc) {
	i.cooldownCancelFuncsMu.Lock()
	defer i.cooldownCancelFuncsMu.Unlock()

	i.cooldownCancelFuncs[u] = c
}

func (i *Instance) removeCooldownCancelFunc(u uuid.UUID) {
	i.cooldownCancelFuncsMu.Lock()
	defer i.cooldownCancelFuncsMu.Unlock()
	delete(i.cooldownCancelFuncs, u)
}

func (i *Instance) cooldownCancelFunc(u uuid.UUID) context.CancelFunc {
	i.cooldownCancelFuncsMu.Lock()
	defer i.cooldownCancelFuncsMu.Unlock()
	return i.cooldownCancelFuncs[u]
}

func (i *Instance) OnExit(pl *instance2.Player, _ *player.Player, _ *world.Tx) {
	i.ResetPearlCooldown(pl)
	i.removeFromCooldownList(pl.UUID())

	i.removeCombat(pl.UUID())
	i.Messagef(text.Colourf(QuitFormat, pl.Name(), i.c.Name, i.playerLen()-1))

	bot.Log(fmt.Sprintf("%s left from the %s", pl.Name(), i.Name()))
}

func (i *Instance) ResetPearlCooldown(pl *instance2.Player) {
	f := i.cooldownCancelFunc(pl.UUID())
	if f == nil {
		return
	}

	f()
	i.removeCooldownCancelFunc(pl.UUID())

	pl.ExecSafe(func(p *player.Player, tx *world.Tx) {
		p.SetExperienceLevel(0)
		p.SetExperienceProgress(0)
	})
}

func (i *Instance) playerLen() int {
	var l int
	for range i.Players() {
		l++
	}

	return l + 1
}

func (i *Instance) CoolDown(u uuid.UUID) (*cooldown.CoolDown, bool) {
	i.cooldownsMu.RLock()
	defer i.cooldownsMu.RUnlock()

	cd, ok := i.cooldowns[u]
	return cd, ok
}

func (i *Instance) addToCooldownList(u uuid.UUID) {
	i.cooldownsMu.Lock()
	defer i.cooldownsMu.Unlock()

	i.cooldowns[u] = cooldown.NewCoolDown()
}

func (i *Instance) removeFromCooldownList(u uuid.UUID) {
	i.cooldownsMu.Lock()
	defer i.cooldownsMu.Unlock()

	delete(i.cooldowns, u)
}

func (i *Instance) MustCoolDown(u uuid.UUID, ifExists func(c *cooldown.CoolDown)) {
	c, ok := i.CoolDown(u)
	if !ok {
		return
	}

	ifExists(c)
}

func (i *Instance) Kit() kit.Playered {
	return i.k
}

func (i *Instance) Closed() bool {
	return i.closed.Load()
}

func (i *Instance) Open() {
	i.closed.Store(false)
}

func (i *Instance) Close(_ *world.Tx) {
	i.closed.Store(true)

	for p := range i.Players() {
		p.Messagef(text.Colourf("<red>This game mode has been closed.</red>"))
		lobby.TransferWithRoutineSimple(p.Player)
	}
}

func (i *Instance) Name() string {
	return i.c.Name
}

func (i *Instance) Icon() string {
	return i.c.Icon
}

func LookupPlayer(id uuid.UUID) (*instance2.Player, *Instance) {
	for _, i := range instance2.AllInstances() {
		if in, isFFA := i.(*Instance); isFFA {
			if in.Active(id) {
				for pl := range in.Players() {
					if pl.UUID() == id {
						return pl, in
					}
				}
				break
			}
		}
	}

	return nil, nil
}
