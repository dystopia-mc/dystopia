package command

import (
	"github.com/df-mc/dragonfly/server/cmd"
	"github.com/df-mc/dragonfly/server/world"
	"github.com/k4ties/dystopia/plugins/practice/src/handlers/knockback"
)

type KnockBackSetForce struct {
	onlyManagerAndConsole
	Set   cmd.SubCommand `cmd:"set"`
	Force cmd.SubCommand `cmd:"force"`
	To    float64        `cmd:"to"`
}

func (k KnockBackSetForce) Run(src cmd.Source, o *cmd.Output, tx *world.Tx) {
	knockback.SetForce(k.To)
	systemMessage(o, "KnockBack force is now: %f", k.To)
}

type KnockBackSetHeight struct {
	onlyManagerAndConsole
	Set    cmd.SubCommand `cmd:"set"`
	Height cmd.SubCommand `cmd:"height"`
	To     float64        `cmd:"to"`
}

func (k KnockBackSetHeight) Run(src cmd.Source, o *cmd.Output, tx *world.Tx) {
	knockback.SetHeight(k.To)
	systemMessage(o, "KnockBack height is now: %f", k.To)
}

type KnockBackSetImmunity struct {
	onlyManagerAndConsole
	Set      cmd.SubCommand `cmd:"set"`
	Immunity cmd.SubCommand `cmd:"immunity"`
	To       int            `cmd:"to"`
}

func (k KnockBackSetImmunity) Run(src cmd.Source, o *cmd.Output, tx *world.Tx) {
	knockback.SetImmunity(int64(k.To))
	systemMessage(o, "KnockBack immunity is now: 0.%ds", k.To)
}

type KnockBackGetForce struct {
	onlyManagerAndConsole
	Get   cmd.SubCommand `cmd:"get"`
	Force cmd.SubCommand `cmd:"force"`
}

func (k KnockBackGetForce) Run(src cmd.Source, o *cmd.Output, tx *world.Tx) {
	systemMessage(o, "Current KnockBack force: %f", knockback.Force())
}

type KnockBackGetHeight struct {
	onlyManagerAndConsole
	Get    cmd.SubCommand `cmd:"get"`
	Height cmd.SubCommand `cmd:"height"`
}

func (k KnockBackGetHeight) Run(src cmd.Source, o *cmd.Output, tx *world.Tx) {
	systemMessage(o, "Current KnockBack height: %f", knockback.Height())
}

type KnockBackGetImmunity struct {
	onlyManagerAndConsole
	Get      cmd.SubCommand `cmd:"get"`
	Immunity cmd.SubCommand `cmd:"immunity"`
}

func (k KnockBackGetImmunity) Run(src cmd.Source, o *cmd.Output, tx *world.Tx) {
	systemMessage(o, "Current KnockBack immunity: 0.%ds", knockback.Immunity())
}
