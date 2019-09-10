package main

import (
	"log"
	"time"

	"github.com/markpotocki/health/pkg/models"
)

func main() {
	for {
		log.Println("checking stats")
		hs := models.MakeHealthStatus()
		log.Println(hs)
		log.Println("waiting")
		<-time.After(time.Duration(10 * time.Second))
	}
}
