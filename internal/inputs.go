package internal

import "github.com/sandertv/gophertunnel/minecraft/protocol"

type OS protocol.DeviceOS

func (o OS) String() string {
	switch o {
	case 1:
		return "Android"
	case 2:
		return "iOS"
	case 3:
		return "Mac OS"
	case 4:
		return "Amazon FireOS"
	case 5:
		return "Gear VR"
	case 6:
		return "VR Hololens"
	case 7:
		return "Windows 10"
	case 8:
		return "Windows 32"
	case 9:
		return "Dedicated"
	case 10:
		return "TV OS"
	case 11:
		return "PlayStation"
	case 12:
		return "Nintendo Switch"
	case 13:
		return "XBOX"
	case 14:
		return "Windows Phone"
	case 15:
		return "Linux"
	default:
		return "Unknown"
	}
}

type InputMode uint16

func (i InputMode) String() string {
	switch i {
	case 1:
		return "Keyboard & Mouse"
	case 2:
		return "Touch Screen"
	case 3:
		return "Controller"
	case 4:
		return "Motion Controller"
	default:
		return "Unknown"
	}
}
