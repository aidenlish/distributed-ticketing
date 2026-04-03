package allocator

import (
	"sync"
	"sync/atomic"
)

type TicketAllocator struct {
	next     int64
	end      int64
	refillMu sync.Mutex
	refill   func() (start, end int64, err error)
}

func New(refill func() (start, end int64, err error)) *TicketAllocator {
	return &TicketAllocator{refill: refill}
}

func (a *TicketAllocator) Next() (int64, error) {
	for {
		id := atomic.AddInt64(&a.next, 1) - 1
		if id < atomic.LoadInt64(&a.end) {
			return id, nil
		}
		if err := a.doRefill(); err != nil {
			return 0, err
		}
	}
}

func (a *TicketAllocator) doRefill() error {
	a.refillMu.Lock()
	defer a.refillMu.Unlock()

	if atomic.LoadInt64(&a.next) < atomic.LoadInt64(&a.end) {
		return nil
	}

	start, end, err := a.refill()
	if err != nil {
		return err
	}
	atomic.StoreInt64(&a.next, start)
	atomic.StoreInt64(&a.end, end)
	return nil
}
