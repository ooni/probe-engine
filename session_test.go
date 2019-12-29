package engine

import (
	"io/ioutil"
	"testing"

	"github.com/apex/log"
)

func TestNewSessionBuilderChecks(t *testing.T) {
	t.Run("with no settings", func(t *testing.T) {
		newSessionMustFail(t, SessionConfig{})
	})
	t.Run("with only assets dir", func(t *testing.T) {
		newSessionMustFail(t, SessionConfig{
			AssetsDir: "testdata",
		})
	})
	t.Run("with also logger", func(t *testing.T) {
		newSessionMustFail(t, SessionConfig{
			AssetsDir: "testdata",
			Logger:    log.Log,
		})
	})
	t.Run("with also software name", func(t *testing.T) {
		newSessionMustFail(t, SessionConfig{
			AssetsDir:    "testdata",
			Logger:       log.Log,
			SoftwareName: "ooniprobe-engine",
		})
	})
	t.Run("with also software version", func(t *testing.T) {
		newSessionMustFail(t, SessionConfig{
			AssetsDir:       "testdata",
			Logger:          log.Log,
			SoftwareName:    "ooniprobe-engine",
			SoftwareVersion: "0.0.1",
		})
	})
}

func TestNewSessionBuilderGood(t *testing.T) {
	newSessionForTesting(t)
}

func newSessionMustFail(t *testing.T, config SessionConfig) {
	sess, err := NewSession(config)
	if err == nil {
		t.Fatal("expected an error here")
	}
	if sess != nil {
		t.Fatal("expected nil session here")
	}
}

func newSessionForTesting(t *testing.T) *Session {
	tempdir, err := ioutil.TempDir("testdata", "enginetests")
	if err != nil {
		t.Fatal(err)
	}
	sess, err := NewSession(SessionConfig{
		AssetsDir:       "testdata",
		Logger:          log.Log,
		SoftwareName:    "ooniprobe-engine",
		SoftwareVersion: "0.0.1",
		TempDir:         tempdir,
	})
	if err != nil {
		t.Fatal(err)
	}
	sess.AddAvailableHTTPSBouncer("https://ps-test.ooni.io")
	sess.AddAvailableHTTPSCollector("https://ps-test.ooni.io")
	sess.SetIncludeProbeASN(true)
	sess.SetIncludeProbeCC(true)
	sess.SetIncludeProbeIP(false)
	if err := sess.MaybeLookupLocation(); err != nil {
		t.Fatal(err)
	}
	log.Infof("Platform: %s", sess.Platform())
	log.Infof("ProbeASN: %d", sess.ProbeASN())
	log.Infof("ProbeASNString: %s", sess.ProbeASNString())
	log.Infof("ProbeCC: %s", sess.ProbeCC())
	log.Infof("ProbeIP: %s", sess.ProbeIP())
	log.Infof("ProbeNetworkName: %s", sess.ProbeNetworkName())
	log.Infof("ResolverASN: %d", sess.ResolverASN())
	log.Infof("ResolverASNString: %s", sess.ResolverASNString())
	log.Infof("ResolverIP: %s", sess.ResolverIP())
	log.Infof("ResolverNetworkName: %s", sess.ResolverNetworkName())
	if err := sess.MaybeLookupBackends(); err != nil {
		t.Fatal(err)
	}
	return sess
}
