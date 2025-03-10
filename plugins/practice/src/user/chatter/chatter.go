package chatter

import (
	"github.com/df-mc/atomic"
	"github.com/df-mc/dragonfly/server/player/chat"
	"github.com/k4ties/dystopia/internal/embeddable"
	"github.com/k4ties/dystopia/plugins/practice/src/bot"
	"github.com/sandertv/gophertunnel/minecraft/text"

	_ "embed"
)

type Config struct {
	Config struct {
		ChatEnabled bool `json:"Chat-Enabled"`
	}
}

//go:embed config.json
var configBytes []byte
var config = embeddable.MustJSON[Config](configBytes)

func init() {
	impl.enabled.Store(config.Config.ChatEnabled)
}

var impl = struct {
	enabled atomic.Bool
}{}

func ChatEnabled() bool {
	return impl.enabled.Load()
}

func ToggleChat() {
	impl.enabled.Store(!ChatEnabled())
}

func Messagef(format string, a ...any) {
	msg := text.Colourf(format, a...)

	_, _ = chat.Global.WriteString(msg)
	bot.Log(msg)
}
