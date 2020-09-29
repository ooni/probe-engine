package resources

const (
	// Version contains the assets version.
	Version = 20200929203018

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
		URLPath:  "/ooni/probe-assets/releases/download/20200929203018/asn.mmdb.gz",
		GzSHA256: "abfed7750af355c2e75feed73cb5a4cf44f4ecb9866199900c65a1e3cb58bda9",
		SHA256:   "3958d1248b13b5aedc5f03e528a3ef2f2ef4ebdc4449bcd6667d03f722988dc9",
	},
	"ca-bundle.pem": {
		URLPath:  "/ooni/probe-assets/releases/download/20200929203018/ca-bundle.pem.gz",
		GzSHA256: "3a99970bc782e5f7899de9011618bbadde5057d380e8fc7dc58ee96335bc1c30",
		SHA256:   "2782f0f8e89c786f40240fc1916677be660fb8d8e25dede50c9f6f7b0c2c2178",
	},
	"country.mmdb": {
		URLPath:  "/ooni/probe-assets/releases/download/20200929203018/country.mmdb.gz",
		GzSHA256: "d76b387d25c5fd408392a28c02938de149116ad7338d4677aec0835d32cd017c",
		SHA256:   "36bbb5057022934ed51c6fe1e093b42be15388e0adc1760250f030df21f6f501",
	},
}
