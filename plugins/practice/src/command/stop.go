package command

import (
	"github.com/df-mc/dragonfly/server/cmd"
	"github.com/df-mc/dragonfly/server/world"
	"os"
	"time"
)

type Stop struct {
	onlyOwnerAndConsole
	CloseFunc func() error `cmd:"-"`
}

func (s Stop) Run(cmd.Source, *cmd.Output, *world.Tx) {
	if err := s.CloseFunc(); err != nil {
		panic(err)
	}
	time.AfterFunc(time.Second, func() {
		os.Exit(0)
	})
}
