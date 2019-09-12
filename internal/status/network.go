package status

import (
	"container/list"
	"math"
)

// ResponseAverager contains a list of vals that is limited by maxN. When a value is added
// passed the maxN value, the oldest entry is removed and the new one added. Ie, vals.Len()
// will never be over maxN.
type ResponseAverager struct {
	vals *list.List
	maxN int
}

// GlobalNetworkInformation contains the Averager for the time to reply on all http
// handlers that utilize the pkg/handlers handler.
var GlobalNetworkInformation *ResponseAverager

func init() {
	GlobalNetworkInformation = &ResponseAverager{
		vals: list.New(),
		maxN: 50,
	}
	GlobalNetworkInformation.vals.PushFront(0) // hopefully fixes NaN issue
}

// AddVal adds a new value, ensuring that the length of the list is not over maxN.
// If it is, it will remove the oldest entry, if not it will add like a normal list.
func (avger *ResponseAverager) AddVal(val int) {
	if avger.vals.Len() >= avger.maxN {
		avger.vals.Remove(avger.vals.Front())
	}
	avger.vals.PushBack(val)
}

// Average gets the mean value of all items in the list.
func (avger *ResponseAverager) Average() float64 {
	var total int
	n := avger.vals.Len()

	for val := avger.vals.Front(); val != nil; val = val.Next() {
		total += val.Value.(int)
	}
	avg := float64(total) / float64(n)
	return math.Round(avg*100) / 100
}
