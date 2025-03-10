package command

import (
	"github.com/df-mc/dragonfly/server/cmd"
	"github.com/df-mc/dragonfly/server/entity"
	"github.com/df-mc/dragonfly/server/world"
	"github.com/k4ties/dystopia/plugins/practice/src/ffa"
	"github.com/k4ties/dystopia/plugins/practice/src/handlers"
	"github.com/k4ties/dystopia/plugins/practice/src/instance/lobby"
	"github.com/k4ties/dystopia/plugins/practice/src/user"
)

type SurRender struct {
	onlyPlayer
}

func (c SurRender) Run(s cmd.Source, o *cmd.Output, tx *world.Tx) {
	if dead(s) {
		o.Errorf("Cannot surrender while you're already dead")
		return
	}

	pl, in := ffa.LookupPlayer(p(s).UUID())
	if pl == nil || in == nil {
		o.Errorf("You're not in ffa.")
		return
	}

	// since we've already checked if he isn't in combat, we are sure that it must not cause errors.
	combat, _ := in.Combat(pl.UUID())
	if combat.With() == nil {
		o.Errorf("You're not in combat.")
		return
	}

	withP, ok := combat.With().(*user.User).Player()
	if !ok {
		panic("should never happen")
	}

	var keepInv bool

	h := handlers.MustPractice(pl.UUID(), pl.Player, lobby.Instance())
	h.HandleDeath(pl.Player, entity.AttackDamageSource{
		Attacker: withP,
	}, &keepInv)

	if !keepInv {
		pl.Inventory().Clear()
		pl.Armour().Clear()
	}

	systemMessage(o, "You've successfully surrendered")
}
