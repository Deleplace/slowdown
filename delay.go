package slowdown

import (
	"net/http"
	"time"
)

// Delay adds an artificial latency at the start of the HandlerFunc.
//
// Usage: myHandlerFunc = slowdown.Delay(myHandlerFunc)
//
// The call to Delay may be chained with other middleware when building a handler.
func Delay(h http.HandlerFunc, opts ...Option) http.HandlerFunc {
	// Default config values
	cfg := config{
		fixedDurationBefore: 1 * time.Second,
	}
	// Apply options
	for _, opt := range opts {
		opt(&cfg)
	}

	return func(w http.ResponseWriter, r *http.Request) {
		cfg.before()
		h(w, r)
		cfg.after()
	}
}

// config holds the parameters used for delaying a HandlerFunc.
// It is not exported. Users must use exported Options and Option providers instead.
type config struct {
	fixedDurationBefore time.Duration
	fixedDurationAfter  time.Duration
}

// Option configures the behavior of the delayed handler.
type Option func(*config)

func (cfg *config) before() {
	time.Sleep(cfg.fixedDurationBefore)
}

func (cfg *config) after() {
	time.Sleep(cfg.fixedDurationAfter)
}

// Fixed sets how long to pause before and after the wrapped HandlerFunc is executed.
func Fixed(before, after time.Duration) Option {
	return func(cfg *config) {
		cfg.fixedDurationBefore = before
		cfg.fixedDurationAfter = after
	}
}
