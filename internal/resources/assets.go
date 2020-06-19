package resources

const (
	// Version contains the assets version.
	Version = 20200619115947

	// ASNDatabaseName is the ASN-DB file name
	ASNDatabaseName = "asn.mmdb"

	// CABundleName is the name of the CA bundle file
	CABundleName = "ca-bundle.pem"

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
		URLPath:  "/ooni/probe-assets/releases/download/20200619115947/asn.mmdb.gz",
		GzSHA256: "2bc9e7d0e445d3f93edad094ef9060288850f34652f7592bdedabb5cf74a2065",
		SHA256:   "d71f0de2fdf17c809e5f60e9b69f874acbe9cde108298ecaedfab16d166cf878",
	},
	"ca-bundle.pem": {
		URLPath:  "/ooni/probe-assets/releases/download/20200619115947/ca-bundle.pem.gz",
		GzSHA256: "08070cbe24c8895d18bb20ccd746ff7409f1947094a1a47aa59993f588474485",
		SHA256:   "adf770dfd574a0d6026bfaa270cb6879b063957177a991d453ff1d302c02081f",
	},
	"country.mmdb": {
		URLPath:  "/ooni/probe-assets/releases/download/20200619115947/country.mmdb.gz",
		GzSHA256: "a3b13c78c149da4fa8bd63bbb0c25ead9737c1da3bf2e3199f2bca86d73c01b9",
		SHA256:   "83f369ddcb560862996848b600ce1e5353659dcb9424c5e9dd7f6d980fc56a60",
	},
}
