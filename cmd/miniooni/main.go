// Command miniooni is simple binary for testing purposes.
package main

import (
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"net/url"
	"os"
	"path"
	"path/filepath"
	"runtime"
	"strings"
	"time"

	"github.com/apex/log"
	engine "github.com/ooni/probe-engine"
	"github.com/pborman/getopt/v2"
)

type options struct {
	annotations  []string
	bouncerURL   string
	collectorURL string
	inputs       []string
	extraOptions []string
	noBouncer    bool
	noGeoIP      bool
	noJSON       bool
	noCollector  bool
	proxy        string
	reportfile   string
	verbose      bool
}

const (
	softwareName    = "miniooni"
	softwareVersion = "0.1.0-dev"
)

var (
	globalOptions options
	startTime     = time.Now()
)

func init() {
	getopt.FlagLong(
		&globalOptions.annotations, "annotation", 'A', "Add annotaton", "KEY=VALUE",
	)
	getopt.FlagLong(
		&globalOptions.bouncerURL, "bouncer", 'b', "Set bouncer base URL", "URL",
	)
	getopt.FlagLong(
		&globalOptions.collectorURL, "collector", 'c',
		"Set collector base URL", "URL",
	)
	getopt.FlagLong(
		&globalOptions.inputs, "input", 'i',
		"Add test-dependent input to the test input", "INPUT",
	)
	getopt.FlagLong(
		&globalOptions.extraOptions, "option", 'O',
		"Pass an option to the experiment", "KEY=VALUE",
	)
	getopt.FlagLong(
		&globalOptions.noBouncer, "no-bouncer", 0, "Don't use the OONI bouncer",
	)
	getopt.FlagLong(
		&globalOptions.noGeoIP, "no-geoip", 'g', "Disable GeoIP lookup",
	)
	getopt.FlagLong(
		&globalOptions.noJSON, "no-json", 'N', "Disable writing to disk",
	)
	getopt.FlagLong(
		&globalOptions.noCollector, "no-collector", 'n', "Don't use a collector",
	)
	getopt.FlagLong(
		&globalOptions.proxy, "proxy", 'P', "Set the proxy URL", "URL",
	)
	getopt.FlagLong(
		&globalOptions.reportfile, "reportfile", 'o',
		"Set the report file path", "PATH",
	)
	getopt.FlagLong(
		&globalOptions.verbose, "verbose", 'v', "Increase verbosity",
	)
}

func split(s string) (string, string, error) {
	v := strings.SplitN(s, "=", 2)
	if len(v) != 2 {
		return "", "", errors.New("invalid key-value pair")
	}
	return v[0], v[1], nil
}

func mustMakeMap(input []string) (output map[string]string) {
	output = make(map[string]string)
	for _, opt := range input {
		key, value, err := split(opt)
		if err != nil {
			log.WithError(err).Fatal("cannot split key-value pair")
		}
		output[key] = value
	}
	return
}

func mustParseURL(URL string) *url.URL {
	rv, err := url.Parse(URL)
	if err != nil {
		log.WithError(err).Fatal("cannot parse URL")
	}
	return rv
}

type logHandler struct {
	io.Writer
}

func (h *logHandler) HandleLog(e *log.Entry) (err error) {
	s := fmt.Sprintf("[%14.6f] <%s> %s",
		time.Since(startTime).Seconds(),
		e.Level, e.Message)
	if len(e.Fields) > 0 {
		s += fmt.Sprintf(": %+v", e.Fields)
	}
	s += "\n"
	_, err = h.Writer.Write([]byte(s))
	return
}

// See https://gist.github.com/miguelmota/f30a04a6d64bd52d7ab59ea8d95e54da
func gethomedir() string {
	if runtime.GOOS == "windows" {
		home := os.Getenv("HOMEDRIVE") + os.Getenv("HOMEPATH")
		if home == "" {
			home = os.Getenv("USERPROFILE")
		}
		return home
	}
	if runtime.GOOS == "linux" {
		home := os.Getenv("XDG_CONFIG_HOME")
		if home != "" {
			return home
		}
		// fallthrough
	}
	return os.Getenv("HOME")
}

func main() {
	getopt.Parse()
	if len(getopt.Args()) != 1 {
		log.Fatal("You must specify the name of the experiment to run")
	}
	extraOptions := mustMakeMap(globalOptions.extraOptions)
	annotations := mustMakeMap(globalOptions.annotations)

	logger := &log.Logger{
		Level:   log.InfoLevel,
		Handler: &logHandler{Writer: os.Stderr},
	}
	if globalOptions.verbose {
		logger.Level = log.DebugLevel
	}
	if globalOptions.reportfile == "" {
		globalOptions.reportfile = "report.jsonl"
	}
	log.Log = logger

	homeDir := gethomedir()
	if homeDir == "" {
		log.Fatal("home directory is empty")
	}
	miniooniDir := path.Join(homeDir, ".miniooni")
	assetsDir := path.Join(miniooniDir, "assets")
	if err := os.MkdirAll(assetsDir, 0700); err != nil {
		log.WithError(err).Fatal("cannot create assets directory")
	}
	log.Infof("miniooni state directory: %s", miniooniDir)
	tempDir, err := ioutil.TempDir("", "miniooni")
	if err != nil {
		log.WithError(err).Fatal("cannot get a temporary directory")
	}
	log.Debugf("miniooni temporary directory: %s", tempDir)

	var proxyURL *url.URL
	if globalOptions.proxy != "" {
		proxyURL = mustParseURL(globalOptions.proxy)
	}

	kvstore2dir := filepath.Join(miniooniDir, "kvstore2")
	kvstore, err := engine.NewFileSystemKVStore(kvstore2dir)
	if err != nil {
		log.WithError(err).Fatal("cannot create kvstore2 directory")
	}

	sess, err := engine.NewSession(engine.SessionConfig{
		AssetsDir:       assetsDir,
		KVStore:         kvstore,
		Logger:          logger,
		ProxyURL:        proxyURL,
		SoftwareName:    softwareName,
		SoftwareVersion: softwareVersion,
		TempDir:         tempDir,
	})
	if err != nil {
		log.WithError(err).Fatal("cannot create measurement session")
	}

	if globalOptions.bouncerURL != "" {
		sess.AddAvailableHTTPSBouncer(globalOptions.bouncerURL)
	}
	if globalOptions.collectorURL != "" {
		// Implementation note: setting the collector before doing the lookup
		// is totally fine because it's a maybe lookup, meaning that any bit
		// of information already available will not be looked up again.
		sess.AddAvailableHTTPSCollector(globalOptions.collectorURL)
	}

	if !globalOptions.noBouncer {
		log.Info("Looking up OONI backends")
		if err := sess.MaybeLookupBackends(); err != nil {
			log.WithError(err).Fatal("cannot lookup OONI backends")
		}
	}
	if !globalOptions.noGeoIP {
		log.Info("Looking up your location")
		if err := sess.MaybeLookupLocation(); err != nil {
			log.WithError(err).Warn("cannot lookup your location")
		} else {
			log.Infof("- IP: %s", sess.ProbeIP())
			log.Infof("- country: %s", sess.ProbeCC())
			log.Infof(
				"- network: %s (%s)", sess.ProbeNetworkName(), sess.ProbeASNString(),
			)
			log.Infof("- resolver's IP: %s", sess.ResolverIP())
			log.Infof(
				"- resolver's network: %s (%s)",
				sess.ResolverNetworkName(),
				sess.ResolverASNString(),
			)
		}
	}

	name := getopt.Args()[0]
	builder, err := sess.NewExperimentBuilder(name)
	if err != nil {
		log.WithError(err).Fatal("cannot create experiment builder")
	}
	if builder.NeedsInput() {
		if len(globalOptions.inputs) <= 0 {
			log.Info("Fetching test lists")
			list, err := sess.QueryTestListsURLs(&engine.TestListsURLsConfig{
				Limit: 16,
			})
			if err != nil {
				log.WithError(err).Fatal("cannot fetch test lists")
			}
			for _, entry := range list.Result {
				globalOptions.inputs = append(globalOptions.inputs, entry.URL)
			}
		}
	} else if len(globalOptions.inputs) != 0 {
		log.Fatal("this experiment does not expect any input")
	} else {
		// Tests that do not expect input internally require an empty input to run
		globalOptions.inputs = append(globalOptions.inputs, "")
	}
	for key, value := range extraOptions {
		err := builder.SetOptionString(key, value)
		if err != nil {
			log.WithError(err).Fatal("cannot set string option")
		}
	}
	experiment := builder.NewExperiment()

	if !globalOptions.noCollector {
		if err := experiment.OpenReport(); err != nil {
			log.WithError(err).Fatal("cannot open report")
		}
		defer experiment.CloseReport()
	}

	inputCount := len(globalOptions.inputs)
	inputCounter := 0
	for _, input := range globalOptions.inputs {
		inputCounter++
		if input != "" {
			log.Infof("[%d/%d] running with input: %s", inputCounter, inputCount, input)
		}
		measurement, err := experiment.Measure(input)
		if err != nil {
			log.WithError(err).Warn("measurement failed")
			// fallthrough and try to submit what we have anyway. Even if it
			// has failed badly, we'd rather see it.
		}
		measurement.AddAnnotations(annotations)
		if !globalOptions.noCollector {
			log.Infof("submitting measurement to OONI collector")
			if err := experiment.SubmitAndUpdateMeasurement(measurement); err != nil {
				log.WithError(err).Warn("submitting measurement failed")
				// fallthrough and save to disk what we have. Not saving is
				// worst because it means we cannot eventually resubmit.
			}
		}
		if !globalOptions.noJSON {
			// Note: must be after submission because submission modifies
			// the measurement to include the report ID.
			log.Infof("saving measurement to disk")
			if err := experiment.SaveMeasurement(
				measurement, globalOptions.reportfile,
			); err != nil {
				log.WithError(err).Warn("saving measurement failed")
				// fallthrough because we're at the bottom of the loop
			}
		}
	}
}
