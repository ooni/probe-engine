package resources

const (
	// Version contains the assets version.
	Version = 20191114173737

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
		URLPath:  "/releases/download/20191114173737/asn.mmdb.gz",
		GzSHA256: "ee7b8b939d74dd75fc58eb1a39eae833c2ba11200c46640977ff8e825e2f0c53",
		SHA256:   "bf9943b5bc6d36573dad1bca1e5d9a3bc523620700b33efbe0261f8fe4aaa967",
	},
	"ca-bundle.pem": ResourceInfo{
		URLPath:  "/releases/download/20191114173737/ca-bundle.pem.gz",
		GzSHA256: "07e47ee9bffe655fdaee5c1a369b9580a53401c1dfde28a02664a40973eb369c",
		SHA256:   "5cd8052fcf548ba7e08899d8458a32942bf70450c9af67a0850b4c711804a2e4",
	},
	"country.mmdb": ResourceInfo{
		URLPath:  "/releases/download/20191114173737/country.mmdb.gz",
		GzSHA256: "33af311275634fb686308e3b7e49b8c0ecdc4e2b152a29abeed948a3a0b82152",
		SHA256:   "1c668e490cdd84b140179d852c19116a35ec8612dd3beabeff7ce1dce536fd30",
	},
}
