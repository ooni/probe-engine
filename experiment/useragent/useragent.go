// Package useragent contains code to set the User-Agent
package useragent

// Random returns a random User-Agent
func Random() string {
	// TODO(bassosimone): this user-agent solution is temporary and we
	// should instead select one among many user agents. We should open
	// an issue before merging this PR to address this defect.
	//
	// 11.8% as of August 24, 2019 according to https://techblog.willshouse.com/2012/01/03/most-common-user-agents/
	const ua = "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/74.0.3729.169 Safari/537.36"
	return ua
}
