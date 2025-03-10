package items

type Shop struct{}

func (Shop) EncodeItem() (name string, meta int16) {
	return "minecraft:clay_ball", 0
}

func (Shop) MaxCount() int16 {
	return 1
}
