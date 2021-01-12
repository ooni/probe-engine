package resources

const (
	// Version contains the assets version.
	Version = 20210112220115

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
		URLPath:  "/ooni/probe-assets/releases/download/20210112220115/asn.mmdb.gz",
		GzSHA256: "3973f23c8f3ac94bf7a7720f12ed42cc1b914d0ed146e19b57427cd5654cc167",
		SHA256:   "1d1ac82491c37757de377d4569a7b2f47683df247a33c314b7cb70df13bd4b3f",
	},
	"country.mmdb": {
		URLPath:  "/ooni/probe-assets/releases/download/20210112220115/country.mmdb.gz",
		GzSHA256: "5d465224ab02242a8a79652161d2768e64dd91fc1ed840ca3d0746f4cd29a914",
		SHA256:   "b4aa1292d072d9b2631711e6d3ac69c1e89687b4d513d43a1c330a92b7345e4d",
	},
}
