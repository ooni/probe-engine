package resources

const (
	// Version contains the assets version.
	Version = 20201207105127

	// ASNDatabaseName is the ASN-DB file name
	ASNDatabaseName = "asn.mmdb"

	// CountryDatabaseName is country-DB file name
	CountryDatabaseName = "country.mmdb"

	// BaseURL is the asset's repository base URL
	BaseURL = "https://github.com/"
)

// ResourceInfo contains information on a resource.
type ResourceInfo struct {
	// URLPath is the resource's URL path.
	URLPath string

	// GzSHA256 is used to validate the downloaded file.
	GzSHA256 string

	// SHA256 is used to check whether the assets file
	// stored locally is still up-to-date.
	SHA256 string
}

// All contains info on all known assets.
var All = map[string]ResourceInfo{
	"asn.mmdb": {
		URLPath:  "/ooni/probe-assets/releases/download/20201207105127/asn.mmdb.gz",
		GzSHA256: "797f875415692d1c0dd4c2e87edefc2e4f37e477b296816ab86c96316faa6a9b",
		SHA256:   "9f8c85ee924d657a25ae0ce98b0b87ff5ff082f0acc57a119bae2d29af106ea4",
	},
	"country.mmdb": {
		URLPath:  "/ooni/probe-assets/releases/download/20201207105127/country.mmdb.gz",
		GzSHA256: "8ba25a0f64151e327c2877e2521eebd1db33ca85918e78861af1d03d5dcaeb8b",
		SHA256:   "87c4a2446e2dc2e29d272d3877a6ba5fb7a5cf551dfe705db13f5a5e63a32cf3",
	},
}
