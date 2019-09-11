package server

import (
	"encoding/json"
	"log"
	"net/http"
	"os"
	"os/signal"
	"time"

	"github.com/markpotocki/health/pkg/models"
)

const port = 9999

type ClientStore interface {
	Save(models.ClientInfo)
	Get() []models.ClientInfo
}

type StatusStore interface {
	SaveAll(...HealthStatus)
	Save(HealthStatus)
	Find(ClientName string) HealthStatus
}

type HealthStatus struct {
	ClientName string
	Data       models.HealthStatus
	Updated    int64
}

type Server struct {
	clientStore ClientStore
	statusStore StatusStore
}

func MakeServer(clientStore ClientStore, statusStore StatusStore) *Server {
	return &Server{clientStore, statusStore}
}

func (srv *Server) Start() {
	http.HandleFunc("/aidi/register", srv.registerHandler)
	http.HandleFunc("/aidi/ready", srv.readyHandler)
	errchan := make(chan error, 1)
	errchan <- http.ListenAndServe(":9900", nil)
	var resetCount int
	sigQuit := make(chan os.Signal, 1)

	signal.Notify(sigQuit, os.Interrupt)
	for {
		select {
		case err := <-errchan:
			log.Println("server: CRITICAL -- the http server has stopped")
			log.Printf("server: error %v", err)
			log.Println("server: attempting restart")
			resetCount++
			if resetCount < 3 {
				errchan <- http.ListenAndServe(":9900", nil)
			}
		case <-time.After(time.Duration(10 * time.Second)):
			log.Println("server: sending ping to all clients")
			go srv.pingAll()
		case <-sigQuit:
			log.Println("server: shutdown process beginning")
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
			// bug here if there is no response the goroutine will hang
			send(cli, respchan)
		}
		close(respchan)
	}()

	for resp := range respchan {
		log.Printf("server: saving to db %v", resp)
		srv.statusStore.Save(resp)
	}

}

func send(cli models.ClientInfo, respchan chan<- HealthStatus) {
	httpcli := http.Client{
		Timeout: time.Duration(5 * time.Second),
	}
	resp, err := httpcli.Get(cli.URL())
	if err != nil {
		respchan <- HealthStatus{cli.Name(), errorStatus(err), time.Now().Unix()}
		return
	}

	if resp.StatusCode != 200 {
		respchan <- HealthStatus{cli.Name(), errorStatus(err), time.Now().Unix()}
		return
	}

	hs := models.HealthStatus{}
	err = json.NewDecoder(resp.Body).Decode(&hs)
	if err != nil {
		respchan <- HealthStatus{cli.Name(), errorStatus(err), time.Now().Unix()}
		return
	}

	respchan <- HealthStatus{cli.Name(), errorStatus(err), time.Now().Unix()}
}
