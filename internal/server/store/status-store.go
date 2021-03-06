package store

import (
	"errors"
	"log"
	"sync"

	"github.com/markpotocki/health/internal/server"
)

// ErrNotFound is returned when the value is not found
type ErrNotFound error

type StatusStore struct {
	db    []server.HealthStatus
	mutex sync.Mutex
}

func MakeStatusStore() *StatusStore {
	return &StatusStore{
		db:    make([]server.HealthStatus, 0),
		mutex: sync.Mutex{},
	}
}

func (ss *StatusStore) Save(hs server.HealthStatus) {
	for i, sshs := range ss.db {
		if sshs.ClientName == hs.ClientName {
			log.Printf("statusstore: match found for %s, updating", sshs.ClientName)
			ss.mutex.Lock()
			ss.db[i] = hs
			ss.mutex.Unlock()
			return
		}
	}
	log.Printf("statusstore: adding new entry for %s", hs.ClientName)
	ss.mutex.Lock()
	ss.db = append(ss.db, hs)
	ss.mutex.Unlock()
}

func (ss *StatusStore) SaveAll(hss ...server.HealthStatus) {
	for _, hs := range hss {
		ss.Save(hs)
	}
}

func (ss *StatusStore) Find(name string) (server.HealthStatus, error) { // might need to return an error here
	for _, hs := range ss.db {
		if hs.ClientName == name {
			return hs, nil
		}
	}
	return server.HealthStatus{}, ErrNotFound(errors.New("value " + name + " not found"))
}

func (ss *StatusStore) FindAll() []server.HealthStatus {
	return ss.db
}
