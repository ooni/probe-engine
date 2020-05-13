package httpheader

// RandomAcceptLanguage returns a random Accept-Language header
func RandomAcceptLanguage() string {
	// This is the same that is returned by MK. We have #147 to
	// ensure that headers are randomized, if needed.
	return "en-US;q=0.8,en;q=0.5"
}
