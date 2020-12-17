package slowdown

import (
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

func TestDelay(t *testing.T) {
	handler := func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprintln(w, "Hello world")
	}
	delayedHandler := Delay(handler)
	// http.HandlerFunc

	s := httptest.NewServer(delayedHandler)
	defer s.Close()

	var res *http.Response
	var err error
	totalResponseTime := clock(func() {
		res, err = http.Get(s.URL)
	})
	if err != nil {
		t.Fatal(err)
	}
	messageBytes, err := ioutil.ReadAll(res.Body)
	res.Body.Close()
	if err != nil {
		log.Fatal(err)
	}
	if expected, actual := "Hello world\n", string(messageBytes); actual != expected {
		t.Errorf("Expected %q, got %q", expected, actual)
	}
	// Test duration, with some tolerance
	if totalResponseTime < 1900*time.Millisecond {
		t.Errorf("Response time too short: %v", totalResponseTime)
	}
	if totalResponseTime > 2600*time.Millisecond {
		t.Errorf("Response time too long: %v", totalResponseTime)
	}
}

// Helper: executes f and return how long it took.
func clock(f func()) time.Duration {
	t := time.Now()
	f()
	return time.Since(t)
}
