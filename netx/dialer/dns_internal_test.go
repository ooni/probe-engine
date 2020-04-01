package dialer

// ReduceErrors exposes the internal function reduceErrors
func ReduceErrors(errorslist []error) error {
	return reduceErrors(errorslist)
}
