// Command miniooni is simple binary for testing purposes.
package main

import (
	"context"
	"crypto/tls"
	"errors"
	"fmt"
	"io"
	"net/url"
	"os"
	"strings"
	"time"

	"github.com/apex/log"

	"github.com/ooni/probe-engine/experiment"
	"github.com/ooni/probe-engine/experiment/dash"
	"github.com/ooni/probe-engine/experiment/example"
	"github.com/ooni/probe-engine/experiment/fbmessenger"
	"github.com/ooni/probe-engine/experiment/hhfm"
	"github.com/ooni/probe-engine/experiment/hirl"
	"github.com/ooni/probe-engine/experiment/ndt"
	"github.com/ooni/probe-engine/experiment/ndt7"
	"github.com/ooni/probe-engine/experiment/psiphon"
	"github.com/ooni/probe-engine/experiment/telegram"
	"github.com/ooni/probe-engine/experiment/web_connectivity"
	"github.com/ooni/probe-engine/experiment/whatsapp"
	"github.com/ooni/probe-engine/httpx/httpx"
	"github.com/ooni/probe-engine/orchestra/testlists"
	"github.com/ooni/probe-engine/session"

	"github.com/pborman/getopt/v2"
)

type options struct {
	annotations  []string
	bouncerURL   string
	caBundlePath string
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
		&globalOptions.caBundlePath, "ca-bundle-path", 0,
		"Set CA bundle path", "PATH",
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

func mustParseCA(caBundlePath string) *tls.Config {
	config, err := httpx.NewTLSConfigWithCABundle(caBundlePath)
	if err != nil {
		log.WithError(err).Fatal("cannot load CA bundle")
	}
	return config
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
	workDir, err := os.Getwd()
	if err != nil {
		log.WithError(err).Fatal("cannot get current working directory")
	}
	log.Log = logger

	ctx := context.Background()
	var tlsConfig *tls.Config
	if globalOptions.caBundlePath != "" {
		tlsConfig = mustParseCA(globalOptions.caBundlePath)
	}
	var proxyURL *url.URL
	if globalOptions.proxy != "" {
		proxyURL = mustParseURL(globalOptions.proxy)
	}
	sess := session.New(
		logger, softwareName, softwareVersion, workDir, proxyURL, tlsConfig,
	)

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
		if err := sess.MaybeLookupBackends(ctx); err != nil {
			log.WithError(err).Fatal("cannot lookup OONI backends")
		}
	}
	if !globalOptions.noGeoIP {
		log.Info("Looking up your location")
		if err := sess.MaybeLookupLocation(ctx); err != nil {
			log.WithError(err).Warn("cannot lookup your location")
		} else {
			log.Infof("your IP: %s, country: %s, ISP name: %s",
				sess.Location.ProbeIP, sess.Location.CountryCode, sess.Location.NetworkName)
		}
	}

	name := getopt.Args()[0]

	if name == "web_connectivity" {
		log.Info("Fetching test lists")
		if len(globalOptions.inputs) <= 0 {
			list, err := testlists.NewClient(sess).Do(ctx, sess.ProbeCC(), 100)
			if err != nil {
				log.WithError(err).Fatal("cannot fetch test lists")
			}
			for _, entry := range list {
				globalOptions.inputs = append(globalOptions.inputs, entry.URL)
			}
		}
	} else if len(globalOptions.inputs) != 0 {
		log.Fatal("this test does not expect any input")
	} else {
		// Tests that do not expect input internally require an empty input to run
		globalOptions.inputs = append(globalOptions.inputs, "")
	}

	var experiment *experiment.Experiment
	if name == "dash" {
		experiment = dash.NewExperiment(sess, dash.Config{})
	} else if name == "facebook_messenger" {
		experiment = fbmessenger.NewExperiment(sess, fbmessenger.Config{})
	} else if name == "http_header_field_manipulation" {
		experiment = hhfm.NewExperiment(sess, hhfm.Config{})
	} else if name == "http_invalid_request_line" {
		experiment = hirl.NewExperiment(sess, hirl.Config{})
	} else if name == "ndt" {
		experiment = ndt.NewExperiment(sess, ndt.Config{})
	} else if name == "ndt7" {
		experiment = ndt7.NewExperiment(sess, ndt7.Config{})
	} else if name == "psiphon" {
		if _, ok := extraOptions["config_file_path"]; !ok {
			log.Fatal("psiphon requires the `-O config_file_path=PATH` option")
		}
		experiment = psiphon.NewExperiment(sess, psiphon.Config{
			ConfigFilePath: extraOptions["config_file_path"],
			WorkDir:        workDir,
		})
	} else if name == "telegram" {
		experiment = telegram.NewExperiment(sess, telegram.Config{})
	} else if name == "web_connectivity" {
		experiment = web_connectivity.NewExperiment(sess, web_connectivity.Config{})
	} else if name == "whatsapp" {
		experiment = whatsapp.NewExperiment(sess, whatsapp.Config{})
	} else if name == "example" {
		experiment = example.NewExperiment(sess, example.Config{2 * time.Second})
	} else {
		log.Fatalf("Unknown experiment: %s", name)
	}

	if !globalOptions.noCollector {
		if err := experiment.OpenReport(ctx); err != nil {
			log.WithError(err).Fatal("cannot open report")
		}
		defer experiment.CloseReport(ctx)
	}

	inputCount := len(globalOptions.inputs)
	inputCounter := 0
	for _, input := range globalOptions.inputs {
		inputCounter++
		if input != "" {
			log.Infof("[%d/%d] running with input: %s", inputCounter, inputCount, input)
		}
		measurement, err := experiment.Measure(ctx, input)
		if err != nil {
			log.WithError(err).Warn("measurement failed")
			continue
		}
		measurement.AddAnnotations(annotations)
		if !globalOptions.noCollector {
			log.Infof("submitting measurement to OONI collector")
			if err := experiment.SubmitMeasurement(ctx, &measurement); err != nil {
				log.WithError(err).Warn("submitting measurement failed")
				continue
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
				continue
			}
		}
	}
}
