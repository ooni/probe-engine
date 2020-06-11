package orchestra_test

import "testing"

func TestUpdateSuccess(t *testing.T) {
	clientID, err := Register()
	if err != nil {
		t.Fatal(err)
	}
	auth, err := Login(clientID)
	if err != nil {
		t.Fatal(err)
	}
	if err := Update(auth, clientID); err != nil {
		t.Fatal(err)
	}
}

func TestUpdateFailure(t *testing.T) {
	if err := Update(nil, "xx"); err == nil {
		t.Fatal("expected an error here")
	}
}
