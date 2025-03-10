package rank

import (
	"github.com/sandertv/gophertunnel/minecraft/text"
)

var Manager = Rank{
	Name:     "Manager",
	Format:   text.Bold + text.Amethyst,
	Priority: PriorityManager,

	DisplayRankName: true,
}
