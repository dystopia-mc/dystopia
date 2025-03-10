package practice

import (
	"github.com/df-mc/dragonfly/server"
	"github.com/df-mc/dragonfly/server/block/cube"
	"github.com/df-mc/dragonfly/server/player"
	"github.com/df-mc/dragonfly/server/world"
	"github.com/google/uuid"
	plugin "github.com/k4ties/df-plugin/df-plugin"
)

type provider struct {
	player.NopProvider

	r   cube.Rotation
	srv *server.Server
}

func Provider(r cube.Rotation, m *plugin.Manager) player.Provider {
	return &provider{r: r, srv: m.Srv()}
}

func (p provider) Load(uuid.UUID, func(world.Dimension) *world.World) (player.Config, *world.World, error) {
	return player.Config{
		GameMode:  world.GameModeAdventure,
		Rotation:  p.r,
		Health:    20,
		MaxHealth: 20,
		Food:      20,
	}, p.srv.World(), nil
}
