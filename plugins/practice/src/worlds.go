package src

import (
	_ "embed"
	"github.com/df-mc/dragonfly/server/player"
	"github.com/df-mc/dragonfly/server/world"
	plugin "github.com/k4ties/df-plugin/df-plugin"
	"github.com/k4ties/dystopia/internal/embeddable"
	"github.com/k4ties/dystopia/plugins/practice/src/ffa"
	"github.com/k4ties/dystopia/plugins/practice/src/ffa/knocker"
	"github.com/k4ties/dystopia/plugins/practice/src/ffa/nodebuff"
	"github.com/k4ties/dystopia/plugins/practice/src/ffa/sumo"
	instance2 "github.com/k4ties/dystopia/plugins/practice/src/instance"
	"github.com/k4ties/dystopia/plugins/practice/src/kit"
	"github.com/k4ties/dystopia/plugins/practice/src/mw"
	"github.com/k4ties/dystopia/plugins/practice/src/user/hud"
	"strings"
	"time"
)

type WorldConfig struct {
	Config struct {
		Name string

		Rotation [2]float64
		Spawn    [3]float64

		HeightThreshold instance2.HeightThresholdConfig
	}
}

var (
	//go:embed lobby.json
	lobbyJson []byte
	//go:embed nodebuff.json
	arenaJson []byte
	//go:embed sumo.json
	sumoJson []byte
	//go:embed knocker.json
	knockerJson []byte
)

var (
	lobbyConfig   = embeddable.MustJSON[WorldConfig](lobbyJson)
	arenaConfig   = embeddable.MustJSON[WorldConfig](arenaJson)
	sumoConfig    = embeddable.MustJSON[WorldConfig](sumoJson)
	knockerConfig = embeddable.MustJSON[WorldConfig](knockerJson)
)

func SetupWorldsManager(m *plugin.Manager, path string) {
	if err := mw.NewManager(m.Srv().World(), path, m.Logger()); err != nil {
		panic(err)
	}

	registerLobby(mw.M(), m)
	registerNodebuff(mw.M(), m)
	registerSumo(mw.M(), m)
	registerKnocker(mw.M(), m)
}

func registerLobby(mn *mw.Manager, m *plugin.Manager) {
	name := lobbyConfig.Config.Name

	w, ok := mn.World(name)
	if !ok {
		panic("no world with specified name on the config")
	}
	if err := mn.SetSpawn(name, lobbyConfig.Config.Spawn); err != nil {
		panic(err)
	}

	hidden := []hud.Element{
		hud.Health, hud.Hunger,
	}

	in := instance2.New("lobby", w, world.GameModeAdventure, m.Logger(), lobbyConfig.Config.Rotation, lobbyConfig.Config.HeightThreshold, hidden...)
	instance2.Register(name, in)
}

func registerNodebuff(mn *mw.Manager, m *plugin.Manager) {
	registerGame(mn, m, arenaConfig, nodebuff.Kit, ffa.Config{
		Name: "NoDebuff",
		Icon: "textures/items/potion_bottle_splash_heal.png",

		PearlCooldown: time.Second * 16,
	})
}

func registerSumo(mn *mw.Manager, m *plugin.Manager) {
	registerGame(mn, m, sumoConfig, sumo.Kit, ffa.Config{
		Name: "Sumo",
		Icon: "textures/dystopia/dog.png",
	})
}

func registerKnocker(mn *mw.Manager, m *plugin.Manager) {
	registerGame(mn, m, knockerConfig, knocker.Kit, ffa.Config{
		Name: "Knocker",
		Icon: "textures/items/wind_charge.png",
	})
}

func registerGame(mn *mw.Manager, m *plugin.Manager, config WorldConfig, k func(*player.Player) kit.Kit, c ffa.Config) {
	w, ok := mn.World(config.Config.Name)
	if !ok {
		panic("no world with specified name on the config")
	}
	if err := mn.SetSpawn(config.Config.Name, config.Config.Spawn); err != nil {
		panic(err)
	}

	in := instance2.New(config.Config.Name, w, world.GameModeAdventure, m.Logger(), config.Config.Rotation, config.Config.HeightThreshold)
	f := ffa.New(in.(*instance2.Impl), k, c)

	instance2.Register(strings.ToLower(c.Name), f)
}
