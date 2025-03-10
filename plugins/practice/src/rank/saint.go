package rank

import "github.com/sandertv/gophertunnel/minecraft/text"

var Saint = Rank{
	Name:     "Saint",
	Format:   text.Bold + text.Aqua,
	Priority: PrioritySaint,

	DisplayRankName: true,
}
