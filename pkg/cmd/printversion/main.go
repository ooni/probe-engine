// Command printversion prints the current version of this repository.
package main

import (
	"fmt"

	"github.com/ooni/probe-engine/pkg/version"
)

func main() {
	fmt.Println(version.Version)
}
