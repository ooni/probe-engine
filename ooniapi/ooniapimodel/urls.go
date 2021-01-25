package ooniapimodel

// URLSRequest is the URLS request.
type URLSRequest struct {
	Categories  string `query:"categories"`
	CountryCode string `query:"country_code"`
	Limit       int64  `query:"limit"`
	RequestType
}

// URLSResponse is the URLS response.
type URLSResponse struct {
	Results []URLSResponseURL `json:"results"`
	ResponseType
}

// URLSResponseURL is a single URL in the URLS response.
type URLSResponseURL struct {
	CategoryCode string `json:"category_code"`
	CountryCode  string `json:"country_code"`
	URL          string `json:"url"`
}

// GETURLS is the GET /api/v1/test-list/urls API call.
type GETURLS struct {
	Method   MethodType  `method:"GET"`
	URLPath  URLPathType `path:"/api/v1/test-list/urls"`
	Request  URLSRequest
	Response URLSResponse
	APIType
}
