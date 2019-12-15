package resources

const (
	// Version contains the assets version.
	Version = 20191213165358

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
		URLPath:  "/releases/download/20191213165358/asn.mmdb.gz",
		GzSHA256: "ff535efc7055179ce121f57dcd0632fbee7a074f01574d01ef8e9b14c0c64c09",
		SHA256:   "97080cfc637f2296aef34b2888ab8f6f3f5818c201489360acdeb109d457eccf",
	},
	"ca-bundle.pem": ResourceInfo{
		URLPath:  "/releases/download/20191213165358/ca-bundle.pem.gz",
		GzSHA256: "fe1c6e28a5553ce47e0c2e1ff8ac4c96793e82085c3958ee5c7df95547b9a10f",
		SHA256:   "0d98a1a961aab523c9dc547e315e1d79e887dea575426ff03567e455fc0b66b4",
	},
	"country.mmdb": ResourceInfo{
		URLPath:  "/releases/download/20191213165358/country.mmdb.gz",
		GzSHA256: "54742c731b86288df61a633209620097b35beee48b9fa1836e6a643c1bc14f5f",
		SHA256:   "dca5e0de1a834366afe09fbaa6b72b2bcc196f2bc4fbbae4c6bf3d048bb29072",
	},
}
