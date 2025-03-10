package rank

import "github.com/sandertv/gophertunnel/minecraft/text"

var OG = Rank{
	Name:     "OG",
	Format:   text.Bold + text.Orange,
	Priority: PriorityOG,

	DisplayRankName: true,
}
