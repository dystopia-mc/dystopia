package entities

import (
	"github.com/df-mc/dragonfly/server/block/cube/trace"
	"github.com/df-mc/dragonfly/server/entity"
	"github.com/df-mc/dragonfly/server/entity/effect"
	"github.com/df-mc/dragonfly/server/event"
	"github.com/df-mc/dragonfly/server/item/potion"
	"github.com/df-mc/dragonfly/server/world"
	"github.com/df-mc/dragonfly/server/world/particle"
	"github.com/df-mc/dragonfly/server/world/sound"
	"github.com/go-gl/mathgl/mgl64"
	"image/color"
)

type PotHandler interface {
	HandleSplash(affected []*event.Context[entity.Living], conf *entity.ProjectileBehaviourConfig)
}

type nopPotHandler struct{}

func (n nopPotHandler) HandleSplash([]*event.Context[entity.Living], *entity.ProjectileBehaviourConfig) {
}

func NewHealPotion(opts world.EntitySpawnOpts, owner world.Entity, colour color.RGBA, h PotHandler) *world.EntityHandle {
	if h == nil {
		h = nopPotHandler{}
	}

	conf := splashPotionConf
	conf.Potion = potion.StrongHealing()
	conf.Particle = particle.Splash{Colour: colour}
	conf.Owner = owner.H()
	conf.Hit = potionSplash(h, &conf)

	return opts.New(entity.SplashPotionType, conf)
}

var splashPotionConf = entity.ProjectileBehaviourConfig{
	Gravity: 0.045,
	Drag:    0.005,
	Damage:  -1,
	Sound:   sound.GlassBreak{},
}

func potionSplash(h PotHandler, conf *entity.ProjectileBehaviourConfig) func(e *entity.Ent, tx *world.Tx, res trace.Result) {
	return func(e *entity.Ent, tx *world.Tx, res trace.Result) {
		pos := e.Position()
		box := e.H().Type().BBox(e).Translate(pos)

		affected := tx.EntitiesWithin(box.GrowVec3(mgl64.Vec3{2.5, 4, 2.5}))

		var ctxList []*event.Context[entity.Living]
		for a := range affected {
			if living, ok := a.(entity.Living); ok {
				context := event.C(living)
				ctxList = append(ctxList, context)
			}
		}

		h.HandleSplash(ctxList, conf)

		var willBeHealed []entity.Living
		for _, context := range ctxList {
			if !context.Cancelled() {
				willBeHealed = append(willBeHealed, context.Val())
			}
		}

		for _, l := range willBeHealed {
			l.Heal(4.5, effect.InstantHealingSource{})
		}
	}
}
