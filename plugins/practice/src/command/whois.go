package command

import (
	"github.com/df-mc/dragonfly/server/cmd"
	"github.com/df-mc/dragonfly/server/world"
	user2 "github.com/k4ties/dystopia/plugins/practice/src/user"
	"math"
	"strconv"
	"strings"
	"time"
)

type Whois struct {
	onlyStaffAndConsole
	Player cmd.Varargs `cmd:"player"`
}

func (w Whois) Run(src cmd.Source, o *cmd.Output, tx *world.Tx) {
	pool := user2.P()
	usr, ok := pool.UserByName(string(w.Player))
	if !ok {
		o.Errorf("Can't find user: %s", w.Player)
		return
	}

	whois(usr, o)
}

type WhoisEnum struct {
	onlyStaff
	Player PlayerEnum `cmd:"player"`
}

func (w WhoisEnum) Run(src cmd.Source, o *cmd.Output, tx *world.Tx) {
	pool := user2.P()
	usr, ok := pool.UserByName(string(w.Player))
	if !ok {
		o.Errorf("Can't find user: %s", w.Player)
		return
	}

	whois(usr, o)
}

func whois(u *user2.User, o *cmd.Output) {
	const removeFlag = "[remove]"

	online := u.Online()
	if _, ok := u.Player(); !ok {
		online = false
	}

	format := []string{
		"Name: " + u.Data().Name(),
		"UUID: " + u.Data().UUID().String(),
		"XUID: " + u.Data().XUID(),
		"Rank: " + u.Data().Rank().Name,
		"Online: " + strconv.FormatBool(u.Online()),
		func() string {
			if !online {
				return removeFlag
			}
			return "Input Mode: " + u.Data().InputMode().String()
		}(),
		func() string {
			if !online {
				return removeFlag
			}
			return "OS: " + u.Data().OS().String()
		}(),
		func() string {
			if !online {
				return removeFlag
			}
			return "Current CPS: " + strconv.Itoa(u.CPS())
		}(),
		func() string {
			if !online || u.Data().FPS() == -1 {
				return removeFlag
			}
			return "Current FPS: " + strconv.Itoa(u.Data().FPS())
		}(),
		"Kill Streak Current: " + strconv.Itoa(int(u.KillStreak().Current())),
		"Kill Streak Max: " + strconv.Itoa(int(u.KillStreak().Max())),
		"Deaths: " + strconv.Itoa(u.Data().Deaths()),
		"Kills: " + strconv.Itoa(u.Data().Kills()),
		"First joined: " + strconv.Itoa(int(math.Floor(time.Since(u.Data().FirstJoin()).Hours()/24))) + " days ago (" + strconv.Itoa(int(math.Floor(time.Since(u.Data().FirstJoin()).Hours()))) + " hours)",
		"Played: " + strconv.Itoa(int(math.Floor(u.Data().Played().Hours()))) + " hours (" + strconv.Itoa(int(math.Floor(u.Data().Played().Minutes()))) + " minutes)",
		"Has tag: " + strconv.FormatBool(u.Data().HasTag()),
		func() string {
			if !u.Data().HasTag() {
				return removeFlag
			}
			return u.Data().Tag()
		}(),
		"IPS: " + strings.Join(u.Data().IPs(), ", "),
		"Device IDs: " + strings.Join(u.Data().DIDs(), ", "),
	}

	replacer := strings.NewReplacer(removeFlag, "")
	msg := replacer.Replace(strings.Join(format, "\n"))

	var r []string
	for _, str := range strings.Split(msg, "\n") {
		if str == "" {
			continue
		}

		r = append(r, str)
	}

	msg = strings.Join(r, "\n")
	systemMessage(o, msg)
}

func remove(slice []string, index int) []string {
	return append(slice[:index], slice[index+1:]...)
}
