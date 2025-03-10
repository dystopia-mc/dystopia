package bot

import (
	"fmt"
	"github.com/df-mc/dragonfly/server/cmd"
	"github.com/df-mc/dragonfly/server/world"
	"github.com/go-gl/mathgl/mgl64"
	"github.com/k4ties/dystopia/plugins/practice/src/console"
	"github.com/sandertv/gophertunnel/minecraft/text"
	"strings"
)

func (b *Bot) requestConsoleCommand(text string, peer int) {
	args := strings.Split(text, " ")

	if len(args) >= 1 && strings.HasPrefix(text, ".") {
		console.ProceedCommand(strings.TrimPrefix(text, "."), &consoleSource{
			b:    b,
			peer: peer,
		})
	}
}

type consoleSource struct {
	b    *Bot
	peer int
}

func (c *consoleSource) IsConsole() {}

func (c *consoleSource) Position() mgl64.Vec3 {
	return mgl64.Vec3{}
}

func (c *consoleSource) World() *world.World {
	return nil
}

func (c *consoleSource) SendCommandOutput(o *cmd.Output) {
	const (
		ReceiveMessageFormat = "Server has an answer:\n%s"
		ReceiveErrorFormat   = "Server has returned an error while executing command:\n%s"
	)

	for _, m := range o.Messages() {
		c.b.sendMessagef(c.peer, fmt.Sprintf(ReceiveMessageFormat, text.Clean(m.String())))
	}
	for _, e := range o.Errors() {
		c.b.sendMessagef(c.peer, fmt.Sprintf(ReceiveErrorFormat, text.Clean(e.Error())))
	}
}

func (c *consoleSource) Name() string {
	return "CONSOLE"
}
