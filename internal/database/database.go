package database

import (
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
	"log/slog"
)

type Database[T comparable] struct {
	l  *slog.Logger
	db *gorm.DB
}

func New[T comparable](path string, l *slog.Logger) (*Database[T], error) {
	db, err := gorm.Open(sqlite.Open(path), &gorm.Config{Logger: logger.Discard})
	if err != nil {
		return nil, err
	}

	var zero T
	if err := db.AutoMigrate(&zero); err != nil {
		return nil, err
	}

	return &Database[T]{l: l, db: db}, nil
}

func (d *Database[T]) DB() *gorm.DB {
	return d.db
}

func (d *Database[T]) Entries() []T {
	var entries []T

	d.db.Find(&entries)
	return entries
}

func (d *Database[T]) DeleteEntry(findEntryArgs ...any) {
	en, ok := d.FindEntry(findEntryArgs)
	if !ok {
		return
	}

	d.db.Delete(en)
}

func (d *Database[T]) NewEntry(entry T) {
	d.db.Create(entry)
}

func (d *Database[T]) FindEntry(args ...any) (T, bool) {
	var entry T
	zeroEntry := entry

	d.db.Find(&entry, args...).First(&entry)
	return entry, zeroEntry != entry
}

func (d *Database[T]) Logger() *slog.Logger {
	return d.l
}
