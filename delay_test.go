package slowdown

import (
	"context"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"
)

// Warning: these tests are SLOW because they need to Sleep a lot.

// helloWorld is a trivial HandlerFunc. It takes very little time to execute.
var helloWorld http.HandlerFunc = func(w http.ResponseWriter, r *http.Request) {
	fmt.Fprintln(w, "Hello world")
}

// h = Delay(h)
func TestDelayNoArguments(t *testing.T) {
	delayedHandler := Delay(helloWorld)
	expectedExtraLatency := 1 * time.Second

	messageBytes, responseTime := call(t, delayedHandler, nil)

	testOuput(t, messageBytes, "Hello world\n")
	testDurationWithTolerance(t, responseTime, expectedExtraLatency)
}

// h = Delay(h, Fixed(d, 0))
func TestFixedDelayBefore(t *testing.T) {
	const extraLatency = 400 * time.Millisecond
	delayedHandler := Delay(helloWorld, Fixed(extraLatency, 0))

	messageBytes, responseTime := call(t, delayedHandler, nil)

	testOuput(t, messageBytes, "Hello world\n")
	testDurationWithTolerance(t, responseTime, extraLatency)
}

// h = Delay(h, Fixed(0, d))
func TestFixedDelayAfter(t *testing.T) {
	const extraLatency = 400 * time.Millisecond
	delayedHandler := Delay(helloWorld, Fixed(0, extraLatency))

	messageBytes, responseTime := call(t, delayedHandler, nil)

	testOuput(t, messageBytes, "Hello world\n")
	testDurationWithTolerance(t, responseTime, extraLatency)
}

// h = Delay(h, Fixed(d1, d2))
func TestFixedDelayBeforeAfter(t *testing.T) {
	const extraLatencyBefore = 300 * time.Millisecond
	const extraLatencyAfter = 700 * time.Millisecond
	delayedHandler := Delay(helloWorld, Fixed(extraLatencyBefore, extraLatencyAfter))

	messageBytes, responseTime := call(t, delayedHandler, nil)

	testOuput(t, messageBytes, "Hello world\n")
	testDurationWithTolerance(t, responseTime, extraLatencyBefore+extraLatencyAfter)
}

// h = Delay(h, Header(prefix))
func TestHeaderDelayBefore(t *testing.T) {
	const extraLatency = 400 * time.Millisecond
	const extraLatencyString = "400ms"
	const prefix = "delay"

	delayedHandler := Delay(helloWorld, Header(prefix))

	headers := http.Header{
		"delay-before": []string{extraLatencyString},
	}
	messageBytes, responseTime := call(t, delayedHandler, headers)

	testOuput(t, messageBytes, "Hello world\n")
	testDurationWithTolerance(t, responseTime, extraLatency)
}

// h = Delay(h, Header(prefix))
func TestHeaderDelayAfter(t *testing.T) {
	const extraLatency = 400 * time.Millisecond
	const extraLatencyString = "400ms"
	const prefix = "delay"

	delayedHandler := Delay(helloWorld, Header(prefix))

	headers := http.Header{
		"delay-after": []string{extraLatencyString},
	}
	messageBytes, responseTime := call(t, delayedHandler, headers)

	testOuput(t, messageBytes, "Hello world\n")
	testDurationWithTolerance(t, responseTime, extraLatency)
}

// h = Delay(h, Header(prefix))
func TestHeaderBeforeDelayAfter(t *testing.T) {
	const extraLatencyBefore = 300 * time.Millisecond
	const extraLatencyBeforeString = "300ms"
	const extraLatencyAfter = 700 * time.Millisecond
	const extraLatencyAfterString = "700ms"
	const prefix = "delay"

	delayedHandler := Delay(helloWorld, Header(prefix))

	headers := http.Header{
		"delay-before": []string{extraLatencyBeforeString},
		"delay-after":  []string{extraLatencyAfterString},
	}
	messageBytes, responseTime := call(t, delayedHandler, headers)

	testOuput(t, messageBytes, "Hello world\n")
	testDurationWithTolerance(t, responseTime, extraLatencyBefore+extraLatencyAfter)
}

// h = Delay(h, Header(prefix), Max(d))
func TestMaxBeforeAfterExceeded(t *testing.T) {
	const maxLatency = 100 * time.Millisecond
	const extraLatencyBeforeString = "300ms"
	const extraLatencyAfterString = "700ms"
	const prefix = "delay"

	delayedHandler := Delay(helloWorld, Header(prefix), Max(maxLatency))

	headers := http.Header{
		"delay-before": []string{extraLatencyBeforeString},
		"delay-after":  []string{extraLatencyAfterString},
	}
	messageBytes, responseTime := call(t, delayedHandler, headers)

	testOuput(t, messageBytes, "Hello world\n")
	// Capping at maxLatency before, and maxLatency after, thus 2*maxLatency
	testDurationWithTolerance(t, responseTime, 2*maxLatency)
}

// h = Delay(h, Header(prefix), Max(d))
func TestMaxBeforeAfterNotExceeded(t *testing.T) {
	const maxLatency = 600 * time.Millisecond
	const extraLatencyBefore = 100 * time.Millisecond
	const extraLatencyBeforeString = "100ms"
	const extraLatencyAfter = 150 * time.Millisecond
	const extraLatencyAfterString = "150ms"
	const prefix = "delay"

	delayedHandler := Delay(helloWorld, Header(prefix), Max(maxLatency))

	headers := http.Header{
		"delay-before": []string{extraLatencyBeforeString},
		"delay-after":  []string{extraLatencyAfterString},
	}
	messageBytes, responseTime := call(t, delayedHandler, headers)

	testOuput(t, messageBytes, "Hello world\n")
	// extraLatencyBefore < maxLatency
	// extraLatencyAfter < maxLatency
	testDurationWithTolerance(t, responseTime, extraLatencyBefore+extraLatencyAfter)
}

// h = Delay(h, Header(prefix), Condition(predicate))
func TestConditionMet(t *testing.T) {
	const extraLatencyBefore = 100 * time.Millisecond
	const extraLatencyBeforeString = "100ms"
	const prefix = "delay"
	const apiKey = "x96f3s6" // valid
	predicate := func(r *http.Request) bool {
		validAPIKeys := map[string]bool{
			"x96f3s6": true,
			"89qWsd2": true,
		}
		return validAPIKeys[r.Header.Get("apikey")]
	}

	delayedHandler := Delay(helloWorld, Header(prefix), Condition(predicate))

	headers := http.Header{
		"delay-before": []string{extraLatencyBeforeString},
		"apikey":       []string{apiKey},
	}
	messageBytes, responseTime := call(t, delayedHandler, headers)

	testOuput(t, messageBytes, "Hello world\n")
	testDurationWithTolerance(t, responseTime, extraLatencyBefore)
}

// h = Delay(h, Header(prefix), Condition(predicate))
func TestConditionUnmet(t *testing.T) {
	const extraLatencyBeforeString = "100ms"
	const prefix = "delay"
	const apiKey = "passw0rd" // invalid
	predicate := func(r *http.Request) bool {
		validAPIKeys := map[string]bool{
			"x96f3s6": true,
			"89qWsd2": true,
		}
		return validAPIKeys[r.Header.Get("apikey")]
	}

	delayedHandler := Delay(helloWorld, Header(prefix), Condition(predicate))

	headers := http.Header{
		"delay-before": []string{extraLatencyBeforeString},
		"apikey":       []string{apiKey},
	}
	messageBytes, responseTime := call(t, delayedHandler, headers)

	testOuput(t, messageBytes, "Hello world\n")
	testDurationWithTolerance(t, responseTime, 0)
}

// h = Delay(h, Header(prefix), Condition(predicate1), Condition(predicate2))
func TestMultipleConditionsMet(t *testing.T) {
	const extraLatencyBefore = 300 * time.Millisecond
	const extraLatencyBeforeString = "300ms"
	const prefix = "delay"
	const apiKey = "x96f3s6" // valid
	predicate1 := func(r *http.Request) bool {
		return r.Header.Get("user") == "admin"
	}
	predicate2 := func(r *http.Request) bool {
		validAPIKeys := map[string]bool{
			"x96f3s6": true,
			"89qWsd2": true,
		}
		return validAPIKeys[r.Header.Get("apikey")]
	}

	delayedHandler := Delay(helloWorld, Header(prefix), Condition(predicate1), Condition(predicate2))

	headers := http.Header{
		"delay-before": []string{extraLatencyBeforeString},
		"apikey":       []string{apiKey},
		"user":         []string{"admin"},
	}
	messageBytes, responseTime := call(t, delayedHandler, headers)

	testOuput(t, messageBytes, "Hello world\n")
	testDurationWithTolerance(t, responseTime, extraLatencyBefore)
}

// h = Delay(h, Header(prefix), Condition(predicate1), Condition(predicate2))
func TestMultipleConditionsUnmet(t *testing.T) {
	const extraLatencyBeforeString = "300ms"
	const prefix = "delay"
	const apiKey = "x96f3s6" // valid
	predicate1 := func(r *http.Request) bool {
		return r.Header.Get("user") == "admin"
	}
	predicate2 := func(r *http.Request) bool {
		validAPIKeys := map[string]bool{
			"x96f3s6": true,
			"89qWsd2": true,
		}
		return validAPIKeys[r.Header.Get("apikey")]
	}

	delayedHandler := Delay(helloWorld, Header(prefix), Condition(predicate1), Condition(predicate2))

	headers := http.Header{
		"delay-before": []string{extraLatencyBeforeString},
		"apikey":       []string{apiKey},
		"user":         []string{"ted"},
	}
	messageBytes, responseTime := call(t, delayedHandler, headers)

	testOuput(t, messageBytes, "Hello world\n")
	testDurationWithTolerance(t, responseTime, 0)
}

// h = Delay(h, Fixed(d, 0))
func TestContextCanceled(t *testing.T) {
	const extraLatency = 500 * time.Millisecond
	const cancelLatency = 200 * time.Millisecond
	var sideEffectHandler http.HandlerFunc = func(w http.ResponseWriter, h *http.Request) {
		t.Errorf("Should have been canceled before hitting the wrapped handler")
	}
	delayedHandler := Delay(sideEffectHandler, Fixed(extraLatency, 0))

	// messageBytes, responseTime := call(t, delayedHandler, nil)

	s := httptest.NewServer(delayedHandler)
	defer s.Close()
	client := &http.Client{}
	req, _ := http.NewRequest("GET", s.URL, nil)
	ctx2, cancel2 := context.WithTimeout(req.Context(), cancelLatency)
	defer cancel2()
	req = req.WithContext(ctx2)

	var err error
	responseTime := clock(func() {
		_, err = client.Do(req)
	})
	if err == nil {
		t.Fatal("Expected err: context deadline exceeded, got nil")
	}
	if !strings.Contains(err.Error(), "context deadline exceeded") {
		t.Fatalf("Expected err: context deadline exceeded, got %T error %q", err, err.Error())
	}
	testDurationWithTolerance(t, responseTime, cancelLatency)
	// Make sure the side effect would have had the time to be triggered
	time.Sleep(extraLatency - cancelLatency + 50*time.Millisecond)
}

func TestHandler(t *testing.T) {
	tmpdir, err := ioutil.TempDir("", "")
	if err != nil {
		t.Fatal(err)
	}
	h := http.FileServer(http.Dir(tmpdir))

	// We're passing an http.Handler h (not an http.HandlerFunc).
	delayedHandler := Delay(h)
	expectedExtraLatency := 1 * time.Second

	_, responseTime := call(t, delayedHandler, nil)

	testDurationWithTolerance(t, responseTime, expectedExtraLatency)
}

// Helper: call the handler with specific request headers, while measuring response time.
func call(t *testing.T, h http.HandlerFunc, headers http.Header) ([]byte, time.Duration) {
	s := httptest.NewServer(h)
	defer s.Close()

	client := &http.Client{}
	req, _ := http.NewRequest("GET", s.URL, nil)
	req.Header = headers

	var res *http.Response
	var err error
	// This is the response time from the client POV, so it is slightly
	// larger than the server-POV service time.
	responseTime := clock(func() {
		res, err = client.Do(req)
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

// Helper: check that elapsed time seems cromulent.
func testDurationWithTolerance(t *testing.T, observed, expected time.Duration) {
	if observed < expected-50*time.Millisecond {
		t.Errorf("Response time too short: %v", observed)
	}
	if observed > expected+100*time.Millisecond {
		t.Errorf("Response time too long: %v", observed)
	}
}
