// Package statefile defines the state file
package statefile

import (
	"encoding/json"
	"time"

	"github.com/ooni/probe-engine/internal/orchestra/login"
	"github.com/ooni/probe-engine/model"
)

// State is the state stored inside the state file
type State struct {
	ClientID string
	Expire   time.Time
	Password string
	Token    string
}

// Auth returns an authentication structure, if possible, otherwise
// it returns nil, meaning that you should login again.
func (s State) Auth() *login.Auth {
	if s.Token == "" {
		return nil
	}
	if time.Now().Add(30 * time.Second).After(s.Expire) {
		return nil
	}
	return &login.Auth{
		Expire: s.Expire,
		Token:  s.Token,
	}
}

// Credentials returns login credentials, if possible, otherwise it
// returns nil, meaning that you should create an account.
func (s State) Credentials() *login.Credentials {
	if s.ClientID == "" {
		return nil
	}
	if s.Password == "" {
		return nil
	}
	return &login.Credentials{
		ClientID: s.ClientID,
		Password: s.Password,
	}
}

// StateFile is the orchestra state file. It is backed by
// a generic key-value store configured by the user.
type StateFile struct {
	key   string
	store model.KeyValueStore
}

// New creates a new state file backed by a key-value store
func New(kvstore model.KeyValueStore) *StateFile {
	return &StateFile{
		key:   "orchestra.state",
		store: kvstore,
	}
}

func (sf *StateFile) set(s State, mf func(interface{}) ([]byte, error)) error {
	data, err := mf(s)
	if err != nil {
		return err
	}
	return sf.store.Set(sf.key, string(data))
}

// Set saves the current state on the key-value store.
func (sf *StateFile) Set(s State) error {
	return sf.set(s, json.Marshal)
}

func (sf *StateFile) get(
	sfget func(string) (string, error),
	unmarshal func([]byte, interface{}) error,
) (State, error) {
	value, err := sfget(sf.key)
	if err != nil {
		return State{}, err
	}
	var state State
	if err := unmarshal([]byte(value), &state); err != nil {
		return State{}, err
	}
	return state, nil
}

// Get returns the current state. In case of any error with the
// underlying key-value store, we return an empty state.
func (sf *StateFile) Get() (state State) {
	state, _ = sf.get(sf.store.Get, json.Unmarshal)
	return
}
