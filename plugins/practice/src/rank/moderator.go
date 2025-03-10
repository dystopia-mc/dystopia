package rank

import (
	"github.com/sandertv/gophertunnel/minecraft/text"
)

var Moderator = Rank{
	Name:     "Moderator",
	Format:   text.Bold + text.Aqua,
	Priority: PrioritySeniorModerator,

	DisplayRankName: true,
}
