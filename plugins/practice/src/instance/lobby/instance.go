package lobby

import (
	"github.com/df-mc/dragonfly/server/player"
	"github.com/df-mc/dragonfly/server/world"
	instance2 "github.com/k4ties/dystopia/plugins/practice/src/instance"
	"github.com/k4ties/dystopia/plugins/practice/src/kit"
)

const name = "lobby"

func Instance() instance2.Instance {
	return instance2.MustByName(name)
}

//func TransferWithRoutine(pl *instance.Player) {
//	transfer := func(p *player.Player, tx *world.Tx) {
//		if pl, in := instance.LookupPlayer(p); pl != nil && in != nil {
//			pl.SendKit(func(*player.Player) kit.Kit {
//				return Kit
//			}, tx)
//
//			Instance().Transfer(pl, tx)
//
//			if Instance().Active(pl.UUID()) {
//				p.Teleport(Instance().World().Spawn().Vec3Centre())
//			}
//		}
//	}
//
//	pl.ExecSafe(func(p *player.Player, tx *world.Tx) {
//		transfer(p, tx)
//	})
//}

func TransferWithRoutineSimple(p *player.Player) {
	if pl, in := instance2.LookupPlayer(p.UUID()); pl != nil && in != nil {
		if pl.Transferring() {
			return
		}

		pl.SendKit(func(*player.Player) kit.Kit {
			return Kit
		}, nil)

		Instance().Transfer(pl, nil)

		if Instance().Active(pl.UUID()) {
			pl.ExecSafe(func(p *player.Player, tx *world.Tx) {
				p.Teleport(Instance().(*instance2.Impl).WorldSpawn())
			})
		}
	}
}
