package statefile

import (
	"testing"
	"time"
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

func TestUnitStateFileMemory(t *testing.T) {
	sf := NewMemory("/tmp")
	if sf == nil {
		t.Fatal("expected non nil pointer here")
	}
	if err := sf.Set(nil); err == nil {
		t.Fatal("expected an error here")
	}
	s := State{
		Expire:   time.Now(),
		Password: "xy",
		Token:    "abc",
		ClientID: "xx",
	}
	if err := sf.Set(&s); err != nil {
		t.Fatal(err)
	}
	os, err := sf.Get()
	if err != nil {
		t.Fatal(err)
	}
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
