package client

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"net/http"
	"time"

	"github.com/markpotocki/health/pkg/models"
)

// Client should be able to answer through http as well as push an update
// of its status to the server
// Need:
// ConnectionConfig
//

const Endpoint string = "/aidi"

type ErrServerNotReady error
type ErrResponder error

type ConnectionConfig struct {
	Host       string
	Port       string
	AuthHeader string
	Interval   time.Duration
}

type Client struct {
	config ConnectionConfig
	name   string
}

func MakeClient(name string, config ConnectionConfig) *Client {
	return &Client{
		config: config,
		name:   name,
	}
}

func (c *Client) Connect(ctx context.Context) chan error {
	errchan := make(chan error, 1)
	hostURL := fmt.Sprintf("http://%s:%s", c.config.Host, c.config.Port)
	log.Println("client: checking if connection is ready")
	// first lets make sure the connection is valid and ready
	// we can do this by sending the server a GET request on
	// $Endpoint/ready
	resp, err := http.Get(hostURL + fmt.Sprintf("%s/ready", Endpoint))
	if err != nil {
		panic(err) // we can connect throw an error
	}
	if resp.StatusCode != 200 {
		errMsg := fmt.Sprintf("Server responded with status %d", resp.StatusCode)
		errchan <- ErrServerNotReady(errors.New(errMsg))
	}

	log.Println("client: aidi server is ready")

	// server is ready for our connections lets setup our pings
	// the server does not know we are here so we will make it aware
	log.Println("client: encoding client info to json")
	buffer := bytes.Buffer{}
	err = json.NewEncoder(&buffer).Encode(models.ClientInfo{CName: c.name})
	log.Println("client: value encoded to json")

	log.Println("client: registering with aidi server")
	resp, err = http.Post(hostURL+
		fmt.Sprintf("%s/register", Endpoint),
		"application/json",
		&buffer,
	)

	if err != nil {
		log.Println("client: could not connect to aidi server, panicking")
		panic(err)
	}

	log.Println("client: registration accepted")

	// we can now listen for requests for our health
	log.Println("client: opening endpoint for metrics")
	go responder(errchan)

	return errchan
}

func responder(errchan chan error) {
	healthHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		crhs := models.MakeHealthStatus()
		jsonErr := json.NewEncoder(w).Encode(crhs)
		if jsonErr != nil {
			errchan <- ErrResponder(jsonErr)
			http.Error(w, "Failed to decode json", http.StatusInternalServerError)
		}
	})

	http.Handle("/metrics/health", healthHandler)

	errchan <- http.ListenAndServe(":9999", nil)
}
