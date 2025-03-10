package sumo

import (
	"github.com/df-mc/dragonfly/server/entity/effect"
	"github.com/df-mc/dragonfly/server/item"
	"github.com/df-mc/dragonfly/server/player"
	"github.com/k4ties/dystopia/plugins/practice/src/kit"
	"github.com/sandertv/gophertunnel/minecraft/text"
)

var Kit = func(*player.Player) kit.Kit {
	var (
		items  = make(kit.Items)
		armour kit.Armour
	)

	items[0] = item.NewStack(item.Stick{}, 1).WithLore(text.Colourf("<red>dystopia</red>"))
	return kit.New(items, armour, effect.NewInfinite(effect.Resistance, 255).WithoutParticles())
}
