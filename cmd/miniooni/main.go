// Command miniooni is simple binary for testing purposes.
package main

import (
	"context"
	"crypto/tls"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"net/url"
	"os"
	"strings"

	"github.com/apex/log"
	"github.com/apex/log/handlers/multi"

	"github.com/ooni/probe-engine/experiment"
	"github.com/ooni/probe-engine/experiment/dash"
	"github.com/ooni/probe-engine/experiment/fbmessenger"
	"github.com/ooni/probe-engine/experiment/harconnectivity"
	"github.com/ooni/probe-engine/experiment/hhfm"
	"github.com/ooni/probe-engine/experiment/hirl"
	"github.com/ooni/probe-engine/experiment/ndt"
	"github.com/ooni/probe-engine/experiment/ndt7"
	"github.com/ooni/probe-engine/experiment/psiphon"
	"github.com/ooni/probe-engine/experiment/telegram"
	"github.com/ooni/probe-engine/experiment/web_connectivity"
	"github.com/ooni/probe-engine/experiment/whatsapp"
	"github.com/ooni/probe-engine/httpx/httpx"
	"github.com/ooni/probe-engine/httpx/minihar"
	"github.com/ooni/probe-engine/model"
	"github.com/ooni/probe-engine/oohar"
	"github.com/ooni/probe-engine/orchestra/testlists"
	"github.com/ooni/probe-engine/session"

	"github.com/pborman/getopt/v2"
)

type options struct {
	annotations  []string
	bouncerURL   string
	caBundlePath string
	collectorURL string
	extraOptions []string
	harfile      string
	inputs       []string
	logfile      string
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
		&globalOptions.extraOptions, "option", 'O',
		"Pass an option to the experiment", "KEY=VALUE",
	)
	getopt.FlagLong(
		&globalOptions.harfile, "harfile", 0,
		"Dump requests in HAR format into the specified file", "PATH",
	)
	getopt.FlagLong(
		&globalOptions.inputs, "input", 0,
		"Add input for tests that use input", "INPUT",
	)
	getopt.FlagLong(
		&globalOptions.logfile, "logfile", 'l', "Set the logfile path", "PATH",
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
	s := fmt.Sprintf("<%s> %s", e.Level, e.Message)
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
	if globalOptions.logfile != "" {
		filep, err := os.Create(globalOptions.logfile)
		if err != nil {
			log.WithError(err).Fatal("cannot open logfile")
		}
		defer filep.Close()
		logger.Handler = multi.New(logger.Handler, &logHandler{Writer: filep})
	}
	workDir, err := os.Getwd()
	if err != nil {
		log.WithError(err).Fatal("cannot get current working directory")
	}
	log.Log = logger // from now on logs may be redirected

	ctx := context.Background()
	if globalOptions.harfile != "" {
		var rs *minihar.RequestSaver
		ctx, rs = minihar.WithRequestSaver(ctx)
		defer func() {
			oolog := oohar.NewLogFromMiniHAR(softwareName, softwareVersion, rs)
			data, err := json.MarshalIndent(oolog, "", "  ")
			if err != nil {
				log.WithError(err).Fatal("Cannot serialize HAR file")
			}
			err = ioutil.WriteFile(globalOptions.harfile, data, 0644)
			if err != nil {
				log.WithError(err).Fatal("Cannot write HAR file")
			}
		}()
	}
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
		sess.SetAvailableHTTPSBouncer(globalOptions.bouncerURL)
	}
	if globalOptions.collectorURL != "" {
		// Implementation note: setting the collector before doing the lookup
		// is totally fine because it's a maybe lookup, meaning that any bit
		// of information already available will not be looked up again.
		sess.SetAvailableHTTPSCollector(globalOptions.collectorURL)
	}

	if !globalOptions.noBouncer {
		if err := sess.MaybeLookupBackends(ctx); err != nil {
			log.WithError(err).Fatal("cannot lookup OONI backends")
		}
	}
	if !globalOptions.noGeoIP {
		if err := sess.MaybeLookupLocation(ctx); err != nil {
			log.WithError(err).Warn("cannot lookup your location")
		}
	}

	name := getopt.Args()[0]
	var inputs []string

	if name == "web_connectivity" || name == "harconnectivity" {
		if len(globalOptions.inputs) <= 0 {
			list, err := testlists.NewClient(sess).Do(ctx, sess.ProbeCC())
			if err != nil {
				log.WithError(err).Fatal("cannot fetch test lists")
			}
			for _, entry := range list {
				inputs = append(inputs, entry.URL)
			}
		} else {
			inputs = globalOptions.inputs
		}
	} else {
		inputs = append(inputs, "")
	}

	var experiment *experiment.Experiment
	if name == "dash" {
		experiment = dash.NewExperiment(sess, dash.Config{})
	} else if name == "facebook_messenger" {
		experiment = fbmessenger.NewExperiment(sess, fbmessenger.Config{})
	} else if name == "harconnectivity" {
		experiment = harconnectivity.NewExperiment(sess, harconnectivity.Config{})
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
	} else {
		log.Fatalf("Unknown experiment: %s", name)
	}

	if !globalOptions.noCollector {
		if err := experiment.OpenReport(ctx); err != nil {
			log.WithError(err).Fatal("cannot open report")
		}
		defer experiment.CloseReport(ctx)
	}

	for _, input := range inputs {
		measurement, err := experiment.Measure(ctx, input)
		if err != nil {
			log.WithError(err).Warn("measurement failed")
			continue
		}
		// Remember to omit the user IP.
		//
		// This is not ooniprobe and there's no need here to increase the
		// complexity to make this option configurable. Still, I don't want
		// this tool to be sharing the IP, because I want to provide the
		// same default level of sharing of ooniprobe to random people that
		// may run this tool for development purposes or exploration.
		measurement.ProbeIP = model.DefaultProbeIP
		measurement.AddAnnotations(annotations)
		if !globalOptions.noCollector {
			if err := experiment.SubmitMeasurement(ctx, &measurement); err != nil {
				log.WithError(err).Warn("submitting measurement failed")
				continue
			}
		}
		if !globalOptions.noJSON {
			if err := experiment.SaveMeasurement(
				measurement, globalOptions.reportfile,
			); err != nil {
				log.WithError(err).Warn("saving measurement failed")
				continue
			}
		}
	}
}
