package knocker

import (
	"github.com/k4ties/dystopia/plugins/practice/src/ffa"
	"github.com/k4ties/dystopia/plugins/practice/src/instance"
)

const name = "knocker"

func Instance() *ffa.Instance {
	return instance.MustByName(name).(*ffa.Instance)
}
