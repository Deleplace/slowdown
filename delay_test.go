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

// helloWorld is a trivial HandlerFunc. It takes very little time to execute.
var helloWorld = func(w http.ResponseWriter, r *http.Request) {
	fmt.Fprintln(w, "Hello world")
}

// h = Delay(h)
func TestDelayNoArguments(t *testing.T) {
	delayedHandler := Delay(helloWorld)
	expectedExtraLatency := 2 * time.Second

	messageBytes, responseTime := call(t, delayedHandler)
	// Test output
	if expected, actual := "Hello world\n", string(messageBytes); actual != expected {
		t.Errorf("Expected %q, got %q", expected, actual)
	}
	// Test duration, with some tolerance
	if responseTime < expectedExtraLatency-100*time.Millisecond {
		t.Errorf("Response time too short: %v", responseTime)
	}
	if responseTime > expectedExtraLatency+600*time.Millisecond {
		t.Errorf("Response time too long: %v", responseTime)
	}
}

// h = Delay(h, Duration(d))
func TestDelay400ms(t *testing.T) {
	const extraLatency = 400 * time.Millisecond
	delayedHandler := Delay(helloWorld, Duration(extraLatency))

	messageBytes, responseTime := call(t, delayedHandler)
	// Test output
	if expected, actual := "Hello world\n", string(messageBytes); actual != expected {
		t.Errorf("Expected %q, got %q", expected, actual)
	}
	// Test duration, with some tolerance
	if responseTime < extraLatency-100*time.Millisecond {
		t.Errorf("Response time too short: %v", responseTime)
	}
	if responseTime > extraLatency+600*time.Millisecond {
		t.Errorf("Response time too long: %v", responseTime)
	}
}

// Helper: call the handler while measuring response time.
func call(t *testing.T, h http.HandlerFunc) ([]byte, time.Duration) {
	s := httptest.NewServer(h)
	defer s.Close()

	var res *http.Response
	var err error
	// This is the response time from the client POV, so it is slightly
	// larger than the server-POV service time.
	responseTime := clock(func() {
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
	return messageBytes, responseTime
}

// Helper: executes f and return how long it took.
func clock(f func()) time.Duration {
	t := time.Now()
	f()
	return time.Since(t)
}
