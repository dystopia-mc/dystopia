package ffa

import (
	_ "embed"
	"github.com/k4ties/dystopia/internal/embeddable"
)

//go:embed closed.json
var closed []byte

type Closer struct {
	Closed []string
}

var Closed = embeddable.MustJSON[Closer](closed)
