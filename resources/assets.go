package resources

const (
	// Version contains the assets version.
	Version = 20191226162429

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
		URLPath:  "/releases/download/20191226162429/asn.mmdb.gz",
		GzSHA256: "f4be54217b08761aa61315680eba2a9371b5b68aeb037e0c901773c43481530c",
		SHA256:   "479218cd19f52f9826bc836646a4bec1b04022e3fc42015d2aa577cdf2017f1a",
	},
	"ca-bundle.pem": ResourceInfo{
		URLPath:  "/releases/download/20191226162429/ca-bundle.pem.gz",
		GzSHA256: "790ad79619bbfc75e4c8b4aef8fe6371d10f36f6071ec774d8cc0c27a848222a",
		SHA256:   "0d98a1a961aab523c9dc547e315e1d79e887dea575426ff03567e455fc0b66b4",
	},
	"country.mmdb": ResourceInfo{
		URLPath:  "/releases/download/20191226162429/country.mmdb.gz",
		GzSHA256: "e694431e766fdf76105b5c14ed43551bc9b4ca2e3bd282a41c3c36175e17a7e4",
		SHA256:   "9aedbe59990318581ec5d880716d3415e40bfc07280107112d6d7328c771e059",
	},
}
