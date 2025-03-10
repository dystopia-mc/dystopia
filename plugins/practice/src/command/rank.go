package command

import (
	"github.com/df-mc/dragonfly/server/cmd"
	"github.com/df-mc/dragonfly/server/world"
	"github.com/k4ties/dystopia/plugins/practice/src/rank"
	"github.com/k4ties/dystopia/plugins/practice/src/user"
	"github.com/sandertv/gophertunnel/minecraft/text"
	"slices"
	"strings"
)

type RankPlayerSet struct {
	onlyManagerAndConsole
	Target string `cmd:"target"`
	Set    cmd.SubCommand
	Rank   RankEnum `cmd:"rank"`
}

type RankPlayerEnumSet struct {
	onlyManager
	Target PlayerEnum     `cmd:"target"`
	Set    cmd.SubCommand `cmd:"set"`
	Rank   RankEnum       `cmd:"rank"`
}

type RankPlayerEnumUpdate struct {
	onlyAdmin
	Target PlayerEnum     `cmd:"target"`
	Update cmd.SubCommand `cmd:"update"`
	Rank   RankEnum       `cmd:"rank"`
}

type RankPlayerUpdate struct {
	onlyAdminAndConsole
	Target string         `cmd:"target"`
	Update cmd.SubCommand `cmd:"update"`
	Rank   RankEnum       `cmd:"rank"`
}

type RankPlayerGet struct {
	onlyManagerAndConsole
	Target string         `cmd:"target"`
	Get    cmd.SubCommand `cmd:"get"`
}

type RankPlayerEnumGet struct {
	onlyManager
	Target PlayerEnum     `cmd:"target"`
	Get    cmd.SubCommand `cmd:"get"`
}

type RankList struct {
	onlyManagerAndConsole
	List cmd.SubCommand `cmd:"list"`
}

func (r RankPlayerEnumUpdate) Run(_ cmd.Source, o *cmd.Output, _ *world.Tx) {
	targetName := string(r.Target)

	usr, ok := user.P().UserByName(targetName)
	if !ok {
		o.Errorf("Can't find user: %s", targetName)
		return
	}

	role, ok2 := rank.ByName(string(r.Rank))
	if !ok2 {
		o.Errorf("Can't find rank: %s", r.Rank)
		return
	}

	if rank.IsDefault(role) {
		role.Name = "Player"
		role.Format = text.Grey
	}

	if role.Priority >= usr.Data().Rank().Priority {
		usr.UpdateRank(role)
		systemMessage(o, "Rank of <grey>%s</grey> is now %s", targetName, role.Format+role.Name)
		return
	}

	o.Errorf("Target rank must be higher by priority than current player rank. (comparation %s to %s)", role.Name, usr.Data().Rank().Name)
	return
}

func (r RankPlayerUpdate) Run(src cmd.Source, o *cmd.Output, tx *world.Tx) {
	RankPlayerEnumUpdate{Target: PlayerEnum(r.Target), Rank: r.Rank}.Run(src, o, tx)
}

func (r RankList) Run(src cmd.Source, o *cmd.Output, _ *world.Tx) {
	list := RankEnum("").Options(src)
	slices.Sort(list)

	format := "Currently there are <red>%d</red> ranks on the server: <grey>%s</grey>"
	systemMessage(o, format, len(list), strings.Join(list, ", "))
}

func (r RankPlayerEnumGet) Run(_ cmd.Source, o *cmd.Output, _ *world.Tx) {
	targetName := string(r.Target)

	usr, ok := user.P().UserByName(targetName)
	if !ok {
		o.Errorf("Can't find user: %s", targetName)
		return
	}

	role := usr.Data().Rank()
	if rank.IsDefault(role) {
		role.Name = "Player"
		role.Format = text.Grey
	}

	systemMessage(o, "Rank of <grey>%s:</grey> %s", targetName, role.Format+role.Name)
}

func (r RankPlayerGet) Run(src cmd.Source, o *cmd.Output, tx *world.Tx) {
	RankPlayerEnumGet{Target: PlayerEnum(r.Target)}.Run(src, o, tx)
}

func (r RankPlayerEnumSet) Run(_ cmd.Source, o *cmd.Output, _ *world.Tx) {
	targetName := string(r.Target)

	usr, ok := user.P().UserByName(targetName)
	if !ok {
		o.Errorf("Can't find user: %s", targetName)
		return
	}

	role, ok2 := rank.ByName(string(r.Rank))
	if !ok2 {
		o.Errorf("Can't find rank: %s", r.Rank)
		return
	}

	if rank.IsDefault(role) {
		role.Name = "Player"
		role.Format = text.Grey
	}

	usr.UpdateRank(role)
	systemMessage(o, "Rank of <grey>%s</grey> is now %s", targetName, role.Format+role.Name)
}

func (r RankPlayerSet) Run(s cmd.Source, o *cmd.Output, tx *world.Tx) {
	RankPlayerEnumSet{
		Target: PlayerEnum(r.Target),
		Rank:   r.Rank,
	}.Run(s, o, tx)
}
