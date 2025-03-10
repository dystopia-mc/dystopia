package bot

import (
	"context"
	"errors"
	"fmt"
	"github.com/SevereCloud/vksdk/v3/api"
	"github.com/SevereCloud/vksdk/v3/api/params"
	"github.com/SevereCloud/vksdk/v3/events"
	"github.com/SevereCloud/vksdk/v3/longpoll-bot"
	"github.com/sandertv/gophertunnel/minecraft/text"
	"log/slog"
	"os"
	"slices"
	"strings"
)

var bot *Bot

func B() *Bot {
	if bot == nil {
		panic("bot is not initialized")
	}
	return bot
}

func Start() {
	key := "DYSTOPIA_VK_BOT_TOKEN"

	const (
		ivan   = 894434272
		sergey = 586471891
		ruslan = 878172584
	)
	var err error

	bot, err = New(slog.Default(), Config{
		Token: os.Getenv(key),
		AllowedUserIDs: []int{
			ivan,
			sergey,
			ruslan,
		},
		LogChat: 2000000003,
	})

	if err != nil {
		panic(err)
	}
	if err := bot.Start(); err != nil {
		panic(err)
	}
}

func Log(msg string) {
	go func() {
		if b := B(); b.HasLogChat() {
			msg := text.Clean(msg)
			b.sendMessagef(b.LogChat(), msg)
		}
	}()
}

type Config struct {
	Token  string
	Prefix string

	AllowedUserIDs []int
	LogChat        int
}

type Bot struct {
	l *slog.Logger
	c Config

	vk *api.VK
	lp *longpoll.LongPoll
}

func (b *Bot) HasLogChat() bool {
	return b.c.LogChat != 0
}

func (b *Bot) LogChat() int {
	return b.c.LogChat
}

func (b *Bot) setup() error {
	b.vk = api.NewVK(b.c.Token)
	if err := b.setupLongPoll(); err != nil {
		return err
	}

	b.lp.MessageNew(b.onMessage)
	return nil
}

func (b *Bot) sendMessagef(peerID int, msg string, a ...any) {
	builder := params.NewMessagesSendBuilder()

	builder.RandomID(0)
	builder.PeerID(peerID)
	builder.DisableMentions(true)
	builder.Message(fmt.Sprintf(msg, a...))

	if _, err := b.vk.MessagesSend(builder.Params); err != nil {
		b.l.Error("failed to send message", "err", err.Error())
		return
	}
}

func (b *Bot) onMessage(_ context.Context, obj events.MessageNewObject) {
	peer := obj.Message.PeerID

	if slices.Contains(b.c.AllowedUserIDs, obj.Message.FromID) {
		args := strings.Split(obj.Message.Text, " ")
		if len(args) == 1 {
			if args[0] == "" {
				return
			}
		}

		b.requestConsoleCommand(obj.Message.Text, peer)
	}
}

func (b *Bot) setupLongPoll() error {
	group, err := b.vk.GroupsGetByID(nil)
	if err != nil {
		return err
	}
	if len(group.Groups) == 0 {
		return errors.New("no groups")
	}

	lp, err2 := longpoll.NewLongPoll(b.vk, group.Groups[0].ID)
	if err2 != nil {
		return err2
	}

	b.lp = lp
	return nil
}

func (b *Bot) startLongPoll() error {
	return b.lp.Run()
}

func (b *Bot) Start() error {
	if err := b.startLongPoll(); err != nil {
		return err
	}
	return nil
}

func (b *Bot) Close() error {
	defer b.l.Info("Closed vk bot.")
	b.l.Info("Shutting down vk bot...")
	// do nothing 💩
	return nil
}

func New(l *slog.Logger, c Config) (*Bot, error) {
	b := &Bot{l: l, c: c}
	if err := b.setup(); err != nil {
		return nil, err
	}

	return b, nil
}
