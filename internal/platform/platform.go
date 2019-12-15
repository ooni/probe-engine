// Package platform returns the platform name. The name returned here
// is compatible with the names returned by Measurement Kit.
package platform

import "runtime"

// Name returns the platform name. The returned value is one of:
//
// 1. "android"
// 2. "ios"
// 3. "linux"
// 5. "macos"
// 4. "windows"
// 5. "unknown"
//
// The android, ios, linux, macos, windows, and unknown strings are
// also returned by Measurement Kit. As a known bug, the detection of
// darwin-based systems relies on the architecture. It returns "ios"
// when using arm{,64} and "macos" when using x86{,_64}.
func Name() string {
	return name(runtime.GOOS, runtime.GOARCH)
}

func name(goos, goarch string) string {
	switch goos {
	case "android", "linux", "windows":
		return goos
	case "darwin":
		return detectDarwin(goarch)
	}
	return "unknown"
}

func detectDarwin(goarch string) string {
	// TODO(bassosimone): consider copying more precise detection from
	// Measurement Kit. Though, using architecture to detect ios vs macos
	// does not seem to be an issue, except for the simulator.
	switch goarch {
	case "386", "amd64":
		return "macos"
	case "arm", "arm64":
		return "ios"
	}
	return "unknown"
}
