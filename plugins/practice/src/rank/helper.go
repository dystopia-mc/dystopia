package rank

import (
	"github.com/sandertv/gophertunnel/minecraft/text"
)

var Helper = Rank{
	Name:     "Helper",
	Format:   text.Bold + text.Lapis,
	Priority: PriorityHelper,

	DisplayRankName: true,
}
