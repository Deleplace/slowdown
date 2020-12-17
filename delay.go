package slowdown

import (
	"net/http"
	"time"
)

// Delay adds an artificial 2s of latency at the start of the HandlerFunc.
//
// Usage: myHandlerFunc = slowdown.Delay(myHandlerFunc)
//
// The call to Delay may be chained with other middleware when building a handler.
func Delay(h http.HandlerFunc, cfg Config) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		time.Sleep(cfg.Duration)
		h(w, r)
	}
}

// Config exposes the parameters used for wrapping a HandlerFunc with Delay.
type Config struct {
	Duration time.Duration
}
