package status

import (
	"container/list"
	"math"
)

type ResponseAverager struct {
	vals *list.List
	maxN int
}

var GlobalNetworkInformation *ResponseAverager

func init() {
	GlobalNetworkInformation = &ResponseAverager{
		vals: list.New(),
		maxN: 50,
	}
	GlobalNetworkInformation.vals.PushFront(0) // hopefully fixes NaN issue
}

func (avger *ResponseAverager) AddVal(val int) {
	if avger.vals.Len() >= avger.maxN {
		avger.vals.Remove(avger.vals.Front())
	}
	avger.vals.PushBack(val)
}

func (avger *ResponseAverager) Average() float64 {
	var total int
	n := avger.vals.Len()

	for val := avger.vals.Front(); val != nil; val = val.Next() {
		total += val.Value.(int)
	}
	avg := float64(total) / float64(n)
	return math.Round(avg*100) / 100
}
