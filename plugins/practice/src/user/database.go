package user

import (
	"errors"
	"github.com/google/uuid"
	"github.com/k4ties/dystopia/internal/database"
	"iter"
)

type DB struct {
	db *database.Database[rawOfflineUser]
}

func NewDB(path string) (*DB, error) {
	d, err := database.New[rawOfflineUser](path, nil)
	if err != nil {
		return nil, err
	}

	return &DB{db: d}, nil
}

func (db *DB) NewAccount(o OfflineUser) {
	db.db.NewEntry(o.raw())
}

func (db *DB) AccountByNameAndUUID(name string, id uuid.UUID) (OfflineUser, bool) {
	e, ok := db.db.FindEntry("name = ? AND uuid = ?", name, id.String())
	return e.offlineUser(), ok
}

func (db *DB) Exists(name string, id uuid.UUID) bool {
	_, ok := db.db.FindEntry("name = ? AND uuid = ?", name, id.String())
	return ok
}

func (db *DB) AccountByName(name string) (OfflineUser, bool) {
	raw, ok := db.db.FindEntry("name = ?", name)
	if !ok {
		return OfflineUser{}, false
	}

	return raw.offlineUser(), true
}

func (db *DB) AccountByUUID(id uuid.UUID) (OfflineUser, bool) {
	raw, ok := db.db.FindEntry("uuid = ?", id.String())
	if !ok {
		return OfflineUser{}, false
	}

	return raw.offlineUser(), true
}

func (db *DB) AccountByXUID(id string) (OfflineUser, bool) {
	raw, ok := db.db.FindEntry("xuid = ?", id)
	if !ok {
		return OfflineUser{}, false
	}

	return raw.offlineUser(), true
}

func (db *DB) SearchArgs(o OfflineUser) []any {
	return []any{
		"name = ? AND uuid = ?",
		o.Name,
		o.UUID,
	}
}

func (db *DB) Update(new OfflineUser) error {
	if _, ok := db.AccountByNameAndUUID(new.Name, new.UUID); !ok {
		return errors.New("no entry with specified search arguments")
	}

	return db.db.DB().Where("name = ? AND uuid = ? AND xuid = ?", new.Name, new.UUID.String(), new.XUID).Updates(new.raw()).Error
}

func (db *DB) Accounts() iter.Seq[OfflineUser] {
	return func(yield func(OfflineUser) bool) {
		for _, raw := range db.db.Entries() {
			if !yield(raw.offlineUser()) {
				return
			}
		}
	}
}
