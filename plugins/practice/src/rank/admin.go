package rank

import (
	"github.com/sandertv/gophertunnel/minecraft/text"
)

var Admin = Rank{
	Name:     "Admin",
	Format:   text.Red,
	Priority: PriorityAdmin,

	DisplayRankName: true,
}
