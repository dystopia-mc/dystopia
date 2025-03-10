package command

import (
	"github.com/df-mc/dragonfly/server/cmd"
	"github.com/df-mc/dragonfly/server/world"
	"github.com/k4ties/dystopia/internal"
	punishment2 "github.com/k4ties/dystopia/plugins/practice/src/punishment"
	"github.com/k4ties/dystopia/plugins/practice/src/user"
	"time"
)

type BanEnum struct {
	onlyStaff
	Target   PlayerEnum  `cmd:"target"`
	Duration string      `cmd:"duration"`
	Reason   cmd.Varargs `cmd:"reason"`
}

func (b BanEnum) Run(src cmd.Source, o *cmd.Output, _ *world.Tx) {
	u, ok := user.P().UserByName(string(b.Target))
	if !ok {
		o.Errorf("User not found: %s", b.Target)
		return
	}
	if !isConsole(src) {
		if self, ok := usr(src); ok {
			if u == self {
				o.Errorf("Cannot ban yourself.")
				return
			}
			if self.Data().Rank().Priority <= u.Data().Rank().Priority {
				o.Errorf("Can only ban players with lower rank.")
				return
			}
		}
	}

	var expire time.Time
	if b.Duration == "-f" {
		expire = punishment2.NilTime
	} else {
		dur, err := internal.ParseDuration(b.Duration)
		if err != nil {
			o.Errorf("Invalid duration")
			return
		}

		expire = time.Now().Add(dur)
	}
	if err := punishment2.P().Ban(u, expire, string(b.Reason), src.(cmd.NamedTarget).Name()); err != nil {
		o.Errorf("Unable to ban: %s", err.Error())
		return
	}

	systemMessage(o, "You've successfully banned <grey>%s</grey> for <grey>%s</grey>", u.Data().Name(), internal.DurationString(time.Until(expire)))
}

type BanOffline struct {
	onlyStaffAndConsole
	Target   string      `cmd:"target"`
	Duration string      `cmd:"duration"`
	Reason   cmd.Varargs `cmd:"reason"`
}

func (b BanOffline) Run(src cmd.Source, o *cmd.Output, tx *world.Tx) {
	BanEnum{Target: PlayerEnum(b.Target), Duration: b.Duration, Reason: b.Reason}.Run(src, o, tx)
}
