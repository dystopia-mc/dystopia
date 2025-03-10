package command

import (
	"github.com/df-mc/dragonfly/server/cmd"
	"github.com/df-mc/dragonfly/server/world"
	"github.com/k4ties/dystopia/internal"
	punishment2 "github.com/k4ties/dystopia/plugins/practice/src/punishment"
	"github.com/k4ties/dystopia/plugins/practice/src/user"
	"time"
)

type MuteEnum struct {
	onlyStaff
	Target   PlayerEnum  `cmd:"target"`
	Duration string      `cmd:"duration"`
	Reason   cmd.Varargs `cmd:"reason"`
}

func (m MuteEnum) Run(src cmd.Source, o *cmd.Output, tx *world.Tx) {
	u, ok := user.P().UserByName(string(m.Target))
	if !ok {
		o.Errorf("User not found: %s", m.Target)
		return
	}
	if !isConsole(src) {
		if self, ok := usr(src); ok {
			if u == self {
				o.Errorf("Cannot mute yourself.")
				return
			}
			if self.Data().Rank().Priority <= u.Data().Rank().Priority {
				o.Errorf("Can only mute players with lower rank.")
				return
			}
		}
	}

	var expire time.Time
	if m.Duration == "-f" {
		expire = punishment2.NilTime
	} else {
		dur, err := internal.ParseDuration(m.Duration)
		if err != nil {
			o.Errorf("Invalid duration")
			return
		}

		expire = time.Now().Add(dur)
	}
	if err := punishment2.P().Mute(u, expire, string(m.Reason), src.(cmd.NamedTarget).Name()); err != nil {
		o.Errorf("Unable to mute: %s", err.Error())
		return
	}

	systemMessage(o, "You've successfully muted <grey>%s</grey> for <grey>%s</grey>", u.Data().Name(), internal.DurationString(time.Until(expire)))
}

type Mute struct {
	onlyStaffAndConsole
	Target   string      `cmd:"target"`
	Duration string      `cmd:"duration"`
	Reason   cmd.Varargs `cmd:"reason"`
}

func (m Mute) Run(src cmd.Source, o *cmd.Output, tx *world.Tx) {
	MuteEnum{Target: PlayerEnum(m.Target), Duration: m.Duration, Reason: m.Reason}.Run(src, o, tx)
}
