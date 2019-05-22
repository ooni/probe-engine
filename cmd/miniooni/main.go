// Command miniooni is simple binary for testing purposes. We try
// to mirror the command line arguments of the original measurement_kit
// binary, to support using miniooni instead of it.
package main

import (
	"context"
	"errors"
	"fmt"
	"io"
	"os"
	"strings"

	"github.com/apex/log"
	"github.com/apex/log/handlers/multi"

	"github.com/ooni/probe-engine/experiment"
	"github.com/ooni/probe-engine/experiment/dash"
	"github.com/ooni/probe-engine/experiment/fbmessenger"
	"github.com/ooni/probe-engine/experiment/hhfm"
	"github.com/ooni/probe-engine/experiment/hirl"
	"github.com/ooni/probe-engine/experiment/ndt"
	"github.com/ooni/probe-engine/experiment/ndt7"
	"github.com/ooni/probe-engine/experiment/psiphon"
	"github.com/ooni/probe-engine/experiment/telegram"
	"github.com/ooni/probe-engine/experiment/web_connectivity"
	"github.com/ooni/probe-engine/experiment/whatsapp"
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
	logfile      string
	noBouncer    bool
	noGeoIP      bool
	noJSON       bool
	noCollector  bool
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
	extraOptions := make(map[string]string)
	for _, opt := range globalOptions.extraOptions {
		key, value, err := split(opt)
		if err != nil {
			log.WithError(err).Fatal("cannot split key-value pair")
		}
		extraOptions[key] = value
	}

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
	sess := session.New(logger, softwareName, softwareVersion)
	sess.WorkDir = workDir
	if !globalOptions.noBouncer {
		if err := sess.LookupBackends(ctx); err != nil {
			log.WithError(err).Fatal("cannot lookup OONI backends")
		}
	}
	if !globalOptions.noGeoIP {
		if err := sess.LookupLocation(ctx); err != nil {
			log.WithError(err).Warn("cannot lookup your location")
		}
	}

	name := getopt.Args()[0]
	var inputs []string

	if name == "web_connectivity" {
		list, err := (&testlists.Client{
			BaseURL:    testlists.DefaultBaseURL,
			HTTPClient: sess.HTTPDefaultClient,
			Logger:     logger,
			UserAgent:  sess.UserAgent(),
		}).Do(ctx, sess.ProbeCC())
		if err != nil {
			log.WithError(err).Fatal("cannot fetch test lists")
		}
		for _, entry := range list {
			inputs = append(inputs, entry.URL)
		}
	} else {
		inputs = append(inputs, "")
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
		if err := experiment.SubmitMeasurement(ctx, &measurement); err != nil {
			log.WithError(err).Warn("submitting measurement failed")
			continue
		}
		if err := experiment.SaveMeasurement(
			measurement, globalOptions.reportfile,
		); err != nil {
			log.WithError(err).Warn("saving measurement failed")
			continue
		}
	}
}
