package store

import (
	"log"
	"sync"

	"github.com/markpotocki/health/internal/server"
)

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

func (ss *StatusStore) Find(name string) server.HealthStatus { // might need to return an error here
	for _, hs := range ss.db {
		if hs.ClientName == name {
			return hs
		}
	}
	return server.HealthStatus{}
}

func (ss *StatusStore) FindAll() []server.HealthStatus {
	return ss.db
}
