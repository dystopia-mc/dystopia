package command

import (
	"github.com/df-mc/dragonfly/server/cmd"
	"github.com/df-mc/dragonfly/server/entity"
	"github.com/df-mc/dragonfly/server/player"
	"github.com/df-mc/dragonfly/server/world"
	"github.com/k4ties/dystopia/plugins/practice/src/ffa"
	"github.com/k4ties/dystopia/plugins/practice/src/handlers"
	"github.com/k4ties/dystopia/plugins/practice/src/instance/lobby"
	"github.com/k4ties/dystopia/plugins/practice/src/user"
)

type Kill struct {
	onlyStaff
	Target []cmd.Target `cmd:"target"`
}

func (c Kill) Run(s cmd.Source, o *cmd.Output, tx *world.Tx) {
	t := c.Target[0]
	tP, ok := t.(*player.Player)

	if !ok {
		o.Errorf("Invalid player")
		return
	}

	pl, in := ffa.LookupPlayer(p(s).UUID())
	if pl == nil || in == nil {
		o.Errorf("You're not in ffa.")
		return
	}

	tPl, tIn := ffa.LookupPlayer(tP.UUID())
	if tPl == nil || tIn == nil {
		o.Errorf("Target player not in ffa.")
		return
	}

	if !in.Active(tP.UUID()) || !tIn.Active(pl.UUID()) {
		o.Errorf("You must be in same instance with target.")
		return
	}

	attacker := p(s)
	if tIn.InCombat(tP.UUID()) {
		combat, _ := tIn.Combat(tP.UUID())

		if combat.With() != nil && combat.With().UUID() != attacker.UUID() {
			attacker, ok = combat.With().(*user.User).Player()

			if !ok {
				panic("should never happen")
			}
		}
	}

	var keepInv bool

	h := handlers.MustPractice(tP.UUID(), tP, lobby.Instance())
	h.HandleDeath(tP, entity.AttackDamageSource{Attacker: attacker}, &keepInv)

	if !keepInv {
		tP.Inventory().Clear()
		tP.Armour().Clear()
	}

	systemMessage(o, "You've successfully killed <grey>%s</grey>", tPl.Name())
}
