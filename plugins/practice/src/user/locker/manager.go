package locker

import (
	"github.com/df-mc/dragonfly/server/player"
	"github.com/df-mc/dragonfly/server/session"
	"github.com/go-gl/mathgl/mgl32"
	"github.com/go-gl/mathgl/mgl64"
	"github.com/sandertv/gophertunnel/minecraft/protocol/packet"
)

type User interface {
	MustConn(func(session.Conn))
	Position() mgl64.Vec3
}

func LockCamera(u User) {
	u.MustConn(func(c session.Conn) {
		_ = c.WritePacket(&packet.UpdateClientInputLocks{
			Locks:    packet.ClientInputLockCamera,
			Position: vec64to32(u.Position().Add(mgl64.Vec3{0, player.Type.NetworkOffset(), 0})),
		})
	})
}

func LockMovement(u User) {
	u.MustConn(func(c session.Conn) {
		_ = c.WritePacket(&packet.UpdateClientInputLocks{
			Locks:    packet.ClientInputLockMovement,
			Position: vec64to32(u.Position().Add(mgl64.Vec3{0, player.Type.NetworkOffset(), 0})),
		})
	})
}

func LockCameraAndMovement(u User) {
	u.MustConn(func(c session.Conn) {
		_ = c.WritePacket(&packet.UpdateClientInputLocks{
			Locks:    packet.ClientInputLockCamera | packet.ClientInputLockMovement,
			Position: vec64to32(u.Position().Add(mgl64.Vec3{0, player.Type.NetworkOffset(), 0})),
		})
	})
}

func ResetLocks(u User) {
	u.MustConn(func(c session.Conn) {
		_ = c.WritePacket(&packet.UpdateClientInputLocks{
			Locks:    0,
			Position: vec64to32(u.Position().Add(mgl64.Vec3{0, player.Type.NetworkOffset(), 0})),
		})
	})
}

func vec64to32(a mgl64.Vec3) mgl32.Vec3 {
	return mgl32.Vec3{
		float32(a[0]), float32(a[1]), float32(a[2]),
	}
}
