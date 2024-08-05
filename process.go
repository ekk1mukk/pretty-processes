package main

import (
	"context"
	"fmt"
	"log"
	"sort"
	"strconv"
	"time"

	"github.com/charmbracelet/bubbles/list"
	"github.com/shirou/gopsutil/v4/process"
)

// processItem represents a single process entry in the list.
type processItem struct {
	pid          int32
	name         string
	cmdline      string
	ram          float32
	ramAmount    uint64
	cpu          float64
	ppid         int32
	creationDate int64
}

// Title returns the formatted title for the list processItem.
func (i processItem) Title() string { return fmt.Sprintf("(%d) %s", i.pid, i.name) }

// Description returns a formatted description for the list processItem, including RAM and CPU usage.
func (i processItem) Description() string {
	ramUsage := fmt.Sprintf("%.2f%%", i.ram)
	ramAmount := formatBytes(i.ramAmount)
	cpuUsage := fmt.Sprintf("%.2f%%", i.cpu)
	timeObj := time.Unix(i.creationDate/1000, 0)
	return fmt.Sprintf("CMD: %s\nRAM: %s (%s) | CPU: %s | PPID: %d | Creation date : %s", i.cmdline, ramUsage, ramAmount, cpuUsage, i.ppid, timeObj.Format("2006-01-01 15:04:05"))
}

// FilterValue returns the filterable string value for the list processItem.
func (i processItem) FilterValue() string { return strconv.Itoa(int(i.pid)) + i.name }

// getProcesses retrieves and returns a list of current processes.
func getProcesses() []list.Item {
	ctx := context.Background()
	processes, err := process.Processes()
	if err != nil {
		log.Fatalf("Error fetching processes: %s", err)
	}

	var processList []processItem

	for _, proc := range processes {
		name, err := proc.NameWithContext(ctx)
		if err != nil {
			log.Printf("Error fetching process name: %s", err)
			continue
		}

		cmdline, err := proc.CmdlineWithContext(ctx)
		if err != nil {
			log.Printf("Error fetching cmdline: %s", err)
			continue
		}

		ramPercentage, err := proc.MemoryPercentWithContext(ctx)
		if err != nil {
			log.Printf("Error fetching RAM usage: %s", err)
			continue
		}

		ramAmount, err := proc.MemoryInfoExWithContext(ctx)
		if err != nil {
			log.Printf("Error fetching RAM amount: %s", err)
			continue
		}

		cpuPercentage, err := proc.CPUPercentWithContext(ctx)
		if err != nil {
			log.Printf("Error fetching CPU usage: %s", err)
			continue
		}

		ppid, err := proc.PpidWithContext(ctx)
		if err != nil {
			log.Printf("Error fetching PPID: %s", err)
			continue
		}

		creationDate, err := proc.CreateTimeWithContext(ctx)
		if err != nil {
			log.Printf("Error fetching the creation date of processes: %s", err)
			continue
		}

		processList = append(processList, processItem{
			pid:          proc.Pid,
			name:         name,
			cmdline:      cmdline,
			ram:          ramPercentage,
			ramAmount:    ramAmount.RSS,
			cpu:          cpuPercentage,
			ppid:         ppid,
			creationDate: creationDate,
		})
	}

	sort.Slice(processList, func(i, j int) bool {
		return processList[i].pid > processList[j].pid
	})

	var processItems []list.Item

	for _, proc := range processList {
		processItems = append(processItems, proc)
	}

	return processItems
}
