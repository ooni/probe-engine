package ooniapi

import (
	"context"
)

// URLSRequest is the URLS request.
type URLSRequest struct {
	Categories  string `query:"categories"`
	CountryCode string `query:"country_code"`
	Limit       int64  `query:"limit"`
	requestType
}

// URLSResponse is the URLS response.
type URLSResponse struct {
	Results []URLSResponseURL `json:"results"`
	responseType
}

// URLSResponseURL is a single URL in the URLS response.
type URLSResponseURL struct {
	CategoryCode string `json:"category_code"`
	CountryCode  string `json:"country_code"`
	URL          string `json:"url"`
}

// URLS implements the URLS API.
func (c Client) URLS(ctx context.Context, in *URLSRequest) (*URLSResponse, error) {
	var out URLSResponse
	err := c.api(ctx, apispec{
		Method:  "GET",
		URLPath: "/api/v1/test-list/urls",
		In:      in,
		Out:     &out,
	})
	return &out, err
}
