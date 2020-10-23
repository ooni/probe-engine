package resources

const (
	// Version contains the assets version.
	Version = 20201023140148

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
		URLPath:  "/ooni/probe-assets/releases/download/20201023140148/asn.mmdb.gz",
		GzSHA256: "b02ff7444835e8d09c5e496f67f304ad9d7bea961da88cc8631d9ffaabec7ed8",
		SHA256:   "3dda76ad582dbb18e6652c8b4ed6c252564625b29973ce8324d7d9e4b487e418",
	},
	"country.mmdb": {
		URLPath:  "/ooni/probe-assets/releases/download/20201023140148/country.mmdb.gz",
		GzSHA256: "c6e71c3d192cab2d18d3f1822698b977ce2d458a3f2df93c20a5a30c95f9ed1b",
		SHA256:   "efc19fe493276f8cdd6df8488ae2f65978ec48529e7524451491ecd7aea8162b",
	},
}
