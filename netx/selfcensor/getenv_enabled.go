// +build selfcensor

package selfcensor

import "os"

func getenv(variable string) string {
	return os.Getenv(variable)
}
