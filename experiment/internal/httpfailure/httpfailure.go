// Package httpfailure groups a bunch of extra HTTP failures.
//
// These failures only matter in the context of processing the results
// of specific experiments, e.g., whatsapp, telegram.
package httpfailure

var (
	// UnexpectedStatusCode indicates that the web interface
	// is not redirecting us with the expected status code.
	UnexpectedStatusCode = "http_unexpected_status_code"

	// UnexpectedRedirectURL indicates that the redirect URL
	// returned by the server is not the expected one.
	UnexpectedRedirectURL = "http_unexpected_redirect_url"
)
