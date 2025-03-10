package instance

import (
	"errors"
	"github.com/df-mc/dragonfly/server/player"
	"github.com/sandertv/gophertunnel/minecraft/text"
	"strings"
)

const kickMessageFormat = "Asynchronous <red>error</red>\nCode: <grey>%s</grey>"

type ErrorCode error

func Kick(p *player.Player, e ErrorCode) {
	p.Disconnect(KickMessage(e))
}

func KickMessage(code ErrorCode) string {
	return text.Colourf(kickMessageFormat, toTitle(code.Error()))
}

// ErrorValagalishe reason: no connection on plugin manager on load user data stage
var ErrorValagalishe = errors.New("valagalishe")

// ErrorSponge reason: fail to sync connection in instance (no connection on plugin manager)
var ErrorSponge = errors.New("sponge")

// ErrorAngus reason: fail to sync user data
var ErrorAngus = errors.New("angus")

// ErrorPeedor reason: too many cps
var ErrorPeedor = errors.New("peedor")

// ErrorMobilizovan reason: touch screen
var ErrorMobilizovan = errors.New("mobilizovan")

// toTitle uppercases the first letter of the string
func toTitle(s string) string {
	return strings.ToUpper(s[:1]) + strings.ToLower(s[1:])
}
