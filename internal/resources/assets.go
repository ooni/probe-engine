package resources

const (
	// Version contains the assets version.
	Version = 20200514135129

	// ASNDatabaseName is the ASN-DB file name
	ASNDatabaseName = "asn.mmdb"

	// CABundleName is the name of the CA bundle file
	CABundleName = "ca-bundle.pem"

	// CountryDatabaseName is country-DB file name
	CountryDatabaseName = "country.mmdb"

	// RepositoryURL is the asset's repository URL
	RepositoryURL = "https://github.com/ooni/probe-assets"
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
	"asn.mmdb": ResourceInfo{
		URLPath:  "/releases/download/20200514135129/asn.mmdb.gz",
		GzSHA256: "f4be54217b08761aa61315680eba2a9371b5b68aeb037e0c901773c43481530c",
		SHA256:   "479218cd19f52f9826bc836646a4bec1b04022e3fc42015d2aa577cdf2017f1a",
	},
	"ca-bundle.pem": ResourceInfo{
		URLPath:  "/releases/download/20200514135129/ca-bundle.pem.gz",
		GzSHA256: "08070cbe24c8895d18bb20ccd746ff7409f1947094a1a47aa59993f588474485",
		SHA256:   "adf770dfd574a0d6026bfaa270cb6879b063957177a991d453ff1d302c02081f",
	},
	"country.mmdb": ResourceInfo{
		URLPath:  "/releases/download/20200514135129/country.mmdb.gz",
		GzSHA256: "5d7b938a66e2018d9cec356a48c1042d85f1a47dbcf2f56247659d5cc6534e4c",
		SHA256:   "c38d94572026fadf86f5ee8f2258c3712d519c98b9f87c0119f6d65cd120d5ad",
	},
}
