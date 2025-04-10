package registry

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"math"
	"os"
	"reflect"
	"testing"

	"github.com/apex/log"
	"github.com/google/go-cmp/cmp"
	"github.com/ooni/probe-engine/pkg/checkincache"
	"github.com/ooni/probe-engine/pkg/experiment/webconnectivitylte"
	"github.com/ooni/probe-engine/pkg/experimentname"
	"github.com/ooni/probe-engine/pkg/kvstore"
	"github.com/ooni/probe-engine/pkg/mocks"
	"github.com/ooni/probe-engine/pkg/model"
	"github.com/ooni/probe-engine/pkg/targetloading"
)

func TestFactoryOptions(t *testing.T) {

	// the fake configuration we're using in this test
	type fakeExperimentConfig struct {
		// values that should be included into the Options return value
		Chan   chan any `ooni:"we cannot set this"`
		String string   `ooni:"a string"`
		Truth  bool     `ooni:"something that no-one knows"`
		Value  int64    `ooni:"a number"`

		// values that should not be included because they're private
		private int64 `ooni:"a private number"`

		// values that should not be included because they lack "ooni"'s tag
		Invisible int64
	}

	t.Run("when config is not a pointer", func(t *testing.T) {
		b := &Factory{
			config: 17,
		}
		options, err := b.Options()
		if !errors.Is(err, ErrConfigIsNotAStructPointer) {
			t.Fatal("expected an error here")
		}
		if options != nil {
			t.Fatal("expected nil here")
		}
	})

	t.Run("when config is not a struct", func(t *testing.T) {
		number := 17
		b := &Factory{
			config: &number,
		}
		options, err := b.Options()
		if !errors.Is(err, ErrConfigIsNotAStructPointer) {
			t.Fatal("expected an error here")
		}
		if options != nil {
			t.Fatal("expected nil here")
		}
	})

	t.Run("when config is a pointer to struct", func(t *testing.T) {
		config := &fakeExperimentConfig{
			Chan:      make(chan any),
			String:    "foobar",
			Truth:     true,
			Value:     177114,
			private:   55,
			Invisible: 9876,
		}
		b := &Factory{
			config: config,
		}
		options, err := b.Options()
		if err != nil {
			t.Fatal(err)
		}

		for name, value := range options {
			switch name {
			case "Chan":
				if value.Doc != "we cannot set this" {
					t.Fatal("invalid doc")
				}
				if value.Type != "chan interface {}" {
					t.Fatal("invalid type", value.Type)
				}
				if value.Value.(chan any) == nil {
					t.Fatal("expected non-nil channel here")
				}

			case "String":
				if value.Doc != "a string" {
					t.Fatal("invalid doc")
				}
				if value.Type != "string" {
					t.Fatal("invalid type", value.Type)
				}
				if v := value.Value.(string); v != "foobar" {
					t.Fatal("unexpected string value", v)
				}

			case "Truth":
				if value.Doc != "something that no-one knows" {
					t.Fatal("invalid doc")
				}
				if value.Type != "bool" {
					t.Fatal("invalid type", value.Type)
				}
				if v := value.Value.(bool); !v {
					t.Fatal("unexpected bool value", v)
				}

			case "Value":
				if value.Doc != "a number" {
					t.Fatal("invalid doc")
				}
				if value.Type != "int64" {
					t.Fatal("invalid type", value.Type)
				}
				if v := value.Value.(int64); v != 177114 {
					t.Fatal("unexpected int64 value", v)
				}

			default:
				t.Fatal("unexpected option name", name)
			}
		}
	})
}

func TestFactorySetOptionAny(t *testing.T) {

	// the fake configuration we're using in this test
	type fakeExperimentConfig struct {
		Chan   chan any `ooni:"we cannot set this"`
		String string   `ooni:"a string"`
		Truth  bool     `ooni:"something that no-one knows"`
		Value  int64    `ooni:"a number"`
	}

	var inputs = []struct {
		TestCaseName  string
		InitialConfig any
		FieldName     string
		FieldValue    any
		ExpectErr     error
		ExpectConfig  any
	}{{
		TestCaseName:  "config is not a pointer",
		InitialConfig: fakeExperimentConfig{},
		FieldName:     "Antani",
		FieldValue:    true,
		ExpectErr:     ErrConfigIsNotAStructPointer,
		ExpectConfig:  fakeExperimentConfig{},
	}, {
		TestCaseName: "config is not a pointer to struct",
		InitialConfig: func() *int {
			v := 17
			return &v
		}(),
		FieldName:  "Antani",
		FieldValue: true,
		ExpectErr:  ErrConfigIsNotAStructPointer,
		ExpectConfig: func() *int {
			v := 17
			return &v
		}(),
	}, {
		TestCaseName:  "for missing field",
		InitialConfig: &fakeExperimentConfig{},
		FieldName:     "Antani",
		FieldValue:    true,
		ExpectErr:     ErrNoSuchField,
		ExpectConfig:  &fakeExperimentConfig{},
	}, {
		TestCaseName:  "[bool] for true value represented as string",
		InitialConfig: &fakeExperimentConfig{},
		FieldName:     "Truth",
		FieldValue:    "true",
		ExpectErr:     nil,
		ExpectConfig: &fakeExperimentConfig{
			Truth: true,
		},
	}, {
		TestCaseName: "[bool] for false value represented as string",
		InitialConfig: &fakeExperimentConfig{
			Truth: true,
		},
		FieldName:  "Truth",
		FieldValue: "false",
		ExpectErr:  nil,
		ExpectConfig: &fakeExperimentConfig{
			Truth: false, // must have been flipped
		},
	}, {
		TestCaseName:  "[bool] for true value",
		InitialConfig: &fakeExperimentConfig{},
		FieldName:     "Truth",
		FieldValue:    true,
		ExpectErr:     nil,
		ExpectConfig: &fakeExperimentConfig{
			Truth: true,
		},
	}, {
		TestCaseName: "[bool] for false value",
		InitialConfig: &fakeExperimentConfig{
			Truth: true,
		},
		FieldName:  "Truth",
		FieldValue: false,
		ExpectErr:  nil,
		ExpectConfig: &fakeExperimentConfig{
			Truth: false, // must have been flipped
		},
	}, {
		TestCaseName:  "[bool] for invalid string representation of bool",
		InitialConfig: &fakeExperimentConfig{},
		FieldName:     "Truth",
		FieldValue:    "xxx",
		ExpectErr:     ErrInvalidStringRepresentationOfBool,
		ExpectConfig:  &fakeExperimentConfig{},
	}, {
		TestCaseName:  "[bool] for value we don't know how to convert to bool",
		InitialConfig: &fakeExperimentConfig{},
		FieldName:     "Truth",
		FieldValue:    make(chan any),
		ExpectErr:     ErrCannotSetBoolOption,
		ExpectConfig:  &fakeExperimentConfig{},
	}, {
		TestCaseName:  "[int] for int",
		InitialConfig: &fakeExperimentConfig{},
		FieldName:     "Value",
		FieldValue:    17,
		ExpectErr:     nil,
		ExpectConfig: &fakeExperimentConfig{
			Value: 17,
		},
	}, {
		TestCaseName:  "[int] for int64",
		InitialConfig: &fakeExperimentConfig{},
		FieldName:     "Value",
		FieldValue:    int64(17),
		ExpectErr:     nil,
		ExpectConfig: &fakeExperimentConfig{
			Value: 17,
		},
	}, {
		TestCaseName:  "[int] for int32",
		InitialConfig: &fakeExperimentConfig{},
		FieldName:     "Value",
		FieldValue:    int32(17),
		ExpectErr:     nil,
		ExpectConfig: &fakeExperimentConfig{
			Value: 17,
		},
	}, {
		TestCaseName:  "[int] for int16",
		InitialConfig: &fakeExperimentConfig{},
		FieldName:     "Value",
		FieldValue:    int16(17),
		ExpectErr:     nil,
		ExpectConfig: &fakeExperimentConfig{
			Value: 17,
		},
	}, {
		TestCaseName:  "[int] for int8",
		InitialConfig: &fakeExperimentConfig{},
		FieldName:     "Value",
		FieldValue:    int8(17),
		ExpectErr:     nil,
		ExpectConfig: &fakeExperimentConfig{
			Value: 17,
		},
	}, {
		TestCaseName:  "[int] for string representation of int",
		InitialConfig: &fakeExperimentConfig{},
		FieldName:     "Value",
		FieldValue:    "17",
		ExpectErr:     nil,
		ExpectConfig: &fakeExperimentConfig{
			Value: 17,
		},
	}, {
		TestCaseName:  "[int] for invalid string representation of int",
		InitialConfig: &fakeExperimentConfig{},
		FieldName:     "Value",
		FieldValue:    "xx",
		ExpectErr:     ErrCannotSetIntegerOption,
		ExpectConfig:  &fakeExperimentConfig{},
	}, {
		TestCaseName:  "[int] for type we don't know how to convert to int",
		InitialConfig: &fakeExperimentConfig{},
		FieldName:     "Value",
		FieldValue:    make(chan any),
		ExpectErr:     ErrCannotSetIntegerOption,
		ExpectConfig:  &fakeExperimentConfig{},
	}, {
		TestCaseName:  "[int] for NaN",
		InitialConfig: &fakeExperimentConfig{},
		FieldName:     "Value",
		FieldValue:    math.NaN(),
		ExpectErr:     ErrCannotSetIntegerOption,
		ExpectConfig:  &fakeExperimentConfig{},
	}, {
		TestCaseName:  "[int] for +Inf",
		InitialConfig: &fakeExperimentConfig{},
		FieldName:     "Value",
		FieldValue:    math.Inf(1),
		ExpectErr:     ErrCannotSetIntegerOption,
		ExpectConfig:  &fakeExperimentConfig{},
	}, {
		TestCaseName:  "[int] for -Inf",
		InitialConfig: &fakeExperimentConfig{},
		FieldName:     "Value",
		FieldValue:    math.Inf(-1),
		ExpectErr:     ErrCannotSetIntegerOption,
		ExpectConfig:  &fakeExperimentConfig{},
	}, {
		TestCaseName:  "[int] for too large value",
		InitialConfig: &fakeExperimentConfig{},
		FieldName:     "Value",
		FieldValue:    float64(jsonMaxInteger + 1),
		ExpectErr:     ErrCannotSetIntegerOption,
		ExpectConfig:  &fakeExperimentConfig{},
	}, {
		TestCaseName:  "[int] for too small value",
		InitialConfig: &fakeExperimentConfig{},
		FieldName:     "Value",
		FieldValue:    float64(jsonMinInteger - 1),
		ExpectErr:     ErrCannotSetIntegerOption,
		ExpectConfig:  &fakeExperimentConfig{},
	}, {
		TestCaseName:  "[int] for float64 with nonzero fractional value",
		InitialConfig: &fakeExperimentConfig{},
		FieldName:     "Value",
		FieldValue:    1.11,
		ExpectErr:     ErrCannotSetIntegerOption,
		ExpectConfig:  &fakeExperimentConfig{},
	}, {
		TestCaseName:  "[int] for float64 with zero fractional value",
		InitialConfig: &fakeExperimentConfig{},
		FieldName:     "Value",
		FieldValue:    float64(16.0),
		ExpectErr:     nil,
		ExpectConfig: &fakeExperimentConfig{
			Value: 16,
		},
	}, {
		TestCaseName:  "[string] for serialized bool value while setting a string value",
		InitialConfig: &fakeExperimentConfig{},
		FieldName:     "String",
		FieldValue:    "true",
		ExpectErr:     nil,
		ExpectConfig: &fakeExperimentConfig{
			String: "true",
		},
	}, {
		TestCaseName:  "[string] for serialized int value while setting a string value",
		InitialConfig: &fakeExperimentConfig{},
		FieldName:     "String",
		FieldValue:    "155",
		ExpectErr:     nil,
		ExpectConfig: &fakeExperimentConfig{
			String: "155",
		},
	}, {
		TestCaseName:  "[string] for any other string",
		InitialConfig: &fakeExperimentConfig{},
		FieldName:     "String",
		FieldValue:    "xxx",
		ExpectErr:     nil,
		ExpectConfig: &fakeExperimentConfig{
			String: "xxx",
		},
	}, {
		TestCaseName:  "[string] for type we don't know how to convert to string",
		InitialConfig: &fakeExperimentConfig{},
		FieldName:     "String",
		FieldValue:    make(chan any),
		ExpectErr:     ErrCannotSetStringOption,
		ExpectConfig:  &fakeExperimentConfig{},
	}, {
		TestCaseName:  "for a field that we don't know how to set",
		InitialConfig: &fakeExperimentConfig{},
		FieldName:     "Chan",
		FieldValue:    make(chan any),
		ExpectErr:     ErrUnsupportedOptionType,
		ExpectConfig:  &fakeExperimentConfig{},
	}}

	for _, input := range inputs {
		t.Run(input.TestCaseName, func(t *testing.T) {
			ec := input.InitialConfig
			b := &Factory{config: ec}
			err := b.SetOptionAny(input.FieldName, input.FieldValue)
			if !errors.Is(err, input.ExpectErr) {
				t.Fatal(err)
			}
			if diff := cmp.Diff(input.ExpectConfig, ec); diff != "" {
				t.Fatal(diff)
			}
		})
	}
}

func TestFactorySetOptionsAny(t *testing.T) {

	// the fake configuration we're using in this test
	type fakeExperimentConfig struct {
		// values that should be included into the Options return value
		Chan   chan any `ooni:"we cannot set this"`
		String string   `ooni:"a string"`
		Truth  bool     `ooni:"something that no-one knows"`
		Value  int64    `ooni:"a number"`
	}

	b := &Factory{config: &fakeExperimentConfig{}}

	t.Run("we correctly handle an empty map", func(t *testing.T) {
		if err := b.SetOptionsAny(nil); err != nil {
			t.Fatal(err)
		}
	})

	t.Run("we correctly handle a map containing options", func(t *testing.T) {
		f := &fakeExperimentConfig{}
		privateb := &Factory{config: f}
		opts := map[string]any{
			"String": "yoloyolo",
			"Value":  "174",
			"Truth":  "true",
		}
		if err := privateb.SetOptionsAny(opts); err != nil {
			t.Fatal(err)
		}
		if f.String != "yoloyolo" {
			t.Fatal("cannot set string value")
		}
		if f.Value != 174 {
			t.Fatal("cannot set integer value")
		}
		if f.Truth != true {
			t.Fatal("cannot set bool value")
		}
	})

	t.Run("we handle mistakes in a map containing string options", func(t *testing.T) {
		opts := map[string]any{
			"String": "yoloyolo",
			"Value":  "xx",
			"Truth":  "true",
		}
		if err := b.SetOptionsAny(opts); !errors.Is(err, ErrCannotSetIntegerOption) {
			t.Fatal("unexpected err", err)
		}
	})
}

func TestFactorySetOptionsJSON(t *testing.T) {

	// PersonRecord is a fake experiment configuration.
	//
	// Note how the `ooni` tag here is missing because we don't care
	// about whether such a tag is present when using JSON.
	type PersonRecord struct {
		Name    string
		Age     int64
		Friends []string
	}

	// testcase is a test case for this function.
	type testcase struct {
		// name is the name of the test case
		name string

		// mutableConfig is the config in which we should unmarshal the JSON
		mutableConfig *PersonRecord

		// rawJSON contains the raw JSON to unmarshal into mutableConfig
		rawJSON json.RawMessage

		// expectErr is the error we expect
		expectErr error

		// expectRecord is what we expectRecord to see in the end
		expectRecord *PersonRecord
	}

	cases := []testcase{
		{
			name: "we correctly accept zero-length options",
			mutableConfig: &PersonRecord{
				Name:    "foo",
				Age:     55,
				Friends: []string{"bar", "baz"},
			},
			rawJSON: []byte{},
			expectRecord: &PersonRecord{
				Name:    "foo",
				Age:     55,
				Friends: []string{"bar", "baz"},
			},
		},

		{
			name:          "we return an error on JSON parsing error",
			mutableConfig: &PersonRecord{},
			rawJSON:       []byte(`{`),
			expectErr:     errors.New("unexpected end of JSON input"),
			expectRecord:  &PersonRecord{},
		},

		{
			name: "we correctly unmarshal into the existing config",
			mutableConfig: &PersonRecord{
				Name:    "foo",
				Age:     55,
				Friends: []string{"bar", "baz"},
			},
			rawJSON:   []byte(`{"Friends":["foo","oof"]}`),
			expectErr: nil,
			expectRecord: &PersonRecord{
				Name:    "foo",
				Age:     55,
				Friends: []string{"foo", "oof"},
			},
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {

			// create the factory to use
			factory := &Factory{config: tc.mutableConfig}

			// unmarshal into the mutableConfig
			err := factory.SetOptionsJSON(tc.rawJSON)

			// make sure the error is the one we actually expect
			switch {
			case err == nil && tc.expectErr == nil:
				if diff := cmp.Diff(tc.expectRecord, tc.mutableConfig); diff != "" {
					t.Fatal(diff)
				}
				return

			case err != nil && tc.expectErr != nil:
				if err.Error() != tc.expectErr.Error() {
					t.Fatal("expected", tc.expectErr, "got", err)
				}
				return

			default:
				t.Fatal("expected", tc.expectErr, "got", err)
			}
		})
	}
}

func TestNewFactory(t *testing.T) {
	// experimentSpecificExpectations contains expectations for an experiment
	type experimentSpecificExpectations struct {
		// enabledByDefault contains the expected value for the enabledByDefault factory field.
		enabledByDefault bool

		// inputPolicy contains the expected value for the input policy.
		inputPolicy model.InputPolicy

		// interruptible contains the expected value for interrupted.
		interruptible bool
	}

	// expectationsMap contains expectations for each experiment that exists
	expectationsMap := map[string]*experimentSpecificExpectations{
		"dash": {
			enabledByDefault: true,
			inputPolicy:      model.InputNone,
			interruptible:    true,
		},
		"dnscheck": {
			enabledByDefault: true,
			inputPolicy:      model.InputOrStaticDefault,
		},
		"dnsping": {
			enabledByDefault: true,
			inputPolicy:      model.InputOrStaticDefault,
		},
		"echcheck": {
			enabledByDefault: true,
			inputPolicy:      model.InputOptional,
		},
		"example": {
			enabledByDefault: true,
			inputPolicy:      model.InputNone,
			interruptible:    true,
		},
		"facebook_messenger": {
			enabledByDefault: true,
			inputPolicy:      model.InputNone,
		},
		"http_header_field_manipulation": {
			enabledByDefault: true,
			inputPolicy:      model.InputNone,
		},
		"http_host_header": {
			enabledByDefault: true,
			inputPolicy:      model.InputOrQueryBackend,
		},
		"http_invalid_request_line": {
			enabledByDefault: true,
			inputPolicy:      model.InputNone,
		},
		"ndt": {
			enabledByDefault: true,
			inputPolicy:      model.InputNone,
			interruptible:    true,
		},
		"openvpn": {
			enabledByDefault: true,
			inputPolicy:      model.InputOrQueryBackend,
			interruptible:    true,
		},
		"portfiltering": {
			enabledByDefault: true,
			inputPolicy:      model.InputNone,
		},
		"psiphon": {
			enabledByDefault: true,
			inputPolicy:      model.InputOptional,
		},
		"quicping": {
			enabledByDefault: true,
			inputPolicy:      model.InputStrictlyRequired,
		},
		"riseupvpn": {
			// Note: riseupvpn is not enabled by default because it has been flaky
			// in the past and we want to be defensive here.
			//enabledByDefault: false,
			inputPolicy: model.InputNone,
		},
		"run": {
			enabledByDefault: true,
			inputPolicy:      model.InputStrictlyRequired,
		},
		"signal": {
			enabledByDefault: true,
			inputPolicy:      model.InputNone,
		},
		"simple_sni": {
			// Note: simple_sni is not enabled by default because it has only been
			// introduced for writing tutorials and should not be used.
			//enabledByDefault: false,
			inputPolicy: model.InputOrQueryBackend,
		},
		"simplequicping": {
			enabledByDefault: true,
			inputPolicy:      model.InputStrictlyRequired,
		},
		"sni_blocking": {
			enabledByDefault: true,
			inputPolicy:      model.InputOrQueryBackend,
		},
		"stunreachability": {
			enabledByDefault: true,
			inputPolicy:      model.InputOrStaticDefault,
		},
		"tcpping": {
			enabledByDefault: true,
			inputPolicy:      model.InputStrictlyRequired,
		},
		"tlsmiddlebox": {
			enabledByDefault: true,
			inputPolicy:      model.InputStrictlyRequired,
		},
		"telegram": {
			enabledByDefault: true,
			inputPolicy:      model.InputNone,
		},
		"tlsping": {
			enabledByDefault: true,
			inputPolicy:      model.InputStrictlyRequired,
		},
		"tlstool": {
			enabledByDefault: true,
			inputPolicy:      model.InputOrQueryBackend,
		},
		"tor": {
			enabledByDefault: true,
			inputPolicy:      model.InputNone,
		},
		"torsf": {
			// We suspect there will be changes in torsf SNI soon. We are not prepared to
			// serve these changes using the check-in API. Hence, disable torsf by default
			// and require enabling it using the check-in API feature flags.
			//enabledByDefault: false,
			inputPolicy: model.InputNone,
		},
		"urlgetter": {
			enabledByDefault: true,
			inputPolicy:      model.InputStrictlyRequired,
		},
		"vanilla_tor": {
			// The experiment crashes on Android and possibly also iOS. We want to
			// control whether and when to run it using check-in.
			//enabledByDefault: false,
			inputPolicy: model.InputNone,
		},
		"web_connectivity": {
			enabledByDefault: true,
			inputPolicy:      model.InputOrQueryBackend,
		},
		"web_connectivity@v0.5": {
			enabledByDefault: true,
			inputPolicy:      model.InputOrQueryBackend,
		},
		"whatsapp": {
			enabledByDefault: true,
			inputPolicy:      model.InputNone,
		},
	}

	// testCase is a test case checked by this func
	type testCase struct {
		// description describes the test case
		description string

		// experimentName is the experiment experimentName
		experimentName string

		// kvStore is the key-value store to use
		kvStore model.KeyValueStore

		// setForceEnableExperiment sets the OONI_FORCE_ENABLE_EXPERIMENT=1 env variable
		setForceEnableExperiment bool

		// expectErr is the error we expect when calling NewFactory
		expectErr error
	}

	// allCases contains all test cases
	allCases := []*testCase{}

	// create test cases for canonical experiment names
	for _, name := range ExperimentNames() {
		allCases = append(allCases, &testCase{
			description:    name,
			experimentName: name,
			kvStore:        &kvstore.Memory{},
			expectErr: (func() error {
				expectations := expectationsMap[name]
				if expectations == nil {
					t.Fatal("no expectations for", name)
				}
				if !expectations.enabledByDefault {
					return ErrRequiresForceEnable
				}
				return nil
			}()),
		})
	}

	// add additional test for the ndt7 experiment name
	allCases = append(allCases, &testCase{
		description:    "the ndt7 name still works",
		experimentName: "ndt7",
		kvStore:        &kvstore.Memory{},
		expectErr:      nil,
	})

	// add additional test for the dns_check experiment name
	allCases = append(allCases, &testCase{
		description:    "the dns_check name still works",
		experimentName: "dns_check",
		kvStore:        &kvstore.Memory{},
		expectErr:      nil,
	})

	// add additional test for the stun_reachability experiment name
	allCases = append(allCases, &testCase{
		description:    "the stun_reachability name still works",
		experimentName: "stun_reachability",
		kvStore:        &kvstore.Memory{},
		expectErr:      nil,
	})

	// add additional test for the web_connectivity@v_0_5 experiment name
	allCases = append(allCases, &testCase{
		description:    "the web_connectivity@v_0_5 name still works",
		experimentName: "web_connectivity@v_0_5",
		kvStore:        &kvstore.Memory{},
		expectErr:      nil,
	})

	// make sure we can create default-not-enabled experiments if we
	// configure the proper environment variable
	for name, expectations := range expectationsMap {
		if expectations.enabledByDefault {
			continue
		}

		allCases = append(allCases, &testCase{
			description:              fmt.Sprintf("we can create %s with OONI_FORCE_ENABLE_EXPERIMENT=1", name),
			experimentName:           name,
			kvStore:                  &kvstore.Memory{},
			setForceEnableExperiment: true,
			expectErr:                nil,
		})
	}

	// make sure we can create default-not-enabled experiments if we
	// configure the proper check-in flags
	for name, expectations := range expectationsMap {
		if expectations.enabledByDefault {
			continue
		}

		// create a check-in configuration with the experiment being enabled
		store := &kvstore.Memory{}
		checkincache.Store(store, &model.OOAPICheckInResult{
			Conf: model.OOAPICheckInResultConfig{
				Features: map[string]bool{
					checkincache.ExperimentEnabledKey(name): true,
				},
			},
		})

		allCases = append(allCases, &testCase{
			description:              fmt.Sprintf("we can create %s with the proper check-in config", name),
			experimentName:           name,
			kvStore:                  store,
			setForceEnableExperiment: false,
			expectErr:                nil,
		})
	}

	// perform checks for each name
	for _, tc := range allCases {
		t.Run(tc.description, func(t *testing.T) {
			// make sure the bypass environment variable is not set
			if os.Getenv(OONI_FORCE_ENABLE_EXPERIMENT) != "" {
				t.Fatal("the OONI_FORCE_ENABLE_EXPERIMENT env variable shouldn't be set")
			}

			// if needed, set the environment variable for the scope of the func
			if tc.setForceEnableExperiment {
				os.Setenv(OONI_FORCE_ENABLE_EXPERIMENT, "1")
				defer os.Unsetenv(OONI_FORCE_ENABLE_EXPERIMENT)
			}

			t.Log("experimentName:", tc.experimentName)

			// get experiment expectations -- note that here we must canonicalize the
			// experiment name otherwise we won't find it into the map when testing non-canonical names
			expectations := expectationsMap[experimentname.Canonicalize(tc.experimentName)]
			if expectations == nil {
				t.Fatal("no expectations for", tc.experimentName)
			}

			t.Logf("expectations: %+v", expectations)

			// get the experiment factory
			factory, err := NewFactory(tc.experimentName, tc.kvStore, model.DiscardLogger)

			t.Logf("NewFactory returned: %+v %+v", factory, err)

			// make sure the returned error makes sense
			switch {
			case tc.expectErr == nil && err != nil:
				t.Fatal(tc.experimentName, ": expected", tc.expectErr, "got", err)

			case tc.expectErr != nil && err == nil:
				t.Fatal(tc.experimentName, ": expected", tc.expectErr, "got", err)

			case tc.expectErr != nil && err != nil:
				if !errors.Is(err, tc.expectErr) {
					t.Fatal(tc.experimentName, ": expected", tc.expectErr, "got", err)
				}
				return

			case tc.expectErr == nil && err == nil:
				// fallthrough
			}

			// make sure the enabled by default field is consistent with expectations
			if factory.enabledByDefault != expectations.enabledByDefault {
				t.Fatal(tc.experimentName, ": expected", expectations.enabledByDefault, "got", factory.enabledByDefault)
			}

			// make sure the input policy is the expected one
			if v := factory.InputPolicy(); v != expectations.inputPolicy {
				t.Fatal(tc.experimentName, ": expected", expectations.inputPolicy, "got", v)
			}

			// make sure the interruptible value is the expected one
			if v := factory.Interruptible(); v != expectations.interruptible {
				t.Fatal(tc.experimentName, ": expected", expectations.interruptible, "got", v)
			}

			// make sure we can create the measurer
			measurer := factory.NewExperimentMeasurer()
			if measurer == nil {
				t.Fatal("expected non-nil measurer, got nil")
			}
		})
	}

	// make sure we create web_connectivity@v0.5 when the check-in says so
	t.Run("we honor check-in flags for web_connectivity@v0.5", func(t *testing.T) {
		// create a keyvalue store with the proper flags
		store := &kvstore.Memory{}
		checkincache.Store(store, &model.OOAPICheckInResult{
			Conf: model.OOAPICheckInResultConfig{
				Features: map[string]bool{
					"webconnectivity_0.5": true,
				},
			},
		})

		// get the experiment factory
		factory, err := NewFactory("web_connectivity", store, model.DiscardLogger)
		if err != nil {
			t.Fatal(err)
		}

		// make sure the enabled by default field is consistent with expectations
		if !factory.enabledByDefault {
			t.Fatal("expected enabledByDefault to be true")
		}

		// make sure the input policy is the expected one
		if factory.InputPolicy() != model.InputOrQueryBackend {
			t.Fatal("expected inputPolicy to be InputOrQueryBackend")
		}

		// make sure the interrupted value is the expected one
		if factory.Interruptible() {
			t.Fatal("expected interruptible to be false")
		}

		// make sure we can create the measurer
		measurer := factory.NewExperimentMeasurer()
		if measurer == nil {
			t.Fatal("expected non-nil measurer, got nil")
		}

		// make sure the type we're creating is the correct one
		if _, good := measurer.(*webconnectivitylte.Measurer); !good {
			t.Fatalf("expected to see an instance of *webconnectivitylte.Measurer, got %T", measurer)
		}
	})

	// add a test case for a nonexistent experiment
	t.Run("we correctly return an error for a nonexistent experiment", func(t *testing.T) {
		// the empty string is a nonexistent experiment
		factory, err := NewFactory("", &kvstore.Memory{}, model.DiscardLogger)
		if !errors.Is(err, ErrNoSuchExperiment) {
			t.Fatal("unexpected err", err)
		}
		if factory != nil {
			t.Fatal("expected nil factory here")
		}
	})
}

// Make sure the target loader for web connectivity is WAI when using no static inputs.
func TestFactoryNewTargetLoaderWebConnectivity(t *testing.T) {
	// construct the proper factory instance
	store := &kvstore.Memory{}
	factory, err := NewFactory("web_connectivity", store, log.Log)
	if err != nil {
		t.Fatal(err)
	}

	// define the expected error.
	expected := errors.New("antani")

	// create suitable loader config.
	config := &model.ExperimentTargetLoaderConfig{
		CheckInConfig: &model.OOAPICheckInConfig{
			// nothing
		},
		Session: &mocks.Session{
			MockCheckIn: func(ctx context.Context, config *model.OOAPICheckInConfig) (*model.OOAPICheckInResult, error) {
				return nil, expected
			},
			MockLogger: func() model.Logger {
				return log.Log
			},
		},
		StaticInputs: nil,
		SourceFiles:  nil,
	}

	// obtain the loader
	loader := factory.NewTargetLoader(config)

	// attempt to load targets
	targets, err := loader.Load(context.Background())

	// make sure we've got the expected error
	if !errors.Is(err, expected) {
		t.Fatal("unexpected error", err)
	}

	// make sure there are no targets
	if len(targets) != 0 {
		t.Fatal("expected zero length targets")
	}
}

// customConfig is a custom config for [TestFactoryCustomTargetLoaderForRicherInput].
type customConfig struct{}

// customTargetLoader is a custom target loader for [TestFactoryCustomTargetLoaderForRicherInput].
type customTargetLoader struct{}

// Load implements [model.ExperimentTargetLoader].
func (c *customTargetLoader) Load(ctx context.Context) ([]model.ExperimentTarget, error) {
	panic("should not be called")
}

func TestFactoryNewTargetLoader(t *testing.T) {
	t.Run("with custom target loader", func(t *testing.T) {
		// create factory creating a custom target loader
		factory := &Factory{
			build:            nil,
			canonicalName:    "",
			config:           &customConfig{},
			enabledByDefault: false,
			inputPolicy:      "",
			interruptible:    false,
			newLoader: func(config *targetloading.Loader, options any) model.ExperimentTargetLoader {
				return &customTargetLoader{}
			},
		}

		// create config for creating a new target loader
		config := &model.ExperimentTargetLoaderConfig{
			CheckInConfig: &model.OOAPICheckInConfig{ /* nothing */ },
			Session: &mocks.Session{
				MockLogger: func() model.Logger {
					return model.DiscardLogger
				},
			},
			StaticInputs: []string{},
			SourceFiles:  []string{},
		}

		// create the loader
		loader := factory.NewTargetLoader(config)

		// make sure the type is the one we expected
		if _, good := loader.(*customTargetLoader); !good {
			t.Fatalf("expected a *customTargetLoader, got %T", loader)
		}
	})

	t.Run("with default target loader", func(t *testing.T) {
		// create factory creating a default target loader
		factory := &Factory{
			build:            nil,
			canonicalName:    "",
			config:           &customConfig{},
			enabledByDefault: false,
			inputPolicy:      "",
			interruptible:    false,
			newLoader:        nil, // explicitly nil
		}

		// create config for creating a new target loader
		config := &model.ExperimentTargetLoaderConfig{
			CheckInConfig: &model.OOAPICheckInConfig{ /* nothing */ },
			Session: &mocks.Session{
				MockLogger: func() model.Logger {
					return model.DiscardLogger
				},
			},
			StaticInputs: []string{},
			SourceFiles:  []string{},
		}

		// create the loader
		loader := factory.NewTargetLoader(config)

		// make sure the type is the one we expected
		if _, good := loader.(*targetloading.Loader); !good {
			t.Fatalf("expected a *targetloading.Loader, got %T", loader)
		}
	})
}

// This test is important because SetOptionsJSON assumes that the experiment
// config is a struct pointer into which it is possible to write
func TestExperimentConfigIsAlwaysAPointerToStruct(t *testing.T) {
	for name, ffunc := range AllExperiments {
		t.Run(name, func(t *testing.T) {
			factory := ffunc()
			config := factory.config
			ctype := reflect.TypeOf(config)
			if ctype.Kind() != reflect.Pointer {
				t.Fatal("expected a pointer")
			}
			ctype = ctype.Elem()
			if ctype.Kind() != reflect.Struct {
				t.Fatal("expected a struct")
			}
		})
	}
}
