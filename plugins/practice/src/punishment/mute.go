package punishment

import (
	"errors"
	"time"
)

type Mute struct {
	punished     User
	expire       time.Time
	givenAt      time.Time
	from, reason string

	forever bool
}

func NewMuteModel(punished User, expire time.Time, givenAt time.Time, whoGave, reason string) Mute {
	return Mute{
		punished: punished,
		expire:   expire,
		givenAt:  givenAt,

		from:   whoGave,
		reason: reason,

		forever: expire == NilTime,
	}
}

func MuteFromModel(m Model) (Mute, error) {
	if m.Type != MuteType {
		return Mute{}, errors.New("invalid model type")
	}
	return Mute{
		punished: m.Punished,
		expire:   m.Expiration,
		givenAt:  m.GivenAt,
		from:     m.GivenFrom,
		reason:   m.Reason,
		forever:  m.Expiration == NilTime,
	}, nil
}

func MuteFromPunishment(p Punishment) (Mute, error) {
	return MuteFromModel(ToModel(p))
}

var (
	MuteType           = "punishment.mute.type"
	MuteNotifyFormat   = "<red><b>>></b></red> You've been muted for <grey>%s</grey>, for reason <grey>%s</grey>"
	UnmuteNotifyFormat = "<red><b>>></b></red> You've been un-muted."
)

func (m Mute) Punished() User {
	return m.punished
}

func (m Mute) Type() string {
	return MuteType
}

func (m Mute) Expiration() time.Time {
	return m.expire
}

func (m Mute) GivenAt() time.Time {
	return m.givenAt
}

func (m Mute) GivenFrom() string {
	return m.from
}

func (m Mute) Reason() string {
	return m.reason
}

func (m Mute) Forever() bool {
	return m.forever || m.expire == NilTime
}
