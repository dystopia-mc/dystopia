package rank

import "github.com/sandertv/gophertunnel/minecraft/text"

var Famous = Rank{
	Name:     "Famous",
	Format:   text.Bold + text.Copper,
	Priority: PriorityFamous,

	DisplayRankName: true,
}
