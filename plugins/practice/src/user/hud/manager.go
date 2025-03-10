package hud

import (
	"github.com/sandertv/gophertunnel/minecraft/protocol/packet"
)

type Writeable interface {
	WritePacket(pk packet.Packet) error
}

func Hide(m Writeable, e ...Element) {
	if len(e) > 0 {
		_ = m.WritePacket(&packet.SetHud{
			Elements:   toBytes(e...),
			Visibility: packet.HudVisibilityHide,
		})
	}
}

func Reset(m Writeable, e ...Element) {
	if len(e) > 0 {
		_ = m.WritePacket(&packet.SetHud{
			Elements:   toBytes(e...),
			Visibility: packet.HudVisibilityReset,
		})
	}
}

func toBytes(e ...Element) []byte {
	var buff []byte

	for _, elem := range e {
		buff = append(buff, byte(elem))
	}

	return buff
}
