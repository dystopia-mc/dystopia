package nodebuff

import (
	"github.com/df-mc/dragonfly/server/entity/effect"
	"github.com/df-mc/dragonfly/server/item"
	"github.com/df-mc/dragonfly/server/item/enchantment"
	"github.com/df-mc/dragonfly/server/player"
	"github.com/k4ties/dystopia/plugins/practice/src/custom/entities"
	items3 "github.com/k4ties/dystopia/plugins/practice/src/custom/items"
	"github.com/k4ties/dystopia/plugins/practice/src/ffa"
	"github.com/k4ties/dystopia/plugins/practice/src/kit"
	"github.com/k4ties/dystopia/plugins/practice/src/user"
	"github.com/sandertv/gophertunnel/minecraft/text"
)

var Kit = func(p *player.Player) kit.Kit {
	var (
		h entities.PotHandler

		items   = make(kit.Items)
		armour  kit.Armour
		effects []effect.Effect

		smoothPearl = false
	)

	if pl, in := ffa.LookupPlayer(p.UUID()); pl != nil && in != nil {
		h = in.CombatHandler(p.UUID())
	}

	if usr, ok := user.P().User(p.UUID()); ok {
		smoothPearl = usr.Settings().FFA.SmoothPearl
	}

	for i := 1; i <= 36; i++ {
		var added item.Stack

		switch i {
		case 1:
			added = kit.ApplyIdentifier(kit.NodebuffSwordIdentifier, item.NewStack(item.Sword{Tier: item.ToolTierDiamond}, 1).WithEnchantments(item.NewEnchantment(enchantment.Sharpness, 6)).WithLore(text.Colourf("<red>dystopia</red>")))
		case 2:
			added = kit.ApplyIdentifier(kit.PearlIdentifier, item.NewStack(items3.Pearl{Smooth: smoothPearl}, 16).WithLore(text.Colourf("<red>dystopia</red>")))
		default:
			added = kit.ApplyIdentifier(kit.PotIdentifier, item.NewStack(items3.NewHealingPotion(h), 1).WithLore(text.Colourf("<red>dystopia</red>")))
		}

		items[i-1] = added
	}

	armour[0] = item.NewStack(item.Helmet{Tier: item.ArmourTierDiamond{}}, 1).WithEnchantments(item.NewEnchantment(enchantment.Protection, 4)).WithLore(text.Colourf("<red>dystopia</red>"))
	armour[1] = item.NewStack(item.Chestplate{Tier: item.ArmourTierDiamond{}}, 1).WithEnchantments(item.NewEnchantment(enchantment.Protection, 4)).WithLore(text.Colourf("<red>dystopia</red>"))
	armour[2] = item.NewStack(item.Leggings{Tier: item.ArmourTierDiamond{}}, 1).WithEnchantments(item.NewEnchantment(enchantment.Protection, 4)).WithLore(text.Colourf("<red>dystopia</red>"))
	armour[3] = item.NewStack(item.Boots{Tier: item.ArmourTierDiamond{}}, 1).WithEnchantments(item.NewEnchantment(enchantment.Protection, 4)).WithLore(text.Colourf("<red>dystopia</red>"))

	effects = append(effects, effect.NewInfinite(effect.Speed, 1).WithoutParticles())
	return kit.New(items, armour, effects...)
}
