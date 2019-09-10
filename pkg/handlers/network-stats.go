package handlers

import (
	"net/http"
	"time"

	"github.com/markpotocki/health/internal/status"
)

func ResponseTimer(next http.Handler) http.Handler {
	times := make(chan int, 100) // buffer a bit for multiple requests at same time
	go ongoingAverage(times)
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now().Unix()
		next.ServeHTTP(w, r)
		end := time.Now().Unix()
		times <- int(end - start)
	})
}

func ongoingAverage(times <-chan int) {
	for time := range times {
		status.GlobalNetworkInformation.AddVal(time)
	}
}
