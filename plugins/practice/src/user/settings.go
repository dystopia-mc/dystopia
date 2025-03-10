package user

import (
	"github.com/df-mc/dragonfly/server/player"
	"github.com/df-mc/dragonfly/server/player/form"
	"github.com/df-mc/dragonfly/server/world"
	"github.com/sandertv/gophertunnel/minecraft/text"
	"sort"
)

func newSettingsButton(title string, img ...string) form.Button {
	if len(img) == 0 {
		img = []string{""}
	}
	return form.NewButton(text.Colourf("<white>%s</white>\n<dark-grey>Click to continue</dark-grey>", title), img[0])
}

func NewSettingsForm(u *User) form.Menu {
	return form.NewMenu(settingsMenuForm{
		u: u,

		Global:    newSettingsButton("Global"),
		Visual:    newSettingsButton("Visual"),
		Arena:     newSettingsButton("Arena"),
		Cosmetics: newSettingsButton("Cosmetics"),
		Close:     newSettingsButton("Close", "textures/ui/redX1.png"),
		Reset:     newSettingsButton("Reset", "textures/items/redstone_dust.png"),
	}, "Settings")
}

type settingsMenuForm struct {
	u *User

	Global    form.Button
	Visual    form.Button
	Arena     form.Button
	Cosmetics form.Button
	Close     form.Button
	Reset     form.Button
}

func (s settingsMenuForm) Submit(submitter form.Submitter, p form.Button, _ *world.Tx) {
	switch p.Text {
	case s.Global.Text:
		submitter.SendForm(newSettingsGlobal(s.u))
	case s.Visual.Text:
		submitter.SendForm(newSettingsVisual(s.u))
	case s.Arena.Text:
		submitter.SendForm(newSettingsFFA(s.u))
	case s.Cosmetics.Text:
		submitter.(*player.Player).Messagef(text.Red + "Coming soon")
	case s.Reset.Text:
		submitter.SendForm(newResetConfirmation(s.u))
	}
}

func newResetConfirmation(u *User) form.Modal {
	return form.NewModal(resetConfirmationForm{
		u: u,

		OK:     form.NewButton(text.Colourf("<green>Yes</green>"), ""),
		Cancel: form.NewButton(text.Colourf("<red>No</red>"), ""),
	}).WithBody("Are you sure that you want to reset settings?")
}

type resetConfirmationForm struct {
	u *User

	OK, Cancel form.Button
}

func (r resetConfirmationForm) Submit(submitter form.Submitter, pressed form.Button, tx *world.Tx) {
	switch pressed.Text {
	case r.OK.Text:
		r.u.Data().resetSettings(submitter.(*player.Player))
		r.u.Messagef(SettingsApplied)
	}
}

var SettingsApplied = text.Colourf("<green>Settings have been successfully applied.</green>")

func newDescriptionLabel(f string, a ...any) form.Label {
	return form.NewLabel(text.Colourf(text.Italic+text.Grey+f+text.Reset, a...))
}

func newSettingsGlobal(u *User) form.Form {
	return form.New(settingsFormGlobal{
		u: u,

		DirectMessages:            form.NewToggle("Direct messages", u.Data().Settings().Global.DirectMessages),
		DirectMessagesDescription: newDescriptionLabel("Allows/disallows someone to message you."),

		ShowJoinAndQuitMessages:            form.NewToggle("Join/Quit messages", u.Data().Settings().Global.ShowJoinAndQuitMessage),
		ShowJoinAndQuitMessagesDescription: newDescriptionLabel("Enables/disables join and quit messages."),
	}, "Global")
}

type settingsFormGlobal struct {
	u *User

	DirectMessages            form.Toggle
	DirectMessagesDescription form.Label

	ShowJoinAndQuitMessages            form.Toggle
	ShowJoinAndQuitMessagesDescription form.Label
}

func (s settingsFormGlobal) Submit(_ form.Submitter, _ *world.Tx) {
	s.u.Data().SetDirectMessages(s.DirectMessages.Value())
	s.u.Data().SetShowJoinQuitMessages(s.ShowJoinAndQuitMessages.Value())

	s.u.Messagef(SettingsApplied)
}

type settingsFormVisual struct {
	u *User

	PersonalTime            form.Dropdown
	PersonalTimeDescription form.Label

	ShowCPS            form.Toggle
	ShowCpsDescription form.Label

	ShowFPS            form.Toggle
	ShowFPSDescription form.Label
}

func newSettingsVisual(u *User) form.Form {
	return form.New(settingsFormVisual{
		u: u,

		PersonalTime:            form.NewDropdown("Personal time", allPersonalTimesNames(), personalTimeIndexByName(u.Data().Settings().Visual.PersonalTime)),
		PersonalTimeDescription: newDescriptionLabel("Changes in-game time only for you (visually)"),

		ShowCPS:            form.NewToggle("Show CPS", u.Data().Settings().Visual.ShowCPS),
		ShowCpsDescription: newDescriptionLabel("Shows cps in the bottom left edge of the screen"),

		ShowFPS:            form.NewToggle("Show FPS", u.Data().Settings().Visual.ShowFPS),
		ShowFPSDescription: newDescriptionLabel("Shows fps in the same spot as cps.\n<red>MUST ENABLE CLIENT DIAGNOSTICS IN THE CREATOR SETTINGS</red>"),
	}, "Visual")
}

func (s settingsFormVisual) Submit(su form.Submitter, _ *world.Tx) {
	s.u.Data().SetPersonalTime(personalTimeByIndex(s.PersonalTime.Value()))
	s.u.Data().SetShowCPS(s.ShowCPS.Value())
	s.u.Data().SetShowFPS(s.ShowFPS.Value(), su.(*player.Player))
	s.u.Messagef(SettingsApplied)
}

type settingsFormFFA struct {
	u *User

	InstantRespawn            form.Toggle
	InstantRespawnDescription form.Label

	RespawnOnArena            form.Toggle
	RespawnOnArenaDescription form.Label

	LightningKill            form.Toggle
	LightningKillDescription form.Label

	ShowOpponentCPS            form.Toggle
	ShowOpponentCPSDescription form.Label

	ShowBossBar            form.Toggle
	ShowBossBarDescription form.Label

	HideNonOpponents            form.Toggle
	HideNonOpponentsDescription form.Label

	SmoothPearl            form.Toggle
	SmoothPearlDescription form.Label

	NightVision            form.Toggle
	NightVisionDescription form.Label

	CriticalHit            form.Toggle
	CriticalHitDescription form.Label
}

func (s settingsFormFFA) Submit(_ form.Submitter, _ *world.Tx) {
	s.u.Data().SetInstantRespawn(s.InstantRespawn.Value())
	s.u.Data().SetRespawnOnArena(s.RespawnOnArena.Value())
	s.u.Data().SetLightningKill(s.LightningKill.Value())
	s.u.Data().SetShowOpponentCPS(s.ShowOpponentCPS.Value())
	s.u.Data().SetShowBossBar(s.ShowBossBar.Value())
	s.u.Data().SetHideNonOpponents(s.HideNonOpponents.Value())
	s.u.Data().SetSmoothPearl(s.SmoothPearl.Value())
	s.u.Data().SetNightVision(s.NightVision.Value())
	s.u.Data().SetCriticalHit(s.CriticalHit.Value())
	s.u.Messagef(SettingsApplied)
}

func newSettingsFFA(u *User) form.Form {
	return form.New(settingsFormFFA{
		u: u,

		InstantRespawn:            form.NewToggle("Instant respawn", u.Data().Settings().FFA.InstantRespawn),
		InstantRespawnDescription: newDescriptionLabel("Instead of waiting 4 seconds after death you will be immediately teleported to the lobby"),

		RespawnOnArena:            form.NewToggle("Respawn on arena", u.Data().Settings().FFA.RespawnOnArena),
		RespawnOnArenaDescription: newDescriptionLabel("After death you will be teleported not to the lobby, you will be teleported to the arena where you dead"),

		LightningKill:            form.NewToggle("Lightning kill", u.Data().Settings().FFA.LightningKill),
		LightningKillDescription: newDescriptionLabel("If disabled it will prevent lightning bolt to spawn after kill"),

		ShowOpponentCPS:            form.NewToggle("Show opponent CPS", u.Data().Settings().FFA.ShowOpponentCPS),
		ShowOpponentCPSDescription: newDescriptionLabel("Shows opponent CPS below his name"),

		ShowBossBar:            form.NewToggle("Show combat boss-bar", u.Data().Settings().FFA.ShowBossBar),
		ShowBossBarDescription: newDescriptionLabel("If disabled, will not show boss-bar in combat"),

		HideNonOpponents:            form.NewToggle("Hide non-opponents", u.Data().Settings().FFA.HideNonOpponents),
		HideNonOpponentsDescription: newDescriptionLabel("If you're in combat, all of the players (excluding your opponent) will be hidden"),

		SmoothPearl:            form.NewToggle("Smooth ender-pearl", u.Data().Settings().FFA.SmoothPearl),
		SmoothPearlDescription: newDescriptionLabel("If enabled, will teleport you smoothly when you throw ender pearl"),

		NightVision:            form.NewToggle("Night vision", u.Data().Settings().FFA.NightVision),
		NightVisionDescription: newDescriptionLabel("Toggles night vision effect, this setting also known as 'FullBright'"),

		CriticalHit:            form.NewToggle("Critical hits", u.Data().Settings().FFA.CriticalHits),
		CriticalHitDescription: newDescriptionLabel("Will spawn critical hit even if it is not critical hit"),
	}, "Arena")
}

func allPersonalTimes() map[string]PersonalTime {
	return map[string]PersonalTime{
		"day":      PersonalTimeDay,
		"midnight": PersonalTimeMidnight,
		"night":    PersonalTimeNight,
		"sunset":   PersonalTimeSunset,
		"sunrise":  PersonalTimeSunrise,
		"noon":     PersonalTimeNoon,
	}
}

func personalTimeIndexByName(name string) int {
	i := 0

	for index, n := range allPersonalTimesNames() {
		if n == name {
			i = index
			break
		}
	}

	return i
}

func personalTimeByIndex(i int) PersonalTime {
	ind := 0

	for index := range allPersonalTimesNames() {
		if index == i {
			ind = index
			break
		}
	}

	return ParseTimeOfDay(allPersonalTimesNames()[ind])
}

func allPersonalTimesNames() []string {
	var names []string

	for name := range allPersonalTimes() {
		names = append(names, name)
	}

	sort.Strings(names)
	return names
}

type PersonalTime int

func (v PersonalTime) Name() string {
	switch v {
	case PersonalTimeDay:
		return "day"
	case PersonalTimeMidnight:
		return "midnight"
	case PersonalTimeNight:
		return "night"
	case PersonalTimeSunset:
		return "sunset"
	case PersonalTimeSunrise:
		return "sunrise"
	case PersonalTimeNoon:
		return "noon"
	default:
		return "unknown"
	}
}

const (
	PersonalTimeDay      PersonalTime = 25000
	PersonalTimeMidnight PersonalTime = 42000
	PersonalTimeNight    PersonalTime = 61000
	PersonalTimeSunset   PersonalTime = 84000
	PersonalTimeSunrise  PersonalTime = 95000
	PersonalTimeNoon     PersonalTime = 102000
)

func ParseTimeOfDay(s string) PersonalTime {
	switch s {
	default:
		fallthrough
	case "day":
		return PersonalTimeDay
	case "midnight":
		return PersonalTimeMidnight
	case "night":
		return PersonalTimeNight
	case "sunset":
		return PersonalTimeSunset
	case "sunrise":
		return PersonalTimeSunrise
	case "noon":
		return PersonalTimeNoon
	}
}

type Settings struct {
	Global struct {
		DirectMessages         bool
		ShowJoinAndQuitMessage bool
	}
	Visual struct {
		PersonalTime string
		ShowCPS      bool
		ShowFPS      bool
	}
	FFA struct {
		InstantRespawn   bool
		RespawnOnArena   bool
		LightningKill    bool
		ShowOpponentCPS  bool
		ShowBossBar      bool
		HideNonOpponents bool
		SmoothPearl      bool
		NightVision      bool
		CriticalHits     bool
	}
}

func DefaultSettings() Settings {
	s := Settings{}
	s.Global.DirectMessages = true
	s.Global.ShowJoinAndQuitMessage = true

	s.Visual.PersonalTime = "day"
	s.Visual.ShowCPS = false
	s.Visual.ShowFPS = false

	s.FFA.InstantRespawn = false
	s.FFA.RespawnOnArena = false
	s.FFA.LightningKill = true
	s.FFA.ShowOpponentCPS = true
	s.FFA.ShowBossBar = true
	s.FFA.HideNonOpponents = true
	s.FFA.SmoothPearl = false
	s.FFA.CriticalHits = false
	s.FFA.NightVision = false
	return s
}
