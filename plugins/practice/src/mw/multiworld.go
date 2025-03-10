package mw

import (
	"fmt"
	"github.com/df-mc/dragonfly/server/block/cube"
	"github.com/df-mc/dragonfly/server/entity"
	"github.com/df-mc/dragonfly/server/player"
	"github.com/df-mc/dragonfly/server/world"
	"github.com/df-mc/dragonfly/server/world/mcdb"
	"github.com/go-gl/mathgl/mgl64"
	"github.com/sasha-s/go-deadlock"
	"golang.org/x/exp/maps"
	"iter"
	"log/slog"
	"os"
	"strings"
)

func M() *Manager {
	return m
}

var m *Manager

type Manager struct {
	path string
	log  *slog.Logger
	w    *world.World

	worldsMu deadlock.Mutex
	worlds   map[string]*world.World
}

func NewManager(w *world.World, path string, log *slog.Logger) error {
	dir, err := os.ReadDir(path)
	if err != nil {
		return fmt.Errorf("error loading world directory %s: %s", path, err)
	}

	manager := &Manager{
		path: path,
		log:  log,
		w:    w,

		worlds: make(map[string]*world.World),
	}

	//track := tracker.New(w)
	//go track.StartTracking()

	w.Handle(&Handler{l: log})

	for _, d := range dir {
		if !d.IsDir() {
			continue
		}
		if _, err := manager.CreateWorld(d.Name()); err != nil {
			return err
		}
	}

	m = manager
	return nil
}

func (m *Manager) DefaultWorld() *world.World {
	return m.w
}

func (m *Manager) World(name string) (*world.World, bool) {
	name = strings.ToLower(name)

	m.worldsMu.Lock()
	w, ok := m.worlds[name]
	m.worldsMu.Unlock()

	return w, ok
}

func (m *Manager) Worlds() []*world.World {
	m.worldsMu.Lock()
	defer m.worldsMu.Unlock()

	return maps.Values(m.worlds)
}

func (m *Manager) CreateWorld(name string) (*world.World, error) {
	name = strings.ToLower(name)

	prov, err := mcdb.Open(m.path + "/" + name)
	if err != nil {
		return nil, fmt.Errorf("error loading world %s: %s", name, err)
	}
	prov.Settings().Name = name

	w := world.Config{
		Log:      m.log,
		Provider: prov,
		Entities: entity.DefaultRegistry,

		ReadOnly: true,
	}.New()

	w.StopWeatherCycle()
	w.SetDefaultGameMode(world.GameModeAdventure)
	w.SetTime(6000)
	w.StopTime()
	w.SetTickRange(0)
	w.StopThundering()
	w.StopRaining()

	//track := tracker.New(w)
	//go track.StartTracking()

	w.Handle(&Handler{l: m.log})

	m.worldsMu.Lock()
	m.worlds[name] = w
	m.worldsMu.Unlock()
	return w, nil
}

func (m *Manager) DeleteWorld(name string) error {
	name = strings.ToLower(name)

	m.worldsMu.Lock()
	w, ok := m.worlds[name]

	if ok {
		for e := range entities(w) {
			if p, ok := e.(*player.Player); ok {
				m.w.Exec(func(tx *world.Tx) {
					tx.AddEntity(p.H())
				})
				p.Teleport(m.w.Spawn().Vec3Middle())
				continue
			}
			_ = e.Close()
		}
		delete(m.worlds, name)
		_ = w.Close()
	}
	m.worldsMu.Unlock()

	if err := os.RemoveAll(m.path + "/" + name); err != nil {
		return fmt.Errorf("error deleting world %s: %s", name, err)
	}

	return nil
}

func entities(w *world.World) iter.Seq[world.Entity] {
	var e iter.Seq[world.Entity]

	<-w.Exec(func(tx *world.Tx) {
		e = tx.Entities()
	})

	return e
}

func (m *Manager) Close() {
	m.worldsMu.Lock()
	defer m.worldsMu.Unlock()
	for _, w := range m.worlds {
		for e := range entities(w) {
			if p, ok := e.(*player.Player); ok {
				m.w.Exec(func(tx *world.Tx) {
					tx.AddEntity(p.H())
				})
				p.Teleport(m.w.Spawn().Vec3Middle())
				continue
			}
			_ = e.Close()
		}
		_ = w.Close()
	}
}

func (m *Manager) SetSpawn(name string, pos mgl64.Vec3) error {
	w, ok := m.World(name)
	if !ok {
		return fmt.Errorf("could not find world %s", name)
	}

	w.SetSpawn(cube.PosFromVec3(pos))
	return nil
}
