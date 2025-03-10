package user

import (
	"github.com/df-mc/dragonfly/server/player"
	"github.com/google/uuid"
	"github.com/k4ties/dystopia/plugins/practice/src/instance"
	"github.com/sasha-s/go-deadlock"
	"iter"
	"strings"
)

var pool *Pool

func P() *Pool {
	return pool
}

func MustUsePool(f func(p *Pool)) {
	if P() != nil {
		f(P())
	}
}

type Pool struct {
	db *DB

	users   map[uuid.UUID]*User
	usersMu deadlock.RWMutex
}

func (p *Pool) DB() *DB {
	return p.db
}

func (p *Pool) MustUser(pl *player.Player, f func(*User)) {
	u, ok := p.User(pl.UUID())
	if !ok {
		instance.Kick(pl, instance.ErrorAngus)
		return
	}

	f(u)
}

func (p *Pool) OnlineUsers() []*User {
	var players []*User

	for u := range p.Users() {
		if _, ok := u.Player(); ok && u.Online() {
			players = append(players, u)
		}
	}

	return players
}

func (p *Pool) UserByName(name string) (*User, bool) {
	for u := range p.Users() {
		if strings.ToLower(u.Data().Name()) == strings.ToLower(name) {
			return u, true
		}
	}

	return nil, false
}

func (p *Pool) User(id uuid.UUID) (*User, bool) {
	p.usersMu.RLock()
	defer p.usersMu.RUnlock()

	u, ok := p.users[id]
	return u, ok
}

func (p *Pool) NewUser(u *User) {
	p.usersMu.Lock()
	defer p.usersMu.Unlock()

	p.users[u.Data().UUID()] = u
}

func (p *Pool) Users() iter.Seq[*User] {
	p.usersMu.RLock()
	defer p.usersMu.RUnlock()

	return func(yield func(*User) bool) {
		for _, u := range p.users {
			if !yield(u) {
				return
			}
		}
	}
}

func (p *Pool) SaveAll() {
	for usr := range p.Users() {
		p.db.db.DB().Save(Offline(usr).raw())
	}
}

func NewPool(d *DB) *Pool {
	pool = &Pool{db: d, users: make(map[uuid.UUID]*User)}
	pool.loadOfflineUsers()
	return pool
}

func (p *Pool) loadOfflineUsers() {
	for acc := range p.db.Accounts() {
		p.NewUser(FromOffline(acc))
	}
}
