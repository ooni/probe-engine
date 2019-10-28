package engine

import (
	"context"

	"github.com/ooni/probe-engine/orchestra/testlists"
)

// TestListsConfig contains settings for TestListsClient
type TestListsConfig struct {
	BaseURL             string   // or use default
	AvailableCategories []string // or ask for all
	CountryCode         string   // set by session
	Limit               int      // or get all
}

// TestListsClient is a test lists client
type TestListsClient struct {
	client *testlists.Client
}

// URLInfo contains info about URLs
type URLInfo interface {
	URL() string
	CategoryCode() string
	CountryCode() string
}

// Fetch fetches the test list
func (c *TestListsClient) Fetch(config *TestListsConfig) ([]URLInfo, error) {
	var out []URLInfo
	if config.BaseURL != "" {
		c.client.BaseURL = config.BaseURL
	}
	list, err := c.client.Do(context.Background(), config.CountryCode, config.Limit)
	if err != nil {
		return nil, err
	}
	for _, entry := range list {
		out = append(out, &urlinfo{u: entry})
	}
	return out, nil
}

type urlinfo struct {
	u testlists.URLInfo
}

func (u *urlinfo) URL() string {
	return u.u.URL
}

func (u *urlinfo) CategoryCode() string {
	return u.u.CategoryCode
}

func (u *urlinfo) CountryCode() string {
	return u.u.CountryCode
}
