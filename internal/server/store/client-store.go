package store

import (
	"log"
	"sync"

	"github.com/markpotocki/health/pkg/models"
)

type ClientStore struct {
	db    []models.ClientInfo
	mutex sync.Mutex
}

func MakeClientStore() *ClientStore {
	return &ClientStore{
		db:    make([]models.ClientInfo, 0),
		mutex: sync.Mutex{},
	}
}

func (cs *ClientStore) Save(info models.ClientInfo) {
	for i, cinfo := range cs.db {
		if info.Name() == cinfo.Name() {
			log.Printf("clientstore: match found on %s, updating entry", info.Name())
			cs.mutex.Lock()
			cs.db[i] = info
			cs.mutex.Unlock()
			return
		}
	}
	log.Printf("clientstore: adding new entry %s", info.Name())
	cs.mutex.Lock()
	cs.db = append(cs.db, info)
	cs.mutex.Unlock()
}

func (cs *ClientStore) Get() []models.ClientInfo {
	return cs.db
}
