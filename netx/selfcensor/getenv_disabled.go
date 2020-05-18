// +build !selfcensor

package selfcensor

func getenv(variable string) string {
	// by returning empty string we basically disable jafar
	return ""
}
