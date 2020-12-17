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
		cfg.before(w, r)
		h(w, r)
		cfg.after(w, r)
	}
}

// config holds the parameters used for delaying a HandlerFunc.
// It is not exported. Users must use exported Options and Option providers instead.
type config struct {
	fixedDurationBefore time.Duration
	fixedDurationAfter  time.Duration
	headerPrefix        string
}

// Option configures the behavior of the delayed handler.
type Option func(*config)

func (cfg *config) before(w http.ResponseWriter, r *http.Request) {
	var d time.Duration
	if cfg.headerPrefix != "" {
		d, _ = readHeaderDuration(r, cfg.headerPrefix+"-before")
	} else {
		d = cfg.fixedDurationBefore
	}
	time.Sleep(d)
}

func (cfg *config) after(w http.ResponseWriter, r *http.Request) {
	var d time.Duration
	if cfg.headerPrefix != "" {
		d, _ = readHeaderDuration(r, cfg.headerPrefix+"-after")
	} else {
		d = cfg.fixedDurationAfter
	}
	time.Sleep(d)
}

// Fixed sets how long to pause before and after the wrapped HandlerFunc is executed.
func Fixed(before, after time.Duration) Option {
	return func(cfg *config) {
		cfg.fixedDurationBefore = before
		cfg.fixedDurationAfter = after
	}
}

// Header sets names of request headers that the client may set to control the
// delay durations.
//
// E.g. "delay" for inspecting "delay-before" and "delay-after"
// The expected format for the header values is a parsable duration, e.g.
// "delay-before: 300ms", "delay-after: 1.5s", etc.
//
// When Header is used and headers are not set by the client for a given request, then
// no delays are applied to this request.
//
// When Header is used, we consider that the client is in charge of deciding the delay
// durations, and any use of Fixed would be ignored. Header and Fixed should not be
// used at the same time.
func Header(prefix string) Option {
	return func(cfg *config) {
		cfg.headerPrefix = prefix
	}
}

func readHeaderDuration(r *http.Request, name string) (time.Duration, bool) {
	value := r.Header.Get(name)
	if value == "" {
		return 0, false
	}
	d, err := time.ParseDuration(value)
	if err != nil {
		return 0, false
	}
	return d, true
}
