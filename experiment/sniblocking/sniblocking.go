// Package sniblocking implements OONI's sni-blocking experiment.
//
// Spec: https://github.com/ooni/spec/blob/master/nettests/ts-024-sni-blocking.md
package sniblocking

import (
	"context"
	"crypto/x509"
	"math/rand"
	"time"

	"github.com/ooni/netx"
	"github.com/ooni/netx/handlers"
	"github.com/ooni/probe-engine/experiment"
	"github.com/ooni/probe-engine/experiment/handler"
	"github.com/ooni/probe-engine/model"
	"github.com/ooni/probe-engine/session"
)

const (
	testName    = "sni_blocking"
	testVersion = "0.1.0"
)

// Config contains the experiment config.
type Config struct{}

// TestKeys contains the experiment test keys
type TestKeys struct {
	Behavior             string  `json:"behavior"`
	FailureWithProperSNI *string `json:"failure_with_proper_sni"`
	FailureWithRandomSNI *string `json:"failure_with_random_sni"`
}

func randomLetters(rnd *rand.Rand, n int) string {
	// See https://stackoverflow.com/a/31832326/4354461
	letters := []rune("abcdefghijklmnopqrstuvwxyz")
	v := make([]rune, n)
	for i := range v {
		v[i] = letters[rnd.Intn(len(letters))]
	}
	return string(v)
}

func randomDomain() string {
	rnd := rand.New(rand.NewSource(time.Now().UnixNano()))
	length := rnd.Intn(8) + 5
	dom := randomLetters(rnd, length)
	dom += "."
	dom += randomLetters(rnd, 3)
	return dom
}

func doWithSNIAndChannel(out chan<- error, input, SNI string) {
	dialer := netx.NewDialer(handlers.StdoutHandler)
	dialer.ForceSpecificSNI(SNI) // empty implies using default
	conn, err := dialer.DialTLS("tcp", input)
	if err != nil {
		out <- err
		return
	}
	conn.Close()
	out <- nil
}

func doWithSNI(input, SNI string) <-chan error {
	ch := make(chan error)
	go doWithSNIAndChannel(ch, input, SNI)
	return ch
}

func doWithContextAndSNI(ctx context.Context, input, SNI string) error {
	select {
	case <-ctx.Done():
		return ctx.Err()
	case err := <-doWithSNI(input, SNI):
		return err
	}
}

func (tk *TestKeys) analyze(properErr, randomErr error) {
	if randomErr == nil {
		tk.Behavior = "interesting"
		return
	}
	_, expected := randomErr.(x509.HostnameError)
	if properErr == nil && expected {
		tk.Behavior = "normal"
		return
	}
	if properErr != nil && expected {
		tk.Behavior = "suspicious"
		return
	}
	tk.Behavior = "🤷"
}

func (tk *TestKeys) fill(properErr, randomErr error) {
	if properErr != nil {
		s := properErr.Error()
		tk.FailureWithProperSNI = &s
	}
	if randomErr != nil {
		s := randomErr.Error()
		tk.FailureWithRandomSNI = &s
	}
}

func measure(
	ctx context.Context, sess *session.Session, measurement *model.Measurement,
	callbacks handler.Callbacks, config Config,
) error {
	ctx, cancel := context.WithTimeout(ctx, 30*time.Second)
	defer cancel()
	tk := &TestKeys{}
	measurement.TestKeys = tk
	properErr := doWithContextAndSNI(ctx, measurement.Input, "")
	randErr := doWithContextAndSNI(ctx, measurement.Input, randomDomain())
	tk.fill(properErr, randErr)
	tk.analyze(properErr, randErr)
	return nil
}

// NewExperiment creates a new experiment.
func NewExperiment(
	sess *session.Session, config Config,
) *experiment.Experiment {
	return experiment.New(
		sess, testName, testVersion,
		func(
			ctx context.Context,
			sess *session.Session,
			measurement *model.Measurement,
			callbacks handler.Callbacks,
		) error {
			return measure(ctx, sess, measurement, callbacks, config)
		})
}
