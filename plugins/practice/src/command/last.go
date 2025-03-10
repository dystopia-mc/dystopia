package command

import (
	"fmt"
	"github.com/df-mc/dragonfly/server/cmd"
	"github.com/df-mc/dragonfly/server/world"
	"github.com/k4ties/dystopia/plugins/practice/src/user"
	"slices"
	"strings"
	"time"
)

type LastMessageEnum struct {
	onlyStaff
	Messages cmd.SubCommand    `cmd:"messages"`
	Target   PlayerEnum        `cmd:"target"`
	Amount   cmd.Optional[int] `cmd:"amount"`
}

func (l LastMessageEnum) Run(_ cmd.Source, o *cmd.Output, _ *world.Tx) {
	u, ok := user.P().UserByName(string(l.Target))
	if !ok {
		o.Errorf("User not found: %s", l.Target)
		return
	}

	amount, hasAmount := l.Amount.Load()
	if !hasAmount {
		amount = 10
	}

	if amount <= 0 || amount > 100 {
		o.Errorf("Amount out of range: %d (1-100)", amount)
		return
	}

	format := "Latest <red>%s</red> <grey>%d</grey> messages:\n%s"
	messages := u.Data().LastMessages()
	if amount > len(messages) {
		amount = len(messages)
	}
	var columns = make([]string, amount)

	messages = messages[len(messages)-amount:]
	slices.Reverse(messages)

	for i, c := range messages {
		columns[i] = fmt.Sprintf("<grey>%d)</grey> %s <dark-grey>(%s)</dark-grey>", i+1, c.Message, c.WrittenAt.Format(time.TimeOnly))
	}
	systemMessage(o, format, u.Data().Name(), amount, strings.Join(columns, "\n"))
}

type LastMessage struct {
	onlyStaffAndConsole
	Messages cmd.SubCommand    `cmd:"messages"`
	Target   string            `cmd:"target"`
	Amount   cmd.Optional[int] `cmd:"amount"`
}

func (l LastMessage) Run(src cmd.Source, o *cmd.Output, tx *world.Tx) {
	LastMessageEnum{Target: PlayerEnum(l.Target), Amount: l.Amount}.Run(src, o, tx)
}

type LastCommandEnum struct {
	onlyStaff
	Commands cmd.SubCommand    `cmd:"commands"`
	Target   PlayerEnum        `cmd:"target"`
	Amount   cmd.Optional[int] `cmd:"amount"`
}

func (l LastCommandEnum) Run(_ cmd.Source, o *cmd.Output, _ *world.Tx) {
	u, ok := user.P().UserByName(string(l.Target))
	if !ok {
		o.Errorf("User not found: %s", l.Target)
		return
	}

	amount, hasAmount := l.Amount.Load()
	if !hasAmount {
		amount = 10
	}

	if amount <= 0 || amount > 100 {
		o.Errorf("Amount out of range: %d (1-100)", amount)
		return
	}

	format := "Latest <red>%s</red> <grey>%d</grey> commands:\n%s"
	commands := u.Data().LastCommands()
	if amount > len(commands) {
		amount = len(commands)
	}
	var columns = make([]string, amount)
	commands = commands[len(commands)-amount:]
	slices.Reverse(commands)

	for i, c := range commands {
		columns[i] = fmt.Sprintf("<grey>%d)</grey> %s <dark-grey>(%s)</dark-grey>", i+1, c.Line, c.WrittenAt.Format("15:04:05"))
	}
	systemMessage(o, format, u.Data().Name(), amount, strings.Join(columns, "\n"))
}

type LastCommand struct {
	onlyStaffAndConsole
	Commands cmd.SubCommand    `cmd:"commands"`
	Target   string            `cmd:"target"`
	Amount   cmd.Optional[int] `cmd:"amount"`
}

func (l LastCommand) Run(src cmd.Source, o *cmd.Output, tx *world.Tx) {
	LastCommandEnum{Target: PlayerEnum(l.Target), Amount: l.Amount}.Run(src, o, tx)
}
