// Package httpheader contains code to set common HTTP headers.
package httpheader

// RandomUserAgent returns a random User-Agent
func RandomUserAgent() string {
	// TODO(bassosimone): this user-agent solution is temporary and we
	// should instead select one among many user agents. See #147.
	//
	// 14.9% as of Jun 25, 2020 according to https://techblog.willshouse.com/2012/01/03/most-common-user-agents/
	const ua = "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/83.0.4103.61 Safari/537.36"
	return ua
}
