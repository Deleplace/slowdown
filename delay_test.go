package slowdown

import (
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestDelay(t *testing.T) {
	handler := func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprintln(w, "Hello world")
	}
	delayedHandler := Delay(handler)
	// http.HandlerFunc

	s := httptest.NewServer(delayedHandler)
	defer s.Close()

	res, err := http.Get(s.URL)
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
}
