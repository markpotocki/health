package server

import (
	"encoding/json"
	"log"
	"net/http"

	"github.com/markpotocki/health/pkg/models"
)

func (srv *Server) registerHandler(w http.ResponseWriter, r *http.Request) {
	clientInfo := models.ClientInfo{}

	err := json.NewDecoder(r.Body).Decode(&clientInfo)

	if err != nil {
		log.Printf("server-register: bad type recieved %v", err)
		http.Error(w, "not expected json", http.StatusBadRequest)
	}

	srv.clientStore.Save(clientInfo)

	w.WriteHeader(http.StatusCreated)

}

func (srv *Server) readyHandler(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusOK) // can add checks for whatever here
}
