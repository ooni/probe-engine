// Package urlpath contains code to manipulate URL paths
package urlpath

import "strings"

// Append appends extPath to basePath. It will properly deal
// with `/` being or being not present respectively at the end
// of basePath and at the beginning of extPath.
func Append(basePath, extPath string) string {
	if strings.HasSuffix(basePath, "/") {
		basePath = basePath[:len(basePath)-1]
	}
	if !strings.HasPrefix(extPath, "/") {
		extPath = "/" + extPath
	}
	return basePath + extPath
}
