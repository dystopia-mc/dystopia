package command

import (
	"github.com/df-mc/dragonfly/server/cmd"
	"github.com/df-mc/dragonfly/server/player"
	"github.com/df-mc/dragonfly/server/world"
	"github.com/k4ties/dystopia/plugins/practice/src/user"
	"github.com/sandertv/gophertunnel/minecraft/text"
)

type Gm struct {
	onlyAdmin
	Mode   GameMode                 `cmd:"GameMode"`
	Target cmd.Optional[PlayerEnum] `cmd:"target"`
}

func (c Gm) Run(src cmd.Source, o *cmd.Output, tx *world.Tx) {
	mode, modeName := parseGameMode(string(c.Mode))
	target, hasTarget := c.Target.Load()

	if !hasTarget {
		// change game mode to yourself
		p(src).SetGameMode(mode)
		p(src).Messagef(text.Colourf("<red><b>>></b></red> Your game mode is now <grey>%s</grey>", modeName))
		return
	}

	usr, ok := user.P().UserByName(string(target))
	if !ok {
		o.Errorf("Cannot find player: %s", target)
		return
	}

	if _, ok := usr.Player(); !ok || !usr.Online() {
		o.Errorf("Cannot find player: %s", target)
		return
	}

	pl, _ := usr.Player()
	go pl.H().ExecWorld(func(tx *world.Tx, e world.Entity) {
		e.(*player.Player).SetGameMode(mode)
	})

	pl.Messagef(text.Colourf("<red><b>>></b></red> Your game mode is now <grey>%s</grey>", modeName))
	systemMessage(o, "You've successfully changed %s game mode to <grey>%s</grey>", target, modeName)
}

type GmConsole struct {
	onlyConsole
	Mode   GameMode `cmd:"GameMode"`
	Player string   `cmd:"player"`
}

func (c GmConsole) Run(src cmd.Source, o *cmd.Output, tx *world.Tx) {
	mode, modeName := parseGameMode(string(c.Mode))
	usr, ok := user.P().UserByName(c.Player)

	if !ok {
		o.Errorf("Cannot find player: %s", c.Player)
		return
	}

	if _, ok := usr.Player(); !ok || !usr.Online() {
		o.Errorf("Cannot find player: %s", c.Player)
		return
	}

	pl, _ := usr.Player()
	go pl.H().ExecWorld(func(tx *world.Tx, e world.Entity) {
		e.(*player.Player).SetGameMode(mode)
	})

	pl.Messagef(text.Colourf("<red><b>>></b></red> Your game mode is now <grey>%s</grey>", modeName))
	systemMessage(o, "You've successfully changed %s game mode to <grey>%s</grey>", c.Player, modeName)
}

type GameMode string

func (GameMode) Type() string {
	return "GameMode"
}
func (GameMode) Options(cmd.Source) []string {
	return []string{
		"0", "survival", "s",
		"1", "creative", "c",
		"2", "adventure", "a",
		"3", "6", "spectator", "sp",
	}
}

func parseGameMode(s string) (world.GameMode, string) {
	switch s {
	case "0", "survival", "s":
		return world.GameModeSurvival, "Survival"
	case "1", "creative", "c":
		return world.GameModeCreative, "Creative"
	default:
		fallthrough
	case "2", "adventure", "a":
		return world.GameModeAdventure, "Adventure"
	case "3", "6", "spectator", "sp":
		return world.GameModeSpectator, "Spectator"
	}
}
