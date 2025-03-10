package command

import (
	"github.com/df-mc/dragonfly/server/cmd"
	"github.com/df-mc/dragonfly/server/player"
	plugin "github.com/k4ties/df-plugin/df-plugin"
	"github.com/k4ties/dystopia/internal"
	"github.com/k4ties/dystopia/plugins/practice/src/instance"
	rank2 "github.com/k4ties/dystopia/plugins/practice/src/rank"
	user2 "github.com/k4ties/dystopia/plugins/practice/src/user"
	"github.com/sandertv/gophertunnel/minecraft/text"
	"strings"
)

type onlyPlayer struct{}

func (onlyPlayer) Allow(src cmd.Source) bool {
	_, ok := src.(*player.Player)
	return ok
}

type onlyOwner struct{}

func (onlyOwner) Allow(src cmd.Source) bool {
	if _, ok := src.(*player.Player); !ok {
		return false
	}

	var found = false

	u(src, func(u *user2.User) {
		found = u.Data().Rank().Priority >= rank2.PriorityOwner
	})

	return found
}

type onlyStaff struct{}

func (onlyStaff) Allow(src cmd.Source) bool {
	if _, ok := src.(*player.Player); !ok {
		return false
	}

	var found = false

	u(src, func(u *user2.User) {
		found = u.Data().Rank().Priority >= rank2.PriorityJuniorHelper
	})

	return found
}

type onlyAdmin struct{}

func (onlyAdmin) Allow(src cmd.Source) bool {
	if _, ok := src.(*player.Player); !ok {
		return false
	}

	var found = false

	u(src, func(u *user2.User) {
		found = u.Data().Rank().Priority >= rank2.PriorityAdmin
	})

	return found
}

type onlyConsole struct{}

func (onlyConsole) Allow(s cmd.Source) bool {
	_, ok := s.(consoler)
	return ok
}

type onlyPlayerAndConsole struct{}

func (onlyPlayerAndConsole) Allow(s cmd.Source) bool {
	_, isConsole := s.(consoler)
	if !isConsole {
		return onlyPlayer{}.Allow(s)
	}
	return isConsole
}

type onlyStaffAndConsole struct{}

func (onlyStaffAndConsole) Allow(s cmd.Source) bool {
	_, isConsole := s.(consoler)
	if !isConsole {
		return onlyStaff{}.Allow(s)
	}
	return isConsole
}

type onlyAdminAndConsole struct{}

func (onlyAdminAndConsole) Allow(s cmd.Source) bool {
	_, isConsole := s.(consoler)
	if !isConsole {
		isAdmin := onlyAdmin{}.Allow(s)
		return isAdmin
	}

	return isConsole
}

type onlyOwnerAndConsole struct{}

func (onlyOwnerAndConsole) Allow(s cmd.Source) bool {
	_, isConsole := s.(consoler)
	if !isConsole {
		isOwner := onlyOwner{}.Allow(s)
		return isOwner
	}

	return isConsole
}

type onlyManager struct{}

func (onlyManager) Allow(s cmd.Source) bool {
	if _, ok := s.(*player.Player); !ok {
		return false
	}

	var found = false

	u(s, func(u *user2.User) {
		found = u.Data().Rank().Priority >= rank2.PriorityManager
	})

	return found
}

type onlyManagerAndConsole struct{}

func (onlyManagerAndConsole) Allow(s cmd.Source) bool {
	_, isConsole := s.(consoler)
	if !isConsole {
		isOwner := onlyManager{}.Allow(s)
		return isOwner
	}

	return isConsole
}

func systemMessage(o *cmd.Output, format string, args ...any) {
	res := text.Colourf(format, args...)
	parts := strings.Split(res, "\n")

	prefix := text.Colourf("<red><b>>></b></red>")

	for i, part := range parts {
		parts[i] = prefix + " " + part
	}

	o.Printf(strings.Join(parts, "\n"))
}

type consoler interface {
	IsConsole()
	cmd.NamedTarget
	cmd.Source
}

func p(s cmd.Source) *player.Player {
	return s.(*player.Player)
}

func c(s cmd.Source) consoler {
	return s.(consoler)
}

func isConsole(s cmd.Source) bool {
	_, ok := s.(consoler)
	return ok
}

func inPl(s cmd.Source) *instance.Player {
	pl, _ := instance.LookupPlayer(p(s).UUID())
	return pl
}

func u(s cmd.Source, f func(user *user2.User), ifDoesntFound ...func()) {
	if pl := p(s); pl != nil {
		user2.MustUsePool(func(p *user2.Pool) {
			p.MustUser(pl, f)
		})
	} else {
		for _, function := range ifDoesntFound {
			function()
		}
	}
}

func usr(s cmd.Source) (*user2.User, bool) {
	p, ok := s.(*player.Player)
	if !ok {
		return nil, false
	}

	return user2.P().User(p.UUID())
}

func dead(s cmd.Source) bool {
	return p(s).GameMode() == internal.FlyingMode
}

type RankEnum string

func (RankEnum) Type() string {
	return "rank"
}

func (RankEnum) Options(cmd.Source) []string {
	var names []string

	for _, r := range rank2.List() {
		names = append(names, strings.ToLower(r.Name))
	}

	return names
}

type PlayerEnum string

func (PlayerEnum) Type() string {
	return "player"
}

func (PlayerEnum) Options(s cmd.Source) []string {
	var names []string
	if isConsole(s) {
		return []string{"YOU CAN SEE THIS MESSAGE ONLY ON CONSOLE. PLEASE DO NOT ALLOW CONSOLE TO USE PLAYER ENUM"}
	}

	for pl := range plugin.M().Srv().Players(s.(*player.Player).Tx()) {
		names = append(names, pl.Name())
	}

	return names
}
