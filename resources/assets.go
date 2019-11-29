package resources

const (
	// Version contains the assets version.
	Version = 20191129161951

	// ASNDatabaseName is the ASN-DB file name
	ASNDatabaseName = "asn.mmdb"

	// CABundleName is the name of the CA bundle file
	CABundleName = "ca-bundle.pem"

	// CountryDatabaseName is country-DB file name
	CountryDatabaseName = "country.mmdb"

	// RepositoryURL is the asset's repository URL
	RepositoryURL = "http://github.com/ooni/probe-assets"
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
		URLPath:  "/releases/download/20191129161951/asn.mmdb.gz",
		GzSHA256: "afbc764c8c0c7a11625c14db2c265183deb0c2c84e998b2d1eb67abd647af13e",
		SHA256:   "aa0b9b39971f047dbfd980eb3ed3ca18c4872a17e1b40e39a482089b83469159",
	},
	"ca-bundle.pem": ResourceInfo{
		URLPath:  "/releases/download/20191129161951/ca-bundle.pem.gz",
		GzSHA256: "fe1c6e28a5553ce47e0c2e1ff8ac4c96793e82085c3958ee5c7df95547b9a10f",
		SHA256:   "0d98a1a961aab523c9dc547e315e1d79e887dea575426ff03567e455fc0b66b4",
	},
	"country.mmdb": ResourceInfo{
		URLPath:  "/releases/download/20191129161951/country.mmdb.gz",
		GzSHA256: "7940b6220e9f6cd6a1c0092141d2ce5ca3408ded497fac3732218a1074be5c24",
		SHA256:   "f7eac8be868b987865e7094764966f046771106183838c121c6364b9a59c0b9a",
	},
}
