package dystopia

import (
	"fmt"
	"github.com/akmalfairuz/legacy-version/legacyver"
	"github.com/df-mc/dragonfly/server/player/chat"
	plugin "github.com/k4ties/df-plugin/df-plugin"
	"github.com/k4ties/dystopia/plugins/practice"
	"github.com/k4ties/dystopia/plugins/practice/src/bot"
	"github.com/k4ties/dystopia/plugins/practice/src/handlers"
	"github.com/k4ties/dystopia/plugins/practice/src/handlers/whitelist"
	"github.com/k4ties/dystopia/plugins/practice/src/punishment"
	"github.com/k4ties/dystopia/plugins/practice/src/trademc"
	"github.com/k4ties/dystopia/plugins/practice/src/user"
	"github.com/sandertv/gophertunnel/minecraft"
	"github.com/sandertv/gophertunnel/minecraft/resource"
	"log/slog"
	"os"
	"os/signal"
	"path/filepath"
	"syscall"
)

type Dystopia struct {
	l *slog.Logger
	c Config

	m *plugin.Manager
}

func New(l *slog.Logger, c Config) *Dystopia {
	d := &Dystopia{c: c, l: l}
	d.setup()

	return d
}

func (d *Dystopia) oldProtocols() []minecraft.Protocol {
	return []minecraft.Protocol{
		legacyver.New766(),
		legacyver.New748(),
		legacyver.New729(),
		legacyver.New712(),
		legacyver.New686(),
		legacyver.New685(),
	}
}

func (d *Dystopia) setup() {
	d.m = plugin.NewManager(plugin.ManagerConfig{
		Logger:     d.l,
		UserConfig: d.c.convert(),
		SubName:    d.c.Dystopia.SubName,
		Packs:      d.loadPacks(),
		Protocols:  d.oldProtocols(),
	})
	whitelist.Setup(d.c.Whitelist.Enabled, d.c.Whitelist.Players...)

	practicePlugin := practice.Plugin(practice.PluginConfig{
		LoginHandler: d.loginHandler(),
		WorldsPath:   d.c.Advanced.CachePath + "/worlds",
		DatabasePath: d.c.Advanced.Database,
		ConfigPath:   d.c.loadedFrom,
		CloseFunc:    d.Close,
	})

	d.m.ToggleStatusCommand()
	d.m.Register(practicePlugin)
}

func (d *Dystopia) loginHandler() *handlers.Login {
	return handlers.NewLoginHandler()
}

func (d *Dystopia) loadPacks() (pool []*resource.Pack) {
	path := d.c.Resources.Path

	dir, err := os.ReadDir(path)
	if err != nil {
		panic("couldn't read dir while loading packs: " + err.Error())
	}

	for i, f := range dir {
		pathTo := filepath.Join(path, f.Name())
		pack, err := resource.ReadPath(pathTo)
		if err != nil {
			d.l.Error("dystopia: resource packs: cannot load directory: " + pathTo)
			continue
		}

		pool = append(pool, pack.WithContentKey(d.c.Resources.ContentKey))
		d.l.Debug(fmt.Sprintf("dystopia: loaded %d/%d packs", i, len(dir)))
	}

	return
}

func (d *Dystopia) Start() {
	chat.Global.Subscribe(chat.StdoutSubscriber{})

	d.CloseAllOnProgramEnd()
	d.m.ListenServer()
}

func (d *Dystopia) CloseAllOnProgramEnd() {
	c := make(chan os.Signal, 2)
	signal.Notify(c, syscall.SIGINT, syscall.SIGTERM)
	go func() {
		<-c
		bot.Log("Closed session")
		_ = d.Close()
	}()
}

func (d *Dystopia) Close() error {
	go func() {
		_ = bot.B().Close()
	}()

	_ = punishment.P().Close()
	_ = trademc.R().Close()

	go user.P().SaveAll()
	go func() {
		_ = d.m.Srv().Close()
	}()
	return nil
}
