package handlers

import (
	"github.com/df-mc/atomic"
	"github.com/df-mc/dragonfly/server/player"
	"github.com/df-mc/dragonfly/server/session"
	"github.com/go-gl/mathgl/mgl32"
	"github.com/go-gl/mathgl/mgl64"
	"github.com/google/uuid"
	ffa2 "github.com/k4ties/dystopia/plugins/practice/src/ffa"
	"github.com/k4ties/dystopia/plugins/practice/src/instance"
	user2 "github.com/k4ties/dystopia/plugins/practice/src/user"
	"github.com/sandertv/gophertunnel/minecraft/protocol"
	"github.com/sandertv/gophertunnel/minecraft/protocol/packet"
	"github.com/sasha-s/go-deadlock"
	"iter"
)

type sessionPlayer struct {
	hidden atomic.Bool

	name string
	uuid uuid.UUID

	entityRuntimeID uint64

	data packet.AddPlayer
}

type Settings struct {
	ffa2.NopCombatHandle
	u *user2.User
	p *player.Player

	players   map[uuid.UUID]*sessionPlayer
	playersMu deadlock.RWMutex

	combatActive atomic.Bool
}

func (s *Settings) HandleCombatStart(ctx *ffa2.Context, _ ffa2.CombatUser) {
	pl, in := ffa2.LookupPlayer(ctx.Val().UUID())
	if pl == nil || in == nil {
		return
	}

	s.combatActive.Store(true)
	s.hideAllPlayersAndStartCombat(pl.Player)
}

func (s *Settings) HandleCombatStop(owner, _ ffa2.CombatUser) {
	pl, in := ffa2.LookupPlayer(owner.UUID())
	if pl == nil || in == nil {
		return
	}

	s.combatActive.Store(false)
	s.hideAllPlayersAndStartCombat(pl.Player)
}

func NewSettings(u *user2.User, p *player.Player) *Settings {
	s := &Settings{u: u, p: p, players: make(map[uuid.UUID]*sessionPlayer)}
	return s
}

func (s *Settings) newPlayer(p *sessionPlayer) {
	s.playersMu.Lock()
	defer s.playersMu.Unlock()

	s.players[p.uuid] = p
}

func (s *Settings) hidden(runtimeID uint64) bool {
	p, ok := s.playerByRuntimeID(runtimeID)
	if !ok {
		return false
	}

	return p.hidden.Load()
}

func (s *Settings) setHidden(runtimeID uint64) {
	p, ok := s.playerByRuntimeID(runtimeID)
	if !ok {
		return
	}

	p.hidden.Store(true)
}

func (s *Settings) setNotHidden(runtimeID uint64) {
	p, ok := s.playerByRuntimeID(runtimeID)
	if !ok {
		return
	}

	p.hidden.Store(false)
}

func (s *Settings) removePlayer(id uuid.UUID) {
	s.playersMu.Lock()
	defer s.playersMu.Unlock()

	delete(s.players, id)
}

func (s *Settings) player(id uuid.UUID) (*sessionPlayer, bool) {
	s.playersMu.RLock()
	defer s.playersMu.RUnlock()

	p, ok := s.players[id]
	return p, ok
}

func (s *Settings) playerByRuntimeID(id uint64) (*sessionPlayer, bool) {
	s.playersMu.RLock()
	defer s.playersMu.RUnlock()

	for _, p := range s.players {
		if p.entityRuntimeID == id {
			return p, true
		}
	}

	return nil, false
}

func (s *Settings) playerByName(name string) (*sessionPlayer, bool) {
	s.playersMu.RLock()
	defer s.playersMu.RUnlock()

	for _, p := range s.players {
		if p.name == name {
			return p, true
		}
	}

	return nil, false
}

func (s *Settings) HandleClientPacket(ctx *player.Context, pk packet.Packet) {
	switch pk := pk.(type) {
	case *packet.PlayerSkin:
		pk.Skin.Trusted = true
	}
}

func RefreshFFACombatHandler() {
	for _, i := range instance.AllInstances() {
		if f, isFFA := i.(*ffa2.Instance); isFFA {
			f.SetCombatHandle(func(usr ffa2.CombatUser) ffa2.CombatHandle {
				u := usr.(*user2.User)
				p, ok := u.Player()
				if !ok {
					return ffa2.NopCombatHandle{}
				}

				return NewSettings(u, p)
			})
		}
	}
}

func (s *Settings) Players() iter.Seq[*sessionPlayer] {
	s.playersMu.RLock()
	defer s.playersMu.RUnlock()

	return func(yield func(*sessionPlayer) bool) {
		for _, p := range s.players {
			if !yield(p) {
				return
			}
		}
	}
}

func (s *Settings) CombatActive() bool {
	return s.combatActive.Load()
}

func (s *Settings) hideAllPlayersAndStartCombat(self *player.Player) {
	s.combatActive.Store(true)

	for p := range s.Players() {
		if s.hidden(p.entityRuntimeID) {
			pl, in := instance.LookupPlayer(self.UUID())
			if pl == nil || in == nil {
				continue
			}

			pl.MustConn(func(conn session.Conn) {
				_ = conn.WritePacket(&packet.RemoveActor{
					EntityUniqueID: int64(p.entityRuntimeID),
				})
			})
		}
		if !s.hidden(p.entityRuntimeID) {
			usr, ok := user2.P().User(p.uuid)
			if !ok {
				continue
			}
			if _, ok := usr.Player(); !ok || !usr.Online() {
				continue
			}

			pl, _ := usr.Player()
			self.Data().Session.StopShowingEntity(pl)
		}
	}
}

func (s *Settings) showAllPlayersAndStopCombat(self *player.Player) {
	s.combatActive.Store(false)

	for p := range s.Players() {
		if s.hidden(p.entityRuntimeID) {
			usr, ok := user2.P().User(p.uuid)
			if !ok {
				continue
			}
			if _, ok := usr.Player(); !ok || !usr.Online() {
				continue
			}

			pl, _ := usr.Player()
			self.Data().Session.StartShowingEntity(pl)

			plr, in := instance.LookupPlayer(self.UUID())
			if plr == nil || in == nil {
				continue
			}

			plr.MustConn(func(conn session.Conn) {
				d := &p.data
				d.Position = mgl64to32(pl.Position())
				d.Yaw = float32(pl.Rotation().Yaw())
				d.Pitch = float32(pl.Rotation().Pitch())

				_ = conn.WritePacket(&p.data)
			})
		}
	}
}

func mgl64to32(m mgl64.Vec3) mgl32.Vec3 {
	return mgl32.Vec3{
		float32(m[0]), float32(m[1]), float32(m[2]),
	}
}

func (s *Settings) HandleServerPacket(ctx *player.Context, pk packet.Packet) {
	switch pk := pk.(type) {
	case *packet.AddPlayer:
		s.newPlayer(&sessionPlayer{
			name:            pk.Username,
			uuid:            pk.UUID,
			entityRuntimeID: pk.EntityRuntimeID,
			data:            *pk,
		})

		if s.CombatActive() {
			s.setHidden(pk.EntityRuntimeID)
			ctx.Cancel()
		}
	case *packet.RemoveActor:
		if sp, ok := s.playerByRuntimeID(uint64(pk.EntityUniqueID)); ok {
			s.setNotHidden(sp.entityRuntimeID)
			s.removePlayer(sp.uuid)
		}
	case *packet.SetActorData:
		if !s.u.Settings().FFA.ShowOpponentCPS {
			pk.EntityMetadata[protocol.EntityDataKeyScore] = ""
		}
	case *packet.LevelSoundEvent:
		switch pk.SoundType {
		case packet.SoundEventExplode, packet.SoundEventThunder:
			if !s.u.Settings().FFA.LightningKill {
				ctx.Cancel()
			}
		case packet.SoundEventAttackStrong, packet.SoundEventAttack, packet.SoundEventAttackNoDamage:
			ctx.Cancel()
		default:
			return
		}
	case *packet.AddActor:
		if !s.u.Settings().FFA.LightningKill {
			if pk.EntityType == "minecraft:lightning_bolt" {
				ctx.Cancel()
			}
		}
	case *packet.SetTime:
		pk.Time = int32(user2.ParseTimeOfDay(s.u.Data().Settings().Visual.PersonalTime))
	}
}
