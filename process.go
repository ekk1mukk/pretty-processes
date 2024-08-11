package main

import (
	"context"
	"log"
	"sort"
	"strconv"

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

// FilterValue returns the filterable string value for the list processItem.
func (i processItem) FilterValue() string { return strconv.Itoa(int(i.pid)) + i.name }

// getProcesses retrieves and returns a list of current processes with some data.
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
