package resources

const (
	// Version contains the assets version.
	Version = 20200610133507

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
	"asn.mmdb": {
		URLPath:  "/releases/download/20200610133507/asn.mmdb.gz",
		GzSHA256: "19c7184dcd6339d33aea22811ed015eeb4d8ce19606bba2cb5bdc1b205b602f0",
		SHA256:   "a89cef5b583607891e2e3fec620efcd3276c51684909351433adc6c870854089",
	},
	"ca-bundle.pem": {
		URLPath:  "/releases/download/20200610133507/ca-bundle.pem.gz",
		GzSHA256: "08070cbe24c8895d18bb20ccd746ff7409f1947094a1a47aa59993f588474485",
		SHA256:   "adf770dfd574a0d6026bfaa270cb6879b063957177a991d453ff1d302c02081f",
	},
	"country.mmdb": {
		URLPath:  "/releases/download/20200610133507/country.mmdb.gz",
		GzSHA256: "a3b13c78c149da4fa8bd63bbb0c25ead9737c1da3bf2e3199f2bca86d73c01b9",
		SHA256:   "83f369ddcb560862996848b600ce1e5353659dcb9424c5e9dd7f6d980fc56a60",
	},
}
