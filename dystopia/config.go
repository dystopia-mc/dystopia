package dystopia

import (
	"github.com/df-mc/dragonfly/server"
	"github.com/sandertv/gophertunnel/minecraft/text"
	"gopkg.in/yaml.v3"
	"os"
	"slices"
	"strconv"
)

type Config struct {
	loadedFrom string

	Dystopia struct {
		MOTD    string `yaml:"MOTD"`
		SubName string `yaml:"SubName"`
	} `yaml:"Dystopia"`

	Whitelist struct {
		Enabled bool     `yaml:"Enabled"`
		Players []string `yaml:"Players"`
	} `yaml:"Whitelist"`

	Advanced struct {
		Port         uint16 `yaml:"Port"`
		XBLAuth      bool   `yaml:"XBOX-Authentication"`
		CachePath    string `yaml:"Cache-Path"`
		DefaultWorld string `yaml:"Default-World"`
		Database     string `yaml:"Database-Path"`
		Ranks        string `yaml:"Ranks-Path"`
	} `yaml:"Advanced"`

	Resources struct {
		Path       string `yaml:"Path"`
		Required   bool   `yaml:"Required"`
		ContentKey string `yaml:"Content-Key"`
	} `yaml:"Resources"`
}

func (c Config) convert() server.UserConfig {
	d := server.DefaultConfig()
	d.Network.Address = ":" + strconv.Itoa(int(c.Advanced.Port))

	d.Resources.Folder = c.Resources.Path
	d.Resources.Required = c.Resources.Required

	d.Resources.AutoBuildPack = false

	d.World.SaveData = false
	d.World.Folder = c.Advanced.CachePath + "/world"

	worldFolder := c.Advanced.DefaultWorld

	if _, err := os.Stat(worldFolder); os.IsNotExist(err) || worldFolder == "" {
		d.World.Folder = worldFolder
	}

	d.Players.SaveData = false
	d.Players.Folder = c.Advanced.CachePath + "/players"

	d.Server.Name = text.Colourf(c.Dystopia.MOTD)
	d.Server.DisableJoinQuitMessages = true
	d.Server.AuthEnabled = c.Advanced.XBLAuth

	return d
}

func MustReadConfig(path string) Config {
	c, err := ReadConfig(path)
	if err != nil {
		// if we cannot read config, we will try to create a new one
		c = DefaultConfig()
		c.MustWrite(path)
		c.loadedFrom = path
	}

	return c
}

func ReadConfig(path string) (Config, error) {
	var nop Config

	d, err := os.ReadFile(path)
	if err != nil {
		return nop, err
	}

	if err := yaml.Unmarshal(d, &nop); err != nil {
		return nop, err
	}

	nop.loadedFrom = path
	return nop, nil
}

func (c Config) MustWrite(path string) {
	if err := c.Write(path); err != nil {
		panic(err)
	}
}

func (c Config) Write(path string) error {
	b, err := yaml.Marshal(c)
	if err != nil {
		return err
	}

	return os.WriteFile(path, b, 0644)
}

func (c Config) AddWhitelistPlayer(p string) Config {
	c.Whitelist.Players = append(c.Whitelist.Players, p)
	return c
}

func (c Config) RemoveWhitelistPlayer(p string) Config {
	c.Whitelist.Players = remove(c.Whitelist.Players, slices.Index(c.Whitelist.Players, p))
	return c
}

func (c Config) ToggleWhitelist() Config {
	c.Whitelist.Enabled = !c.Whitelist.Enabled
	return c
}

func DefaultConfig() Config {
	c := Config{}

	c.Dystopia.MOTD = "<red>Dystopia</red>"
	c.Dystopia.SubName = ""

	c.Advanced.CachePath = "assets"
	c.Advanced.Port = 19132
	c.Advanced.XBLAuth = true
	c.Advanced.DefaultWorld = ""
	c.Advanced.Database = "dystopia.db"
	c.Advanced.Ranks = c.Advanced.CachePath + "/ranks"

	c.Whitelist.Enabled = false
	c.Whitelist.Players = []string{"player1", "player2"}

	c.Resources.Path = c.Advanced.CachePath + "/resources"
	c.Resources.Required = true
	c.Resources.ContentKey = "enter key here"

	return c
}

func (c Config) syncWithFile(path string) {
	c.MustWrite(path)
}

func remove(slice []string, index int) []string {
	return append(slice[:index], slice[index+1:]...)
}
