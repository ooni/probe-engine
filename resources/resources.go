// Package resources contains code to download resources.
package resources

import (
	"bytes"
	"compress/gzip"
	"context"
	"crypto/sha256"
	"fmt"
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

type resource struct {
	urlPath  string
	gzsha256 string
	sha256   string
}

var resources = map[string]resource{
	ASNDatabaseName: resource{
		urlPath:  "/releases/download/20190822135402/asn.mmdb.gz",
		gzsha256: "6cd343757dc4e3fe26de6f9f5a5b3e07a8b5949df99f3efd06e8d8d85e7031d1",
		sha256:   "d2649105e32b3a924ecac416b9d441fcf7f56186b4a3fe5bb3891e5d7c2c2d46",
	},
	CountryDatabaseName: resource{
		urlPath:  "/releases/download/20190822135402/country.mmdb.gz",
		gzsha256: "c29a631d448bace064d3ce675664714c3f4bec5839b14e618e07b18631189584",
		sha256:   "20bef853dd1288d55c9fd474cfe02f899f000f355c05252f355cbcedeba843b5",
	},
	CABundleName: resource{
		urlPath:  "/releases/download/20190822135402/ca-bundle.pem.gz",
		gzsha256: "d5a6aa2290ee18b09cc4fb479e2577ed5ae66c253870ba09776803a5396ea3ab",
		sha256:   "cb2eca3fbfa232c9e3874e3852d43b33589f27face98eef10242a853d83a437a",
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
		fullpath := filepath.Join(c.WorkDir, name)
		data, err := ioutil.ReadFile(fullpath)
		if err == nil {
			sha256sum := fmt.Sprintf("%x", sha256.Sum256(data))
			if sha256sum == resource.sha256 {
				continue
			}
			c.Logger.Debugf("resources: %s is outdated", fullpath)
		} else {
			c.Logger.Debugf("resources: can't read %s: %s", fullpath, err.Error())
		}
		URL := repository + resource.urlPath
		c.Logger.Debugf("resources: fetch %s", URL)
		data, err = (&fetch.Client{
			HTTPClient: c.HTTPClient,
			Logger:     c.Logger,
			UserAgent:  c.UserAgent,
		}).FetchAndVerify(ctx, URL, resource.gzsha256)
		if err != nil {
			return err
		}
		c.Logger.Debugf("resources: uncompress %s", fullpath)
		gzreader, err := gzip.NewReader(bytes.NewReader(data))
		if err != nil {
			return err
		}
		defer gzreader.Close()               // we already have a sha256 for it
		data, err = ioutil.ReadAll(gzreader) // small file
		if err != nil {
			return err
		}
		sha256sum := fmt.Sprintf("%x", sha256.Sum256(data))
		if sha256sum != resource.sha256 {
			return fmt.Errorf("resources: %s sha256 mismatch", fullpath)
		}
		c.Logger.Debugf("resources: overwrite %s", fullpath)
		err = ioutil.WriteFile(fullpath, data, 0600)
		if err != nil {
			return err
		}
	}
	return nil
}
