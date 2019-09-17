package status

import (
	"io/ioutil"
	"log"
	"path/filepath"
	"regexp"
	"runtime"
	"strconv"
	"strings"
)

// CPUUtilizationStats holds the usage information on each core of the CPU as well as the
// total CPU usage. The information is represented as a percentage of utilization.
type CPUUtilizationStats struct {
	Total uint
	Cores []uint
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

type utilStats struct {
	total uint64
	idle  uint64
}

var prev = struct {
	total utilStats
	cores []utilStats
}{
	utilStats{0, 0},
	make([]utilStats, runtime.NumCPU()),
}

func getUtilization() (CPUUtilizationStats, error) {
	// pass # 1
	path, err := filepath.Abs("/proc/stat")
	if err != nil {
		return CPUUtilizationStats{}, err
	}
	fil, err := ioutil.ReadFile(path)
	if err != nil {
		return CPUUtilizationStats{}, err
	}
	lines := strings.Split(string(fil), "\n")
	tot, cores := calculate(lines)
	defer func() {
		prev.total = tot
		prev.cores = cores
	}()

	ret := CPUUtilizationStats{}

	total := tot.total - prev.total.total // sets the total
	idle := tot.idle - prev.total.idle
	log.Printf("total: %d \t idle: %d", total, idle)
	ret.Total = uint(float64(total-idle) / float64(total) * 100)

	for i, core := range cores {
		total = core.total - prev.cores[i].total // sets the total
		idle := core.idle - prev.cores[i].idle
		ret.Cores = append(ret.Cores, uint(float64(total-idle)/float64(total)*100))
	}

	return ret, nil

}

func calculate(lines []string) (total utilStats, cores []utilStats) {
	//var total, idle uint64
	cores = make([]utilStats, runtime.NumCPU())
	for _, line := range lines {
		fields := strings.Fields(line)
		if len(fields) < 1 {
			continue
		}
		if match, err := regexp.MatchString("^(cpu){1}$", fields[0]); match {
			if err != nil {
				log.Printf("stats: encountered regex error -- %#v", err)
				continue
			}

			for i := 1; i < len(fields); i++ {
				conv, err := strconv.ParseUint(fields[i], 10, 64)
				if err != nil {
					log.Printf("stats: encountered error -- %#v", err)
					continue
				}
				total.total += conv
				if i == 4 {
					total.idle = conv
				}
			}

		} else if match, err := regexp.MatchString("^(cpu\\d+){1}$", fields[0]); match {
			log.Println("match found!")
			if err != nil {
				log.Printf("stats: encountered regex error -- %#v", err)
				continue
			}
			log.Printf("parsing value: %s", strings.ReplaceAll(string(fields[0]), "cpu", ""))
			coreNum, err := strconv.ParseInt(strings.ReplaceAll(string(fields[0]), "cpu", ""), 10, 64)
			if err != nil {
				panic(err)
			}

			for i := 1; i < len(fields); i++ {
				conv, err := strconv.ParseUint(fields[i], 10, 64)
				if err != nil {
					log.Printf("stats: encountered error -- %#v", err)
					continue
				}
				cores[coreNum].total += conv
				if i == 4 {
					cores[coreNum].idle = conv
				}
			}
		}
	}
	return
}
