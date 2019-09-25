package status

import (
	"math"
)

// ResponseAverager contains a list of vals that is limited by maxN. When a value is added
// passed the maxN value, the oldest entry is removed and the new one added. Ie, vals.Len()
// will never be over maxN.
type ResponseAverager struct {
	vals []int
	curr int
	div  int
}

// GlobalNetworkInformation contains the Averager for the time to reply on all http
// handlers that utilize the pkg/handlers handler.
var GlobalNetworkInformation *ResponseAverager

func init() {
	GlobalNetworkInformation = &ResponseAverager{
		vals: make([]int, 50),
		div:  0,
	}
}

// AddVal adds a new value, ensuring that the length of the list is not over maxN.
// If it is, it will remove the oldest entry, if not it will add like a normal list.
func (avger *ResponseAverager) AddVal(val int) {
	ind := avger.div % 50         // max supported is 50
	avger.vals[ind] = val         // store the previous value in the next free index
	avger.curr = avger.curr + val // get our running total
	avger.div++                   // increment our division counter
}

// Average gets the mean value of all items in the list.
func (avger *ResponseAverager) Average() float64 {
	avg := float64(avger.curr) / float64(avger.div)
	return math.Round(avg*100) / 100
}

func (avger *ResponseAverager) AverageLastN(n int) {

}
