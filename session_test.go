package engine

import (
	"io/ioutil"
	"testing"

	"github.com/apex/log"
)

func TestNewSessionBuilder(t *testing.T) {
	newSessionForTesting(t)
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
	if err := sess.MaybeLookupLocation(); err != nil {
		t.Fatal(err)
	}
	log.Infof("ProbeASN: %d", sess.ProbeASN())
	log.Infof("ProbeASNString: %s", sess.ProbeASNString())
	log.Infof("ProbeCC: %s", sess.ProbeCC())
	log.Infof("ProbeIP: %s", sess.ProbeIP())
	log.Infof("ProbeNetworkName: %s", sess.ProbeNetworkName())
	log.Infof("ResolverIP: %s", sess.ResolverIP())
	if err := sess.MaybeLookupBackends(); err != nil {
		t.Fatal(err)
	}
	return sess
}
