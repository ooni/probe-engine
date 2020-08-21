package resources

const (
	// Version contains the assets version.
	Version = 20200821081345

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
		URLPath:  "/ooni/probe-assets/releases/download/20200821081345/asn.mmdb.gz",
		GzSHA256: "b118504866e29bc3643c2aec54bf7605fdbac153d60491b0df6960b11837be90",
		SHA256:   "abcaa0a4bc7d9014affd72f501e64a9b708e701fb0733f72706c4c5ce55897d4",
	},
	"ca-bundle.pem": {
		URLPath:  "/ooni/probe-assets/releases/download/20200821081345/ca-bundle.pem.gz",
		GzSHA256: "3a99970bc782e5f7899de9011618bbadde5057d380e8fc7dc58ee96335bc1c30",
		SHA256:   "2782f0f8e89c786f40240fc1916677be660fb8d8e25dede50c9f6f7b0c2c2178",
	},
	"country.mmdb": {
		URLPath:  "/ooni/probe-assets/releases/download/20200821081345/country.mmdb.gz",
		GzSHA256: "c594dea05462f8013d8416cdcf65af61e9adb004d1d6aa9971e021f28cba35b7",
		SHA256:   "4570ac1701e52aa3d30b2512a3297e028cac6b66a5dac10715db4db8dcd84043",
	},
}
