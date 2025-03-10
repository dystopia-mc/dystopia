package entities

import (
	"github.com/df-mc/dragonfly/server/block/cube/trace"
	"github.com/df-mc/dragonfly/server/entity"
	"github.com/df-mc/dragonfly/server/player"
	"github.com/df-mc/dragonfly/server/session"
	"github.com/df-mc/dragonfly/server/world"
	"github.com/df-mc/dragonfly/server/world/sound"
	"github.com/go-gl/mathgl/mgl32"
	"github.com/go-gl/mathgl/mgl64"
	"github.com/google/uuid"
	plugin "github.com/k4ties/df-plugin/df-plugin"
	"github.com/sandertv/gophertunnel/minecraft/protocol/packet"
	"net"
)

func NewEnderPearl(opts world.EntitySpawnOpts, owner world.Entity, smooth bool) *world.EntityHandle {
	conf := enderPearlConf
	conf.Owner = owner.H()
	conf.Hit = func(e *entity.Ent, tx *world.Tx, target trace.Result) {
		teleport(e, tx, target, smooth)
	}

	return opts.New(entity.EnderPearlType, conf)
}

//size: 0.5
//speed: 2.8
//gravity: 0.05
//drag: 0.009

var enderPearlConf = entity.ProjectileBehaviourConfig{
	Gravity: 0.05,
	Drag:    0.009,
	Damage:  0,
	Sound:   sound.Teleport{},
}

func teleport(e *entity.Ent, tx *world.Tx, target trace.Result, smooth bool) {
	owner, _ := e.Behaviour().(*entity.ProjectileBehaviour).Owner().Entity(tx)
	if user, ok := owner.(user); ok {
		if smooth {
			m := plugin.M()
			if conn, ok := m.Conn(user.Name()); ok {
				_ = conn.WritePacket(&packet.MovePlayer{
					EntityRuntimeID: 1,
					Position:        vec64to32(target.Position()),
					Pitch:           float32(user.Rotation().Pitch()),
					Yaw:             float32(user.Rotation().Yaw()),
					HeadYaw:         float32(user.Rotation().Yaw()),
					OnGround:        user.OnGround(),
				})

				for _, v := range tx.Viewers(user.Position()) {
					if s, ok := v.(*session.Session); ok {
						if s.ClientData().DeviceID == user.DeviceID() {
							continue
						}

						v.ViewEntityTeleport(user, target.Position())
					}
				}
			}
		} else {
			user.Teleport(target.Position())
		}

		tx.PlaySound(user.Position(), sound.Teleport{})
		// user.Hurt(5, FallDamageSource{})
	}
}

func vec64to32(a mgl64.Vec3) mgl32.Vec3 {
	return mgl32.Vec3{
		float32(a[0]), float32(a[1]) + float32(player.Type.NetworkOffset()), float32(a[2]),
	}
}

type user interface {
	Teleport(pos mgl64.Vec3)
	Name() string
	UUID() uuid.UUID
	OnGround() bool
	Addr() net.Addr
	Data() player.Config
	DeviceID() string

	entity.Living
}
