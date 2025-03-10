package hud

import "strconv"

type Element byte

const (
	PaperDoll Element = iota
	Armour
	ToolTips
	TouchControls
	Crosshair
	HotBar
	Health
	ProgressBar
	Hunger
	AirBubbles
	HorseHealth
)

func (e Element) String() string {
	switch e {
	case PaperDoll:
		return "PaperDoll"
	case Armour:
		return "Armour"
	case ToolTips:
		return "Tool Tips"
	case TouchControls:
		return "Touch Controls"
	case Crosshair:
		return "Crosshair"
	case HotBar:
		return "HotBar"
	case Health:
		return "Health"
	case ProgressBar:
		return "Progress Bar"
	case Hunger:
		return "Hunger"
	case AirBubbles:
		return "Air Bubbles"
	case HorseHealth:
		return "Horse Health"
	}

	panic("unknown element " + strconv.Itoa(int(e)))
}

func All() []Element {
	return []Element{0, 1, 2, 3, 4, 5, 6, 7, 8, 9, 10}
}
