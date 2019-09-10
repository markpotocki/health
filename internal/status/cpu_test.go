package status

import (
	"log"
	"testing"
)

func TestGetUtilization(t *testing.T) {
	goLimit := 100
	// run a go routine
	for i := 0; i < goLimit; i++ {
		go func() {
			i := 1 + 2
			log.Println(i)
		}()
	}
	stats, err := getUtilization()

	if err != nil {
		t.Log("test failed with error")
		t.Log(err)
		t.FailNow()
	}

	if stats.Total == 0 {
		t.Log("got 0 for utilization, could be nil value")
		t.Fail()
	}
}
