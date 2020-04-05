package resources

const (
	// Version contains the assets version.
	Version = 20200405230536

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
		URLPath:  "/releases/download/20200405230536/asn.mmdb.gz",
		GzSHA256: "f4be54217b08761aa61315680eba2a9371b5b68aeb037e0c901773c43481530c",
		SHA256:   "479218cd19f52f9826bc836646a4bec1b04022e3fc42015d2aa577cdf2017f1a",
	},
	"ca-bundle.pem": ResourceInfo{
		URLPath:  "/releases/download/20200405230536/ca-bundle.pem.gz",
		GzSHA256: "08070cbe24c8895d18bb20ccd746ff7409f1947094a1a47aa59993f588474485",
		SHA256:   "adf770dfd574a0d6026bfaa270cb6879b063957177a991d453ff1d302c02081f",
	},
	"country.mmdb": ResourceInfo{
		URLPath:  "/releases/download/20200405230536/country.mmdb.gz",
		GzSHA256: "fab3b477d9683247865c1233646118b5807205c342be81ef56b862e5197ace5a",
		SHA256:   "a2bfa672934f9fd2b6077adb30293fb85e559b3aacccc2a7b879a471f77f464b",
	},
}
