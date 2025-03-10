package command

import (
	"github.com/df-mc/dragonfly/server/cmd"
	"github.com/df-mc/dragonfly/server/entity/effect"
	"github.com/df-mc/dragonfly/server/world"
	"github.com/k4ties/dystopia/plugins/practice/src/ffa"
)

type ReKit struct {
	onlyPlayer
}

func (ReKit) Run(s cmd.Source, o *cmd.Output, tx *world.Tx) {
	if dead(s) {
		o.Errorf("Cannot rekit while you're dead")
		return
	}

	pl, in := ffa.LookupPlayer(p(s).UUID())
	if pl == nil || in == nil {
		o.Errorf("You're not in ffa.")
		return
	}

	if c, ok := in.Combat(pl.UUID()); ok {
		if c.Active() {
			o.Errorf("Cannot rekit while you're in combat")
			return
		}
	}

	in.ReKit(pl.UUID(), tx)
	p(s).Heal(p(s).MaxHealth(), effect.InstantHealingSource{})

	systemMessage(o, "You've successfully rekitted.")
}
