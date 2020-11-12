package engine

import (
	"time"

	"github.com/ooni/probe-engine/experiment/dash"
	"github.com/ooni/probe-engine/experiment/dnscheck"
	"github.com/ooni/probe-engine/experiment/example"
	"github.com/ooni/probe-engine/experiment/fbmessenger"
	"github.com/ooni/probe-engine/experiment/hhfm"
	"github.com/ooni/probe-engine/experiment/hirl"
	"github.com/ooni/probe-engine/experiment/httphostheader"
	"github.com/ooni/probe-engine/experiment/ndt7"
	"github.com/ooni/probe-engine/experiment/psiphon"
	"github.com/ooni/probe-engine/experiment/riseupvpn"
	"github.com/ooni/probe-engine/experiment/sniblocking"
	"github.com/ooni/probe-engine/experiment/stunreachability"
	"github.com/ooni/probe-engine/experiment/telegram"
	"github.com/ooni/probe-engine/experiment/tlstool"
	"github.com/ooni/probe-engine/experiment/tor"
	"github.com/ooni/probe-engine/experiment/urlgetter"
	"github.com/ooni/probe-engine/experiment/webconnectivity"
	"github.com/ooni/probe-engine/experiment/whatsapp"
)

var experimentsByName = map[string]func(*Session) *ExperimentBuilder{
	"dash": func(session *Session) *ExperimentBuilder {
		return &ExperimentBuilder{
			build: func(config interface{}) *Experiment {
				return NewExperiment(session, dash.NewExperimentMeasurer(
					*config.(*dash.Config),
				))
			},
			config:        &dash.Config{},
			interruptible: true,
			inputPolicy:   InputNone,
		}
	},

	"dnscheck": func(session *Session) *ExperimentBuilder {
		return &ExperimentBuilder{
			build: func(config interface{}) *Experiment {
				return NewExperiment(session, dnscheck.NewExperimentMeasurer(
					*config.(*dnscheck.Config),
				))
			},
			config:      &dnscheck.Config{},
			inputPolicy: InputRequired,
		}
	},

	"example": func(session *Session) *ExperimentBuilder {
		return &ExperimentBuilder{
			build: func(config interface{}) *Experiment {
				return NewExperiment(session, example.NewExperimentMeasurer(
					*config.(*example.Config), "example",
				))
			},
			config: &example.Config{
				Message:   "Good day from the example experiment!",
				SleepTime: int64(time.Second),
			},
			interruptible: true,
			inputPolicy:   InputNone,
		}
	},

	"example_with_input": func(session *Session) *ExperimentBuilder {
		return &ExperimentBuilder{
			build: func(config interface{}) *Experiment {
				return NewExperiment(session, example.NewExperimentMeasurer(
					*config.(*example.Config), "example_with_input",
				))
			},
			config: &example.Config{
				Message:   "Good day from the example with input experiment!",
				SleepTime: int64(time.Second),
			},
			interruptible: true,
			inputPolicy:   InputRequired,
		}
	},

	// TODO(bassosimone): when we can set experiment options using the JSON
	// we need to get rid of all these multiple experiments.
	//
	// See https://github.com/ooni/probe-engine/issues/413
	"example_with_input_non_interruptible": func(session *Session) *ExperimentBuilder {
		return &ExperimentBuilder{
			build: func(config interface{}) *Experiment {
				return NewExperiment(session, example.NewExperimentMeasurer(
					*config.(*example.Config), "example_with_input_non_interruptible",
				))
			},
			config: &example.Config{
				Message:   "Good day from the example with input experiment!",
				SleepTime: int64(time.Second),
			},
			interruptible: false,
			inputPolicy:   InputRequired,
		}
	},

	"example_with_failure": func(session *Session) *ExperimentBuilder {
		return &ExperimentBuilder{
			build: func(config interface{}) *Experiment {
				return NewExperiment(session, example.NewExperimentMeasurer(
					*config.(*example.Config), "example_with_failure",
				))
			},
			config: &example.Config{
				Message:     "Good day from the example with failure experiment!",
				ReturnError: true,
				SleepTime:   int64(time.Second),
			},
			interruptible: true,
			inputPolicy:   InputNone,
		}
	},

	"facebook_messenger": func(session *Session) *ExperimentBuilder {
		return &ExperimentBuilder{
			build: func(config interface{}) *Experiment {
				return NewExperiment(session, fbmessenger.NewExperimentMeasurer(
					*config.(*fbmessenger.Config),
				))
			},
			config:      &fbmessenger.Config{},
			inputPolicy: InputNone,
		}
	},

	"http_header_field_manipulation": func(session *Session) *ExperimentBuilder {
		return &ExperimentBuilder{
			build: func(config interface{}) *Experiment {
				return NewExperiment(session, hhfm.NewExperimentMeasurer(
					*config.(*hhfm.Config),
				))
			},
			config:      &hhfm.Config{},
			inputPolicy: InputNone,
		}
	},

	"http_host_header": func(session *Session) *ExperimentBuilder {
		return &ExperimentBuilder{
			build: func(config interface{}) *Experiment {
				return NewExperiment(session, httphostheader.NewExperimentMeasurer(
					*config.(*httphostheader.Config),
				))
			},
			config:      &httphostheader.Config{},
			inputPolicy: InputRequired,
		}
	},

	"http_invalid_request_line": func(session *Session) *ExperimentBuilder {
		return &ExperimentBuilder{
			build: func(config interface{}) *Experiment {
				return NewExperiment(session, hirl.NewExperimentMeasurer(
					*config.(*hirl.Config),
				))
			},
			config:      &hirl.Config{},
			inputPolicy: InputNone,
		}
	},

	"ndt": func(session *Session) *ExperimentBuilder {
		return &ExperimentBuilder{
			build: func(config interface{}) *Experiment {
				return NewExperiment(session, ndt7.NewExperimentMeasurer(
					*config.(*ndt7.Config),
				))
			},
			config:        &ndt7.Config{},
			interruptible: true,
			inputPolicy:   InputNone,
		}
	},

	"psiphon": func(session *Session) *ExperimentBuilder {
		return &ExperimentBuilder{
			build: func(config interface{}) *Experiment {
				return NewExperiment(session, psiphon.NewExperimentMeasurer(
					*config.(*psiphon.Config),
				))
			},
			config:      &psiphon.Config{},
			inputPolicy: InputOptional,
		}
	},

	"sni_blocking": func(session *Session) *ExperimentBuilder {
		return &ExperimentBuilder{
			build: func(config interface{}) *Experiment {
				return NewExperiment(session, sniblocking.NewExperimentMeasurer(
					*config.(*sniblocking.Config),
				))
			},
			config: &sniblocking.Config{},
			inputPolicy: InputRequired,
		}
	},

	"stun_reachability": func(session *Session) *ExperimentBuilder {
		return &ExperimentBuilder{
			build: func(config interface{}) *Experiment {
				return NewExperiment(session, stunreachability.NewExperimentMeasurer(
					*config.(*stunreachability.Config),
				))
			},
			config:      &stunreachability.Config{},
			inputPolicy: InputOptional,
		}
	},

	"riseupvpn": func(session *Session) *ExperimentBuilder {
		return &ExperimentBuilder{
			build: func(config interface{}) *Experiment {
				return NewExperiment(session, riseupvpn.NewExperimentMeasurer(
					*config.(*riseupvpn.Config),
				))
			},
			config:      &riseupvpn.Config{},
			inputPolicy: InputNone,
		}
	},

	"telegram": func(session *Session) *ExperimentBuilder {
		return &ExperimentBuilder{
			build: func(config interface{}) *Experiment {
				return NewExperiment(session, telegram.NewExperimentMeasurer(
					*config.(*telegram.Config),
				))
			},
			config:      &telegram.Config{},
			inputPolicy: InputNone,
		}
	},

	"tlstool": func(session *Session) *ExperimentBuilder {
		return &ExperimentBuilder{
			build: func(config interface{}) *Experiment {
				return NewExperiment(session, tlstool.NewExperimentMeasurer(
					*config.(*tlstool.Config),
				))
			},
			config:      &tlstool.Config{},
			inputPolicy: InputRequired,
		}
	},

	"tor": func(session *Session) *ExperimentBuilder {
		return &ExperimentBuilder{
			build: func(config interface{}) *Experiment {
				return NewExperiment(session, tor.NewExperimentMeasurer(
					*config.(*tor.Config),
				))
			},
			config:      &tor.Config{},
			inputPolicy: InputNone,
		}
	},

	"urlgetter": func(session *Session) *ExperimentBuilder {
		return &ExperimentBuilder{
			build: func(config interface{}) *Experiment {
				return NewExperiment(session, urlgetter.NewExperimentMeasurer(
					*config.(*urlgetter.Config),
				))
			},
			config:      &urlgetter.Config{},
			inputPolicy: InputRequired,
		}
	},

	"web_connectivity": func(session *Session) *ExperimentBuilder {
		return &ExperimentBuilder{
			build: func(config interface{}) *Experiment {
				return NewExperiment(session, webconnectivity.NewExperimentMeasurer(
					*config.(*webconnectivity.Config),
				))
			},
			config:      &webconnectivity.Config{},
			inputPolicy: InputRequired,
		}
	},

	"whatsapp": func(session *Session) *ExperimentBuilder {
		return &ExperimentBuilder{
			build: func(config interface{}) *Experiment {
				return NewExperiment(session, whatsapp.NewExperimentMeasurer(
					*config.(*whatsapp.Config),
				))
			},
			config:      &whatsapp.Config{},
			inputPolicy: InputNone,
		}
	},
}

// AllExperiments returns the name of all experiments
func AllExperiments() []string {
	var names []string
	for key := range experimentsByName {
		names = append(names, key)
	}
	return names
}
