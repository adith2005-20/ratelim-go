package limiter

import (
	"sync"
	"time"
)

type Limiter struct {
	rate	int 
	window	time.Duration
	mu		sync.Mutex
	allowance int
	lastCheck	time.Time
}

func New(rate int, window time.Duration) *Limiter {
	return &Limiter{
		rate:	rate,
		window:	window,
		allowance: rate,
		lastCheck: time.Now(),
	}
}

func (l *Limiter) Allow() bool {
	l.mu.Lock()
	defer l.mu.Unlock()


	now:= time.Now()
	elapsed:= now.Sub(l.lastCheck)
	l.lastCheck= now

	l.allowance= int(float64(elapsed)/float64(l.window) * float64(l.rate))

	if l.allowance > l.rate {
		l.allowance = l.rate
	}

	if l.allowance< 1 {
		return false
	}

	l.allowance--
	return true
}