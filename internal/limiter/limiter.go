package limiter
// INT64 based atomic rate limiter
import (
	"sync/atomic"
	"time"
	"github.com/benbjohnson/clock"
)

type Limiter interface {
	Take() time.Time
}

type atomicInt64Limiter struct {
	prepadding [64]byte
	state int64
	postpadding [56]byte

	perRequest time.Duration
	maxSlack time.Duration
	clock clock.Clock
}

type config struct {
	clock clock.Clock 
	slack int
	per time.Duration
}

func New(rate int, opts ...Option) Limiter {
	return newAtomicInt64Based(rate, opts...)
}

func buildConfig(opts []Option) config{
	c := config{
		clock: clock.New(),
		slack: 10,
		per:   time.Second,
	}

	for _, opt := range opts {
		opt.apply(&c)
	}
	return c
}

func newAtomicInt64Based(rate int, opts ...Option) *atomicInt64Limiter {
	config:=buildConfig(opts)
	perRequest:=config.per / time.Duration(rate)
	l:=&atomicInt64Limiter{
		perRequest: perRequest,
		maxSlack: time.Duration(config.slack)*perRequest,
		clock: config.clock,
	}
	atomic.StoreInt64(&l.state, 0)
	return l
}

func (t *atomicInt64Limiter) Take() time.Time {
	var (
		newTimeOfNextPermissionIssue int64
		now                          int64
	)
	for {
		now = t.clock.Now().UnixNano()
		timeOfNextPermissionIssue := atomic.LoadInt64(&t.state)

		switch {
		case timeOfNextPermissionIssue == 0 || (t.maxSlack == 0 && now-timeOfNextPermissionIssue > int64(t.perRequest)):
			newTimeOfNextPermissionIssue = now
		case t.maxSlack > 0 && now-timeOfNextPermissionIssue > int64(t.maxSlack)+int64(t.perRequest):
			
			newTimeOfNextPermissionIssue = now - int64(t.maxSlack)
		default:
			
			newTimeOfNextPermissionIssue = timeOfNextPermissionIssue + int64(t.perRequest)
		}

		if atomic.CompareAndSwapInt64(&t.state, timeOfNextPermissionIssue, newTimeOfNextPermissionIssue) {
			break
		}
	}

	sleepDuration := time.Duration(newTimeOfNextPermissionIssue - now)
	if sleepDuration > 0 {
		t.clock.Sleep(sleepDuration)
		return time.Unix(0, newTimeOfNextPermissionIssue)
	}
	return time.Unix(0, now)
}


///////////////


type Option interface {
	apply(*config)
}

type clockOption struct {
	clock clock.Clock
}

func (o clockOption) apply(c *config) {
	c.clock = o.clock
}

// WithClock returns an option for ratelimit.New that provides an alternate
// Clock implementation, typically a mock Clock for testing.
func WithClock(clock clock.Clock) Option {
	return clockOption{clock: clock}
}

type slackOption int

func (o slackOption) apply(c *config) {
	c.slack = int(o)
}

// WithoutSlack configures the limiter to be strict and not to accumulate
// previously "unspent" requests for future bursts of traffic.
var WithoutSlack Option = slackOption(0)

// WithSlack configures custom slack.
// Slack allows the limiter to accumulate "unspent" requests
// for future bursts of traffic.
func WithSlack(slack int) Option {
	return slackOption(slack)
}

type perOption time.Duration

func (p perOption) apply(c *config) {
	c.per = time.Duration(p)
}

// Per allows configuring limits for different time windows.
//
// The default window is one second, so New(100) produces a one hundred per
// second (100 Hz) rate limiter.
//
// New(2, Per(60*time.Second)) creates a 2 per minute rate limiter.
func Per(per time.Duration) Option {
	return perOption(per)
}

type unlimited struct{}

// NewUnlimited returns a RateLimiter that is not limited.
func NewUnlimited() Limiter {
	return unlimited{}
}

func (unlimited) Take() time.Time {
	return time.Now()
}