package gorm

import "github.com/jinzhu/gorm"

func New(db *gorm.DB) *Store { return &Store{db: db} }

type Store struct {
	db *gorm.DB
}

func (s *Store) Where(query string, args ...interface{}) {
	s.db = s.db.Where(query, args...)
}
func (s *Store) Order(order string) {
	s.db = s.db.Order(order)
}
func (s *Store) Limit(limit int64) {
	s.db = s.db.Limit(limit)
}
func (s *Store) Offset(offset int64) {
	s.db = s.db.Offset(offset)
}
func (s *Store) Find(dest interface{}) error {
	s.db = s.db.Find(dest)
	return s.db.Error
}
func (s *Store) Count(count *int64) error {
	return s.db.Count(count).Error
}
