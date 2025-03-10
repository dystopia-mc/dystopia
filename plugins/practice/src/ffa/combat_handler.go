package ffa

import (
	"github.com/df-mc/dragonfly/server/entity"
	"github.com/df-mc/dragonfly/server/event"
	"github.com/df-mc/dragonfly/server/item"
	"github.com/df-mc/dragonfly/server/player"
	"github.com/df-mc/dragonfly/server/world"
	"github.com/google/uuid"
	"github.com/k4ties/dystopia/internal"
	"github.com/k4ties/dystopia/plugins/practice/src/custom/items"
	"github.com/k4ties/dystopia/plugins/practice/src/instance"
	"github.com/k4ties/dystopia/plugins/practice/src/user"
	"github.com/sandertv/gophertunnel/minecraft/text"
	"time"
)

type CombatHandler interface {
	HandleQuit(*player.Player)
	HandleAttackEntity(*player.Context, world.Entity, *float64, *float64, *bool) (canContinueChecks bool)
	HandleDeath(*player.Player, world.DamageSource, *bool)
	HandleSplash(affected []*event.Context[entity.Living], conf *entity.ProjectileBehaviourConfig)
	HandleHurt(ctx *player.Context, damage *float64, immune bool, attackImmunity *time.Duration, src world.DamageSource)
}

type combatHandler struct {
	u *user.User
	p *player.Player

	i  *Instance
	pl *instance.Player
}

func (c *combatHandler) HandleHurt(shakil *player.Context, _ *float64, _ bool, _ *time.Duration, src world.DamageSource) {
	_, isAttackSrc := src.(entity.AttackDamageSource)
	proj, isProjectileSrc := src.(entity.ProjectileDamageSource)

	if !isAttackSrc && !isProjectileSrc {
		return
	}

	// if player got hit by pearl/snowball or any other thing, we redirect it to HandleAttackEntity
	if isProjectileSrc {
		if o := proj.Owner; o != nil {
			attacker, in := LookupPlayer(shakil.Val().UUID())
			if attacker == nil || in == nil {
				// we don't care if it is not ffa
				return
			}

			c.HandleAttackEntity(shakil, o, nil, nil, nil)
		}
	}
}

func (i *Instance) CombatHandler(id uuid.UUID) CombatHandler {
	p, ok := i.Player(id)
	if !ok {
		return NopCombatHandler{}
	}

	var u *user.User

	user.MustUsePool(func(pool *user.Pool) {
		pool.MustUser(p.Player, func(usr *user.User) {
			u = usr
		})
	})

	if u == nil {
		return NopCombatHandler{}
	}

	return &combatHandler{u: u, p: p.Player, i: i, pl: p}
}

func (c *combatHandler) HandleQuit(p *player.Player) {
	if pl, in := LookupPlayer(p.UUID()); pl != nil && in != nil {
		c.i.MustCombat(p.UUID(), func(ct *Combat, _ *instance.Player, _ *Instance) {
			if ct.Active() {
				if w, ok := ct.With().(*user.User); ok {
					wp, ok := w.Player()
					if !ok {
						return
					}

					if wPl, wIn := LookupPlayer(wp.UUID()); wPl != nil && wIn != nil {
						if wIn.InCombat(wp.UUID()) {
							wIn.StopCombat(wp.UUID())
						}
						if withPl, withIn := LookupPlayer(w.UUID()); withIn != nil && withPl != nil {
							myPots := internal.Count[items.HealingPotion](p.Inventory())
							myGApples := internal.Count[item.GoldenApple](p.Inventory())

							var msg = internal.FormatDeadMessage(internal.DeadMessageBlank, p.Name(), w.Name())

							if internal.KitIncludes[items.HealingPotion](in.Kit()(wp)) {
								msg = internal.FormatDeadMessage(internal.DeadMessageCount, p.Name(), myPots, "POTS", w.Name(), internal.Count[items.HealingPotion](wp.Inventory()), "POTS")
							} else if internal.KitIncludes[item.GoldenApple](in.Kit()(wp)) {
								msg = internal.FormatDeadMessage(internal.DeadMessageCount, p.Name(), myGApples, "GAPPLES", w.Name(), internal.Count[item.GoldenApple](wp.Inventory()), "GAPPLES")
							}

							wIn.Messagef(msg)
							defer wIn.ReKit(wPl.UUID(), nil) // do after stop of combat
						}
					}
				}

				c.i.StopCombat(p.UUID())
			}
		})
	}
}

func (c *combatHandler) HandleAttackEntity(ctx *player.Context, a world.Entity, _, _ *float64, _ *bool) (canContinueChecks bool) {
	canContinueChecks = true

	if ctx.Val().GameMode() != c.i.GameMode() {
		canContinueChecks = false
		return
	}

	selfPl, in := LookupPlayer(ctx.Val().UUID())
	if selfPl == nil || in == nil {
		canContinueChecks = false
		return
	}

	selfU, ok := user.P().User(ctx.Val().UUID())
	if !ok {
		canContinueChecks = false
		return
	}

	ap, ok1 := a.(*player.Player)
	if !ok1 {
		canContinueChecks = false
		return
	}

	if ap.GameMode() != c.i.GameMode() {
		canContinueChecks = false
		return
	}

	if ap.UUID() == ctx.Val().UUID() {
		// can't be in combat with yourself
		canContinueChecks = false
		return
	}

	self, ok2 := c.i.Combat(selfPl.UUID())
	if !ok2 {
		canContinueChecks = false
		return
	}

	attacked, ok3 := c.i.Combat(ap.UUID())
	if !ok3 {
		canContinueChecks = false
		return
	}

	user.MustUsePool(func(pl *user.Pool) {
		pl.MustUser(ap, func(attackedU *user.User) {
			if self.Active() {
				if !self.InCombatWith(attackedU) {
					ctx.Cancel()
					canContinueChecks = false
					c.p.Messagef(text.Red + "Interrupting is not allowed")
					return
				} else { // if my combat active and with attacked
					self.RenewCoolDown()
					attacked.RenewCoolDown()
				}
			} else { // if my combat is not active
				if attacked.Active() {
					if !self.InCombatWith(attackedU) {
						ctx.Cancel()
						canContinueChecks = false
						c.p.Messagef(text.Red + "Interrupting is not allowed")
						return
					} else { // if attacker combat is active and self combat is in with attacker
						self.RenewCoolDown()
						attacked.RenewCoolDown()
					}
				} else { // if attacked and our combat is not active
					self.Start(attackedU)
					attacked.Start(selfU)

					ap.Messagef(text.Colourf(CombatNotifyFormat, self.owner.Name()))
					c.p.Messagef(text.Colourf(CombatNotifyFormat, attacked.owner.Name()))
				}
			}
		})
	})

	return canContinueChecks
}

const CombatNotifyFormat = "<red><b>>></b></red> You're now in combat with <grey>%s</grey>"

func (c *combatHandler) HandleDeath(p *player.Player, src world.DamageSource, _ *bool) {
	if deadPl, deadIn := LookupPlayer(p.UUID()); deadIn != nil && deadPl != nil {
		if a, deathByAttack := src.(entity.AttackDamageSource); deathByAttack {
			if killerP, ok := a.Attacker.(*player.Player); ok {
				user.MustUsePool(func(pl *user.Pool) {
					pl.MustUser(killerP, func(killer *user.User) {
						deadIn.StopCombat(p.UUID())
						deadIn.StopCombat(killer.UUID())

						deadIn.reKit(killerP.UUID(), nil)
					})
				})
			}
		}
	}
}

func (c *combatHandler) HandleSplash(affected []*event.Context[entity.Living], conf *entity.ProjectileBehaviourConfig) {
	owner := conf.Owner
	ownerPl, ownerIn := LookupPlayer(owner.UUID()) // entity handle uuid is equals to player uuid
	if ownerPl == nil || ownerIn == nil {
		// cancel all
		for _, a := range affected {
			a.Cancel()
		}
		return
	}

	var ownerU *user.User
	var ownerIsUser = false

	user.P().MustUser(ownerPl.Player, func(u *user.User) {
		ownerU = u
		ownerIsUser = true
	})

	for _, ctx := range affected {
		l := ctx.Val()
		canBeHealed := false

		if p, ok := l.(*player.Player); ok {
			if p.UUID() == ownerPl.UUID() {
				canBeHealed = true
				continue
			}

			if pl, in := LookupPlayer(p.UUID()); pl != nil || in != nil {
				if !in.InCombat(p.UUID()) && !ownerIn.InCombat(owner.UUID()) {
					canBeHealed = true
				}

				if in.InCombat(p.UUID()) {
					c, _ := in.Combat(p.UUID())
					if ownerIsUser {
						if c.InCombatWith(ownerU) {
							canBeHealed = true
						}
					}
				}
			}
		}

		if !canBeHealed {
			ctx.Cancel()
		}
	}
}

type NopCombatHandler struct{}

func (n NopCombatHandler) HandleQuit(*player.Player)                             {}
func (n NopCombatHandler) HandleDeath(*player.Player, world.DamageSource, *bool) {}
func (n NopCombatHandler) HandleSplash([]*event.Context[entity.Living], *entity.ProjectileBehaviourConfig) {
}
func (n NopCombatHandler) HandleAttackEntity(*player.Context, world.Entity, *float64, *float64, *bool) bool {
	return false
}
func (n NopCombatHandler) HandleHurt(*player.Context, *float64, bool, *time.Duration, world.DamageSource) {
}
