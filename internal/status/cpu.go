package status

import (
	"io/ioutil"
	"log"
	"path/filepath"
	"strconv"
	"strings"
)

// CPUUtilizationStats holds the usage information on each core of the CPU as well as the
// total CPU usage. The information is represented as a percentage of utilization.
type CPUUtilizationStats struct {
	Total uint64
	Cores []uint64
}

// CPUUtilization gets the percentage of cpu utilization from the computer it is running on.
// It utilizes /proc/stat so as a consequence, this requires a linux OS to work correctly. The
// return value is a channel to listen on for the result, or it will close if there is an error.
func CPUUtilization() CPUUtilizationStats {
	stats, err := getUtilization()
	if err != nil {
		panic(err)
	}
	return stats

}

var prev = struct {
	total uint64
	idle  uint64
}{0, 0}

func getUtilization() (CPUUtilizationStats, error) {
	// pass # 1
	path, err := filepath.Abs("/proc/stat")
	log.Printf("cpu: trying to find on path %s", path)
	if err != nil {
		return CPUUtilizationStats{}, err
	}
	fil, err := ioutil.ReadFile(path)
	if err != nil {
		return CPUUtilizationStats{}, err
	}
	lines := strings.Split(string(fil), "\n")
	totPass, idlPass := calculate(lines)
	defer func() {
		prev.total = totPass
		prev.idle = idlPass
	}()

	total := totPass - prev.total
	idle := idlPass - prev.idle

	totPerc := uint64(float64(total-idle) / float64(total) * 100)
	log.Printf("cpu: got percentage %d", totPerc)
	return CPUUtilizationStats{
		Total: totPerc,
		Cores: make([]uint64, 0), // TODO get core stats
	}, nil

}

func calculate(lines []string) (total uint64, idle uint64) {
	//var total, idle uint64
	for _, line := range lines {
		fields := strings.Fields(line)
		log.Printf("fields: %#v", fields)
		if len(fields) < 1 {
			continue
		}
		if fields[0] == "cpu" {
			for i := 1; i < len(fields); i++ {
				conv, err := strconv.ParseUint(fields[i], 10, 64)
				if err != nil {
					log.Printf("stats: encountered error -- %#v", err)
					continue
				}
				total += conv
				log.Printf("cpu: adding to total at %d", total)
				if i == 4 {
					idle = conv
					log.Printf("cpu: idle set to %d", idle)
				}
			}
			break
		}
	}
	return
}
