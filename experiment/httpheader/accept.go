package httpheader

// RandomAccept returns a random Accept header.
func RandomAccept() string {
	// This is the same that is returned by MK. We have #147 to
	// ensure that headers are randomized, if needed.
	return "text/html,application/xhtml+xml,application/xml;q=0.9,*/*;q=0.8"
}
