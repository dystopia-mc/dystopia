package knocker

import (
	"github.com/df-mc/dragonfly/server/entity/effect"
	"github.com/df-mc/dragonfly/server/item"
	"github.com/df-mc/dragonfly/server/item/enchantment"
	"github.com/df-mc/dragonfly/server/player"
	items2 "github.com/k4ties/dystopia/plugins/practice/src/custom/items"
	kit2 "github.com/k4ties/dystopia/plugins/practice/src/kit"
	"github.com/k4ties/dystopia/plugins/practice/src/user"
	"github.com/sandertv/gophertunnel/minecraft/text"
)

var Kit = func(p *player.Player) kit2.Kit {
	var (
		items  = make(kit2.Items)
		armour kit2.Armour

		smoothPearl bool
	)

	if usr, ok := user.P().User(p.UUID()); ok {
		smoothPearl = usr.Settings().FFA.SmoothPearl
	}

	items[0] = kit2.ApplyIdentifier(kit2.KBSwordIdentifier, item.NewStack(item.Sword{Tier: item.ToolTierWood}, 1).WithEnchantments(item.NewEnchantment(enchantment.Knockback, 2)).WithLore(text.Colourf("<red>dystopia</red>")))
	items[1] = kit2.ApplyIdentifier(kit2.PearlIdentifier, item.NewStack(items2.Pearl{Smooth: smoothPearl}, 5).WithLore(text.Colourf("<red>dystopia</red>")))
	return kit2.New(items, armour, effect.NewInfinite(effect.Resistance, 255).WithoutParticles())
}
