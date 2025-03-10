package announcement

import (
	_ "embed"
	"github.com/df-mc/dragonfly/server"
	"github.com/k4ties/dystopia/internal/embeddable"
	"github.com/k4ties/dystopia/plugins/practice/src/user/chatter"
	"slices"
	"time"
)

type messages struct {
	Messages struct {
		List []string
	}
}

//go:embed announcement_messages.json
var messagesJson []byte

var messagesConf = embeddable.MustJSON[messages](messagesJson)

func DoTask(srv *server.Server) {
	ticker := time.NewTicker(time.Minute * 5)

	for range ticker.C {
		if playersOnline(2, srv) {
			globalMessage(getMessage())
		}
	}
}

func playersOnline(amount int, srv *server.Server) bool {
	if len(slices.Collect(srv.Players(nil))) >= amount {
		return true
	}

	return false
}

func globalMessage(format string, args ...any) {
	chatter.Messagef(format, args...)
}

var currentMessage = 0

func getMessage() string {
	currentMessage++
	list := messagesConf.Messages.List

	if currentMessage == len(list)+1 {
		currentMessage = 1
	}

	return list[currentMessage-1] // make sure that this is indexed
}
