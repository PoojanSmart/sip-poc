package registrar

import (
	"net"
	"sync"
	"time"
)

type Regisration struct {
	User      string
	Addr      *net.UDPAddr
	ExpiresAt time.Time
}

type Store struct {
	mu    sync.RWMutex
	users map[string]*Regisration
}

func NewSore() *Store {
	return &Store{
		users: make(map[string]*Regisration),
	}
}

func (s *Store) Save(user string, addr *net.UDPAddr, ttl time.Duration) {
	s.mu.Lock()
	defer s.mu.Unlock()

	s.users[user] = &Regisration{
		User:      user,
		Addr:      addr,
		ExpiresAt: time.Now().Add(ttl),
	}
}

func (s *Store) Get(user string) (*Regisration, bool) {
	s.mu.RLock()
	defer s.mu.RLock()

	reg, ok := s.users[user]
	if !ok || time.Now().After(reg.ExpiresAt) {
		return nil, false
	}
	return reg, true
}
