package user

import (
	"encoding/json"
	"github.com/google/uuid"
	"github.com/k4ties/dystopia/internal"
	"github.com/k4ties/dystopia/plugins/practice/src/rank"
	"strings"
	"time"
)

type OfflineUser struct {
	Name string
	XUID string

	IPs  []string
	DIDs []string

	UUID uuid.UUID
	Rank rank.Rank

	Deaths, Kills int
	TimePlayed    time.Duration

	FirstJoin time.Time
	Tag       string

	Settings Settings

	KillStreak struct {
		Max, Current int
	}
}

// rawOfflineUser is used for database to exclude all hard types
type rawOfflineUser struct {
	Name      string `gorm:"column:name"`
	XUID      string `gorm:"column:xuid"`
	UUID      string `gorm:"column:uuid"`
	Rank      string `gorm:"column:rank"`
	FirstJoin int64  `gorm:"column:first_join"`

	IPs  string `gorm:"column:ips"`
	DIDs string `gorm:"column:device_ids"`

	Deaths int64 `gorm:"column:deaths"`
	Kills  int64 `gorm:"column:kills"`

	TimePlayed string `gorm:"column:time_played"`

	KillStreakCurrent int `gorm:"column:kill_streak_current"`
	KillStreakMax     int `gorm:"column:kill_streak_max"`

	Tag string `gorm:"column:tag"`

	//
	// Settings
	//
	// for some reason gorm cant work with booleans, so I'll use bool strings instead

	DirectMessages          string `gorm:"column:direct_messages"`
	ShowJoinAndQuitMessages string `gorm:"column:show_join_quit_messages"`

	PersonalTime string `gorm:"column:personal_time"`
	ShowCPS      string `gorm:"column:show_cps"`
	ShowFPS      string `gorm:"column:show_fps"`

	InstantRespawn   string `gorm:"column:instant_respawn"`
	RespawnOnArena   string `gorm:"column:respawn_on_arena"`
	LightningKill    string `gorm:"column:lightning_kill"`
	ShowOpponentCPS  string `gorm:"column:show_opponent_cps"`
	ShowBossBar      string `gorm:"column:show_boss_bar"`
	HideNonOpponents string `gorm:"column:hide_non_opponents"`
	SmoothPearl      string `gorm:"column:smooth_pearl"`
	NightVision      string `gorm:"column:night_vision"`
	CriticalHits     string `gorm:"column:critical_hits"`
}

func (rawOfflineUser) json() []byte {
	j, err := json.Marshal(rawOfflineUser{})
	if err != nil {
		return nil
	}

	return j
}

func (r rawOfflineUser) offlineUser() OfflineUser {
	dur, err := internal.ParseDuration(r.TimePlayed)
	if err != nil {
		dur = time.Duration(0)
	}

	u := OfflineUser{
		Name: r.Name,
		XUID: r.XUID,
		UUID: uuid.MustParse(r.UUID),
		Rank: rank.MustByName(r.Rank),

		Tag: r.Tag,

		IPs:  strings.Split(r.IPs, ","),
		DIDs: strings.Split(r.DIDs, ","),

		Deaths: int(r.Deaths),
		Kills:  int(r.Kills),

		TimePlayed: dur,
		FirstJoin:  time.Unix(r.FirstJoin, 0),

		Settings: settingsFromRaw(r),
	}

	u.KillStreak.Current = r.KillStreakCurrent
	u.KillStreak.Max = r.KillStreakMax

	return u
}

func (u OfflineUser) raw() rawOfflineUser {
	return rawOfflineUser{
		Name: u.Name,
		Rank: strings.ToLower(u.Rank.Name),

		XUID: u.XUID,
		UUID: u.UUID.String(),

		FirstJoin: u.FirstJoin.Unix(),

		IPs:  strings.Join(u.IPs, ","),
		DIDs: strings.Join(u.DIDs, ","),

		Deaths: int64(u.Deaths),
		Kills:  int64(u.Kills),

		TimePlayed: u.TimePlayed.String(),

		KillStreakCurrent: u.KillStreak.Current,
		KillStreakMax:     u.KillStreak.Max,

		Tag: u.Tag,

		ShowJoinAndQuitMessages: formatBool(u.Settings.Global.ShowJoinAndQuitMessage),
		DirectMessages:          formatBool(u.Settings.Global.DirectMessages),

		PersonalTime: u.Settings.Visual.PersonalTime,
		ShowCPS:      formatBool(u.Settings.Visual.ShowCPS),
		ShowFPS:      formatBool(u.Settings.Visual.ShowFPS),

		InstantRespawn:   formatBool(u.Settings.FFA.InstantRespawn),
		RespawnOnArena:   formatBool(u.Settings.FFA.RespawnOnArena),
		LightningKill:    formatBool(u.Settings.FFA.LightningKill),
		ShowOpponentCPS:  formatBool(u.Settings.FFA.ShowOpponentCPS),
		ShowBossBar:      formatBool(u.Settings.FFA.ShowBossBar),
		HideNonOpponents: formatBool(u.Settings.FFA.HideNonOpponents),
		SmoothPearl:      formatBool(u.Settings.FFA.SmoothPearl),
		NightVision:      formatBool(u.Settings.FFA.NightVision),
		CriticalHits:     formatBool(u.Settings.FFA.CriticalHits),
	}
}

func settingsFromRaw(raw rawOfflineUser) Settings {
	s := Settings{}
	s.Global.DirectMessages = parseBool(raw.DirectMessages)
	s.Global.ShowJoinAndQuitMessage = parseBool(raw.ShowJoinAndQuitMessages)

	s.Visual.PersonalTime = raw.PersonalTime
	s.Visual.ShowCPS = parseBool(raw.ShowCPS)
	s.Visual.ShowFPS = parseBool(raw.ShowFPS)

	s.FFA.ShowOpponentCPS = parseBool(raw.ShowOpponentCPS)
	s.FFA.LightningKill = parseBool(raw.LightningKill)
	s.FFA.RespawnOnArena = parseBool(raw.RespawnOnArena)
	s.FFA.InstantRespawn = parseBool(raw.InstantRespawn)
	s.FFA.ShowBossBar = parseBool(raw.ShowBossBar)
	s.FFA.HideNonOpponents = parseBool(raw.HideNonOpponents)
	s.FFA.SmoothPearl = parseBool(raw.SmoothPearl)
	s.FFA.NightVision = parseBool(raw.NightVision)
	s.FFA.CriticalHits = parseBool(raw.CriticalHits)
	return s
}

func parseBool(s string) bool {
	if s == "true" {
		return true
	}

	return false
}

func formatBool(b bool) string {
	if b {
		return "true"
	}

	return "false"
}
