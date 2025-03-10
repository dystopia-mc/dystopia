package command

import (
	"github.com/df-mc/dragonfly/server/cmd"
	"github.com/df-mc/dragonfly/server/world"
	"github.com/k4ties/dystopia/plugins/practice/src/ffa"
	"github.com/k4ties/dystopia/plugins/practice/src/instance"
	"strings"
)

type Closer struct {
	onlyStaffAndConsole
	Instance FFAInstance `cmd:"ffa"`
}

type FFAInstance string

func (FFAInstance) Type() string {
	return "ffa"
}

func (FFAInstance) Options(cmd.Source) []string {
	var names []string

	for _, i := range instance.AllInstances() {
		if f, ok := i.(*ffa.Instance); ok {
			names = append(names, f.Name())
		}
	}

	return names
}

func (c Closer) Run(_ cmd.Source, o *cmd.Output, tx *world.Tx) {
	i, ok := instance.ByName(strings.ToLower(string(c.Instance)))
	if !ok {
		o.Errorf("No instances with name %s", c.Instance)
		return
	}

	f, isFfa := i.(*ffa.Instance)
	if !isFfa {
		o.Errorf("Instance %s is not ffa.", c.Instance)
		return
	}

	if f.Closed() {
		f.Open()
		systemMessage(o, "You've successfully opened <grey>%s</grey>", c.Instance)
		return
	}

	f.Close(tx)
	systemMessage(o, "You've successfully closed <grey>%s</grey>", c.Instance)
}
