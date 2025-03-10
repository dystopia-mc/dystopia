package whitelist

import (
	"github.com/sasha-s/go-deadlock"
	"slices"
	"strings"
	"sync/atomic"
)

var w = struct {
	enabled atomic.Bool

	players   []string
	playersMu deadlock.RWMutex
}{}

func Toggle() {
	w.enabled.Store(!w.enabled.Load())
}

func Setup(whitelisted bool, players ...string) {
	players = strings.Split(strings.ToLower(strings.Join(players, " ")), " ") // lowercase all names

	w.enabled.Store(whitelisted)
	w.players = append(w.players, players...)
}

func Add(player string) {
	w.playersMu.Lock()
	defer w.playersMu.Unlock()

	w.players = append(w.players, strings.ToLower(player))
}

func Enabled() bool {
	return w.enabled.Load()
}

func Whitelisted(player string) bool {
	if w.enabled.Load() {
		w.playersMu.RLock()
		defer w.playersMu.RUnlock()
		return slices.Contains(w.players, strings.ToLower(player))
	}

	return false
}

func Remove(player string) {
	w.playersMu.Lock()
	defer w.playersMu.Unlock()

	w.players = remove(w.players, slices.Index(w.players, strings.ToLower(player)))
}

func remove(slice []string, index int) []string {
	return append(slice[:index], slice[index+1:]...)
}
