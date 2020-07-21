package resources

const (
	// Version contains the assets version.
	Version = 20200721121920

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
		URLPath:  "/ooni/probe-assets/releases/download/20200721121920/asn.mmdb.gz",
		GzSHA256: "e0f6adfbbbb565a04b97d9c78d77023e22f8a30477d36acab55169911cb6ba51",
		SHA256:   "600222c58a464c4cece3096b0b553c0db674276e81e5cd84e76b18acef1c2fce",
	},
	"ca-bundle.pem": {
		URLPath:  "/ooni/probe-assets/releases/download/20200721121920/ca-bundle.pem.gz",
		GzSHA256: "96de2f6469ce24c1909c82704df519566ec9ecddb7fd8b6ae635dba2a55c0e8c",
		SHA256:   "726889705b00f736200ed7999f7a50021b8735d53228d679c4e6665aa3b44987",
	},
	"country.mmdb": {
		URLPath:  "/ooni/probe-assets/releases/download/20200721121920/country.mmdb.gz",
		GzSHA256: "5f23c9155ce6e786c2c94d154d67411f33b39d67f1206fee8d6c61b829acdad5",
		SHA256:   "cffb80632d05ea76188bacaf6c085bb41540b283b7189b5f5fed9ae7173e7967",
	},
}
