package punishment

import (
	"github.com/google/uuid"
	"strings"
	"time"
)

type Punishment interface {
	Punished() User
	Type() string
	Expiration() time.Time
	GivenAt() time.Time
	GivenFrom() string
	Reason() string
}

func ToModel(p Punishment) Model {
	return Model{
		Punished:   p.Punished(),
		Type:       p.Type(),
		Expiration: p.Expiration(),
		GivenAt:    p.GivenAt(),
		GivenFrom:  p.GivenFrom(),
		Reason:     p.Reason(),
	}
}

var NilTime time.Time

func FromModel(m Model) Punishment {
	return impl{m}
}

type impl struct {
	Model
}

func (i impl) Punished() User {
	return i.Model.Punished
}

func (i impl) Type() string {
	return i.Model.Type
}

func (i impl) Expiration() time.Time {
	return i.Model.Expiration
}

func (i impl) GivenAt() time.Time {
	return i.Model.GivenAt
}

func (i impl) GivenFrom() string {
	return i.Model.GivenFrom
}

func (i impl) Reason() string {
	return i.Model.Reason
}

func Expired(p Punishment) bool {
	if p.Expiration() == NilTime {
		return false
	}
	if p.Expiration().Before(time.Now()) {
		return true
	}
	return false
}

type Model struct {
	Punished   User
	Type       string
	Expiration time.Time
	GivenAt    time.Time
	GivenFrom  string
	Reason     string
}

var Nil = Model{}

func (s Model) raw() rawModel {
	exp := s.Expiration.Unix()
	if s.Expiration == NilTime {
		exp = -1
	}
	pu := s.Punished

	return rawModel{
		Name:      pu.Name,
		UUID:      pu.UUID.String(),
		XUID:      pu.XUID,
		DeviceIDs: strings.Join(pu.DeviceIDs, ","),
		IPs:       strings.Join(pu.IPs, ","),

		PunishmentType: s.Type,
		Expiration:     exp,
		GivenAt:        s.GivenAt.Unix(),
		GivenFrom:      s.GivenFrom,
		Reason:         s.Reason,
	}
}

type rawModel struct {
	Name string
	UUID string
	XUID string

	DeviceIDs string
	IPs       string

	PunishmentType string `gorm:"column:type"`
	Expiration     int64  `gorm:"column:expiration"`
	GivenAt        int64  `gorm:"column:given_at"`
	GivenFrom      string `gorm:"column:given_from"`
	Reason         string `gorm:"column:reason"`
}

func (r rawModel) model() Model {
	exp := time.Unix(r.Expiration, 0)
	if r.Expiration <= -1 {
		exp = NilTime
	}

	u := User{
		Name:      r.Name,
		UUID:      uuid.MustParse(r.UUID),
		XUID:      r.XUID,
		DeviceIDs: strings.Split(r.DeviceIDs, ","),
		IPs:       strings.Split(r.IPs, ","),
	}

	return Model{
		Punished:   u,
		Type:       r.PunishmentType,
		Expiration: exp,
		GivenAt:    time.Unix(r.GivenAt, 0),
		GivenFrom:  r.GivenFrom,
		Reason:     r.Reason,
	}
}
