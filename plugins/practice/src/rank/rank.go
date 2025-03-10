package rank

import (
	"errors"
	"github.com/sasha-s/go-deadlock"
	"gopkg.in/yaml.v3"
	"maps"
	"os"
	"path/filepath"
	"slices"
	"strings"
)

func init() {
	Register(Player, Owner, Builder, Manager, Moderator, Admin, Helper, Saint, Vip, OG, Famous)
}

type Rank struct {
	Name   string
	Format string

	DisplayRankName bool
	Priority        Priority
}

type YAMLRank struct {
	Rank struct {
		Name     string `yaml:"Name"`
		Format   string `yaml:"Format"`
		Priority int    `yaml:"Priority"`

		DisplayRankName bool `yaml:"Display-Rank-Name"`
	}
}

func ToYaml(r Rank) YAMLRank {
	return func() YAMLRank {
		y := YAMLRank{}

		y.Rank.Name = r.Name
		y.Rank.Format = r.Format
		y.Rank.DisplayRankName = r.DisplayRankName
		y.Rank.Priority = int(r.Priority)

		return y
	}()
}

func FromYAML(r YAMLRank) Rank {
	return func() Rank {
		return Rank{
			Name:     r.Rank.Name,
			Format:   r.Rank.Format,
			Priority: Priority(r.Rank.Priority),

			DisplayRankName: r.Rank.DisplayRankName,
		}
	}()
}

var ranks = struct {
	ranks map[string]Rank
	mu    deadlock.RWMutex
}{
	ranks: make(map[string]Rank),
}

func Register(r ...Rank) {
	ranks.mu.Lock()
	defer ranks.mu.Unlock()

	for _, rk := range r {
		ranks.ranks[rk.Name] = rk
	}
}

func ByName(name string) (Rank, bool) {
	for _, r := range List() {
		if strings.ToLower(r.Name) == strings.ToLower(name) {
			return r, true
		}
	}

	return Rank{}, false
}

func MustByName(name string) Rank {
	r, ok := ByName(name)
	if !ok {
		return Player
	}

	return r
}

func List() []Rank {
	ranks.mu.RLock()
	defer ranks.mu.RUnlock()

	return slices.Collect(maps.Values(ranks.ranks))
}

func RegisterFromFiles(path string) error {
	dir, err := os.ReadDir(path)
	if err != nil {
		return err
	}

	for _, entry := range dir {
		if entry.IsDir() {
			continue
		}
		if filepath.Ext(entry.Name()) != ".yaml" {
			continue
		}

		pathToRank := path + "/" + entry.Name()

		d, err := os.ReadFile(pathToRank)
		if err != nil {
			return err
		}

		var r YAMLRank
		if err := yaml.Unmarshal(d, &r); err != nil {
			return err
		}

		Register(FromYAML(r))
	}

	return nil
}

func MustRegisterFromFiles(path string) {
	if err := RegisterFromFiles(path); err != nil {
		panic(err)
	}
}

func Write(r Rank, dir string) error {
	_, err := os.ReadDir(dir)
	if err != nil {
		return errors.New("not a directory")
	}

	d, err := yaml.Marshal(ToYaml(r))
	if err != nil {
		return err
	}

	return os.WriteFile(dir+"/"+r.Name+"Player.yaml", d, 0644)
}

func MustWrite(r Rank, dir string) {
	if err := Write(r, dir); err != nil {
		panic(err)
	}
}

func IsDefault(r Rank) bool {
	return r.Name == "" && r.Priority == 5 || r == Player
}
