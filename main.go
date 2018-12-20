package main

import (
	"encoding/json"
	"fmt"
	"os"
	"runtime"
	"syscall"
	"time"

	"github.com/c9s/goprocinfo/linux"
)

type segment struct {
	Text          string `json:"full_text"`
	Color         string `json:"color,omitempty"`
	Urgent        bool   `json:"urgent"`
	SeperateAfter bool   `json:"separator"`
}

const (
	red    = "#C83737"
	orange = "#DA7D3E"
	green  = "#A5CB76"
	white  = "#E6E6E6"
)

func main() {
	// i3bar expects an infinitely expanding JSON array as output from this
	// program. First we tell i3bar which protocol we want to use
	println(`{"version":1}`)

	// Then initialize the array
	println("[")

	// And print an empty array so we can start the next line with a comma while
	// perserving valid JSON syntax
	println("[]")

	// == Recurring variables ==

	// Used for measuring the duration of information gathering. We'll subtract
	// the difference from the final sleep duration
	var startTime time.Time
	var sleepDur = time.Second // Interval at which the status gets printed
	var err error

	var segments []segment
	var out []byte
	var color string

	// Disk
	var statfs = &syscall.Statfs_t{}
	var dused, dtotal uint64
	var dpercent float64

	// Memory
	var meminfo *linux.MemInfo
	var mused, mtotal uint64
	var mpercent float64

	// Network
	var netstat *linux.NetStat
	var oldRx, newRx, oldTx, newTx uint64

	// CPU usage
	var stat *linux.Stat
	var oldBusy, newBusy, newIdle, oldIdle uint64
	var cpercent float64

	// System load
	var loadAvg *linux.LoadAvg

	for {
		startTime = time.Now()
		// Clear the bar
		segments = []segment{}

		// Disk
		err = syscall.Statfs("/", statfs)
		if err != nil {
			segments = append(segments, segment{
				Text:          "fs: " + err.Error(),
				Color:         red,
				SeperateAfter: true,
			})
		} else {
			dused = (statfs.Blocks - statfs.Bavail) * uint64(statfs.Bsize)
			dtotal = statfs.Blocks * uint64(statfs.Bsize)
			dpercent = float64(dused) / float64(dtotal)

			if dpercent > .9 {
				color = red
			} else if dpercent > .75 {
				color = orange
			} else {
				color = green
			}

			segments = append(segments,
				segment{
					Text:          "fs",
					Color:         white,
					SeperateAfter: false,
				},
				segment{
					Text: fmt.Sprintf(
						"%9s / %-9s",
						formatData(dused),
						formatData(dtotal),
					),
					Color:         color,
					SeperateAfter: true,
				})
		}

		// Memory
		meminfo, err = linux.ReadMemInfo("/proc/meminfo")
		if err != nil {
			segments = append(segments, segment{
				Text:          "mem: " + err.Error(),
				Color:         red,
				SeperateAfter: true,
			})
		} else {
			mused = (meminfo.MemTotal * 1024) - (meminfo.MemAvailable * 1024)
			mtotal = meminfo.MemTotal * 1024
			mpercent = float64(mused) / float64(mtotal)

			if mpercent > .9 {
				color = red
			} else if mpercent > .75 {
				color = orange
			} else {
				color = green
			}

			segments = append(segments,
				segment{
					Text:          "mem",
					Color:         white,
					SeperateAfter: false,
				},
				segment{
					Text: fmt.Sprintf(
						"%9s / %-9s",
						formatData((meminfo.MemTotal*1024)-(meminfo.MemAvailable*1024)),
						formatData(meminfo.MemTotal*1024),
					),
					Color:         color,
					SeperateAfter: true,
				})
		}

		// Network speed
		netstat, err = linux.ReadNetStat("/proc/net/netstat")
		if err != nil {
			segments = append(segments, segment{
				Text:          "net: " + err.Error(),
				Color:         red,
				SeperateAfter: true,
			})
		} else {
			newRx = netstat.InOctets
			newTx = netstat.OutOctets
			segments = append(segments,
				segment{
					Text:          "net",
					Color:         white,
					SeperateAfter: false,
				},
				segment{
					Text: fmt.Sprintf("↓ %s ↑ %s",
						formatData((newRx-oldRx)/uint64(sleepDur.Seconds()))+"/s",
						formatData((newTx-oldTx)/uint64(sleepDur.Seconds()))+"/s",
					),
					Color:         white,
					SeperateAfter: true,
				})
			oldRx = newRx
			oldTx = newTx
		}

		// CPU usage
		stat, err = linux.ReadStat("/proc/stat")
		if err != nil {
			segments = append(segments, segment{
				Text:          "cpu: " + err.Error(),
				Color:         red,
				SeperateAfter: true,
			})
		} else {
			newBusy = (stat.CPUStatAll.User +
				stat.CPUStatAll.Nice +
				stat.CPUStatAll.System +
				stat.CPUStatAll.IOWait) - oldBusy
			newIdle = stat.CPUStatAll.Idle - oldIdle
			cpercent = float64(newBusy) / float64(newBusy+newIdle)

			if cpercent > .9 {
				color = red
			} else if cpercent > .75 {
				color = orange
			} else {
				color = green
			}

			segments = append(segments,
				segment{
					Text:          "cpu",
					Color:         white,
					SeperateAfter: false,
				},
				segment{
					Text:          fmt.Sprintf("%5.1f%%", cpercent*100),
					Color:         color,
					SeperateAfter: true,
				},
			)

			oldBusy = stat.CPUStatAll.User +
				stat.CPUStatAll.Nice +
				stat.CPUStatAll.System +
				stat.CPUStatAll.IOWait
			oldIdle = stat.CPUStatAll.Idle
		}

		// System load
		loadAvg, err = linux.ReadLoadAvg("/proc/loadavg")
		if err != nil {
			segments = append(segments, segment{
				Text:          "load: " + err.Error(),
				Color:         red,
				SeperateAfter: true,
			})
		} else {
			if loadAvg.Last1Min > float64(runtime.NumCPU()) {
				color = red
			} else if loadAvg.Last1Min > float64(runtime.NumCPU())/2 {
				color = orange
			} else {
				color = green
			}

			segments = append(segments,
				segment{
					Text:          "load",
					Color:         white,
					SeperateAfter: false,
				},
				segment{
					Text: fmt.Sprintf(
						"%.2f %.2f %.2f",
						loadAvg.Last1Min, loadAvg.Last5Min, loadAvg.Last15Min,
					),
					Color:         color,
					SeperateAfter: true,
				})
		}

		// Clock
		segments = append(segments, segment{
			Text:          time.Now().Format("2006-01-02 15:04:05"),
			Color:         white,
			SeperateAfter: true,
		})

		// Print the array
		out, err = json.Marshal(segments)
		if err != nil {
			println(`,[{"full_text":"error: ` + err.Error() + `"}]`)
		}
		println("," + string(out))
		time.Sleep(sleepDur - time.Since(startTime))
	}
}

// For some reason the default println prints to stderr, so we override it
func println(s string) {
	os.Stdout.Write(append([]byte(s), byte('\n')))
}

// FormatData converts a raw amount of bytes to an easily readable string
func formatData(v uint64) string {
	var fmtSize = func(n float64, u string) string {
		var f string
		if n > 100 {
			f = "%5.1f"
		} else if n > 10 {
			f = "%5.2f"
		} else {
			f = "%5.3f"
		}
		return fmt.Sprintf(f+" "+u, n)
	}
	if v > 1e12 {
		return fmtSize(float64(v)/1e12, "TB")
	} else if v > 1e9 {
		return fmtSize(float64(v)/1e9, "GB")
	} else if v > 1e6 {
		return fmtSize(float64(v)/1e6, "MB")
	} else if v > 1e3 {
		return fmtSize(float64(v)/1e3, "kB")
	}
	return fmt.Sprintf("%5d  B", v)
}
