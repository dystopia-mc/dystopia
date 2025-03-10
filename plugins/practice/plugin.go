package practice

import (
	_ "embed"
	"github.com/df-mc/dragonfly/server/cmd"
	plugin "github.com/k4ties/df-plugin/df-plugin"
	"github.com/k4ties/dystopia/internal/restarter"
	root "github.com/k4ties/dystopia/plugins/practice/src"
	"github.com/k4ties/dystopia/plugins/practice/src/announcement"
	"github.com/k4ties/dystopia/plugins/practice/src/bot"
	"github.com/k4ties/dystopia/plugins/practice/src/command"
	"github.com/k4ties/dystopia/plugins/practice/src/console"
	"github.com/k4ties/dystopia/plugins/practice/src/handlers"
	"github.com/k4ties/dystopia/plugins/practice/src/instance/lobby"
	"github.com/k4ties/dystopia/plugins/practice/src/lcn"
	"github.com/k4ties/dystopia/plugins/practice/src/punishment"
	"github.com/k4ties/dystopia/plugins/practice/src/trademc"
	"github.com/k4ties/dystopia/plugins/practice/src/user"
	"time"
)

//go:embed plugin.toml
var config []byte

type PluginConfig struct {
	LoginHandler *handlers.Login

	WorldsPath   string
	DatabasePath string
	ConfigPath   string

	CloseFunc func() error
}

const (
	ShopKey = "..."
)

func Plugin(c PluginConfig) plugin.Plugin {
	m := plugin.M()
	root.SetupWorldsManager(m, c.WorldsPath)

	pool := regUserPool(c.DatabasePath)
	punishment.NewPool("punishments.db", m.Logger())

	practiceHandler := handlers.NewTranslator(pool, lobby.Instance())
	userHandler := handler(c.LoginHandler, practiceHandler)

	h := []plugin.PlayerHandler{
		userHandler,
	}

	go bot.Start()
	task := m.NewTask(func(m *plugin.Manager) {
		registerCommands(c.CloseFunc)
		restarter.DoOnRestart(func() {
			_ = c.CloseFunc()
		})

		go trademc.New(trademc.Config{
			ListenPort: 13337,
			Logger:     m.Logger(),
			Password:   ShopKey,
		})
		go lcn.Task()(m)
		time.AfterFunc(time.Second/2, func() {
			bot.Log("Started session")
		})
		go announcement.DoTask(m.Srv())
		go console.Start(console.Config{
			Logger:     m.Logger(),
			ConfigPath: c.ConfigPath,
			Server:     m.Srv(),
		})

		handlers.RefreshFFACombatHandler()
	})

	return plugin.New(
		plugin.MustUnmarshalConfig(config),
		task,
		h...,
	)
}

func regUserPool(databasePath string) *user.Pool {
	d, err := user.NewDB(databasePath)
	if err != nil {
		panic(err)
	}

	pool := user.NewPool(d)
	return pool
}

func handler(loginHandler, practiceHandler plugin.PlayerHandler) plugin.PlayerHandler {
	return punishment.NewHandler(user.NewHandler(loginHandler, practiceHandler.(*handlers.Translator)))
}

func registerCommands(c func() error) {
	for _, c := range []cmd.Command{
		cmd.New("ffa", "Requests FFA form", nil, command.FFA{}),
		cmd.New("hub", "Teleports you to the lobby", []string{"spawn", "lobby"}, command.Hub{}),
		cmd.New("rekit", "Resends kit if you're in ffa", nil, command.ReKit{}),
		cmd.New("close", "Opens/closes specified ffa instance", nil, command.Closer{}),
		cmd.New("kill", "Kills specified player in ffa", nil, command.Kill{}),
		cmd.New("surrender", "Kills you if you're in combat", []string{"leave", "suicide"}, command.SurRender{}),
		cmd.New("chat", "Disables/enables chat for players that rank is lower than admin", nil, command.Chat{}),
		cmd.New("broadcast", "Broadcasts any message to all of users", []string{"say"}, command.Broadcast{}),
		cmd.New("op", "Toggles owner rank of the specified player", nil, command.Op{}),
		cmd.New("reconnect", "Re connects you to the server", []string{"rejoin"}, command.ReConnect{}),
		cmd.New("rank", "Manages rank of the specified player", []string{"role"}, command.RankPlayerSet{}, command.RankPlayerEnumSet{}, command.RankPlayerGet{}, command.RankPlayerEnumGet{}, command.RankList{}, command.RankPlayerEnumUpdate{}, command.RankPlayerUpdate{}),
		cmd.New("msg", "Whispers to player specified message", []string{"w", "m", "whisper", "tell"}, command.Msg{}),
		cmd.New("whitelist", "Manages session whitelist", []string{"wl"}, command.WhitelistToggle{}, command.WhitelistAdd{}, command.WhitelistAddEnum{}, command.WhitelistRemove{}, command.WhitelistRemoveEnum{}),
		cmd.New("whois", "Gives information about player", nil, command.Whois{}, command.WhoisEnum{}),
		cmd.New("knockback", "Changes knockback in real-time", []string{"kb"}, command.KnockBackSetForce{}, command.KnockBackSetHeight{}, command.KnockBackSetImmunity{}, command.KnockBackGetForce{}, command.KnockBackGetHeight{}, command.KnockBackGetImmunity{}),
		cmd.New("ping", "Shows your or other player latency", []string{"latency", "connection"}, command.PingEnum{}, command.Ping{}),
		cmd.New("platform", "Shows your or other player platform", nil, command.Platform{}, command.PlatformConsole{}),
		cmd.New("gm", "Changes your or other player game mode", []string{"gamemode"}, command.Gm{}, command.GmConsole{}),
		cmd.New("clear", "Clears your or other player inventory", nil, command.ClearEnum{}, command.Clear{}),
		cmd.New("tp", "Teleports you or target to any destination", nil, command.TeleportToPos{}, command.TeleportToTarget{}),
		cmd.New("transfer", "Transfers player to specified instance", nil, command.TransferEnum{}, command.Transfer{}),
		cmd.New("ban", "Bans specified player", nil, command.BanEnum{}, command.BanOffline{}),
		cmd.New("unban", "Unbans specified player", nil, command.UnbanEnum{}, command.Unban{}),
		cmd.New("mute", "Mutes specified player", nil, command.MuteEnum{}, command.Mute{}),
		cmd.New("unmute", "Un-mutes specified player", nil, command.UnmuteEnum{}, command.Unmute{}),
		cmd.New("kick", "Kicks specified player", nil, command.KickEnum{}, command.Kick{}),
		cmd.New("list", "Shows a list of all players that currently online", []string{"players", "online"}, command.List{}),
		cmd.New("freeze", "Freezes specified player", nil, command.FreezeEnum{}, command.Freeze{}),
		cmd.New("unfreeze", "Unfreezes specified player", nil, command.UnFreezeEnum{}, command.UnFreeze{}),
		cmd.New("last", "Shows latest commands/messages of specified player", []string{"latest"}, command.LastMessageEnum{}, command.LastMessage{}, command.LastCommandEnum{}, command.LastCommand{}),
		cmd.New("restart", "Restarts server", nil, command.Restart{}),
		cmd.New("stop", "Closes and stops the server", nil, command.Stop{CloseFunc: c}),
	} {
		cmd.Register(c)
	}
}
