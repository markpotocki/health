package server

import (
	"bytes"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/markpotocki/health/pkg/models"
)

// Client Information Handler
// Responses:
// 	200 - client is found and returned
// 	404 - client not found
// 	500 - error decoding json from store (not tested)
func TestClientInfoHandler(t *testing.T) {
	t.Run("success-many", cihsuccessAll)
	t.Run("success-one", cihsuccess)
	t.Run("not-found", cihnotfound)
}

func cihsuccessAll(t *testing.T) {
	// setup
	srv := Server{
		clientStore: &mockClientStore{},
		statusStore: &mockStatusStore{},
	}
	recorder := httptest.NewRecorder()
	request := httptest.NewRequest("GET", "/aidi/info/", nil)
	request.Header.Set("Content-Type", "application/json")

	handler := http.HandlerFunc(srv.allClientInfoHandler)
	handler.ServeHTTP(recorder, request)

	resp := recorder.Result()
	defer resp.Body.Close()

	// check
	assert(t, resp.StatusCode, 200) // status is 200
	cli := []models.ClientInfo{}
	err := json.NewDecoder(resp.Body).Decode(&cli)
	check(err)
	assert(t, cli, defaultClient)
}

func cihsuccess(t *testing.T) {
	// setup
	srv := Server{
		clientStore: &mockClientStore{},
		statusStore: &mockStatusStore{},
	}
	recorder := httptest.NewRecorder()
	request := httptest.NewRequest("GET", "/aidi/info/test", nil)
	request.Header.Set("Content-Type", "application/json")

	handler := http.HandlerFunc(srv.clientInfoHandler)
	handler.ServeHTTP(recorder, request)

	resp := recorder.Result()
	defer resp.Body.Close()

	// check
	assert(t, resp.StatusCode, 200) // status is 200
	cli := models.ClientInfo{}
	err := json.NewDecoder(resp.Body).Decode(&cli)
	check(err)
	assert(t, cli, defaultClient)
}

func cihnotfound(t *testing.T) {
	// setup
	srv := Server{
		clientStore: &mockClientStore{},
		statusStore: &mockStatusStore{},
	}
	recorder := httptest.NewRecorder()
	request := httptest.NewRequest("GET", "/aidi/info/"+notFoundClient, nil)
	request.Header.Set("Content-Type", "application/json")

	handler := http.HandlerFunc(srv.clientInfoHandler)
	handler.ServeHTTP(recorder, request)

	resp := recorder.Result()
	defer resp.Body.Close()

	// check
	assert(t, resp.StatusCode, 404) // status is 404
}

// Test Register Handler
// Responses:
//	200 - registered successfully (could this be created?)
// 	400 - invalid json format for client
func TestRegisterHandler(t *testing.T) {
	t.Run("succeess", rhsuccess)
	t.Run("bad-request", rhbadrequest)
}

func rhsuccess(t *testing.T) {
	// setup
	srv := Server{
		clientStore: &mockClientStore{},
		statusStore: &mockStatusStore{},
	}
	buf := bytes.Buffer{}
	err := json.NewEncoder(&buf).Encode(defaultClient)
	check(err)

	recorder := httptest.NewRecorder()
	request := httptest.NewRequest("POST", "/register", &buf)
	request.Header.Set("Content-Type", "application/json")

	handler := http.HandlerFunc(srv.registerHandler)
	handler.ServeHTTP(recorder, request)

	resp := recorder.Result()

	assert(t, resp.StatusCode, 200)
}

func rhbadrequest(t *testing.T) {
	// setup
	srv := Server{
		clientStore: &mockClientStore{},
		statusStore: &mockStatusStore{},
	}
	buf := bytes.Buffer{}
	err := json.NewEncoder(&buf).Encode(defaultStatus)
	check(err)

	recorder := httptest.NewRecorder()
	request := httptest.NewRequest("POST", "/register", &buf)
	request.Header.Set("Content-Type", "application/json")

	handler := http.HandlerFunc(srv.registerHandler)
	handler.ServeHTTP(recorder, request)

	resp := recorder.Result()

	assert(t, resp.StatusCode, 400)
}

func TestReadyHandler(t *testing.T) {
	t.Run("success", readysuccess)
}

func readysuccess(t *testing.T) {
	// setup
	srv := Server{
		clientStore: &mockClientStore{},
		statusStore: &mockStatusStore{},
	}

	recorder := httptest.NewRecorder()
	request := httptest.NewRequest("GET", "/ready", nil)

	handler := http.HandlerFunc(srv.registerHandler)
	handler.ServeHTTP(recorder, request)

	resp := recorder.Result()

	assert(t, resp.StatusCode, 200)
}

// Utils

func check(err error) {
	if err != nil {
		panic(err)
	}
}

func assert(t *testing.T, actual interface{}, expect interface{}) {
	if actual != expect {
		t.Logf("assert: actual[%v] did not match expected[%v]", actual, expect)
	}
}

// Mocks
var defaultClient = models.ClientInfo{
	CName: "test",
	CPort: 1,
	CURL:  "http://test",
	Key:   "blah",
}

var defaultStatus = models.HealthStatus{
	CPU: models.HealthStatusCpu{
		Cores:           2,
		Utilization:     50,
		CoreUtilization: []uint{50, 10},
	},
	Memory: models.HealthStatusMem{
		ProcUsed:  10,
		ProcTotal: 20,
		SysTotal:  15,
	},
	Network: models.HealthStatusNetwork{
		AverageTime: 10,
	},
	Down:   true,
	Status: "testing",
}

type mockClientStore struct{}

func (mcs *mockClientStore) Save(ci models.ClientInfo) {}
func (mcs *mockClientStore) Get() []models.ClientInfo {
	return []models.ClientInfo{defaultClient}
}

type mockStatusStore struct{}

const notFoundClient string = "notfound"

func (mss *mockStatusStore) Save(hs HealthStatus)       {}
func (mss *mockStatusStore) SaveAll(hs ...HealthStatus) {}
func (mss *mockStatusStore) Find(ClientName string) (HealthStatus, error) {
	if ClientName == notFoundClient {
		return HealthStatus{}, errors.New("not found")
	}
	return HealthStatus{
		ClientName: ClientName,
		Data:       defaultStatus,
		Updated:    1,
	}, nil
}
func (mss *mockStatusStore) FindAll() []HealthStatus {
	foo, _ := mss.Find("test")
	return []HealthStatus{foo}
}
