package items

import (
	"github.com/df-mc/dragonfly/server/item"
	"github.com/df-mc/dragonfly/server/item/potion"
	"github.com/df-mc/dragonfly/server/world"
	"github.com/df-mc/dragonfly/server/world/sound"
	"github.com/k4ties/dystopia/plugins/practice/src/custom/entities"
	"image/color"
)

func NewHealingPotion(h entities.PotHandler) world.Item {
	return HealingPotion{h: h}
}

type HealingPotion struct {
	h entities.PotHandler
}

func (s HealingPotion) MaxCount() int {
	return 1
}

func (s HealingPotion) Use(tx *world.Tx, usr item.User, ctx *item.UseContext) bool {
	create := entities.NewHealPotion
	opts := world.EntitySpawnOpts{Position: eyePosition(usr), Velocity: usr.Rotation().Vec3().Mul(0.505)}

	tx.AddEntity(create(opts, usr, color.RGBA{A: 255}, s.h))
	tx.PlaySound(usr.Position(), sound.ItemThrow{})

	ctx.SubtractFromCount(1)
	return true
}

// EncodeItem ...
func (s HealingPotion) EncodeItem() (name string, meta int16) {
	return "minecraft:splash_potion", int16(potion.StrongHealing().Uint8())
}
