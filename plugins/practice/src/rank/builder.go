package rank

import (
	"github.com/sandertv/gophertunnel/minecraft/text"
)

var Builder = Rank{
	Name:     "Builder",
	Format:   text.Bold + text.Orange,
	Priority: PriorityHelper,

	DisplayRankName: true,
}
