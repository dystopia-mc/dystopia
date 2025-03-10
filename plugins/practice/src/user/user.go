package user

import (
	"context"
	atomic2 "github.com/df-mc/atomic"
	"github.com/df-mc/dragonfly/server/player"
	"github.com/df-mc/dragonfly/server/player/bossbar"
	"github.com/df-mc/dragonfly/server/session"
	"github.com/df-mc/dragonfly/server/world"
	"github.com/google/uuid"
	plugin "github.com/k4ties/df-plugin/df-plugin"
	"github.com/k4ties/dystopia/internal"
	"github.com/k4ties/dystopia/plugins/practice/src/instance"
	"github.com/k4ties/dystopia/plugins/practice/src/rank"
	"github.com/sandertv/gophertunnel/minecraft/protocol/packet"
	"github.com/sandertv/gophertunnel/minecraft/text"
	"github.com/sasha-s/go-deadlock"
	"sync/atomic"
	"time"
)

type online struct {
	is atomic.Bool
	p  atomic2.Value[*player.Player]

	cps *cps
}

type User struct {
	online *online
	ks     *KillStreak
	data   *data
}

func (u *User) HideEntity(e world.Entity) {
	if _, ok := u.Player(); !ok || !u.Online() {
		return
	}

	p, _ := u.Player()
	p.HideEntity(e)
}

func (u *User) ShowEntity(e world.Entity) {
	if _, ok := u.Player(); !ok || !u.Online() {
		return
	}

	p, _ := u.Player()
	p.ShowEntity(e)
}

func (u *User) CPS() int {
	if _, ok := u.Player(); !ok || !u.Online() {
		return -1
	}
	return u.online.cps.Amount()
}

func (u *User) SendTip(a ...any) {
	if p := u.online.p.Load(); p != nil {
		p.SendTip(a...)
	}
}

func (u *User) Name() string {
	if p := u.online.p.Load(); p != nil {
		return p.Name()
	}

	return ""
}

func (u *User) Messagef(s string, a ...any) {
	if p := u.online.p.Load(); p != nil {
		p.Messagef(s, a...)
	}
}

func (u *User) GameMode() world.GameMode {
	if p := u.online.p.Load(); p != nil {
		return p.GameMode()
	}

	return nil
}

func (u *User) Settings() Settings {
	return u.data.Settings()
}

func (u *User) UUID() uuid.UUID {
	if p := u.online.p.Load(); p != nil {
		return p.UUID()
	}

	return uuid.Nil
}

func (u *User) InputMode() internal.InputMode {
	if u.Online() {
		return u.Data().InputMode()
	}

	return 65535
}

func (u *User) OS() internal.OS {
	if u.Online() {
		return u.Data().OS()
	}

	return -1
}

func (u *User) Latency() time.Duration {
	if p := u.online.p.Load(); p != nil {
		return p.Latency()
	}

	return -1
}

func (u *User) SendBossBar(bar bossbar.BossBar) {
	if p := u.online.p.Load(); p != nil {
		p.SendBossBar(bar)
	}
}

func (u *User) RemoveBossBar() {
	if p := u.online.p.Load(); p != nil {
		p.RemoveBossBar()
	}
}

func (u *User) UpdateRank(new rank.Rank) {
	u.Data().setRank(new)

	if !u.Online() {
		MustUsePool(func(p *Pool) {
			if err := p.DB().Update(Offline(u)); err != nil {
				l := plugin.M().Logger()
				l.Error("failed to update rank", "player", u.Data().Name(), "error", err)
			}
		})
	} else {
		if p, ok := u.Player(); ok {
			// safe execution
			go p.H().ExecWorld(func(tx *world.Tx, e world.Entity) {
				p := e.(*player.Player)

				p.SetNameTag(new.Format + p.Name())
				p.Messagef(text.Colourf("<i><dark-grey>Your rank has been updated</dark-grey></i>"))
			})
		}
	}
}

func Offline(u *User) OfflineUser {
	return OfflineUser{
		Name: u.data.Name(),
		XUID: u.data.XUID(),
		IPs:  u.data.IPs(),
		DIDs: u.data.DIDs(),
		UUID: u.data.UUID(),
		Rank: u.data.Rank(),

		Tag: u.data.ColouredTag(),

		Kills:  u.data.Kills(),
		Deaths: u.data.Deaths(),

		Settings: u.data.Settings(),

		FirstJoin:  u.data.FirstJoin(),
		TimePlayed: u.data.Played(),

		KillStreak: struct{ Max, Current int }{
			Max: int(u.KillStreak().Max()), Current: int(u.KillStreak().Current()),
		},
	}
}

type data struct {
	name string
	xuid string
	uuid uuid.UUID

	ips    []string
	ipsMu  deadlock.RWMutex
	dids   []string
	didsMu deadlock.RWMutex

	lastMessages   []Message
	lastMessagesMu deadlock.RWMutex

	lastCommands   []Command
	lastCommandsMu deadlock.RWMutex

	played           atomic2.Duration
	playedTicker     *time.Ticker
	playedCancelFunc context.CancelFunc

	rank      atomic2.Value[rank.Rank]
	firstJoin time.Time

	fps atomic.Int64

	// settings

	directMessages         atomic.Bool
	showJoinAndQuitMessage atomic.Bool

	personalTime atomic2.Value[PersonalTime]
	showCPS      atomic.Bool
	showFPS      atomic.Bool

	instantRespawn   atomic.Bool
	respawnOnArena   atomic.Bool
	lightningKill    atomic.Bool
	showOpponentCPS  atomic.Bool
	showBossBar      atomic.Bool
	hideNonOpponents atomic.Bool
	smoothPearl      atomic.Bool

	nightVision atomic.Bool
	criticalHit atomic.Bool

	// settings

	inputMode atomic2.Value[internal.InputMode]
	os        atomic2.Value[internal.OS]

	kills  atomic.Int64
	deaths atomic.Int64

	tag atomic2.Value[string]
}

func (d *data) setFPS(new int) {
	d.fps.Store(int64(new))
}

func (d *data) FPS() int {
	return int(d.fps.Load())
}

func (d *data) SetDirectMessages(b bool) {
	d.directMessages.Store(b)
}

func (d *data) SetShowJoinQuitMessages(b bool) {
	d.showJoinAndQuitMessage.Store(b)
}

func (d *data) SetPersonalTime(p PersonalTime) {
	d.personalTime.Store(p)
}

func (d *data) SetShowCPS(b bool) {
	d.showCPS.Store(b)
}

func (d *data) SetShowFPS(b bool, p *player.Player) {
	d.showFPS.Store(b)

	if b {
		p.Messagef(text.Red + "You have enabled FPS counter. If it isn't working, you must enable client diagnostics in creator settings (need to log out)")
	}
}

type Message struct {
	WrittenAt time.Time
	Message   string
}

type Command struct {
	WrittenAt time.Time
	Line      string
}

func (d *data) WriteMessage(m string) {
	d.lastMessagesMu.Lock()
	defer d.lastMessagesMu.Unlock()
	if len(d.lastMessages) >= 100 {
		d.lastMessages = d.lastMessages[1:]
	}
	d.lastMessages = append(d.lastMessages, Message{
		WrittenAt: time.Now(),
		Message:   m,
	})
}

func (d *data) LastMessages() []Message {
	d.lastMessagesMu.RLock()
	defer d.lastMessagesMu.RUnlock()
	return d.lastMessages
}

func (d *data) WriteCommand(cmd string) {
	d.lastCommandsMu.Lock()
	defer d.lastCommandsMu.Unlock()
	if len(d.lastCommands) >= 100 {
		d.lastCommands = d.lastCommands[1:]
	}
	d.lastCommands = append(d.lastCommands, Command{
		WrittenAt: time.Now(),
		Line:      cmd,
	})
}

func (d *data) LastCommands() []Command {
	d.lastCommandsMu.RLock()
	defer d.lastCommandsMu.RUnlock()
	return d.lastCommands
}

func (d *data) SetInstantRespawn(b bool) {
	d.instantRespawn.Store(b)
}

func (d *data) SetRespawnOnArena(b bool) {
	d.respawnOnArena.Store(b)
}

func (d *data) SetLightningKill(b bool) {
	d.lightningKill.Store(b)
}

func (d *data) SetShowOpponentCPS(b bool) {
	d.showOpponentCPS.Store(b)
}

func (d *data) SetShowBossBar(b bool) {
	d.showBossBar.Store(b)
}

func (d *data) SetHideNonOpponents(b bool) {
	d.hideNonOpponents.Store(b)
}

func (d *data) SetSmoothPearl(b bool) {
	d.smoothPearl.Store(b)
}

func (d *data) SetNightVision(b bool) {
	d.nightVision.Store(b)

	if conn, ok := plugin.M().Conn(d.Name()); ok {
		if b {
			_ = conn.WritePacket(&packet.MobEffect{
				EntityRuntimeID: 1,
				Operation:       packet.MobEffectAdd,
				EffectType:      packet.EffectNightVision,
				Duration:        -1,
			})
		} else {
			_ = conn.WritePacket(&packet.MobEffect{
				EntityRuntimeID: 1,
				Operation:       packet.MobEffectRemove,
				EffectType:      packet.EffectNightVision,
			})
		}
	}
}

func (d *data) SetCriticalHit(b bool) {
	d.criticalHit.Store(b)
}

func (d *data) resetSettings(p *player.Player) {
	def := DefaultSettings()
	d.SetDirectMessages(def.Global.DirectMessages)
	d.SetShowJoinQuitMessages(def.Global.ShowJoinAndQuitMessage)

	d.SetPersonalTime(ParseTimeOfDay(def.Visual.PersonalTime))
	d.SetShowCPS(def.Visual.ShowCPS)
	d.SetShowFPS(def.Visual.ShowFPS, p)

	d.SetInstantRespawn(def.FFA.InstantRespawn)
	d.SetRespawnOnArena(def.FFA.RespawnOnArena)
	d.SetLightningKill(def.FFA.LightningKill)
	d.SetShowOpponentCPS(def.FFA.ShowOpponentCPS)
	d.SetShowBossBar(def.FFA.ShowBossBar)
	d.SetHideNonOpponents(def.FFA.HideNonOpponents)
	d.SetSmoothPearl(def.FFA.SmoothPearl)
	d.SetNightVision(def.FFA.NightVision)
	d.SetCriticalHit(def.FFA.CriticalHits)
}

func (d *data) Settings() Settings {
	s := Settings{}
	s.Global.DirectMessages = d.directMessages.Load()
	s.Global.ShowJoinAndQuitMessage = d.showJoinAndQuitMessage.Load()

	s.Visual.PersonalTime = d.personalTime.Load().Name()
	s.Visual.ShowCPS = d.showCPS.Load()
	s.Visual.ShowFPS = d.showFPS.Load()

	s.FFA.LightningKill = d.lightningKill.Load()
	s.FFA.ShowOpponentCPS = d.showOpponentCPS.Load()
	s.FFA.ShowBossBar = d.showBossBar.Load()
	s.FFA.RespawnOnArena = d.respawnOnArena.Load()
	s.FFA.InstantRespawn = d.instantRespawn.Load()
	s.FFA.HideNonOpponents = d.hideNonOpponents.Load()
	s.FFA.SmoothPearl = d.smoothPearl.Load()
	s.FFA.NightVision = d.nightVision.Load()
	s.FFA.CriticalHits = d.criticalHit.Load()
	return s
}

func (d *data) HasTag() bool {
	return d.Tag() != ""
}

func (d *data) ColouredTag() string {
	return d.tag.Load()
}

func (d *data) Tag() string {
	return text.Clean(d.tag.Load())
}

func (d *data) SetTag(tag string) {
	d.tag.Store(tag)
}

func (d *data) Played() time.Duration {
	return d.played.Load()
}

func (d *data) Kills() int {
	return int(d.kills.Load())
}

func (d *data) Kill() {
	d.kills.Add(1)
}

func (d *data) Dead() {
	d.deaths.Add(1)
}

func (d *data) Deaths() int {
	return int(d.deaths.Load())
}

func (d *data) SwitchInputMode(new internal.InputMode) {
	d.inputMode.Store(new)
}

func (d *data) setRank(new rank.Rank) {
	d.rank.Store(new)
}

func (d *data) InputMode() internal.InputMode {
	return d.inputMode.Load()
}

func (d *data) OS() internal.OS {
	return d.os.Load()
}

func (d *data) Name() string {
	return d.name
}

func (d *data) XUID() string {
	return d.xuid
}

func (d *data) UUID() uuid.UUID {
	return d.uuid
}

func (d *data) IPs() []string {
	d.ipsMu.RLock()
	defer d.ipsMu.RUnlock()
	return d.ips
}

func (d *data) addIp(ip string) {
	d.ipsMu.Lock()
	defer d.ipsMu.Unlock()
	d.ips = append(d.ips, ip)
}

func (d *data) DIDs() []string {
	d.didsMu.RLock()
	defer d.didsMu.RUnlock()
	return d.dids
}

func (d *data) addDID(did string) {
	d.didsMu.Lock()
	defer d.didsMu.Unlock()
	d.dids = append(d.dids, did)
}

func (d *data) Rank() rank.Rank {
	return d.rank.Load()
}

func (d *data) FirstJoin() time.Time {
	return d.firstJoin
}

func (d *data) startPlayedTicker() {
	if d.playedTicker != nil {
		return
	}

	ctx, cancel := context.WithCancel(context.Background())
	d.playedTicker = time.NewTicker(time.Second)

	go func() {
		defer d.playedTicker.Stop()
		for {
			select {
			case <-ctx.Done():
				return
			case <-d.playedTicker.C:
				d.played.Add(time.Second)
			}
		}
	}()

	d.playedCancelFunc = cancel
}

func (d *data) stopPlayedTicker() {
	if d.playedTicker != nil {
		d.playedTicker.Stop()
	}

	if d.playedCancelFunc != nil {
		d.playedCancelFunc()
	}
}

func (u *User) SetOnline(p *player.Player, c session.Conn) {
	u.online.p.Store(p)
	u.online.is.Store(true)

	u.data.os.Store(internal.OS(c.ClientData().DeviceOS))
	u.data.inputMode.Store(internal.InputMode(c.ClientData().CurrentInputMode))

	u.data.startPlayedTicker()
	if u.Settings().FFA.NightVision {
		conn, ok := plugin.M().Conn(p.Name())
		if !ok {
			instance.Kick(p, instance.ErrorSponge)
			return
		}

		_ = conn.WritePacket(&packet.MobEffect{
			EntityRuntimeID: 1,
			Operation:       packet.MobEffectAdd,
			EffectType:      packet.EffectNightVision,
			Amplifier:       1,
			Duration:        -1,
		})
	}
}

func (u *User) SetOffline() {
	u.online.is.Store(false)
	u.online.p.Store(nil)

	u.data.os.Store(-1)
	u.data.inputMode.Store(65535)
	u.data.fps.Store(-1)

	u.data.stopPlayedTicker()
}

func ValidOS(o internal.OS) bool {
	return o != -1
}

func ValidInputMode(o internal.InputMode) bool {
	return o != 65535
}

func (u *User) Online() bool {
	return u.online.is.Load()
}

func (u *User) Player() (*player.Player, bool) {
	if !u.Online() {
		return nil, false
	}

	p := u.online.p.Load()
	return p, p != nil
}

func (u *User) KillStreak() *KillStreak {
	return u.ks
}

func (u *User) Data() *data {
	return u.data
}

func FromOffline(o OfflineUser) *User {
	u := &User{
		online: &online{cps: &cps{}},
		ks:     &KillStreak{},
		data:   dataFromOfflineUser(o),
	}

	u.online.is.Store(false)
	u.ks.current.Store(int64(o.KillStreak.Current))
	u.ks.max.Store(int64(o.KillStreak.Max))

	u.data.fps.Store(-1)
	return u
}

func dataFromOfflineUser(o OfflineUser) *data {
	d := &data{
		name:      o.Name,
		xuid:      o.XUID,
		uuid:      o.UUID,
		ips:       o.IPs,
		dids:      o.DIDs,
		firstJoin: o.FirstJoin,
	}

	d.rank.Store(o.Rank)
	d.kills.Store(int64(o.Kills))
	d.deaths.Store(int64(o.Deaths))
	d.tag.Store(o.Tag)
	d.played.Store(o.TimePlayed)

	d.directMessages.Store(o.Settings.Global.DirectMessages)
	d.showJoinAndQuitMessage.Store(o.Settings.Global.ShowJoinAndQuitMessage)

	d.personalTime.Store(ParseTimeOfDay(o.Settings.Visual.PersonalTime))
	d.showCPS.Store(o.Settings.Visual.ShowCPS)
	d.showFPS.Store(o.Settings.Visual.ShowFPS)

	d.instantRespawn.Store(o.Settings.FFA.InstantRespawn)
	d.respawnOnArena.Store(o.Settings.FFA.RespawnOnArena)
	d.lightningKill.Store(o.Settings.FFA.LightningKill)
	d.showOpponentCPS.Store(o.Settings.FFA.ShowOpponentCPS)
	d.showBossBar.Store(o.Settings.FFA.ShowBossBar)
	d.hideNonOpponents.Store(o.Settings.FFA.HideNonOpponents)
	d.smoothPearl.Store(o.Settings.FFA.SmoothPearl)
	d.criticalHit.Store(o.Settings.FFA.CriticalHits)
	d.nightVision.Store(o.Settings.FFA.NightVision)

	d.fps.Store(-1)
	return d
}
