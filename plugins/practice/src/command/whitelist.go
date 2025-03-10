package command

import (
	"github.com/df-mc/dragonfly/server/cmd"
	"github.com/df-mc/dragonfly/server/world"
	"github.com/k4ties/dystopia/plugins/practice/src/handlers/whitelist"
)

type WhitelistToggle struct {
	onlyManagerAndConsole
	Toggle cmd.SubCommand `cmd:"toggle"`
}

func (w WhitelistToggle) Run(_ cmd.Source, o *cmd.Output, _ *world.Tx) {
	whitelist.Toggle()
	systemMessage(o, "You've successfully toggled whitelist")
}

type WhitelistAdd struct {
	onlyManagerAndConsole
	Add    cmd.SubCommand `cmd:"add"`
	Player cmd.Varargs    `cmd:"player"`
}

func (w WhitelistAdd) Run(_ cmd.Source, o *cmd.Output, _ *world.Tx) {
	whitelist.Add(string(w.Player))
	systemMessage(o, "You've successfully added player <grey>%s</grey> to whitelist", w.Player)
}

type WhitelistAddEnum struct {
	onlyManager
	Add    cmd.SubCommand `cmd:"add"`
	Player PlayerEnum     `cmd:"player"`
}

func (w WhitelistAddEnum) Run(_ cmd.Source, o *cmd.Output, _ *world.Tx) {
	whitelist.Add(string(w.Player))
	systemMessage(o, "You've successfully added player <grey>%s</grey> to whitelist", w.Player)
}

type WhitelistRemove struct {
	onlyManagerAndConsole
	Remove cmd.SubCommand `cmd:"remove"`
	Player cmd.Varargs    `cmd:"player"`
}

func (w WhitelistRemove) Run(_ cmd.Source, o *cmd.Output, _ *world.Tx) {
	whitelist.Remove(string(w.Player))
	systemMessage(o, "You've successfully removed %s from the whitelist", w.Player)
}

type WhitelistRemoveEnum struct {
	onlyManager
	Remove cmd.SubCommand `cmd:"remove"`
	Player PlayerEnum     `cmd:"player"`
}

func (w WhitelistRemoveEnum) Run(_ cmd.Source, o *cmd.Output, _ *world.Tx) {
	whitelist.Remove(string(w.Player))
	systemMessage(o, "You've successfully removed %s from the whitelist", w.Player)
}
