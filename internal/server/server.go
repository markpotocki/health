package server

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"sync"
	"syscall"
	"time"

	"github.com/markpotocki/health/pkg/client"
	"github.com/markpotocki/health/pkg/handlers"
	"github.com/markpotocki/health/pkg/models"
)

const port = 9999

// ClientStore is an object that is able to hold records of ClientInfo. It is used as an
// interface to allow for a database backed solution instead of the memory back one
// provided.
type ClientStore interface {
	Save(models.ClientInfo)
	Get() []models.ClientInfo
}

// StatusStore is an object that is able to hold records of ClientInfo. It is used as an
// interface to allow for a database backed solution instead of the memory back one
// provided.
type StatusStore interface {
	SaveAll(...HealthStatus)
	Save(HealthStatus)
	Find(ClientName string) (HealthStatus, error)
	FindAll() []HealthStatus
}

// HealthStatus contains the data that will be saved into the StatusStore. Contains the
// health data supplied by the client, the name of the client, and when it was last updated.
type HealthStatus struct {
	ClientName string
	Data       models.HealthStatus
	Updated    int64
}

// Server is an aidi server that is able to take in health data from clients that register
// with it.
type Server struct {
	clientStore ClientStore
	statusStore StatusStore
	connections sync.Map
}

// MakeServer provides a new Server pointer with the provided ClientStore and StatusStore.
func MakeServer(clientStore ClientStore, statusStore StatusStore) *Server {
	return &Server{clientStore, statusStore, sync.Map{}}
}

// Start registers the servers handlers, starts up the http server, and registers with
// itself. If all this is successful, it will run until shutdown, pinging clients at the
// set interval of 10s.
func (srv *Server) Start() {
	log.Println("server: starting health server")
	http.Handle("/aidi/register", handlers.ResponseTimer(http.HandlerFunc(srv.registerHandler)))
	http.Handle("/aidi/ready", handlers.ResponseTimer(http.HandlerFunc(srv.readyHandler)))
	http.Handle("/aidi/health/", http.HandlerFunc(srv.clientInfoHandler))
	log.Println("server: registering handlers")
	errchan := make(chan error, 1)

	go func() {
		errchan <- http.ListenAndServe(":9900", nil)
	}()

	log.Println("server: server started correctly")

	// register client data with self
	log.Println("server: registering health data with self")
	selfInfo := client.ConnectionConfig{
		Host: "localhost",
		Port: "9900",
	}

	cli := client.MakeClient("aidi", 9901, selfInfo)
	log.Println("server: self client created")

	cli.Connect(context.Background()) // the error is being ignored

	// running
	var resetCount int
	sigQuit := make(chan os.Signal, 1)

	signal.Notify(sigQuit, os.Interrupt)
	signal.Notify(sigQuit, syscall.SIGTERM)
	signal.Notify(sigQuit, syscall.SIGINT)
	for {
		select {
		case err := <-errchan:
			log.Println("server: CRITICAL -- the http server has stopped")
			log.Printf("server: error %v", err)
			log.Println("server: attempting restart")
			resetCount++
			if resetCount < 3 {
				go func() {
					errchan <- http.ListenAndServe(":9900", nil)
				}()
			}
		case <-time.After(time.Duration(1 * time.Second)):
			log.Println("server: sending ping to all clients")
			go srv.pingAll()
		case <-sigQuit:
			log.Println("server: shutdown process beginning")
			os.Exit(0)
		}
	}
}

func errorStatus(err error) models.HealthStatus {
	return models.HealthStatus{
		Down:   true,
		Status: err.Error(),
	}
}

func (srv *Server) pingAll() {
	respchan := make(chan HealthStatus, 50)
	go func() {
		for _, cli := range srv.clientStore.Get() {
			send(cli, respchan)
		}
		close(respchan)
	}()

	for resp := range respchan {
		log.Printf("server: saving to db %v", resp)
		srv.statusStore.Save(resp)
	}

	// we saved it all, notify there is new data
	go func() {
		srv.connections.Range(func(key interface{}, val interface{}) bool {
			log.Printf("DEBUG: sending to %v", key)
			if ch, ok := val.(chan struct{}); ok {
				ch <- struct{}{}
			}
			return true
		})
	}()

}

var httpcli = http.Client{
	Timeout: time.Duration(7 * time.Second),
}

func send(cli models.ClientInfo, respchan chan<- HealthStatus) {
	resp, err := httpcli.Get(cli.URL())
	if err != nil {
		respchan <- HealthStatus{cli.Name(), errorStatus(err), time.Now().Unix()}
		return
	}

	if resp.StatusCode != 200 {
		log.Printf("server: tried to reach %s but got bad status", cli.URL())
		respchan <- HealthStatus{cli.Name(), errorStatus(fmt.Errorf("did not get 200 response, got %d", resp.StatusCode)), time.Now().Unix()}
		return
	}

	hs := models.HealthStatus{}
	err = json.NewDecoder(resp.Body).Decode(&hs)
	if err != nil {
		respchan <- HealthStatus{cli.Name(), errorStatus(err), time.Now().Unix()}
		return
	}

	respchan <- HealthStatus{cli.Name(), hs, time.Now().Unix()}
}
