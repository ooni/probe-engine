package orchestra_test

import (
	"testing"

	"github.com/ooni/probe-engine/internal/orchestra/testorchestra"
)

func TestIntegrationSuccess(t *testing.T) {
	clientID, err := testorchestra.Register()
	if err != nil {
		t.Fatal(err)
	}
	auth, err := testorchestra.Login(clientID)
	if err != nil {
		t.Fatal(err)
	}
	if err := testorchestra.Update(auth, clientID); err != nil {
		t.Fatal(err)
	}
}

func TestIntegrationFailure(t *testing.T) {
	if err := testorchestra.Update(nil, "xx"); err == nil {
		t.Fatal("expected an error here")
	}
}
