// Package slowdown provides a debugging middleware for adding delays
// to an http.HandlerFunc.
//
// Sample usage:
//     h = slowdown.Delay(h, slowdown.Header("delay"), slowdown.Max(5*time.Second))
// which means "Accept request headers 'delay-before' and 'delay-after'
// and pause the request processing accordingly, but never more than 5s."
//
// This is useful to test client-side race conditions that depend on the specific
// processing order of concurrent requests.
// (This is not related to Go's memory model data races and Go's race detector)
package slowdown
