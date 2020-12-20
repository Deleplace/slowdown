package slowdown

import (
	"context"
	"net/http"
	"time"
)

// Delay adds artificial latencies before and after an http.HandlerFunc. By default it
// adds 1s of latency before the execution of the wrapped handler.
//
// Sample usage:
//     myHandlerFunc = slowdown.Delay(myHandlerFunc)
//
// The call to Delay may be chained with other middleware when building a handler
// func. By default the added latency may not exceed 40s per request (20s before
// and 20s after). If you need to add more latency, set a higher cap with Max.
func Delay(h http.HandlerFunc, opts ...Option) http.HandlerFunc {
	// Default config values
	cfg := config{
		fixedDurationBefore: 1 * time.Second,
		max:                 20 * time.Second,
	}
	// Apply options
	for _, opt := range opts {
		opt(&cfg)
	}

	return func(w http.ResponseWriter, r *http.Request) {
		cfg.before(w, r)

		if isDone(r.Context()) {
			return
		}

		h(w, r)

		if isDone(r.Context()) {
			return
		}

		cfg.after(w, r)
	}
}

// config holds the parameters used for delaying a HandlerFunc.
// It is not exported. Users must use exported Options and Option providers instead.
type config struct {
	fixedDurationBefore time.Duration
	fixedDurationAfter  time.Duration
	headerPrefix        string
	max                 time.Duration
	conditions          []func(*http.Request) bool
}

// Option configures the behavior of the delayed handler.
type Option func(*config)

func (cfg *config) before(w http.ResponseWriter, r *http.Request) {
	cfg.sleep(w, r, "before")
}

func (cfg *config) after(w http.ResponseWriter, r *http.Request) {
	cfg.sleep(w, r, "after")
}

func (cfg *config) sleep(w http.ResponseWriter, r *http.Request, beforeOrAfter string) {
	if !cfg.checkConditions(r) {
		// When at least one condition is not met, there is no delay added.
		return
	}

	var d time.Duration
	if cfg.headerPrefix != "" {
		d, _ = readHeaderDuration(r, cfg.headerPrefix+"-"+beforeOrAfter)
	} else {
		switch beforeOrAfter {
		case "before":
			d = cfg.fixedDurationBefore
		case "after":
			d = cfg.fixedDurationAfter
		}
	}
	if d > cfg.max {
		d = cfg.max
	}
	// Like time.Sleep(d), but Context-aware
	select {
	case <-r.Context().Done(): //context cancelled
	case <-time.After(d):
	}
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
// E.g. "delay" for inspecting "delay-before" and "delay-after".
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

// Max sets an upper limit on each delay.
// The total delay (before + after) will not exceed 2*maxDuration.
// The default is 20s.
func Max(maxDuration time.Duration) Option {
	return func(cfg *config) {
		cfg.max = maxDuration
	}
}

// Condition adds a predicate to determine if a given request should be delayed or not.
//
// Condition only has the power to prevent the delays to be applied, when the condition
// is not met (i.e. when the predicate returns false).
//
// Multiple conditions may be combined.
func Condition(predicate func(*http.Request) bool) Option {
	return func(cfg *config) {
		cfg.conditions = append(cfg.conditions, predicate)
	}
}

func (cfg *config) checkConditions(r *http.Request) bool {
	for _, predicate := range cfg.conditions {
		if !predicate(r) {
			return false
		}
	}
	return true
}

// Helper to determine if a Context is already done (cancelled), in an imperative style.
func isDone(ctx context.Context) bool {
	select {
	case <-ctx.Done():
		return true
	default:
		return false
	}
}
