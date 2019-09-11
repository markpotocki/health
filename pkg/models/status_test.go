package models

import (
	"testing"
	"time"
)

func TestMakeHealthStatus(t *testing.T) {
	hschan := make(chan HealthStatus)

	go func() {
		hschan <- MakeHealthStatus()
	}()

	// Timeout: 5s
	select {
	case <-hschan:
		return
	case <-time.After(time.Duration(5 * time.Second)):
		t.Log("test timed out before health status was made")
		t.Fail()
	}
}
