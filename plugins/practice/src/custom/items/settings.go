package items

type Settings struct{}

func (Settings) EncodeItem() (name string, meta int16) {
	return "minecraft:coal", 0
}
