package engine

import (
	"bufio"
	"context"
	"errors"

	"github.com/ooni/probe-engine/internal/fsx"
	"github.com/ooni/probe-engine/model"
)

// The following errors are returned by the InputLoader.
var (
	ErrNoInputExpected = errors.New("we did not expect any input")
)

// InputLoaderSession is the session according to an InputLoader.
type InputLoaderSession interface {
	MaybeLookupLocationContext(ctx context.Context) error
	NewOrchestraClient(ctx context.Context) (model.ExperimentOrchestraClient, error)
	ProbeCC() string
}

// InputLoader loads input according to the specified policy
// from the specified sources and OONI services. The behaviour
// depends on the input policy as described below.
//
// InputNone
//
// We fail if there is any StaticInput or any SourceFiles. If
// there's no input, we return a single, empty entry that causes
// experiments that don't require input to run once.
//
// InputOptional
//
// We gather input from StaticInput and SourceFiles. If there is
// input, we return it. Otherwise we return a single, empty entry
// that causes experiments that don't require input to run once.
//
// InputRequired
//
// We gather input from StaticInput and SourceFiles. If there is
// input, we return it. Otherwise, we use OONI's probe services
// to gather input using the test lists API.
type InputLoader struct {
	// StaticInputs contains optional input to be added
	// to the resulting input list if possible.
	StaticInputs []string

	// SourceFiles contains optional files to read input
	// from. Each file should contain a single input string
	// per line. We will fail if any file is unreadable.
	SourceFiles []string

	// InputPolicy specifies the input policy for the
	// current experiment. We will not load any input if
	// the policy says we should not.
	InputPolicy InputPolicy

	// Session is the current measurement session.
	Session InputLoaderSession

	// URLLimit is the optional limit on the number of URLs
	// that probe services should return to us.
	URLLimit int64

	// URLCategories limits the categories of URLs that
	// probe services should return to us.
	URLCategories []string
}

// Load attempts to load input using the specified input loader. We will
// return a list of URLs because this is the only input we support.
func (il InputLoader) Load(ctx context.Context) ([]model.URLInfo, error) {
	switch il.InputPolicy {
	case InputOptional:
		return il.loadOptional()
	case InputRequired:
		return il.loadRequired(ctx)
	default:
		return il.loadNone()
	}
}

func (il InputLoader) loadNone() ([]model.URLInfo, error) {
	if len(il.StaticInputs) > 0 || len(il.SourceFiles) > 0 {
		return nil, ErrNoInputExpected
	}
	return []model.URLInfo{{}}, nil
}

func (il InputLoader) loadOptional() ([]model.URLInfo, error) {
	inputs, err := il.loadLocal()
	if err == nil && len(inputs) <= 0 {
		inputs = []model.URLInfo{{}}
	}
	return inputs, err
}

func (il InputLoader) loadRequired(ctx context.Context) ([]model.URLInfo, error) {
	inputs, err := il.loadLocal()
	if err != nil || len(inputs) > 0 {
		return inputs, err
	}
	return il.loadRemote(loadRemoteConfig{ctx: ctx, session: il.Session})
}

func (il InputLoader) loadLocal() ([]model.URLInfo, error) {
	inputs := []model.URLInfo{}
	for _, input := range il.StaticInputs {
		inputs = append(inputs, model.URLInfo{URL: input})
	}
	for _, filepath := range il.SourceFiles {
		extra, err := il.readfile(filepath, fsx.Open)
		if err != nil {
			return nil, err
		}
		inputs = append(inputs, extra...)
	}
	return inputs, nil
}

func (il InputLoader) readfile(filepath string, open func(string) (fsx.File, error)) ([]model.URLInfo, error) {
	inputs := []model.URLInfo{}
	filep, err := open(filepath)
	if err != nil {
		return nil, err
	}
	defer filep.Close()
	// Implementation note: when you save file with vim, you have newline at
	// end of file and you don't want to consider that an input line. While there
	// ignore any other empty line that may occur inside the file.
	scanner := bufio.NewScanner(filep)
	for scanner.Scan() {
		line := scanner.Text()
		if line != "" {
			inputs = append(inputs, model.URLInfo{URL: line})
		}
	}
	if scanner.Err() != nil {
		return nil, scanner.Err()
	}
	return inputs, nil
}

type loadRemoteConfig struct {
	ctx     context.Context
	session InputLoaderSession
}

func (il InputLoader) loadRemote(conf loadRemoteConfig) ([]model.URLInfo, error) {
	if err := conf.session.MaybeLookupLocationContext(conf.ctx); err != nil {
		return nil, err
	}
	client, err := conf.session.NewOrchestraClient(conf.ctx)
	if err != nil {
		return nil, err
	}
	list, err := client.FetchURLList(conf.ctx, model.URLListConfig{
		CountryCode: conf.session.ProbeCC(),
		Limit:       il.URLLimit,
		Categories:  il.URLCategories,
	})
	if err != nil {
		return nil, err
	}
	inputs := []model.URLInfo{}
	for _, entry := range list {
		inputs = append(inputs, entry)
	}
	return inputs, nil
}
