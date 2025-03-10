package command

import (
	"github.com/df-mc/dragonfly/server/cmd"
	"github.com/df-mc/dragonfly/server/world"
	plugin "github.com/k4ties/df-plugin/df-plugin"
	"slices"
	"strings"
)

type List struct {
	onlyPlayerAndConsole
}

func (l List) Run(_ cmd.Source, o *cmd.Output, tx *world.Tx) {
	var players []string
	srv := plugin.M().Srv()
	for p := range srv.Players(tx) {
		players = append(players, p.Name())
	}
	slices.Sort(players)

	format := "Currently there are <red>%d/%d</red> players: <grey>%s</grey>"
	systemMessage(o, format, srv.PlayerCount(), srv.MaxPlayerCount(), strings.Join(players, ", "))
}
