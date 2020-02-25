package resources

const (
	// Version contains the assets version.
	Version = 20200225143707

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
		URLPath:  "/releases/download/20200225143707/asn.mmdb.gz",
		GzSHA256: "f4be54217b08761aa61315680eba2a9371b5b68aeb037e0c901773c43481530c",
		SHA256:   "479218cd19f52f9826bc836646a4bec1b04022e3fc42015d2aa577cdf2017f1a",
	},
	"ca-bundle.pem": ResourceInfo{
		URLPath:  "/releases/download/20200225143707/ca-bundle.pem.gz",
		GzSHA256: "08070cbe24c8895d18bb20ccd746ff7409f1947094a1a47aa59993f588474485",
		SHA256:   "adf770dfd574a0d6026bfaa270cb6879b063957177a991d453ff1d302c02081f",
	},
	"country.mmdb": ResourceInfo{
		URLPath:  "/releases/download/20200225143707/country.mmdb.gz",
		GzSHA256: "e82eed4d73ebd4f47a73d15a6c5c53aaec45f2d6faf50c87f8a7a8e66fb57910",
		SHA256:   "48bf9c397d1d8af6a89c30ca047401cc04bb0d768b93bb682f48a13ca1455ac5",
	},
}
