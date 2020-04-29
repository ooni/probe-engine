package oonimkall

// IsRunning returns true if the background goroutine is still running. This
// function is an extension over Measurement Kit's API.
func (t *Task) IsRunning() bool {
	return t.isstopped.Load() == 0
}
