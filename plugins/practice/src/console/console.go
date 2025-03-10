package console

import (
	"bufio"
	"github.com/df-mc/dragonfly/server"
	"github.com/df-mc/dragonfly/server/cmd"
	"github.com/df-mc/dragonfly/server/world"
	"github.com/go-gl/mathgl/mgl64"
	"github.com/sandertv/gophertunnel/minecraft/text"
	"log/slog"
	"os"
	"strings"
	"sync/atomic"
)

type console struct {
	l   *slog.Logger
	srv *server.Server
}

func (c *console) Name() string {
	return "CONSOLE"
}

func (c *console) Position() mgl64.Vec3 {
	return mgl64.Vec3{}
}

func (c *console) SendCommandOutput(o *cmd.Output) {
	for _, e := range o.Errors() {
		c.l.Error(text.ANSI(e.Error() + text.Reset))
	}
	for _, m := range o.Messages() {
		c.l.Info(text.ANSI(m.String() + text.Reset))
	}
}

func (c *console) World() *world.World {
	return nil
}

func (c *console) IsConsole() {}

type Config struct {
	Logger     *slog.Logger
	ConfigPath string

	Server *server.Server
}

var status = struct {
	active atomic.Bool
	srv    *server.Server
	l      *slog.Logger
}{}

func Start(c Config) {
	console := &console{l: c.Logger, srv: c.Server}
	scanner := bufio.NewScanner(os.Stdin)

	status.active.Store(true)
	status.srv = c.Server
	status.l = c.Logger

	for scanner.Scan() {
		msg := scanner.Text()
		msg = strings.Join(strings.Fields(msg), " ")

		if len(msg) > 0 {
			ProceedCommand(msg, console)
		}
	}
}

func ProceedCommand(line string, src cmd.Source) {
	if !status.active.Load() {
		return
	}

	srv := status.srv
	args := strings.Split(line, " ")

	if len(line) > 0 {
		name := args[0]
		command, ok := cmd.ByAlias(name)

		if !ok {
			o := new(cmd.Output)
			o.Errorf("Unknown command: %s", name)

			src.SendCommandOutput(o)
			return
		}

		srv.World().Exec(func(tx *world.Tx) {
			command.Execute(strings.Join(args[1:], " "), src, tx)
		})
	}
}
