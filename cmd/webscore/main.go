// The webscore tool takes in input a OONI report containing Web Connectivity
// measurements and emits in output the score of such measurement according to
// OONI Probe Engine.
//
// We use this tool to process old measurements and verify what is the score
// that OONI Probe Engine would have applied to such measurement
package main

import (
	"bufio"
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"net/url"
	"os"
	"regexp"

	"github.com/apex/log"
	"github.com/apex/log/handlers/cli"
	"github.com/ooni/probe-engine/experiment/webconnectivity"
	"github.com/ooni/probe-engine/internal/runtimex"
)

// These are the accepted options
var (
	flagFile        = flag.String("file", "", "JSONL file containing the report")
	flagInputFilter = flag.String("input-filter", "", "Pattern for filtering by input")
	flagVerbose     = flag.Bool("v", false, "Run in verbose mode")
)

// Measurement is a WebConnectivity measurement. We do not care much about
// understanding the whole measurement here, rather we only filter out just
// the fields we need to process and score the measurement itself.
type Measurement struct {
	Input    string                   `json:"input"`
	ReportID string                   `json:"report_id"`
	TestKeys webconnectivity.TestKeys `json:"test_keys"`
	TestName string                   `json:"test_name"`
}

// LoadReport loads the report from the JSONL file at filepath. Returns a
// list of measurements on success, an error on failure. Measurements in the
// input file will be skipped if they're not Web Connectivity measurements
// as well as if they have an empty input or report ID. In such cases, this
// function will just emit a warning and continue processing its input.
func LoadReport(filepath string, inputFilter *regexp.Regexp) ([]Measurement, error) {
	var out []Measurement
	filep, err := os.Open(filepath)
	if err != nil {
		return nil, err
	}
	scanner := bufio.NewScanner(filep)
	scanner.Buffer(nil, 1<<24) // we need a large buffer
	for scanner.Scan() {
		var measurement Measurement
		text := []byte(scanner.Text())
		if err := json.Unmarshal(text, &measurement); err != nil {
			log.WithError(err).Warn("cannot parse measurement")
			continue
		}
		if measurement.TestName != "web_connectivity" {
			log.Warn("skipping non Web Connectivity measurement")
			continue
		}
		if measurement.Input == "" {
			log.Warn("skipping measurement with empty input")
			continue
		}
		if measurement.ReportID == "" {
			log.Warn("skipping measurement with empty report ID")
			continue
		}
		if !inputFilter.MatchString(measurement.Input) {
			log.Debug("skipping measurement because input does not match filter")
			continue
		}
		out = append(out, measurement)
		log.Debugf("loaded %s", measurement.ExplorerURL())
	}
	if err := scanner.Err(); err != nil {
		return nil, err
	}
	return out, nil
}

// ExplorerURL constructs the OONI Explorer measurement URL.
func (m Measurement) ExplorerURL() string {
	query := url.Values{}
	query["input"] = []string{m.Input}
	URL := &url.URL{
		Scheme:   "https",
		Host:     "explorer.ooni.org",
		Path:     fmt.Sprintf("/measurement/%s", m.ReportID),
		RawQuery: query.Encode(),
	}
	return URL.String()
}

// Summary returns the Measurement summary according to OONI Probe Engine.
func (m Measurement) Summary() webconnectivity.Summary {
	return webconnectivity.Summarize(&m.TestKeys)
}

// ClassificationTable contains the table for Measurement.Classification.
var ClassificationTable = map[int64]string{
	webconnectivity.StatusSuccessSecure:             "sucess_secure",
	webconnectivity.StatusSuccessCleartext:          "success_cleartext",
	webconnectivity.StatusSuccessNXDOMAIN:           "success_nxdomain",
	webconnectivity.StatusAnomalyControlUnreachable: "anomaly_control_unreachable",
	webconnectivity.StatusAnomalyControlFailure:     "anomaly_control_failure",
	webconnectivity.StatusAnomalyDNS:                "anomaly_dns",
	webconnectivity.StatusAnomalyHTTPDiff:           "anomaly_http_diff",
	webconnectivity.StatusAnomalyConnect:            "anomaly_connect",
	webconnectivity.StatusAnomalyReadWrite:          "anomaly_read_write",
	webconnectivity.StatusAnomalyUnknown:            "anomaly_unknown",
	webconnectivity.StatusAnomalyTLSHandshake:       "anomaly_tls_handshake",
	webconnectivity.StatusExperimentDNS:             "experiment_dns",
	webconnectivity.StatusExperimentConnect:         "experiment_connect",
	webconnectivity.StatusExperimentHTTP:            "experiment_http",
	webconnectivity.StatusBugNoRequests:             "bug_no_requests",
}

// Classification is the classification according to Probe Engine.
func (m Measurement) Classification() (out []string) {
	summary := m.Summary()
	for key, value := range ClassificationTable {
		if (summary.Status & key) != 0 {
			out = append(out, value)
		}
	}
	return
}

// FormatAccessible formats the value of TestKeys.Accessible.
func (m Measurement) FormatAccessible() string {
	accessible := m.TestKeys.Accessible
	if accessible == nil {
		return "nil"
	}
	return fmt.Sprintf("%+v", *accessible)
}

// FormatBlocking formats the value of TestKeys.Blocking.
func (m Measurement) FormatBlocking() string {
	return fmt.Sprintf("%+v", m.TestKeys.Blocking)
}

// Summarize prints the measurement summary on the standard output.
func (m Measurement) Summarize() {
	fmt.Printf("\n")
	fmt.Printf("ReportID   : %s\n", m.ReportID)
	fmt.Printf("Input      : %s\n", m.Input)
	fmt.Printf("Accessible : %+v\n", m.FormatAccessible())
	fmt.Printf("Blocking   : %+v\n", m.FormatBlocking())
	fmt.Printf("Classes    : %+v\n", m.Classification())
	fmt.Printf("\n")
}

// LogLevels maps the verbosity flag to log levels.
var LogLevels = map[bool]log.Level{
	false: log.InfoLevel,
	true:  log.DebugLevel,
}

// Config contains the settings for running this program.
type Config struct {
	Filepath    string
	InputFilter *regexp.Regexp
	Verbose     bool
}

// RunWithFlags processes all the suitable measurements in the input report
// file and outputs their classification according to us.
func RunWithFlags(config Config) error {
	if config.Filepath == "" {
		return errors.New("missing input file")
	}
	log.SetHandler(cli.Default)
	log.SetLevel(LogLevels[config.Verbose])
	measurements, err := LoadReport(config.Filepath, config.InputFilter)
	if err != nil {
		return fmt.Errorf("cannot parse input file: %w", err)
	}
	log.Infof("loaded %d measurements", len(measurements))
	for _, measurement := range measurements {
		measurement.Summarize()
	}
	return nil
}

func main() {
	flag.Parse()
	runtimex.PanicOnError(RunWithFlags(Config{
		Filepath:    *flagFile,
		InputFilter: regexp.MustCompile(*flagInputFilter),
		Verbose:     *flagVerbose,
	}), "RunWithFlags failed")
}
