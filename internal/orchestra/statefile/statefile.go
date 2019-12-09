// Package statefile defines the state file
package statefile

import (
	"errors"
	"sync"
	"time"

	"github.com/ooni/probe-engine/internal/orchestra/login"
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

// StateFile is a generic state file
type StateFile interface {
	Set(*State) error
	Get() (*State, error)
}

type memory struct {
	state State
	mu    sync.Mutex
}

// NewMemory creates a new state file in memory
func NewMemory(workdir string) StateFile {
	return &memory{}
}

func (sf *memory) Set(s *State) error {
	if s == nil {
		return errors.New("passed nil pointer")
	}
	sf.mu.Lock()
	defer sf.mu.Unlock()
	sf.state = *s
	return nil
}

func (sf *memory) Get() (*State, error) {
	sf.mu.Lock()
	defer sf.mu.Unlock()
	state := sf.state
	return &state, nil
}
