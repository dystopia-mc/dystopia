package punishment

import (
	"errors"
	"fmt"
	"github.com/google/uuid"
	"github.com/k4ties/dystopia/internal"
	"github.com/k4ties/dystopia/plugins/practice/src/bot"
	"github.com/k4ties/dystopia/plugins/practice/src/user"
	"github.com/k4ties/dystopia/plugins/practice/src/user/chatter"
	"github.com/sandertv/gophertunnel/minecraft/text"
	"github.com/sasha-s/go-deadlock"
	"log/slog"
	"slices"
	"strings"
	"sync"
	"time"
)

var p *Pool

func P() *Pool {
	if p == nil {
		panic("pool is not initialized")
	}

	return p
}

type Pool struct {
	d *Database

	once sync.Once

	models   map[uuid.UUID]Model
	modelsMu deadlock.RWMutex
}

func NewPool(path string, l *slog.Logger) *Pool {
	db, err := NewDatabase(path, l)
	if err != nil {
		panic(err)
	}

	p = &Pool{d: db, models: make(map[uuid.UUID]Model)}
	p.loadFromDB()
	return p
}

var KickFormat = []string{
	"<red>You've been kicked!</red> <dark-grey>(%s)</dark-grey>",
	"",
	"Reason: <grey>%s</grey>",
	"Issued by: <grey>%s</grey>",
	"Date: <grey>%s</grey>",
}

var (
	GlobalMessageMuteFormat = "<yellow>%s was muted</yellow>"
	GlobalMessageBanFormat  = "<yellow>%s was banned</yellow>"
	GlobalMessageKickFormat = "<yellow>%s was kicked</yellow>"

	GlobalMessageUnmuteFormat = "<yellow>%s was unmuted</yellow>"
	GlobalMessageUnbanFormat  = "<yellow>%s was unbanned</yellow>"
)

func KickReason(u *user.User, reason, whoGave string, whenGave time.Time) string {
	return text.Colourf(
		strings.Join(KickFormat, "\n"),
		u.Data().Name(),
		reason,
		whoGave,
		whenGave.Format("02.01.2006"),
	)
}

func (p *Pool) Kick(u *user.User, reason, whoGave string) error {
	pl, ok := u.Player()
	if !u.Online() || !ok {
		return errors.New("player not online")
	}

	msg := text.Colourf(GlobalMessageKickFormat, u.Data().Name())
	chatter.Messagef(msg)

	pl.Disconnect(KickReason(u, reason, whoGave, time.Now()))
	bot.Log(fmt.Sprintf("%s was kicked by %s for reason %s", u.Data().Name(), whoGave, reason))
	return nil
}

func (p *Pool) Ban(u *user.User, expire time.Time, reason, whoGave string) error {
	if ban, banned := p.Banned(u); banned {
		if Expired(ban) {
			_ = p.delete(ban)
		} else {
			return errors.New("already banned")
		}
	}
	//if _, muted := p.Muted(u); muted {
	//	return errors.New("please unmute player before ban")
	//}

	pu := User{
		Name:      u.Data().Name(),
		UUID:      u.Data().UUID(),
		XUID:      u.Data().XUID(),
		DeviceIDs: u.Data().DIDs(),
		IPs:       u.Data().IPs(),
	}

	ban := NewBanModel(pu, expire, time.Now(), whoGave, reason)
	if err := p.add(ban); err != nil {
		return err
	}
	if p, ok := u.Player(); ok && u.Online() {
		p.Disconnect(BanReason(ban))
	}

	msg := text.Colourf(GlobalMessageBanFormat, u.Data().Name())
	chatter.Messagef(msg)

	bot.Log(fmt.Sprintf("%s was banned by %s for %s for reason %s", u.Data().Name(), whoGave, internal.DurationString(time.Until(expire)), reason))
	return nil
}

func (p *Pool) Unban(u *user.User, by string) error {
	ban, banned := p.Banned(u)
	if !banned {
		return errors.New("not banned")
	}
	if Expired(ban) {
		_ = p.delete(ban)
		return errors.New("not banned")
	}
	if err := p.delete(ban); err != nil {
		return err
	}

	msg := text.Colourf(GlobalMessageUnbanFormat, u.Data().Name())
	chatter.Messagef(msg)

	bot.Log(fmt.Sprintf("%s was unbanned by %s", u.Data().Name(), by))
	return nil
}

func (p *Pool) Banned(u *user.User) (Ban, bool) {
	m, ok := p.modelExists(u.Data().Name(), u.Data().XUID(), u.Data().UUID(), u.Data().IPs(), u.Data().DIDs(), BanType)
	if !ok {
		return Ban{}, false
	}
	mu, err := BanFromModel(m.Model)
	if err != nil {
		return Ban{}, false
	}

	return mu, true
}

func (p *Pool) Mute(u *user.User, expire time.Time, reason, whoGave string) error {
	if mute, muted := p.Muted(u); muted {
		if Expired(mute) {
			_ = p.delete(mute)
		} else {
			return errors.New("already muted")
		}
	}
	//if _, banned := p.Banned(u); banned {
	//	return errors.New("please unban player before mute")
	//}

	pu := User{
		Name:      u.Data().Name(),
		UUID:      u.Data().UUID(),
		XUID:      u.Data().XUID(),
		DeviceIDs: u.Data().DIDs(),
		IPs:       u.Data().IPs(),
	}

	mute := NewMuteModel(pu, expire, time.Now(), whoGave, reason)
	if err := p.add(mute); err != nil {
		return err
	}
	if p, ok := u.Player(); ok && u.Online() {
		p.Messagef(text.Colourf(MuteNotifyFormat, internal.DurationString(time.Until(expire)), reason))
	}

	msg := text.Colourf(GlobalMessageMuteFormat, u.Data().Name())
	chatter.Messagef(msg)

	bot.Log(fmt.Sprintf("%s was muted by %s for %s for reason %s", u.Data().Name(), whoGave, internal.DurationString(time.Until(expire)), reason))
	return nil
}

func (p *Pool) Unmute(u *user.User, by string) error {
	mute, muted := p.Muted(u)
	if !muted {
		return errors.New("not muted")
	}
	if Expired(mute) {
		_ = p.delete(mute)
		return errors.New("not muted")
	}
	if err := p.delete(mute); err != nil {
		return err
	}
	if p, ok := u.Player(); ok && u.Online() {
		p.Messagef(text.Colourf(UnmuteNotifyFormat))
	}

	chatter.Messagef(GlobalMessageUnmuteFormat, u.Data().Name())
	bot.Log(fmt.Sprintf("%s was unmuted by %s", u.Data().Name(), by))
	return nil
}

func (p *Pool) Muted(u *user.User) (Mute, bool) {
	m, ok := p.modelExists(u.Data().Name(), u.Data().XUID(), u.Data().UUID(), u.Data().IPs(), u.Data().DIDs(), MuteType)
	if !ok {
		return Mute{}, false
	}
	mu, err := MuteFromModel(m.Model)
	if err != nil {
		return Mute{}, false
	}
	if Expired(mu) {
		_ = p.delete(mu)
		return Mute{}, false
	}

	return mu, true
}

func (p *Pool) modelExists(name, xuid string, id uuid.UUID, ips, dids []string, Type ...string) (model identifiedModel, exists bool) {
	if m, ok := p.modelByName(name, Type...); ok {
		if !Expired(FromModel(m.Model)) {
			return m, true
		}
	}
	if m, ok := p.modelByXUID(xuid, Type...); ok {
		if !Expired(FromModel(m.Model)) {
			return m, true
		}
	}
	if m, ok := p.modelByUUID(id, Type...); ok {
		if !Expired(FromModel(m.Model)) {
			return m, true
		}
	}
	for _, ip := range ips {
		if m, ok := p.modelByIP(ip, Type...); ok {
			if !Expired(FromModel(m.Model)) {
				return m, true
			}
		}
	}
	for _, did := range dids {
		if m, ok := p.modelByDID(did, Type...); ok {
			if !Expired(FromModel(m.Model)) {
				return m, true
			}
		}
	}
	return identifiedModel{}, false
}

func (p *Pool) model(i uuid.UUID) (Model, bool) {
	p.modelsMu.RLock()
	defer p.modelsMu.RUnlock()

	model, ok := p.models[i]
	if !ok {
		return Model{}, false
	}

	return model, ok
}

func (p *Pool) modelUUID(m Model) (uuid.UUID, bool) {
	for _, u := range p.modelsList() {
		if m.Punished.Equals(u.Model.Punished) && m.Type == u.Model.Type && m.Expiration == u.Model.Expiration && m.Reason == u.Model.Reason && m.GivenAt == u.Model.GivenAt && m.GivenFrom == u.Model.GivenFrom {
			return u.UUID, true
		}
	}
	return uuid.Nil, false
}

type identifiedModel struct {
	UUID  uuid.UUID
	Model Model
}

func (p *Pool) modelByName(name string, Type ...string) (identifiedModel, bool) {
	hasType := false
	var t string

	if len(Type) >= 1 {
		hasType = true
		t = Type[0]
	}

	for _, m := range p.modelsList() {
		if m.Model.Punished.Name == name {
			if !hasType {
				return m, true
			}
			if m.Model.Type != t {
				continue
			}
			return m, true
		}
	}
	return identifiedModel{}, false
}

func (p *Pool) modelByXUID(xuid string, Type ...string) (identifiedModel, bool) {
	hasType := false
	var t string

	if len(Type) >= 1 {
		hasType = true
		t = Type[0]
	}

	for _, m := range p.modelsList() {
		if m.Model.Punished.XUID == xuid {
			if !hasType {
				return m, true
			}
			if m.Model.Type != t {
				continue
			}
			return m, true
		}
	}
	return identifiedModel{}, false
}

func (p *Pool) modelByUUID(id uuid.UUID, Type ...string) (identifiedModel, bool) {
	hasType := false
	var t string

	if len(Type) >= 1 {
		hasType = true
		t = Type[0]
	}

	for _, m := range p.modelsList() {
		if m.Model.Punished.UUID == id {
			if !hasType {
				return m, true
			}
			if m.Model.Type != t {
				continue
			}
			return m, true
		}
	}
	return identifiedModel{}, false
}

func (p *Pool) modelByIP(ip string, Type ...string) (identifiedModel, bool) {
	hasType := false
	var t string

	if len(Type) >= 1 {
		hasType = true
		t = Type[0]
	}

	for _, m := range p.modelsList() {
		if slices.Contains(m.Model.Punished.IPs, ip) {
			if !hasType {
				return m, true
			}
			if m.Model.Type != t {
				continue
			}
			return m, true
		}
	}
	return identifiedModel{}, false
}

func (p *Pool) modelByDID(did string, Type ...string) (identifiedModel, bool) {
	hasType := false
	var t string

	if len(Type) >= 1 {
		hasType = true
		t = Type[0]
	}

	for _, m := range p.modelsList() {
		if slices.Contains(m.Model.Punished.DeviceIDs, did) {
			if !hasType {
				return m, true
			}
			if m.Model.Type != t {
				continue
			}
			return m, true
		}
	}
	return identifiedModel{}, false
}

func (p *Pool) modelsList() (models []identifiedModel) {
	p.modelsMu.RLock()
	defer p.modelsMu.RUnlock()

	for id, m := range p.models {
		models = append(models, identifiedModel{Model: m, UUID: id})
	}
	return models
}

func (p *Pool) add(pu Punishment) error {
	if m, ok := p.modelUUID(ToModel(pu)); ok {
		if m, ok := p.model(m); ok {
			if m.Type == pu.Type() {
				return errors.New("already exists")
			}
		}
	}

	pun := pu.Punished()

	if _, exists := p.d.Exists(pun.Name, pun.XUID, pun.UUID, pun.IPs, pun.DeviceIDs, pu.Type()); !exists {
		if err := p.d.CreateEntry(pu); err != nil {
			return err
		}
	}

	p.models[uuid.New()] = ToModel(pu)
	return nil
}

func (p *Pool) delete(pu Punishment) error {
	m := ToModel(pu)
	id, ok := p.modelUUID(m)
	if !ok {
		return errors.New("model not found")
	}

	p.modelsMu.RLock()
	if _, ok := p.models[id]; !ok {
		p.modelsMu.RUnlock()
		return errors.New("model not found")
	}
	p.modelsMu.RUnlock()

	pun := pu.Punished()
	if m, exists := p.d.Exists(pun.Name, pun.XUID, pun.UUID, pun.IPs, pun.DeviceIDs, pu.Type()); exists {
		if m.Type == pu.Type() {
			if err := p.d.DeleteEntry(pu); err != nil {
				return err
			}
		}
	}

	p.modelsMu.Lock()
	defer p.modelsMu.Unlock()

	delete(p.models, id)
	return nil
}

func (p *Pool) update(new Punishment) error {
	pu := new.Punished()
	m, exists := p.modelExists(pu.Name, pu.XUID, pu.UUID, pu.IPs, pu.DeviceIDs, new.Type())
	if !exists {
		return errors.New("model not found")
	}
	if m.Model.Type != new.Type() {
		return errors.New("model not found")
	}
	if err := p.delete(FromModel(m.Model)); err != nil {
		return err
	}
	if err := p.add(new); err != nil {
		return err
	}
	pun := new.Punished()
	if _, exists := p.d.Exists(pun.Name, pun.XUID, pun.UUID, pun.IPs, pun.DeviceIDs, new.Type()); exists {
		if err := p.d.UpdateEntry(new); err != nil {
			return err
		}
	}
	return nil
}

func (p *Pool) loadFromDB() {
	p.once.Do(func() {
		p.modelsMu.Lock()
		defer p.modelsMu.Unlock()

		for m := range p.d.All() {
			p.models[uuid.New()] = m
		}
	})
}

func (p *Pool) l() *slog.Logger {
	return p.d.db.Logger()
}

func (p *Pool) Close() error {
	p.l().Info("Closing punishment pool...")
	p.close()
	return nil
}

func (p *Pool) close() {
	defer p.l().Info("Closed punishment pool.")
	for m := range p.d.All() {
		p.d.db.DB().Save(m.raw())
	}
}
