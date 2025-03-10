package punishment

import (
	"errors"
	"github.com/k4ties/dystopia/internal"
	"github.com/sandertv/gophertunnel/minecraft/text"
	"strings"
	"time"
)

type Ban struct {
	punished     User
	expire       time.Time
	givenAt      time.Time
	from, reason string

	forever bool
}

func NewBanModel(punished User, expire time.Time, givenAt time.Time, whoGave, reason string) Ban {
	return Ban{
		punished: punished,
		expire:   expire,
		givenAt:  givenAt,

		from:   whoGave,
		reason: reason,

		forever: expire == NilTime,
	}
}

func BanFromModel(m Model) (Ban, error) {
	if m.Type != BanType {
		return Ban{}, errors.New("invalid model type")
	}
	return Ban{
		punished: m.Punished,
		expire:   m.Expiration,
		givenAt:  m.GivenAt,
		from:     m.GivenFrom,
		reason:   m.Reason,
		forever:  m.Expiration == NilTime,
	}, nil
}

func BanFromPunishment(p Punishment) (Ban, error) {
	return BanFromModel(ToModel(p))
}

var (
	BanType       = "punishment.ban.type"
	BanKickFormat = []string{
		"<red><b>You are banned!</b></red> <dark-grey>(%s)</dark-grey>",
		"",
		"Reason: <grey>%s</grey>",
		"Expires in: <grey>%s</grey>",
		"Date: <grey>%s</grey>",
	}
)

func BanReason(b Ban) string {
	f := strings.Join(BanKickFormat, "\n")
	return text.Colourf(
		f,
		b.Punished().Name,
		b.Reason(),
		func() string {
			if b.Forever() {
				return "never"
			}
			return internal.DurationString(time.Until(b.Expiration()))
		}(),
		b.GivenAt().Format("02.01.2006"),
	)
}

func (b Ban) Punished() User {
	return b.punished
}

func (b Ban) Type() string {
	return BanType
}

func (b Ban) Expiration() time.Time {
	return b.expire
}

func (b Ban) GivenAt() time.Time {
	return b.givenAt
}

func (b Ban) GivenFrom() string {
	return b.from
}

func (b Ban) Reason() string {
	return b.reason
}

func (b Ban) Forever() bool {
	return b.forever || b.expire == NilTime
}
