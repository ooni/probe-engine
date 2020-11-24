package resources

const (
	// Version contains the assets version.
	Version = 20201112191431

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
		URLPath:  "/ooni/probe-assets/releases/download/20201112191431/asn.mmdb.gz",
		GzSHA256: "44945b2c056c05b7a48b13659bcdb92ea0b6d511e94cc615dd1467400ea57637",
		SHA256:   "b4b2d3c33ff4167a95850295421c13e7e3d0bd429c269cda86341c465c4e76e1",
	},
	"country.mmdb": {
		URLPath:  "/ooni/probe-assets/releases/download/20201112191431/country.mmdb.gz",
		GzSHA256: "2cc5f66f6eada914a01b1f9da032beb7fef2a14e546de51028a9662d58cbf6b0",
		SHA256:   "c28169745af88c0f512072dcc6091e26a7969e4755ef1bb061e9eaee15900106",
	},
}
