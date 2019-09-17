package server

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"strings"

	"github.com/markpotocki/health/pkg/models"
)

func (srv *Server) registerHandler(w http.ResponseWriter, r *http.Request) {
	clientInfo := models.ClientInfo{}

	err := json.NewDecoder(r.Body).Decode(&clientInfo)

	if clientInfo.CPort == 0 {
		clientInfo.CPort = 9999 // for backwards compatability
	}

	if err != nil {
		log.Printf("server-register: bad type recieved %v", err)
		http.Error(w, "not expected json", http.StatusBadRequest)
	}

	clientAddr := r.RemoteAddr
	splited := strings.Split(clientAddr, ":")
	if baseurl := splited[0]; baseurl != "" {
    clientInfo.CURL = fmt.Sprintf("http://%s:%d/metrics/health", baseurl, clientInfo.CPort)
	}

	srv.clientStore.Save(clientInfo)

	w.WriteHeader(http.StatusCreated)

}

func (srv *Server) readyHandler(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusOK) // can add checks for whatever here
}

func (srv *Server) clientInfoHandler(w http.ResponseWriter, r *http.Request) {
	// start with http://localhost:0/info/param
	httpTrim := strings.TrimPrefix(r.RequestURI, "http://")
	httpTrim = strings.TrimPrefix(r.RequestURI, "https://")
	httpTrim = strings.TrimPrefix(httpTrim, "/")
	log.Println("FIND ME::  " + httpTrim)
	// localhost:0/aidi/info/param/
	split := strings.Split(httpTrim, "/")
	if len(split) > 4 {
		log.Println("server: invalid path in info handler")
		http.Error(w, "Not Found", http.StatusNotFound)
	} else if len(split) == 4 {
		info := srv.statusStore.Find(split[3])
		if info.ClientName == "" {
			http.Error(w, "could not find the requested client", http.StatusNotFound)
		} else {
			err := json.NewEncoder(w).Encode(&info)
			if err != nil {
				log.Printf("server: encountered error decoding json: %v", err)
				http.Error(w, "internal error", http.StatusInternalServerError)
			}
		}
	} else {
		info := srv.statusStore.FindAll()
		err := json.NewEncoder(w).Encode(&info)
		if err != nil {
			log.Printf("server: encountered error decoding json: %v", err)
			http.Error(w, "internal error", http.StatusInternalServerError)
		}
	}

}
