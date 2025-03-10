package rank

import (
	"github.com/sandertv/gophertunnel/minecraft/text"
)

var Owner = Rank{
	Name:     "Owner",
	Format:   text.Bold + text.Red,
	Priority: PriorityOwner,

	DisplayRankName: true,
}
