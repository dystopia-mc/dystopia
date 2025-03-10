package handlers

import (
	"bytes"
	_ "embed"
	"fmt"
	"github.com/bedrock-gophers/cooldown/cooldown"
	"github.com/df-mc/atomic"
	"github.com/df-mc/dragonfly/server/block/cube"
	"github.com/df-mc/dragonfly/server/cmd"
	"github.com/df-mc/dragonfly/server/entity"
	"github.com/df-mc/dragonfly/server/entity/effect"
	"github.com/df-mc/dragonfly/server/item"
	"github.com/df-mc/dragonfly/server/player"
	"github.com/df-mc/dragonfly/server/player/skin"
	"github.com/df-mc/dragonfly/server/player/title"
	"github.com/df-mc/dragonfly/server/session"
	"github.com/df-mc/dragonfly/server/world"
	"github.com/df-mc/dragonfly/server/world/sound"
	"github.com/df-mc/npc"
	"github.com/go-gl/mathgl/mgl64"
	"github.com/google/uuid"
	plugin "github.com/k4ties/df-plugin/df-plugin"
	"github.com/k4ties/dystopia/internal"
	"github.com/k4ties/dystopia/internal/embeddable"
	"github.com/k4ties/dystopia/plugins/practice/src/bot"
	"github.com/k4ties/dystopia/plugins/practice/src/custom/items"
	"github.com/k4ties/dystopia/plugins/practice/src/ffa"
	"github.com/k4ties/dystopia/plugins/practice/src/handlers/knockback"
	"github.com/k4ties/dystopia/plugins/practice/src/instance"
	"github.com/k4ties/dystopia/plugins/practice/src/instance/lobby"
	"github.com/k4ties/dystopia/plugins/practice/src/kit"
	"github.com/k4ties/dystopia/plugins/practice/src/rank"
	"github.com/k4ties/dystopia/plugins/practice/src/user"
	"github.com/k4ties/dystopia/plugins/practice/src/user/chatter"
	"github.com/sandertv/gophertunnel/minecraft/protocol/packet"
	"github.com/sandertv/gophertunnel/minecraft/text"
	"github.com/sasha-s/go-deadlock"
	"regexp"
	"strings"
	"time"

	_ "unsafe"
)

func newPractice(i instance.Instance, u *user.User, p *player.Player) *Practice {
	return &Practice{i: i, u: u, s: NewSettings(u, p)}
}

var pool = struct {
	v  map[uuid.UUID]*Practice
	mu deadlock.RWMutex
}{
	v: make(map[uuid.UUID]*Practice),
}

func MustPractice(id uuid.UUID, p *player.Player, i instance.Instance) *Practice {
	// firstly search to existing handler
	pool.mu.RLock()
	if p, ok := pool.v[id]; ok {
		pool.mu.RUnlock()
		return p
	}
	pool.mu.RUnlock()

	// if there's no handler, we're creating a new one
	pool.mu.Lock()
	defer pool.mu.Unlock()

	var u *user.User

	user.P().MustUser(p, func(usr *user.User) {
		u = usr
	})

	if u == nil {
		return nil
	}

	pr := newPractice(i, u, p)
	pool.v[id] = pr

	return pr
}

func RemovePractice(id uuid.UUID) {
	pool.mu.Lock()
	defer pool.mu.Unlock()

	delete(pool.v, id)
}

type Practice struct {
	i instance.Instance // default instance

	lastChattedAt atomic.Value[time.Time]
	lastCommandAt atomic.Value[time.Time]

	lastMessage atomic.Value[string]

	u *user.User
	plugin.NopPlayerHandler

	s *Settings
}

const (
	JoinFormat = "<dark-grey>(<green>+</green>)</dark-grey> %s"
	QuitFormat = "<dark-grey>(<red>-</red>)</dark-grey> %s"
)

func globalJoinOrQuitMessage(format string, r rank.Rank, pl *player.Player) {
	msg := text.Colourf(format, r.Format+pl.Name()+text.Reset)

	user.MustUsePool(func(p *user.Pool) {
		for _, u := range p.OnlineUsers() {
			if p, ok := u.Player(); ok {
				if u.Settings().Global.ShowJoinAndQuitMessage {
					p.Message(msg)
				}
			}
		}
	})

	fmt.Println(text.ANSI(msg))
}

func (pr *Practice) HandleScoreTagTask(_ float64, cps int, p *player.Player) {
	if cps >= 1 {
		go p.H().ExecWorld(func(tx *world.Tx, e world.Entity) {
			e.(*player.Player).SetScoreTag(text.Colourf("<grey>CPS: <white>%d</white></grey>", cps))
		})
	}
}

//go:embed default_skin.png
var defaultSkinData []byte
var defaultSkin skin.Skin

func init() {
	texture, err := npc.ReadTexture(bytes.NewBuffer(defaultSkinData))
	if err != nil {
		panic(err)
	}

	sk, err := npc.Skin(texture, npc.DefaultModel)
	if err != nil {
		panic(err)
	}

	defaultSkin = sk
}

func (pr *Practice) HandleSpawn(pl *player.Player) {
	user.MustUsePool(func(p *user.Pool) {
		p.MustUser(pl, func(u *user.User) {
			pr.u = u
			r := u.Data().Rank()

			if !r.DisplayRankName {
				r.Format = text.Grey
			}
			globalJoinOrQuitMessage(JoinFormat, r, pl)

			go pl.H().ExecWorld(func(tx *world.Tx, e world.Entity) {
				p := e.(*player.Player)
				p.SetNameTag(r.Format + pl.Name())

				if p.Data().Session.ClientData().TrustedSkin {
					p.SetSkin(defaultSkin)
				}
			})

			if c, ok := plugin.M().Conn(pl.Name()); ok {
				instance.FadeInCamera(c, 1.5, false)
			}

			welcomePlayer(pl)
			_ = pr.spawnRoutine(pl)
		})
	})
}

func (pr *Practice) spawnRoutine(p *player.Player) *instance.Player {
	// player must be in lobby instance on spawn
	pl, in := instance.LookupPlayer(p.UUID())
	if pl == nil || in == nil {
		pl = instance.NewPlayer(p, lobby.Instance())
	}

	pl.ExecSafe(func(p *player.Player, tx *world.Tx) {
		pr.i.Transfer(pl, tx)
	})

	if pr.i.Active(p.UUID()) {
		pl.ExecSafe(func(p *player.Player, tx *world.Tx) {
			p.Teleport(pr.i.(instance.WorldSpawn).WorldSpawn())
		})
	}

	pl.SendKit(func(p *player.Player) kit.Kit {
		return lobby.Kit
	}, nil)
	return pl
}

func (pr *Practice) HandleSkinChange(ctx *player.Context, _ *skin.Skin) {
	ctx.Val().Messagef(text.Colourf("<green>Skin successfully applied.</green>"))
}

func (pr *Practice) HandleMove(ctx *player.Context, newPos mgl64.Vec3, _ cube.Rotation) {
	if p, i := instance.LookupPlayer(ctx.Val().UUID()); p != nil && i != nil {
		if p.Frozen() {
			ctx.Cancel()
			return
		}

		if i := p.Instance(); i != instance.Nop {
			if i.HeightThresholdEnabled() {
				if newPos.Y() <= float64(i.HeightThreshold()) && p.GameMode() == i.GameMode() && !p.Transferring() {
					switch i.HeightThresholdMode() {
					case instance.EventTeleportToSpawn:
						i.Transfer(p, ctx.Val().Tx())
						pr.spawnRoutine(ctx.Val())
					case instance.EventDeath:
						pr.HandleDeath(ctx.Val(), internal.IntersectThresholdSource{}, new(bool))
					}
				}
			}
		}
	}
}

func (pr *Practice) HandleQuit(p *player.Player) {
	bot.Log(fmt.Sprintf("%s has quited", p.Name()))
	r := pr.u.Data().Rank()

	if !r.DisplayRankName {
		r.Format = text.Grey
	}

	globalJoinOrQuitMessage(QuitFormat, r, p)

	if fpl, in := ffa.LookupPlayer(p.UUID()); fpl != nil && in != nil {
		//in.ResetPearlCooldown(fpl)
		if c, ok := in.Combat(p.UUID()); ok {
			if c.Active() && c.With() != nil {
				if killerPl, killerIn := ffa.LookupPlayer(c.With().UUID()); killerPl != nil && killerIn != nil {
					Kill(p, killerPl.Name(), killerPl, killerIn, internal.Count[items.HealingPotion](p.Inventory()), internal.Count[item.GoldenApple](p.Inventory()), false)
					killerIn.StopCombat(killerPl.UUID())
				}
			}
		}
		//in.StopCombat(p.UUID())
	}

	if pl, in := instance.LookupPlayer(p.UUID()); pl != nil && in != nil {
		if i := pl.Instance(); i != instance.Nop {
			instance.Nop.Transfer(pl, nil) // remove player from instance
		}
		defer pl.CloseTransactions()
	}
}

func (pr *Practice) HandleDeath(deadPlayer *player.Player, src world.DamageSource, keepInv *bool) {
	userPool := user.P()
	deadPlayer.Respawn()

	deathPosition := deadPlayer.Position()

	if pl, in := ffa.LookupPlayer(deadPlayer.UUID()); pl != nil && in != nil {
		in.CombatHandler(deadPlayer.UUID()).HandleDeath(deadPlayer, src, keepInv)
	}

	*keepInv = true
	deadPlayerPl, deadPlayerIn := ffa.LookupPlayer(deadPlayer.UUID()) // player can only be dead on ffa

	deadPlayerGapplesBeforeDeath := internal.Count[item.GoldenApple](deadPlayer.Inventory())
	deadPlayerPotsBeforeDeath := internal.Count[items.HealingPotion](deadPlayer.Inventory())

	if deadPlayerPl == nil || deadPlayerIn == nil {
		return
	}

	deadPlayerIn.ResetPearlCooldown(deadPlayerPl)

	var killerName = "..."
	var killerIsInstancePlayer = false

	var (
		killerPl *instance.Player
		killerIn *ffa.Instance
	)

	if a, ok := src.(entity.AttackDamageSource); ok {
		attackerP, pok := a.Attacker.(*player.Player)
		attackerId := attackerP.UUID()

		if pok {
			pl, in := ffa.LookupPlayer(attackerId)
			if pl != nil && in != nil {
				killerIsInstancePlayer = true
				killerName = pl.Name()

				killerPl = pl
				killerIn = in

				killerIn.StopCombat(killerPl.UUID())
			}
		}
	}

	var (
		deadPlayerU      *user.User
		deadPlayerIsUser = false
	)
	user.MustUsePool(func(p *user.Pool) {
		p.MustUser(deadPlayer, func(u *user.User) {
			deadPlayerU = u
			deadPlayerIsUser = true
		})
	})

	instantRespawn := deadPlayerU.Settings().FFA.InstantRespawn
	respawnOnArena := deadPlayerU.Settings().FFA.RespawnOnArena

	wait := time.Second * 4
	if instantRespawn {
		wait = time.Second
	}

	deadPlayerPl.Reset(nil, func(p *player.Player) {
		p.SetGameMode(internal.FlyingMode)
		p.SetInvisible()
		p.StartFlying()
		p.SetScale(0)
	}, func(p *player.Player) {
		t := title.New(text.Red + "WASTED")
		t = t.WithSubtitle(text.Grey + killerName)
		t = t.WithDuration(wait)
		p.SendTitle(t)
	})

	_, isIntersectThresholdCause := src.(internal.IntersectThresholdSource)
	if isIntersectThresholdCause {
		if c, ok := deadPlayerIn.Combat(deadPlayer.UUID()); ok && deadPlayerIn.InCombat(deadPlayer.UUID()) {
			if !killerIsInstancePlayer {
				if c.With() != nil && c.With().UUID() != uuid.Nil {
					if kPl, kIn := ffa.LookupPlayer(c.With().UUID()); kPl != nil && kIn != nil {
						killerIsInstancePlayer = true
						killerName = kPl.Name()

						killerPl = kPl
						killerIn = kIn

						killerIn.StopCombat(killerPl.UUID())
						killerIn.StopCombat(deadPlayer.UUID())

						deadPlayerIn.ResetPearlCooldown(deadPlayerPl)
						killerIn.ResetPearlCooldown(killerPl)
					}

					// player is not in combat, and player have intersected height threshold. So, that means he fell into the void.
					deadPlayerIn.Messagef(text.Colourf(voidFallFormat, deadPlayer.Name()))
				}
			}
		}
	}

	const (
		killStreakNotify = 10
	)

	userPool.MustUser(deadPlayer, func(u *user.User) {
		deadPlayerIsUser = true
		deadPlayerU = u

		u.Data().Dead()

		if u.KillStreak().Current() >= killStreakNotify {
			if killerIsInstancePlayer {
				killerIn.Messagef(KillStreakStopFormat, killerName, deadPlayer.Name(), u.KillStreak().Current())
			}
		}
		if u.KillStreak().Current() > 0 {
			u.KillStreak().ResetCurrent()
		}
	})

	if killerIsInstancePlayer {
		if killerPl != nil {
			userPool.MustUser(killerPl.Player, func(u *user.User) {
				u.Data().Kill()
				u.KillStreak().Kill()

				if u.KillStreak().Current()%killStreakNotify == 0 {
					// handle kill streak
					killerIn.Messagef(KillStreakFormat, killerName, u.KillStreak().Current())
				}
			})
		}

		Kill(deadPlayer, killerName, killerPl, killerIn, deadPlayerPotsBeforeDeath, deadPlayerGapplesBeforeDeath, isIntersectThresholdCause)
	}

	if !deadPlayerIn.InCombat(deadPlayer.UUID()) && isIntersectThresholdCause {
		deadPlayerIn.Messagef(text.Colourf(voidFallFormat, deadPlayer.Name()))
	}

	deadPlayerPl.ExecSafe(func(p *player.Player, tx *world.Tx) {
		lightning := entity.NewLightningWithDamage(world.EntitySpawnOpts{Position: deathPosition}, -1, false, -1)
		tx.AddEntity(lightning)
	})

	if deadPlayerIsUser {
		go func() {
			timer := time.NewTimer(wait)
			defer timer.Stop()
			<-timer.C

			if deadPlayerU.Online() {
				if !respawnOnArena {
					if pl, in := instance.LookupPlayer(deadPlayerPl.UUID()); pl != nil && in != nil {
						lobby.TransferWithRoutineSimple(deadPlayer)
						return
					}
				}
				deadPlayerPl.ExecSafe(func(p *player.Player, tx *world.Tx) {
					deadPlayerIn.Transfer(deadPlayerPl, tx)
					p.Teleport(deadPlayerIn.WorldSpawn())
				})
			}
		}()
	}
}

func (pr *Practice) HandlePunchAir(ctx *player.Context) {
	if pl, in := instance.LookupPlayer(ctx.Val().UUID()); pl != nil && in != nil {
		if pl.Frozen() {
			ctx.Cancel()
			return
		}
	}
	if pl, in := ffa.LookupPlayer(ctx.Val().UUID()); pl != nil && in != nil {
		pr.HandleScoreTagTask(ctx.Val().Health(), pr.u.CPS(), ctx.Val())
	}
}

const (
	KillStreakFormat     = "<red>%s</red> now has <red>%d</red> kills in a row"
	KillStreakStopFormat = "<red>%s</red> ruined <red>%s</red> <grey>%d</grey> kill streak"

	voidFallFormat         = "<red>%s</red> fell into the void"
	voidFallInCombatFormat = "<red>%s</red> fell into the void while running from <red>%s</red>"
)

func Kill(deadPlayer *player.Player, killerName string, killerPl *instance.Player, killerIn *ffa.Instance, deadPlayerPotsBeforeDeath, deadPlayerGapplesBeforeDeath int, isIntersectThresholdCause bool) {
	var msg = internal.FormatDeadMessage(internal.DeadMessageBlank, deadPlayer.Name(), killerName)

	if internal.KitIncludes[items.HealingPotion](killerIn.Kit()(killerPl.Player)) {
		msg = internal.FormatDeadMessage(internal.DeadMessageCount, deadPlayer.Name(), deadPlayerPotsBeforeDeath, "POTS", killerName, internal.Count[items.HealingPotion](killerPl.Inventory()), "POTS")
	} else if internal.KitIncludes[item.GoldenApple](killerIn.Kit()(killerPl.Player)) {
		msg = internal.FormatDeadMessage(internal.DeadMessageCount, deadPlayer.Name(), deadPlayerGapplesBeforeDeath, "GAPPLES", killerName, internal.Count[item.GoldenApple](killerPl.Inventory()), "GAPPLES")
	} else if killerIn.InCombat(killerPl.UUID()) {
		msg = text.Colourf(voidFallInCombatFormat, deadPlayer.Name(), killerName)
	} else if isIntersectThresholdCause && killerIn.InCombat(killerPl.UUID()) {
		msg = text.Colourf(voidFallFormat, deadPlayer.Name())
	}

	killerIn.Messagef(msg)
	reKit(killerIn, killerPl.UUID(), nil)
	killRoutine(killerPl.Player, deadPlayer)
}

func killRoutine(p *player.Player, deadPlayer *player.Player) {
	t := title.New(text.Colourf("<red>KILL!</red>"))
	t = t.WithDuration(time.Second * 2)
	t = t.WithSubtitle(text.Colourf("<grey>%s</grey>", deadPlayer.Name()))

	p.SendTitle(t)
	p.PlaySound(sound.LevelUp{})
}

//go:linkname reKit github.com/k4ties/dystopia/plugins/practice/src/ffa.(*Instance).reKit
func reKit(_ *ffa.Instance, _ uuid.UUID, _ *world.Tx)

func (pr *Practice) HandleItemUseOnBlock(ctx *player.Context, _ cube.Pos, _ cube.Face, _ mgl64.Vec3) {
	if pl, in := instance.LookupPlayer(ctx.Val().UUID()); pl != nil && in != nil {
		if pl.Frozen() {
			ctx.Cancel()
			return
		}
	}
	if notCreative(ctx.Val()) {
		ctx.Cancel()
	}
}

func (pr *Practice) HandleItemUse(ctx *player.Context) {
	if pl, in := instance.LookupPlayer(ctx.Val().UUID()); pl != nil && in != nil {
		if pl.Frozen() {
			ctx.Cancel()
			return
		}
	}
	reHandleItemUse(ctx, pr.u)
}

func notCreative(p *player.Player) bool {
	return p.GameMode() != world.GameModeCreative
}

func reHandleItemUse(ctx *player.Context, u *user.User) {
	p := ctx.Val()
	i, _ := ctx.Val().HeldItems()

	switch kit.LoadIdentifier(i) {
	case kit.PotIdentifier, kit.KnockerPearlIdentifier:
		// blank
	default:
		if notCreative(p) {
			ctx.Cancel()
		}
	case kit.SettingsIdentifier:
		if notCreative(p) {
			ctx.Cancel()
		}
		go ctx.Val().SendForm(user.NewSettingsForm(u))
	case kit.StatisticsIdentifier:
		if notCreative(p) {
			ctx.Cancel()
		}
		go ctx.Val().SendForm(u.NewStatisticsForm())
	case kit.FFAIdentifier:
		if notCreative(p) {
			ctx.Cancel()
		}
		go ctx.Val().SendForm(ffa.NewForm())
	case kit.PearlIdentifier:
		p := ctx.Val()
		p.SetCooldown(items.Pearl{}, -1)

		if pl, in := ffa.LookupPlayer(p.UUID()); pl != nil && in != nil {
			if in.HasPearCooldown() {
				in.MustCoolDown(pl.UUID(), func(c *cooldown.CoolDown) {
					if c.Active() {
						p.SendTip(text.Colourf("<red>Please wait %d more seconds to use ender pearl again", int(c.Remaining().Seconds())+1))
						ctx.Cancel()
						return
					}

					in.StartPearlCoolDown(pl, nil)
					p.SendTip(text.Colourf("<red>Ender Pearl cooldown has started</red>"))
				})
			}
		}
	}
}

func (pr *Practice) HandleHeal(ctx *player.Context, health *float64, src world.HealingSource) {
	if pl, in := instance.LookupPlayer(ctx.Val().UUID()); pl != nil && in != nil {
		if pl.Frozen() {
			ctx.Cancel()
			return
		}
	}

	if pl, in := ffa.LookupPlayer(ctx.Val().UUID()); pl != nil && in != nil {
		if _, ok := src.(effect.InstantHealingSource); ok {
			*health *= 2.5
		}

		pr.HandleScoreTagTask(*health, pr.u.CPS(), ctx.Val())
	}
}

func (pr *Practice) HandleClientPacket(ctx *player.Context, pk packet.Packet) {
	switch pk := pk.(type) {
	case *packet.MobEffect:
		if pk.Operation == packet.MobEffectRemove && pr.u.Settings().FFA.NightVision {
			if pk.EffectType == packet.EffectNightVision {
				ctx.Cancel()
			}
		}
	case *packet.ClientCacheStatus:
		pk.Enabled = false
	case *packet.PlayerAuthInput:
		if pr.u != nil {
			current := internal.InputMode(pk.InputMode)
			d := pr.u.Data()

			if pk.InputMode == packet.InputModeTouch {
				go ctx.Val().Disconnect(instance.KickMessage(instance.ErrorMobilizovan))
				return
			}

			if d.InputMode() != current {
				d.SwitchInputMode(current)
			}
		}
	case *packet.Text:
		p := ctx.Val()

		if pk.SourceName != p.Name() {
			pk.SourceName = p.Name()
		}
	case *packet.CommandRequest:
		p := ctx.Val()
		bot.Log(fmt.Sprintf("%s executed command line: %s", p.Name(), pk.CommandLine))
	}

	pr.s.HandleClientPacket(ctx, pk)
}

func (pr *Practice) HandleServerPacket(ctx *player.Context, pk packet.Packet) {
	pr.s.HandleServerPacket(ctx, pk)

	switch pk := pk.(type) {
	case *packet.ClientStartItemCooldown:
		if pk.Category == "ender_pearl" {
			pk.Duration = -1
		}
	}
}

var (
	chatCoolDown   = time.Second / 2
	characterLimit = 130
)

func (pr *Practice) HandleChat(ctx *player.Context, msg *string) {
	ctx.Cancel()

	if pr.u.Data().Rank().Priority <= rank.PriorityAdmin {
		if time.Since(pr.lastChattedAt.Load()) <= chatCoolDown {
			ctx.Val().Messagef(text.Colourf("<red>Please don't spam</red>"))
			bot.Log(fmt.Sprintf("Cancelled %s message: '%s'", ctx.Val().Name(), *msg))
			return
		}

		if len(*msg) > characterLimit {
			ctx.Val().Messagef(text.Colourf("<red>You've ran out of characters. (%d/%d)</red>", len(*msg), characterLimit))
			bot.Log(fmt.Sprintf("Cancelled %s message: '%s'", ctx.Val().Name(), *msg))
			return
		}
	}

	if notAlphaOnly(*msg) {
		bot.Log(fmt.Sprintf("Cancelled %s message: '%s'", ctx.Val().Name(), *msg))
		return
	}

	*msg = removeExtraSpaces(*msg)
	*msg = alphaReplacer(*msg)
	*msg = strings.TrimSpace(*msg)

	if pr.u.Data().Rank().Priority <= rank.PriorityAdmin {
		if pr.lastMessage.Load() == *msg {
			ctx.Val().Messagef(text.Colourf("<red>Please don't send identical messages</red>"))
			bot.Log(fmt.Sprintf("Cancelled %s message: '%s'", ctx.Val().Name(), *msg))
			return
		}
	}

	pr.lastChattedAt.Store(time.Now())

	if *msg != "" {
		r := pr.u.Data().Rank()
		f := r.Format

		t := f + r.Name + " "
		if rank.IsDefault(r) || !r.DisplayRankName {
			t = ""
			f = text.Grey
		}

		user.MustUsePool(func(p *user.Pool) {
			p.MustUser(ctx.Val(), func(u *user.User) {
				tag := ""

				if u.Data().HasTag() {
					tag = u.Data().ColouredTag() + text.Reset + " "
				}

				pr.lastMessage.Store(*msg)

				finalMsg := text.Colourf("%s%s:%s %s", tag+t+text.Reset, text.Grey+ctx.Val().Name(), text.Reset, *msg)
				chatter.Messagef(finalMsg)

				u.Data().WriteMessage(*msg)
			})
		})

	}
}

func notAlphaOnly(s string) bool {
	return len(alphaReplacer(s)) == 0
}

func alphaReplacer(s string) string {
	re := regexp.MustCompile(`[\s§<>]+`)
	return re.ReplaceAllString(s, " ")
}

func removeExtraSpaces(s string) string {
	words := strings.Fields(s)
	return strings.Join(words, " ")
}

func (pr *Practice) HandleCommandExecution(ctx *player.Context, cmd cmd.Command, arg []string) {
	if time.Since(pr.lastCommandAt.Load()) <= chatCoolDown && pr.u.Data().Rank().Priority < rank.PriorityAdmin {
		ctx.Val().Messagef(text.Colourf("<red>Please don't spam</red>"))
		ctx.Cancel()
		return
	}
	if pl, in := instance.LookupPlayer(ctx.Val().UUID()); pl != nil && in != nil {
		if pl.Frozen() {
			ctx.Val().Messagef(text.Colourf("<red>Cannot execute commands while you're frozen</red>"))
			ctx.Cancel()
			return
		}
	}
	if ctx.Val().GameMode() == internal.FlyingMode {
		ctx.Val().Messagef(text.Colourf("<red>Cannot execute commands while you're dead</red>"))
		ctx.Cancel()
		return
	}

	pr.lastCommandAt.Store(time.Now())
	pr.u.Data().WriteCommand(fmt.Sprintf("/%s %s", cmd.Name(), strings.Join(arg, " ")))
}

type kbConfig struct {
	KnockBack struct {
		Force    float64
		Height   float64
		Immunity int
	}
}

//go:embed knockback.json
var knockbackConfig []byte

func init() {
	conf := embeddable.MustJSON[kbConfig](knockbackConfig)
	knockback.Setup(conf.KnockBack.Height, conf.KnockBack.Force, int64(conf.KnockBack.Immunity))
}

func milliseconds(amount int) time.Duration {
	return time.Millisecond * time.Duration(amount)
}

func (pr *Practice) HandleHurt(ctx *player.Context, damage *float64, _ bool, immunity *time.Duration, src world.DamageSource) {
	attackSrc, isAttackSource := src.(entity.AttackDamageSource)
	proj, isProjectileSource := src.(entity.ProjectileDamageSource)

	attacker, isPlayer := attackSrc.Attacker.(*player.Player)

	if lobby.Instance().Active(ctx.Val().UUID()) {
		ctx.Cancel()
		return
	}

	if isPlayer && attacker.GameMode() == internal.FlyingMode {
		ctx.Cancel()
		return
	}

	if !isAttackSource && !isProjectileSource {
		ctx.Cancel()
		return
	}

	if isProjectileSource {
		p, ok := proj.Owner.(*player.Player)
		if !ok {
			ctx.Cancel()
			return
		} else {
			if p.GameMode() == internal.FlyingMode {
				ctx.Cancel()
				return
			}
		}
	}

	if pl, in := instance.LookupPlayer(ctx.Val().UUID()); pl != nil && in != nil {
		pl.MustConn(func(conn session.Conn) {
			_ = conn.WritePacket(&packet.CameraShake{
				Intensity: 4,
				Duration:  -1,
				Type:      packet.CameraShakeTypeRotational,
				Action:    0,
			})
		})
	}

	*immunity = milliseconds(int(knockback.Immunity()))

	if isPlayer && isCritical(attacker) {
		reHandleCritical(attacker, ctx.Val(), ctx.Val().Tx())
		*damage *= 1.5
	}

	if pl, in := ffa.LookupPlayer(ctx.Val().UUID()); pl != nil && in != nil {
		in.CombatHandler(ctx.Val().UUID()).HandleHurt(ctx, damage, false, immunity, src)
	}

	if *damage >= 1 {
		pr.HandleScoreTagTask(ctx.Val().Health(), pr.u.CPS(), ctx.Val())
	}
}

func (pr *Practice) HandleHeldSlotChange(ctx *player.Context, _ int, _ int) {
	if pl, in := instance.LookupPlayer(ctx.Val().UUID()); pl != nil && in != nil {
		if pl.Frozen() {
			ctx.Cancel()
			return
		}
	}
}

func (pr *Practice) HandleAttackEntity(ctx *player.Context, attacked world.Entity, force, height *float64, critical *bool) {
	*critical = false // prevents double hit 🤡

	if pl, in := instance.LookupPlayer(ctx.Val().UUID()); pl != nil && in != nil {
		if pl.Frozen() {
			ctx.Cancel()
			return
		}
	}

	if pl, in := ffa.LookupPlayer(ctx.Val().UUID()); pl != nil && in != nil {
		pr.HandleScoreTagTask(ctx.Val().Health(), pr.u.CPS(), ctx.Val())
		if !isCritical(ctx.Val()) && pr.u.Settings().FFA.CriticalHits {
			ctx.Val().Data().Session.ViewEntityAction(attacked, entity.CriticalHitAction{})
		}

		if in.CombatHandler(ctx.Val().UUID()).HandleAttackEntity(ctx, attacked, force, height, critical) {
			*force = knockback.Force()
			*height = knockback.Height()

			i, _ := ctx.Val().HeldItems()
			if kit.LoadIdentifier(i) == kit.KBSwordIdentifier {
				*force = knockback.Force() * 5
				*height = knockback.Height() / 5
			}
		}
		return
	}

	if lobby.Instance().Active(ctx.Val().UUID()) {
		if ap, ok := attacked.(*player.Player); ok {
			if lobby.Instance().Active(ap.UUID()) {
				ap.KnockBack(ctx.Val().Position(), 0.25, 0.25)
			}
		}
	}
}

func reHandleCritical(p *player.Player, attacked world.Entity, tx *world.Tx) {
	for _, v := range tx.Viewers(p.Position()) {
		v.ViewEntityAction(attacked, entity.CriticalHitAction{})
	}
}

func isCritical(p *player.Player) bool {
	_, slowFalling := p.Effect(effect.SlowFalling)
	_, blind := p.Effect(effect.Blindness)

	return !p.Sprinting() && !p.Flying() && p.FallDistance() > 0 && !slowFalling && !blind
}

func welcomePlayer(pl *player.Player) {
	_ = pl.SetHeldSlot(4)

	welcomeTitle := title.New(text.Colourf("<red><b>Dystopia</b></red>"))
	welcomeTitle = welcomeTitle.WithFadeInDuration(time.Second * 2).WithDuration(time.Second).WithFadeOutDuration(time.Second)
	welcomeTitle = welcomeTitle.WithSubtitle(text.Colourf("<white>Welcome, <grey>%s</grey>!</white>", pl.Name()))

	pl.SendTitle(welcomeTitle)
	bot.Log(fmt.Sprintf("%s has joined", pl.Name()))
}
