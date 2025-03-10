package punishment

import (
	"errors"
	"github.com/google/uuid"
	"github.com/k4ties/dystopia/internal/database"
	"iter"
	"log/slog"
	"slices"
)

type Database struct {
	db *database.Database[rawModel]
}

func NewDatabase(path string, l *slog.Logger) (*Database, error) {
	db, err := database.New[rawModel](path, l)
	if err != nil {
		return nil, err
	}

	return &Database{db: db}, nil
}

func (db *Database) Exists(name, xuid string, id uuid.UUID, ips, dids []string, Type ...string) (model Model, exists bool) {
	hasType := false
	var t string

	if len(Type) >= 1 {
		hasType = true
		t = Type[0]
	}

	if m, ok := db.ByMainInfo(name, xuid, id, Type...); ok {
		return m, true
	}
	for _, ip := range ips {
		for _, did := range dids {
			if m, ok := db.ByIpAndDID(ip, did, Type...); ok {
				if !hasType {
					return m, true
				}
				if m.Type != t {
					continue
				}
				return m, true
			}
		}
	}
	return Model{}, false
}

func (db *Database) ByMainInfo(name, xuid string, id uuid.UUID, Type ...string) (Model, bool) {
	m, ok := db.ByName(name, Type...)
	if ok {
		return m, ok
	}
	m, ok = db.ByXUID(xuid, Type...)
	if ok {
		return m, ok
	}
	m, ok = db.ByUUID(id, Type...)
	if ok {
		return m, ok
	}

	return Model{}, false
}

func (db *Database) ByName(name string, Type ...string) (Model, bool) {
	hasType := false
	var t string

	if len(Type) >= 1 {
		hasType = true
		t = Type[0]
	}

	for m := range db.All() {
		if m.Punished.Name == name {
			if !hasType {
				return m, true
			}
			if m.Type != t {
				continue
			}
			return m, true
		}
	}
	return Model{}, false
}

func (db *Database) ByXUID(xuid string, Type ...string) (Model, bool) {
	hasType := false
	var t string

	if len(Type) >= 1 {
		hasType = true
		t = Type[0]
	}

	for m := range db.All() {
		if m.Punished.XUID == xuid {
			if !hasType {
				return m, true
			}
			if m.Type != t {
				continue
			}
			return m, true
		}
	}
	return Model{}, false
}

func (db *Database) ByUUID(id uuid.UUID, Type ...string) (Model, bool) {
	hasType := false
	var t string

	if len(Type) >= 1 {
		hasType = true
		t = Type[0]
	}

	for m := range db.All() {
		if m.Punished.UUID == id {
			if !hasType {
				return m, true
			}
			if m.Type != t {
				continue
			}
			return m, true
		}
	}

	return Model{}, false
}

func (db *Database) ByIP(ip string, Type ...string) (Model, bool) {
	hasType := false
	var t string

	if len(Type) >= 1 {
		hasType = true
		t = Type[0]
	}

	for m := range db.All() {
		if slices.Contains(m.Punished.IPs, ip) {
			if !hasType {
				return m, true
			}
			if m.Type != t {
				continue
			}
			return m, true
		}
	}

	return Model{}, false
}

func (db *Database) ByDID(did string, Type ...string) (Model, bool) {
	hasType := false
	var t string

	if len(Type) >= 1 {
		hasType = true
		t = Type[0]
	}

	for m := range db.All() {
		if slices.Contains(m.Punished.DeviceIDs, did) {
			if !hasType {
				return m, true
			}
			if m.Type != t {
				continue
			}
			return m, true
		}
	}

	return Model{}, false
}

func (db *Database) ByIpAndDID(ip, did string, Type ...string) (Model, bool) {
	m, ok := db.ByIP(ip, Type...)
	if ok {
		return m, true
	}

	m, ok = db.ByDID(did, Type...)
	if ok {
		return m, true
	}

	return Model{}, false
}

func (db *Database) All() iter.Seq[Model] {
	return func(yield func(Model) bool) {
		for _, r := range db.db.Entries() {
			if !yield(r.model()) {
				return
			}
		}
	}
}

func (db *Database) CreateEntry(p Punishment) error {
	pu := p.Punished()
	if _, exists := db.Exists(pu.Name, pu.XUID, pu.UUID, pu.IPs, pu.DeviceIDs, p.Type()); exists {
		return errors.New("already exists")
	}

	db.db.NewEntry(ToModel(p).raw())
	return nil
}

func (db *Database) UpdateEntry(new Punishment) error {
	pu := new.Punished()
	m, exists := db.Exists(pu.Name, pu.XUID, pu.UUID, pu.IPs, pu.DeviceIDs, new.Type())
	if !exists {
		return errors.New("not exists")
	}

	model := m.raw()
	if err := db.db.DB().Where(&model).Updates(ToModel(new).raw()).Error; err != nil {
		return err
	}

	return nil
}

func (db *Database) DeleteEntry(p Punishment) error {
	pu := p.Punished()
	m, exists := db.Exists(pu.Name, pu.XUID, pu.UUID, pu.IPs, pu.DeviceIDs, p.Type())
	if !exists {
		return errors.New("not exists")
	}

	model := m.raw()
	if err := db.db.DB().Where(&model).Delete(model).Error; err != nil {
		return err
	}

	return nil
}
