package main

import (
	"net/http"
	"testing"
)

func TestSmoke(t *testing.T) {
	// just check whether we can start and then tear down the server
	*endpoint = ":54321"
	go main()
	resp, err := http.Get("http://127.0.0.1:54321")
	if err != nil {
		t.Fatal(err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != 400 {
		t.Fatal("unexpected status code")
	}
	srvcancel()  // kills the listener
	srvwg.Wait() // joined
}
