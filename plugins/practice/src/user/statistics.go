package user

import (
	"github.com/df-mc/dragonfly/server/player/form"
	"github.com/df-mc/dragonfly/server/world"
	"github.com/k4ties/dystopia/internal"
	"github.com/k4ties/dystopia/plugins/practice/src/rank"
	"github.com/sandertv/gophertunnel/minecraft/text"
	"strings"
	"time"
)

type statisticsForm struct{}

func (s statisticsForm) Submit(form.Submitter, form.Button, *world.Tx) {}

func (u *User) NewStatisticsForm() form.Menu {
	m := form.NewMenu(statisticsForm{}, "Statistics")

	format := []string{
		"Rank: %s",
		"Name: <grey>%s</grey>",
		"",
		"First joined: <red>%s ago</red>",
		"Played: <red>%s</red>",
		"",
		"Kills: <red>%d</red>",
		"Deaths: <red>%d</red>",
		"",
		"Kill Streak:",
		"<dark-grey>-</dark-grey> Current: <red>%d</red>",
		"<dark-grey>-</dark-grey> Max: <red>%d</red>",
	}

	d := u.Data()
	k := u.KillStreak()

	name := d.Name()

	r := d.Rank()
	if rank.IsDefault(r) {
		r.Name = "Player"
		r.Format = text.Grey
	}

	role := r.Format + r.Name + text.Reset

	kills := d.Kills()
	deaths := d.Deaths()
	killStreakCurrent := k.Current()
	killStreakMax := k.Max()

	formatted := strings.Join(format, "\n")
	formatted = text.Colourf(formatted, role, name, internal.DurationString(time.Since(d.FirstJoin())), internal.DurationString(d.Played()), kills, deaths, killStreakCurrent, killStreakMax)

	backButton := form.Button{
		Text:  "Close",
		Image: "textures/ui/redX1.png",
	}

	return m.WithBody(formatted).WithButtons(backButton)
}
