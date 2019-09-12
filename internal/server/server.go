package server

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/markpotocki/health/pkg/client"
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
	log.Println("server: starting health server")
	http.HandleFunc("/aidi/register", srv.registerHandler)
	http.HandleFunc("/aidi/ready", srv.readyHandler)
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

	cli := client.MakeClient("aidi", selfInfo)
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
		case <-time.After(time.Duration(10 * time.Second)):
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

}

func send(cli models.ClientInfo, respchan chan<- HealthStatus) {
	httpcli := http.Client{
		Timeout: time.Duration(7 * time.Second),
	}
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
