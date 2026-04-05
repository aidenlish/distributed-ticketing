package integration_test

import (
	"database/sql"
	"os"
	"sync"
	"testing"

	"github.com/aidenlish/distributed-ticketing/internal/db"
	"github.com/aidenlish/distributed-ticketing/internal/ticketserver"
	_ "github.com/go-sql-driver/mysql"
)

func openTestDB(t *testing.T) *sql.DB {
	t.Helper()
	dsn := os.Getenv("DB_DSN")
	if dsn == "" {
		t.Skip("DB_DSN not set; skipping integration test")
	}
	sqlDB, err := sql.Open("mysql", dsn)
	if err != nil {
		t.Fatalf("sql.Open: %v", err)
	}
	if err := sqlDB.Ping(); err != nil {
		t.Fatalf("DB ping failed: %v", err)
	}
	t.Cleanup(func() { sqlDB.Close() })
	return sqlDB
}

// TestReserve_UpdatesDB verifies that a single Reserve() call increments
// end_range in MySQL by exactly rangeSize and returns the correct [start, end).
func TestReserve_UpdatesDB(t *testing.T) {
	sqlDB := openTestDB(t)
	reserver := ticketserver.NewRangeReserver(db.New(sqlDB))

	var before int64
	if err := sqlDB.QueryRow(`SELECT end_range FROM tickets WHERE id = 1`).Scan(&before); err != nil {
		t.Fatalf("reading end_range before: %v", err)
	}

	start, end, err := reserver.Reserve()
	if err != nil {
		t.Fatalf("Reserve(): %v", err)
	}

	const rangeSize = 20
	if start != before {
		t.Errorf("start = %d, want %d (old end_range)", start, before)
	}
	if end != before+rangeSize {
		t.Errorf("end = %d, want %d (old end_range + %d)", end, before+rangeSize, rangeSize)
	}

	var after int64
	if err := sqlDB.QueryRow(`SELECT end_range FROM tickets WHERE id = 1`).Scan(&after); err != nil {
		t.Fatalf("reading end_range after: %v", err)
	}
	if after != end {
		t.Errorf("DB end_range = %d after reserve, want %d", after, end)
	}

	t.Logf("end_range: %d → %d (returned [%d, %d))", before, after, start, end)
}

// TestReserve_ConcurrentRangesUnique fires multiple concurrent Reserve() calls
// and checks that the returned ranges do not overlap.
func TestReserve_ConcurrentRangesUnique(t *testing.T) {
	sqlDB := openTestDB(t)
	reserver := ticketserver.NewRangeReserver(db.New(sqlDB))

	const workers = 10
	type rangeResult struct{ start, end int64 }
	results := make([]rangeResult, workers)

	var wg sync.WaitGroup
	for i := 0; i < workers; i++ {
		wg.Add(1)
		go func(idx int) {
			defer wg.Done()
			start, end, err := reserver.Reserve()
			if err != nil {
				t.Errorf("worker %d Reserve(): %v", idx, err)
				return
			}
			results[idx] = rangeResult{start, end}
		}(i)
	}
	wg.Wait()

	// Check no two ranges overlap.
	for i := 0; i < workers; i++ {
		for j := i + 1; j < workers; j++ {
			a, b := results[i], results[j]
			if a.start < b.end && b.start < a.end {
				t.Errorf("ranges overlap: [%d,%d) and [%d,%d)", a.start, a.end, b.start, b.end)
			}
		}
	}

	t.Logf("all %d concurrent ranges are non-overlapping", workers)
}
