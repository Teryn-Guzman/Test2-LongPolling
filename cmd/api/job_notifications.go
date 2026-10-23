package main

import "sync"

// jobNotifier broadcasts coalescible wake-ups. Persisted PostgreSQL rows remain
// authoritative; observers always reload the representation after a wake-up.
type jobNotifier struct {
	mu          sync.Mutex
	subscribers map[string]map[chan struct{}]struct{}
}

func newJobNotifier() *jobNotifier {
	return &jobNotifier{subscribers: make(map[string]map[chan struct{}]struct{})}
}

func (n *jobNotifier) subscribe(jobID string) (<-chan struct{}, func()) {
	ch := make(chan struct{}, 1)
	n.mu.Lock()
	if n.subscribers[jobID] == nil {
		n.subscribers[jobID] = make(map[chan struct{}]struct{})
	}
	n.subscribers[jobID][ch] = struct{}{}
	n.mu.Unlock()

	return ch, func() {
		n.mu.Lock()
		delete(n.subscribers[jobID], ch)
		if len(n.subscribers[jobID]) == 0 {
			delete(n.subscribers, jobID)
		}
		n.mu.Unlock()
	}
}

func (n *jobNotifier) publish(jobID string) {
	n.mu.Lock()
	defer n.mu.Unlock()
	for ch := range n.subscribers[jobID] {
		select {
		case ch <- struct{}{}:
		default: // Coalesce repeated updates while an observer is rechecking SQL.
		}
	}
}
