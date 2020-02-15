package measurementkit

import "errors"

func start(settings []byte) (<-chan []byte, error) {
	return nil, errors.New("Measurement Kit not available")
}

func available() bool {
	return false
}
