package rank

import "github.com/sandertv/gophertunnel/minecraft/text"

var Vip = Rank{
	Name:     "VIP",
	Format:   text.Bold + text.Green,
	Priority: PriorityVip,

	DisplayRankName: true,
}
