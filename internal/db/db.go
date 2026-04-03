package db

import "database/sql"

// Wrapper of the MySQL leader DB.
type Store struct {
	db *sql.DB
}

func New(db *sql.DB) *Store {
	return &Store{db: db}
}

func (s *Store) IncrementEndRange(delta int64) (int64, error) {
	res, err := s.db.Exec(
		`UPDATE tickets SET end_range = end_range + ? WHERE id = 1`, delta,
	)
	if err != nil {
		return 0, err
	}
	_ = res

	var end int64
	err = s.db.QueryRow(`SELECT end_range FROM tickets WHERE id = 1`).Scan(&end)
	return end, err
}
