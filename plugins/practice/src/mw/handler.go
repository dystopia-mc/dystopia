package mw

import (
	"github.com/df-mc/dragonfly/server/block/cube"
	"github.com/df-mc/dragonfly/server/world"
	"github.com/df-mc/dragonfly/server/world/sound"
	"github.com/go-gl/mathgl/mgl64"
	"log/slog"
)

type Handler struct {
	world.NopHandler
	l *slog.Logger
}

//func (h Handler) HandleClose(_ *world.Tx) {
//	if err := h.tr.Close(); err != nil {
//		h.l.Error(err.Error())
//	}
//}12

func (h Handler) HandleExplosion(ctx *world.Context, _ mgl64.Vec3, _ *[]world.Entity, _ *[]cube.Pos, _ *float64, _ *bool) {
	ctx.Cancel()
}

func (h Handler) HandleLiquidFlow(ctx *world.Context, _, _ cube.Pos, _ world.Liquid, _ world.Block) {
	ctx.Cancel()
}

func (h Handler) HandleLiquidDecay(ctx *world.Context, _ cube.Pos, _, _ world.Liquid) {
	ctx.Cancel()
}

func (h Handler) HandleLiquidHarden(ctx *world.Context, _ cube.Pos, _, _, _ world.Block) {
	ctx.Cancel()
}

func (h Handler) HandleSound(ctx *world.Context, s world.Sound, _ mgl64.Vec3) {
	if _, ok := s.(sound.Attack); ok {
		ctx.Cancel()
		return
	}
}

func (h Handler) HandleFireSpread(ctx *world.Context, _, _ cube.Pos) {
	ctx.Cancel()
}

func (h Handler) HandleBlockBurn(ctx *world.Context, _ cube.Pos) {
	ctx.Cancel()
}

func (h Handler) HandleCropTrample(ctx *world.Context, _ cube.Pos) {
	ctx.Cancel()
}

func (h Handler) HandleLeavesDecay(ctx *world.Context, _ cube.Pos) {
	ctx.Cancel()
}
