// Command miniooni is a simple binary for research and QA purposes
// with a CLI interface similar to MK and OONI Probe v2.x.
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
