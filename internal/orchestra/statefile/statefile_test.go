package statefile

import (
	"encoding/json"
	"errors"
	"testing"
	"time"

	"github.com/ooni/probe-engine/internal/kvstore"
)

func TestUnitStateAuth(t *testing.T) {
	t.Run("with no Token", func(t *testing.T) {
		state := State{Expire: time.Now().Add(10 * time.Hour)}
		if state.Auth() != nil {
			t.Fatal("expected nil here")
		}
	})
	t.Run("with expired Token", func(t *testing.T) {
		state := State{
			Expire: time.Now().Add(-1 * time.Hour),
			Token:  "xx-x-xxx-xx",
		}
		if state.Auth() != nil {
			t.Fatal("expected nil here")
		}
	})
	t.Run("with good Token", func(t *testing.T) {
		state := State{
			Expire: time.Now().Add(10 * time.Hour),
			Token:  "xx-x-xxx-xx",
		}
		if state.Auth() == nil {
			t.Fatal("expected valid auth here")
		}
	})
}

func TestUnitStateCredentials(t *testing.T) {
	t.Run("with no ClientID", func(t *testing.T) {
		state := State{}
		if state.Credentials() != nil {
			t.Fatal("expected nil here")
		}
	})
	t.Run("with no Password", func(t *testing.T) {
		state := State{
			ClientID: "xx-x-xxx-xx",
		}
		if state.Credentials() != nil {
			t.Fatal("expected nil here")
		}
	})
	t.Run("with all good", func(t *testing.T) {
		state := State{
			ClientID: "xx-x-xxx-xx",
			Password: "xx",
		}
		if state.Credentials() == nil {
			t.Fatal("expected valid auth here")
		}
	})
}

func TestIntegrationStateFileMemory(t *testing.T) {
	sf := New(kvstore.NewMemoryKeyValueStore())
	if sf == nil {
		t.Fatal("expected non nil pointer here")
	}
	s := State{
		Expire:   time.Now(),
		Password: "xy",
		Token:    "abc",
		ClientID: "xx",
	}
	if err := sf.Set(s); err != nil {
		t.Fatal(err)
	}
	os := sf.Get()
	if s.ClientID != os.ClientID {
		t.Fatal("the ClientID field has changed")
	}
	if !s.Expire.Equal(os.Expire) {
		t.Fatal("the Expire field has changed")
	}
	if s.Password != os.Password {
		t.Fatal("the Password field has changed")
	}
	if s.Token != os.Token {
		t.Fatal("the Token field has changed")
	}
}

func TestUnitStateFileSetMarshalError(t *testing.T) {
	sf := New(kvstore.NewMemoryKeyValueStore())
	if sf == nil {
		t.Fatal("expected non nil pointer here")
	}
	s := State{
		Expire:   time.Now(),
		Password: "xy",
		Token:    "abc",
		ClientID: "xx",
	}
	expected := errors.New("mocked error")
	failingfunc := func(v interface{}) ([]byte, error) {
		return nil, expected
	}
	if err := sf.set(s, failingfunc); !errors.Is(err, expected) {
		t.Fatal("not the error we expected")
	}
}

func TestUnitStateFileGetKVStoreGetError(t *testing.T) {
	sf := New(kvstore.NewMemoryKeyValueStore())
	if sf == nil {
		t.Fatal("expected non nil pointer here")
	}
	expected := errors.New("mocked error")
	failingfunc := func(string) ([]byte, error) {
		return nil, expected
	}
	s, err := sf.get(failingfunc, json.Unmarshal)
	if !errors.Is(err, expected) {
		t.Fatal("not the error we expected")
	}
	if s.ClientID != "" {
		t.Fatal("unexpected ClientID field")
	}
	if !s.Expire.IsZero() {
		t.Fatal("unexpected Expire field")
	}
	if s.Password != "" {
		t.Fatal("unexpected Password field")
	}
	if s.Token != "" {
		t.Fatal("unexpected Token field")
	}
}

func TestUnitStateFileGetUnmarshalError(t *testing.T) {
	sf := New(kvstore.NewMemoryKeyValueStore())
	if sf == nil {
		t.Fatal("expected non nil pointer here")
	}
	if err := sf.Set(State{}); err != nil {
		t.Fatal(err)
	}
	expected := errors.New("mocked error")
	failingfunc := func([]byte, interface{}) error {
		return expected
	}
	s, err := sf.get(sf.store.Get, failingfunc)
	if !errors.Is(err, expected) {
		t.Fatal("not the error we expected")
	}
	if s.ClientID != "" {
		t.Fatal("unexpected ClientID field")
	}
	if !s.Expire.IsZero() {
		t.Fatal("unexpected Expire field")
	}
	if s.Password != "" {
		t.Fatal("unexpected Password field")
	}
	if s.Token != "" {
		t.Fatal("unexpected Token field")
	}
}
