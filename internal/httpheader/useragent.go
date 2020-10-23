// Package httpheader contains code to set common HTTP headers.
package httpheader

// UserAgent returns the User-Agent header used for measuring.
func UserAgent() string {
	// 13.1% as of Oct 23, 2020 according to https://techblog.willshouse.com/2012/01/03/most-common-user-agents/
	const ua = "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/85.0.4183.121 Safari/537.36"
	return ua
}
