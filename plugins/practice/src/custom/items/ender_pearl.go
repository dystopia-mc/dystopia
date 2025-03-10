package items

import (
	"github.com/df-mc/dragonfly/server/item"
	"github.com/df-mc/dragonfly/server/world"
	"github.com/df-mc/dragonfly/server/world/sound"
	"github.com/go-gl/mathgl/mgl64"
	"github.com/k4ties/dystopia/plugins/practice/src/custom/entities"
	"time"
)

type Pearl struct {
	Smooth bool
}

func eyePosition(i item.User) mgl64.Vec3 {
	if p, ok := i.(interface {
		EyeHeight() float64
	}); ok {
		return i.Position().Add(mgl64.Vec3{0, p.EyeHeight(), 0})
	}

	return i.Position()
}

const pearlSpeed = 4.0

func (p Pearl) Use(tx *world.Tx, user item.User, ctx *item.UseContext) bool {
	r := user.Rotation()
	e := entities.NewEnderPearl(world.EntitySpawnOpts{Position: eyePosition(user), Velocity: r.Vec3().Mul(pearlSpeed)}, user, p.Smooth)

	tx.AddEntity(e)
	tx.PlaySound(user.Position(), sound.ItemThrow{})

	ctx.SubtractFromCount(1)
	return true
}

func (Pearl) Cooldown() time.Duration {
	return -1
}

func (Pearl) MaxCount() int {
	return 64
}

func (Pearl) EncodeItem() (name string, meta int16) {
	return "minecraft:ender_pearl", 0
}
