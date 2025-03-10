package lobby

import (
	"github.com/df-mc/dragonfly/server/item"
	"github.com/k4ties/dystopia/plugins/practice/src/custom/items"
	kit2 "github.com/k4ties/dystopia/plugins/practice/src/kit"
)

var Kit = func() kit2.Kit {
	ffaStack := item.NewStack(item.Sword{Tier: item.ToolTierDiamond}, 1)
	settingsStack := item.NewStack(items.Settings{}, 1)
	statisticsStack := item.NewStack(item.Book{}, 1)

	//shopStack := item.NewStack(items.Shop{}, 1)

	var items = make(kit2.Items)

	//items[0] = kit.ApplyIdentifier(kit.ShopIdentifier, kit.FillNames("Shop", shopStack))
	items[1] = kit2.ApplyIdentifier(kit2.SettingsIdentifier, kit2.FillNames("Settings", settingsStack))
	items[4] = kit2.ApplyIdentifier(kit2.FFAIdentifier, kit2.FillNames("FFA", ffaStack))
	items[7] = kit2.ApplyIdentifier(kit2.StatisticsIdentifier, kit2.FillNames("Statistics", statisticsStack))

	return kit2.New(items, kit2.NopArmour)
}()
