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

	"github.com/ooni/probe-engine/internal/fetch"
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
		urlPath:  "/releases/download/20190520205742/asn.mmdb.gz",
		gzsha256: "9d2c7a1bafd626492645638cafab0617d689d5f18b8cd0659b9d1907ae34b2a6",
		sha256:   "b8a1ae0910de88ba3f2dc0a01219fcb23426385e4c91c20da9aaf4d41e5c4a5a",
	},
	CountryDatabaseName: resource{
		urlPath:  "/releases/download/20190520205742/country.mmdb.gz",
		gzsha256: "d9d5b5aeed2b4f7a73d01804e12332990cb0d9b4afcef150d56053baeeb9b35a",
		sha256:   "2a3c85df9dde8431e699411c28e85d716625e3f0850b717f8d61ba17190cbc88",
	},
	CABundleName: resource{
		urlPath:  "/releases/download/20190520205742/ca-bundle.pem.gz",
		gzsha256: "9f2d176644b779710cf3d4f975edc9d64897ef49867459d2ab94b3540943f30f",
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
				c.Logger.Debugf("resources: %s is up to date", fullpath)
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
		err = ioutil.WriteFile(fullpath, data, 600)
		if err != nil {
			return err
		}
	}
	return nil
}
