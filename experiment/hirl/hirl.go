// Package hirl contains the HTTP Invalid Request Line network experiment.
package hirl

import (
	"context"
	"errors"
	"fmt"
	"net"
	"strings"
	"time"

	"github.com/ooni/probe-engine/internal/randx"
	"github.com/ooni/probe-engine/model"
	"github.com/ooni/probe-engine/netx/archival"
	"github.com/ooni/probe-engine/netx/httptransport"
	"github.com/ooni/probe-engine/netx/modelx"
)

const (
	testName    = "http_invalid_request_line"
	testVersion = "0.1.0"
)

// Config contains the experiment config.
type Config struct{}

// TestKeys contains the experiment test keys.
type TestKeys struct {
	FailureList   []*string                   `json:"failure_list"`
	Received      []archival.MaybeBinaryValue `json:"received"`
	Sent          []string                    `json:"sent"`
	TamperingList []bool                      `json:"tampering_list"`
	Tampering     bool                        `json:"tampering"`
}

// NewExperimentMeasurer creates a new ExperimentMeasurer.
func NewExperimentMeasurer(config Config) model.ExperimentMeasurer {
	return Measurer{config: config}
}

// Measurer performs the measurement.
type Measurer struct {
	config Config
}

// ExperimentName implements ExperimentMeasurer.ExperiExperimentName.
func (m Measurer) ExperimentName() string {
	return testName
}

// ExperimentVersion implements ExperimentMeasurer.ExperimentVersion.
func (m Measurer) ExperimentVersion() string {
	return testVersion
}

var (
	// ErrNoAvailableTestHelpers is emitted when there are no available test helpers.
	ErrNoAvailableTestHelpers = errors.New("no available helpers")

	// ErrInvalidHelperType is emitted when the helper type is invalid.
	ErrInvalidHelperType = errors.New("invalid helper type")
)

// Run implements ExperimentMeasurer.Run.
func (m Measurer) Run(
	ctx context.Context, sess model.ExperimentSession,
	measurement *model.Measurement, callbacks model.ExperimentCallbacks,
) error {
	tk := new(TestKeys)
	measurement.TestKeys = tk
	const helperName = "tcp-echo"
	helpers, ok := sess.GetTestHelpersByName(helperName)
	if !ok || len(helpers) < 1 {
		return ErrNoAvailableTestHelpers
	}
	helper := helpers[0]
	if helper.Type != "legacy" {
		return ErrInvalidHelperType
	}
	out := make(chan MethodResult)
	methods := []Method{
		randomInvalidMethod{},
		randomInvalidFieldCount{},
		randomBigRequestMethod{},
		randomInvalidVersionNumber{},
		squidCacheManager{},
	}
	for _, method := range methods {
		callbacks.OnProgress(0.0, fmt.Sprintf("%s...", method.Name()))
		go method.Run(ctx, MethodConfig{
			Address: helper.Address,
			Logger:  sess.Logger(),
			Out:     out,
		})
	}
	var completed int
	for {
		result := <-out
		failure := archival.NewFailure(result.Err)
		tk.FailureList = append(tk.FailureList, failure)
		tk.Received = append(tk.Received, result.Received)
		tk.Sent = append(tk.Sent, result.Sent)
		tk.TamperingList = append(tk.TamperingList, result.Tampering)
		tk.Tampering = (tk.Tampering || result.Tampering)
		completed++
		percentage := float64(completed) / float64(len(methods))
		callbacks.OnProgress(percentage, fmt.Sprintf("%s... %+v", result.Name, result.Err))
		if completed >= len(methods) {
			break
		}
	}
	return nil
}

// MethodConfig contains the settings for a specific measuring method.
type MethodConfig struct {
	Address string
	Logger  model.Logger
	Out     chan<- MethodResult
}

// MethodResult is the result of one of the methods implemented by this experiment.
type MethodResult struct {
	Err       error
	Name      string
	Received  archival.MaybeBinaryValue
	Sent      string
	Tampering bool
}

// Method is one of the methods implemented by this experiment.
type Method interface {
	Name() string
	Run(ctx context.Context, config MethodConfig)
}

type randomInvalidMethod struct{}

func (randomInvalidMethod) Name() string {
	return "random_invalid_method"
}

func (meth randomInvalidMethod) Run(ctx context.Context, config MethodConfig) {
	RunMethod(ctx, RunMethodConfig{
		MethodConfig: config,
		Name:         meth.Name(),
		RequestLine:  randx.LettersUppercase(4) + " / HTTP/1.1\n\r",
	})
}

type randomInvalidFieldCount struct{}

func (randomInvalidFieldCount) Name() string {
	return "random_invalid_field_count"
}

func (meth randomInvalidFieldCount) Run(ctx context.Context, config MethodConfig) {
	RunMethod(ctx, RunMethodConfig{
		MethodConfig: config,
		Name:         meth.Name(),
		RequestLine: strings.Join([]string{
			randx.LettersUppercase(5),
			" ",
			randx.LettersUppercase(5),
			" ",
			randx.LettersUppercase(5),
			" ",
			randx.LettersUppercase(5),
			"\r\n",
		}, ""),
	})
}

type randomBigRequestMethod struct{}

func (randomBigRequestMethod) Name() string {
	return "random_big_request_method"
}

func (meth randomBigRequestMethod) Run(ctx context.Context, config MethodConfig) {
	RunMethod(ctx, RunMethodConfig{
		MethodConfig: config,
		Name:         meth.Name(),
		RequestLine: strings.Join([]string{
			randx.LettersUppercase(1024),
			" / HTTP/1.1\r\n",
		}, ""),
	})
}

type randomInvalidVersionNumber struct{}

func (randomInvalidVersionNumber) Name() string {
	return "random_invalid_version_number"
}

func (meth randomInvalidVersionNumber) Run(ctx context.Context, config MethodConfig) {
	RunMethod(ctx, RunMethodConfig{
		MethodConfig: config,
		Name:         meth.Name(),
		RequestLine: strings.Join([]string{
			"GET / HTTP/",
			randx.LettersUppercase(3),
			"\r\n",
		}, ""),
	})
}

type squidCacheManager struct{}

func (squidCacheManager) Name() string {
	return "squid_cache_manager"
}

func (meth squidCacheManager) Run(ctx context.Context, config MethodConfig) {
	RunMethod(ctx, RunMethodConfig{
		MethodConfig: config,
		Name:         meth.Name(),
		RequestLine:  "GET cache_object://localhost/ HTTP/1.0\n\r",
	})
}

// RunMethodConfig contains the config for RunMethod
type RunMethodConfig struct {
	MethodConfig
	Name        string
	NewDialer   func(config httptransport.Config) httptransport.Dialer
	RequestLine string
}

// RunMethod runs the specific method using the given config and context
func RunMethod(ctx context.Context, config RunMethodConfig) {
	ctx, cancel := context.WithTimeout(ctx, 10*time.Second)
	defer cancel()
	result := MethodResult{Name: config.Name}
	defer func() {
		config.Out <- result
	}()
	if config.NewDialer == nil {
		config.NewDialer = httptransport.NewDialer
	}
	dialer := config.NewDialer(httptransport.Config{
		ContextByteCounting: true,
		Logger:              config.Logger,
	})
	conn, err := dialer.DialContext(ctx, "tcp", net.JoinHostPort(config.Address, "80"))
	if err != nil {
		result.Err = err
		return
	}
	deadline := time.Now().Add(5 * time.Second)
	if err := conn.SetDeadline(deadline); err != nil {
		result.Err = err
		return
	}
	if _, err := conn.Write([]byte(config.RequestLine)); err != nil {
		result.Err = err
		return
	}
	result.Sent = config.RequestLine
	data := make([]byte, 4096)
	defer func() {
		result.Tampering = (result.Sent != result.Received.Value)
	}()
	for {
		count, err := conn.Read(data)
		if err != nil {
			// We expect this method to terminate w/ timeout
			if err.Error() == modelx.FailureGenericTimeoutError {
				err = nil
			}
			result.Err = err
			return
		}
		result.Received.Value += string(data[:count])
	}
}
