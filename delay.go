package slowdown

import (
	"net/http"
	"time"
)

// Delay adds an artificial 2s of latency at the start
//
// Usage: myHandlerFunc = slowdown.Delay(myHandlerFunc)
//
// The call to Delay may be chained with other middleware when building a handler.
func Delay(h http.HandlerFunc) http.HandlerFunc {
	const duration = 2 * time.Second
	return func(w http.ResponseWriter, r *http.Request) {
		time.Sleep(duration)
		h(w, r)
	}
}
