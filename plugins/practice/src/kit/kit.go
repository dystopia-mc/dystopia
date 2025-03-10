package kit

import (
	"github.com/df-mc/dragonfly/server/entity/effect"
	"github.com/df-mc/dragonfly/server/item"
	"github.com/df-mc/dragonfly/server/player"
	"github.com/sandertv/gophertunnel/minecraft/text"
	"github.com/sasha-s/go-deadlock"
)

type Playered func(*player.Player) Kit

var kits = struct {
	v  map[string]Kit
	mu deadlock.RWMutex
}{
	v: make(map[string]Kit),
}

func Register(kit Kit, name string) {
	kits.mu.Lock()
	defer kits.mu.Unlock()
	kits.v[name] = kit
}

func MustByName(name string) Kit {
	k, ok := ByName(name)
	if !ok {
		panic("unknown kit " + name)
	}

	return k
}

func ByName(name string) (Kit, bool) {
	kits.mu.RLock()
	defer kits.mu.RUnlock()

	v, ok := kits.v[name]
	return v, ok
}

func UnRegister(name string) {
	kits.mu.Lock()
	defer kits.mu.Unlock()

	delete(kits.v, name)
}

type Armour [4]item.Stack
type Items map[int]item.Stack

var NopArmour = Armour{}
var NopItems = Items{}

func NewArmour(helmet, chestplate, leggings, boots item.Stack) Armour {
	return Armour{
		helmet,
		chestplate,
		leggings,
		boots,
	}
}

func NewItems(items ...item.Stack) Items {
	if len(items) > 36 {
		panic("cant have more than 36 items")
	}

	var m = make(Items)

	for i, entry := range items {
		m[i] = entry
	}

	return m
}

type Kit interface {
	Items() Items
	Armour() Armour
	Effects() []effect.Effect
}

func Send(k Kit, p *player.Player) {
	SendItems(k, p)
	SendArmour(k, p)
	SendEffects(k, p)
}

func SendItems(k Kit, p *player.Player) {
	for slot, i := range k.Items() {
		_ = p.Inventory().SetItem(slot, i)
	}
}

func SendArmour(k Kit, p *player.Player) {
	armour := p.Armour()

	armour.SetHelmet(k.Armour()[0])
	armour.SetChestplate(k.Armour()[1])
	armour.SetLeggings(k.Armour()[2])
	armour.SetBoots(k.Armour()[3])
}

func SendEffects(k Kit, p *player.Player) {
	for _, e := range k.Effects() {
		p.AddEffect(e)
	}
}

func New(i Items, a Armour, e ...effect.Effect) Kit {
	return kit{i, a, e}
}

type kit struct {
	i Items
	a Armour
	e []effect.Effect
}

var Empty = func() Kit {
	return New(Items{}, Armour{})
}()

func (k kit) Items() Items {
	return k.i
}

func (k kit) Armour() Armour {
	return k.a
}

func (k kit) Effects() []effect.Effect {
	return k.e
}

const identifier = "identifier"

func LoadIdentifier(i item.Stack) string {
	v, ok := i.Value(identifier)
	if !ok {
		return ""
	}

	str, ok := v.(string)
	if !ok {
		return ""
	}

	return str
}

func ApplyIdentifier(id string, i item.Stack) item.Stack {
	return i.WithValue(identifier, id)
}

func FillNames(title string, i item.Stack) item.Stack {
	return i.WithCustomName(text.Colourf("<white>%s</white>\n<red>Click</red>", title)).WithLore(text.Colourf(text.Reset + "<red>dystopia</red>"))
}
