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
		duration: 2 * time.Second,
	}
	// Apply options
	for _, opt := range opts {
		opt(&cfg)
	}

	return func(w http.ResponseWriter, r *http.Request) {
		time.Sleep(cfg.duration)
		h(w, r)
	}
}

// config holds the parameters used for delaying a HandlerFunc.
// It is not exported. Users must use exported Options and Option providers instead.
type config struct {
	duration time.Duration
}

// Option configures the behavior of the delayed handler.
type Option func(*config)

// Duration sets how long to delay the wrapped HandlerFunc.
func Duration(d time.Duration) Option {
	return func(cfg *config) {
		cfg.duration = d
	}
}
