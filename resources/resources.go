// Package resources contains code to download resources.
package resources

import (
	"bytes"
	"compress/gzip"
	"context"
	"crypto/sha256"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"path/filepath"

	"github.com/ooni/probe-engine/httpx/fetch"
	"github.com/ooni/probe-engine/log"
)

const (
	// ASNDatabaseName is the name of the ASN database file
	ASNDatabaseName = "asn.mmdb"

	// CABundleName is the name of the CA bundle file
	CABundleName = "ca-bundle.pem"

	// CountryDatabaseName is the name of the country database file
	CountryDatabaseName = "country.mmdb"

	repository = "https://github.com/measurement-kit/generic-assets"
)

// ResourceInfo contains information on a resource.
type ResourceInfo struct {
	URLPath  string
	GzSHA256 string
	SHA256   string
}

var resources = map[string]ResourceInfo{
	ASNDatabaseName: ResourceInfo{
		URLPath:  "/releases/download/20190822135402/asn.mmdb.gz",
		GzSHA256: "6cd343757dc4e3fe26de6f9f5a5b3e07a8b5949df99f3efd06e8d8d85e7031d1",
		SHA256:   "d2649105e32b3a924ecac416b9d441fcf7f56186b4a3fe5bb3891e5d7c2c2d46",
	},
	CountryDatabaseName: ResourceInfo{
		URLPath:  "/releases/download/20190822135402/country.mmdb.gz",
		GzSHA256: "c29a631d448bace064d3ce675664714c3f4bec5839b14e618e07b18631189584",
		SHA256:   "20bef853dd1288d55c9fd474cfe02f899f000f355c05252f355cbcedeba843b5",
	},
	CABundleName: ResourceInfo{
		URLPath:  "/releases/download/20190822135402/ca-bundle.pem.gz",
		GzSHA256: "d5a6aa2290ee18b09cc4fb479e2577ed5ae66c253870ba09776803a5396ea3ab",
		SHA256:   "cb2eca3fbfa232c9e3874e3852d43b33589f27face98eef10242a853d83a437a",
	},
}

// Client is a client for fetching resources.
type Client struct {
	// HTTPClient is the HTTP client to use.
	HTTPClient *http.Client

	// Logger is the logger to use.
	Logger log.Logger

	// UserAgent is the user agent to use.
	UserAgent string

	// WorkDir is the directory where to save resources.
	WorkDir string
}

// Ensure ensures that resources are downloaded and current.
func (c *Client) Ensure(ctx context.Context) error {
	for name, resource := range resources {
		if err := c.EnsureForSingleResource(
			ctx, name, resource, func(real, expected string) bool {
				return real == expected
			},
			gzip.NewReader, ioutil.ReadAll,
		); err != nil {
			return err
		}
	}
	return nil
}

// EnsureForSingleResource ensures that a single resource
// is downloaded and is current.
func (c *Client) EnsureForSingleResource(
	ctx context.Context, name string, resource ResourceInfo,
	equal func(real, expected string) bool,
	gzipNewReader func(r io.Reader) (*gzip.Reader, error),
	ioutilReadAll func(r io.Reader) ([]byte, error),
) error {
	fullpath := filepath.Join(c.WorkDir, name)
	data, err := ioutil.ReadFile(fullpath)
	if err == nil {
		sha256sum := fmt.Sprintf("%x", sha256.Sum256(data))
		if equal(sha256sum, resource.SHA256) {
			return nil
		}
		c.Logger.Debugf("resources: %s is outdated", fullpath)
	} else {
		c.Logger.Debugf("resources: can't read %s: %s", fullpath, err.Error())
	}
	URL := repository + resource.URLPath
	c.Logger.Debugf("resources: fetch %s", URL)
	data, err = (&fetch.Client{
		HTTPClient: c.HTTPClient,
		Logger:     c.Logger,
		UserAgent:  c.UserAgent,
	}).FetchAndVerify(ctx, URL, resource.GzSHA256)
	if err != nil {
		return err
	}
	c.Logger.Debugf("resources: uncompress %s", fullpath)
	gzreader, err := gzipNewReader(bytes.NewReader(data))
	if err != nil {
		return err
	}
	defer gzreader.Close()              // we already have a sha256 for it
	data, err = ioutilReadAll(gzreader) // small file
	if err != nil {
		return err
	}
	sha256sum := fmt.Sprintf("%x", sha256.Sum256(data))
	if equal(sha256sum, resource.SHA256) == false {
		return fmt.Errorf("resources: %s sha256 mismatch", fullpath)
	}
	c.Logger.Debugf("resources: overwrite %s", fullpath)
	return ioutil.WriteFile(fullpath, data, 0600)
}
