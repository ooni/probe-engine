package ooniapimodel

// TestHelpersRequest is the TestHelpers request.
type TestHelpersRequest struct {
	RequestType
}

// TestHelpersResponse is the TestHelpers response.
type TestHelpersResponse struct {
	HTTPReturnJSONHeaders []TestHelpersHelperInfo `json:"http-return-json-headers"`
	TCPEcho               []TestHelpersHelperInfo `json:"tcp-echo"`
	WebConnectivity       []TestHelpersHelperInfo `json:"web-connectivity"`
	ResponseType
}

// TestHelpersHelperInfo is a single helper within the
// response returned by the TestHelpers API.
type TestHelpersHelperInfo struct {
	Address string `json:"address"`
	Type    string `json:"type"`
	Front   string `json:"front,omitempty"`
}

// GETTestHelpers is the GET /api/v1/test-helpers API call.
type GETTestHelpers struct {
	Method   MethodType  `method:"GET"`
	URLPath  URLPathType `path:"/api/v1/test-helpers"`
	Request  TestHelpersRequest
	Response TestHelpersResponse
	APIType
}
