package ticketserver

import "github.com/aidenlish/distributed-ticketing/internal/db"

const rangeSize = 20

// RangeReserver reserves the next ID range from the DB and returns [start, end).
type RangeReserver struct {
	store *db.Store
}

func NewRangeReserver(store *db.Store) *RangeReserver {
	return &RangeReserver{store: store}
}

func (r *RangeReserver) Reserve() (start, end int64, err error) {
	end, err = r.store.IncrementEndRange(rangeSize)
	if err != nil {
		return 0, 0, err
	}
	return end - rangeSize, end, nil
}
