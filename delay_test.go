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

// Warning: these tests are SLOW because they need to Sleep a lot.

// helloWorld is a trivial HandlerFunc. It takes very little time to execute.
var helloWorld = func(w http.ResponseWriter, r *http.Request) {
	fmt.Fprintln(w, "Hello world")
}

// h = Delay(h)
func TestDelayNoArguments(t *testing.T) {
	delayedHandler := Delay(helloWorld)
	expectedExtraLatency := 1 * time.Second

	messageBytes, responseTime := call(t, delayedHandler)

	testOuput(t, messageBytes, "Hello world\n")
	testDurationWithTolerance(t, responseTime, expectedExtraLatency)
}

// h = Delay(h, Fixed(d, 0))
func TestDelayBefore(t *testing.T) {
	const extraLatency = 400 * time.Millisecond
	delayedHandler := Delay(helloWorld, Fixed(extraLatency, 0))

	messageBytes, responseTime := call(t, delayedHandler)

	testOuput(t, messageBytes, "Hello world\n")
	testDurationWithTolerance(t, responseTime, extraLatency)
}

// h = Delay(h, Fixed(0, d))
func TestDelayAfter(t *testing.T) {
	const extraLatency = 400 * time.Millisecond
	delayedHandler := Delay(helloWorld, Fixed(0, extraLatency))

	messageBytes, responseTime := call(t, delayedHandler)

	testOuput(t, messageBytes, "Hello world\n")
	testDurationWithTolerance(t, responseTime, extraLatency)
}

// h = Delay(h, Fixed(d1, d2))
func TestDelayBeforeAfter(t *testing.T) {
	const extraLatencyBefore = 300 * time.Millisecond
	const extraLatencyAfter = 700 * time.Millisecond
	delayedHandler := Delay(helloWorld, Fixed(extraLatencyBefore, extraLatencyAfter))

	messageBytes, responseTime := call(t, delayedHandler)

	testOuput(t, messageBytes, "Hello world\n")
	testDurationWithTolerance(t, responseTime, extraLatencyBefore+extraLatencyAfter)
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

// Helper: inspect the HTTP response body.
func testOuput(t *testing.T, messageBytes []byte, expected string) {
	if actual := string(messageBytes); actual != expected {
		t.Errorf("Expected %q, got %q", expected, actual)
	}
}

// Helper: check that elapsed time seem cromulent.
func testDurationWithTolerance(t *testing.T, observed, expected time.Duration) {
	if observed < expected-100*time.Millisecond {
		t.Errorf("Response time too short: %v", observed)
	}
	if observed > expected+600*time.Millisecond {
		t.Errorf("Response time too long: %v", observed)
	}
}
