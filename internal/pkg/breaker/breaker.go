package breaker

import (
	"SneakerFlash/internal/pkg/metrics"
	"sync"
	"time"
)

const (
	StateClosed   = "closed"
	StateOpen     = "open"
	StateHalfOpen = "half_open"
)

type entry struct {
	state            string
	failures         int
	openedAt         time.Time
	halfOpenAcquired bool
}

type CircuitBreaker struct {
	mu        sync.Mutex
	entries   map[string]*entry
	threshold int
	cooldown  time.Duration
}

func New(threshold int, cooldown time.Duration) *CircuitBreaker {
	if threshold <= 0 {
		threshold = 3
	}
	if cooldown <= 0 {
		cooldown = 5 * time.Second
	}
	return &CircuitBreaker{
		entries:   make(map[string]*entry),
		threshold: threshold,
		cooldown:  cooldown,
	}
}

func (b *CircuitBreaker) Allow(name string) bool {
	b.mu.Lock()
	defer b.mu.Unlock()

	e := b.get(name)
	switch e.state {
	case StateOpen:
		if time.Since(e.openedAt) < b.cooldown {
			metrics.IncBreakerReject(name, e.state)
			return false
		}
		if e.halfOpenAcquired {
			metrics.IncBreakerReject(name, StateHalfOpen)
			return false
		}
		e.state = StateHalfOpen
		e.halfOpenAcquired = true
		metrics.IncBreakerTransition(name, StateOpen, StateHalfOpen)
		return true
	case StateHalfOpen:
		if e.halfOpenAcquired {
			metrics.IncBreakerReject(name, e.state)
			return false
		}
		e.halfOpenAcquired = true
		return true
	default:
		return true
	}
}

func (b *CircuitBreaker) ReportSuccess(name string) {
	b.mu.Lock()
	defer b.mu.Unlock()

	e := b.get(name)
	prevState := e.state
	e.failures = 0
	e.halfOpenAcquired = false
	if e.state != StateClosed {
		e.state = StateClosed
		metrics.IncBreakerTransition(name, prevState, StateClosed)
	}
}

func (b *CircuitBreaker) ReportFailure(name string) {
	b.mu.Lock()
	defer b.mu.Unlock()

	e := b.get(name)
	prevState := e.state
	e.failures++
	switch e.state {
	case StateHalfOpen:
		e.state = StateOpen
		e.openedAt = time.Now()
		e.halfOpenAcquired = false
		metrics.IncBreakerTransition(name, prevState, StateOpen)
	case StateClosed:
		if e.failures >= b.threshold {
			e.state = StateOpen
			e.openedAt = time.Now()
			e.halfOpenAcquired = false
			metrics.IncBreakerTransition(name, prevState, StateOpen)
		}
	}
}

func (b *CircuitBreaker) get(name string) *entry {
	e, ok := b.entries[name]
	if ok {
		return e
	}
	e = &entry{state: StateClosed}
	b.entries[name] = e
	return e
}

var Default = New(3, 5*time.Second)
