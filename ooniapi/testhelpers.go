package ooniapi

import "context"

// TestHelpersRequest is the TestHelpers Request.
type TestHelpersRequest struct {
	requestType
}

// TestHelpersResponse is the TestHelpers Response.
type TestHelpersResponse struct {
	HTTPReturnJSONHeaders []TestHelpersHelperInfo `json:"http-return-json-headers"`
	TCPEcho               []TestHelpersHelperInfo `json:"tcp-echo"`
	WebConnectivity       []TestHelpersHelperInfo `json:"web-connectivity"`
	responseType
}

// TestHelpersHelperInfo is a single helper within the
// response returned by the TestHelpers API.
type TestHelpersHelperInfo struct {
	Address string `json:"address"`
	Type    string `json:"type"`
	Front   string `json:"front,omitempty"`
}

// TestHelpers implements the TestHelpers API.
func (c Client) TestHelpers(ctx context.Context, in *TestHelpersRequest) (*TestHelpersResponse, error) {
	var out TestHelpersResponse
	err := c.api(ctx, apispec{
		Method:  "GET",
		URLPath: "/api/v1/test-helpers",
		In:      in,
		Out:     &out,
	})
	return &out, err
}
