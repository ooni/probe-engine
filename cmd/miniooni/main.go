// Command miniooni is simple binary for testing purposes.
package main

import (
	"log"

	"github.com/ooni/probe-engine/libminiooni"
)

func main() {
	defer func() {
		if s := recover(); s != nil {
			log.Fatal(s)
		}
	}()
	libminiooni.Main()
}
