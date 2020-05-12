// Command miniooni is simple binary for testing purposes.
package main

import (
	"os"

	"github.com/ooni/probe-engine/libminiooni"
)

func main() {
	defer func() {
		if recover() != nil {
			os.Exit(1)
		}
	}()
	libminiooni.Main()
}
