package lcn

import (
	"github.com/df-mc/dragonfly/server/cmd"
	plugin "github.com/k4ties/df-plugin/df-plugin"
)

func Task() plugin.TaskFunc {
	return func(m *plugin.Manager) {
		cmd.Register(cmd.New("prefix", "Prefix for clan members", nil, command{}))
	}
}
