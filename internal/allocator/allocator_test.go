package allocator_test

import (
	"sync"
	"testing"

	"github.com/aidenlish/distributed-ticketing/internal/allocator"
)

// TestDistributedUniqueIDs simulates multiple app servers sharing a single
// ticket server. Each app server has its own TicketAllocator; the shared
// refill function acts as the ticket server backed by an in-memory counter
// (equivalent to the MySQL end_range increment).
func TestDistributedUniqueIDs(t *testing.T) {
	var mu sync.Mutex
	var globalEnd int64
	const rangeSize = 20

	// Shared refill simulates the ticket server + DB.
	sharedRefill := func() (int64, int64, error) {
		mu.Lock()
		defer mu.Unlock()
		start := globalEnd
		globalEnd += rangeSize
		return start, globalEnd, nil
	}

	const (
		numServers     = 3   // simulated app servers
		goroutinesEach = 100 // concurrent goroutines per server
		idsEach        = 50  // IDs requested per goroutine
	)

	total := numServers * goroutinesEach * idsEach
	ids := make(chan int64, total)

	var wg sync.WaitGroup
	for s := 0; s < numServers; s++ {
		alloc := allocator.New(sharedRefill)
		for g := 0; g < goroutinesEach; g++ {
			wg.Add(1)
			go func() {
				defer wg.Done()
				for i := 0; i < idsEach; i++ {
					id, err := alloc.Next()
					if err != nil {
						t.Errorf("Next() error: %v", err)
						return
					}
					ids <- id
				}
			}()
		}
	}

	wg.Wait()
	close(ids)

	seen := make(map[int64]bool, total)
	for id := range ids {
		if seen[id] {
			t.Fatalf("duplicate ID: %d", id)
		}
		seen[id] = true
	}

	t.Logf("generated %d unique IDs across %d simulated app servers", len(seen), numServers)
}
