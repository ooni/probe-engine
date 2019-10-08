package psiphon

const (
	testName    = "psiphon"
	testVersion = "0.3.0"
)

// Config contains the experiment's configuration.
type Config struct {
	// ConfigFilePath is the path where Psiphon config file is located.
	ConfigFilePath string

	// WorkDir is the directory where Psiphon should store
	// its configuration database.
	WorkDir string
}
