package knockback

import (
	"github.com/df-mc/atomic"
)

var w = struct {
	height   atomic.Float64
	force    atomic.Float64
	immunity atomic.Int64
}{}

func Setup(height, force float64, immunity int64) {
	SetHeight(height)
	SetForce(force)
	SetImmunity(immunity)
}

func SetForce(force float64) {
	w.force.Store(force)
}

func SetHeight(height float64) {
	w.height.Store(height)
}

func SetImmunity(immunity int64) {
	w.immunity.Store(immunity)
}

func Height() float64 {
	return w.height.Load()
}

func Force() float64 {
	return w.force.Load()
}

func Immunity() int64 {
	return w.immunity.Load()
}
