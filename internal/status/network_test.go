package status

import "testing"

func TestOngoingAverages(t *testing.T) {
	var testCases = []struct {
		values  []int
		average float64
	}{
		{[]int{1, 2, 3, 4, 5, 6, 7, 8, 9, 10}, 5.5},
		{[]int{45, 34, 213}, 97.333},
		{[]int{1, 1, 1, 1}, 1},
	}

	for _, test := range testCases {
		t.Run("Base", func(t *testing.T) {
			valchan := make(chan int, len(test.values))
			for _, val := range test.values {
				valchan <- val
			}

			avg := GlobalNetworkInformation.Average()

			if avg < (test.average-0.01) || avg > (test.average+0.01) {
				t.Logf("averages did not match. wanted: %v, got: %v", test.average, avg)
				t.Fail()
			}
		})
	}
}
