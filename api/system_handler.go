package api

import (
	"net/http"
	"os"
	"runtime"
	"syscall"
	"time"

	"github.com/elastic/go-sysinfo"
	"github.com/gin-gonic/gin"
	"github.com/shirou/gopsutil/v3/cpu"
	gopsHost "github.com/shirou/gopsutil/v3/host"
	"github.com/docker/docker/client"
	"github.com/docker/docker/api/types"
	"context"
)

type DockerContainer struct {
	ID      string `json:"id"`
	Name    string `json:"name"`
	Image   string `json:"image"`
	State   string `json:"state"`
	Status  string `json:"status"`
	Created int64  `json:"created"`
}

// SystemStats System statistics structure
type SystemStats struct {
	OS           string  `json:"os"`
	Arch         string  `json:"arch"`
	NumCPU       int     `json:"num_cpu"`
	GoRoutines   int     `json:"go_routines"`
	MemoryUsed   uint64  `json:"memory_used"`   // Bytes
	MemoryTotal  uint64  `json:"memory_total"`  // Bytes
	MemoryUsage  float64 `json:"memory_usage"`  // Percentage
	DiskTotal    uint64  `json:"disk_total"`    // Bytes
	DiskUsed     uint64  `json:"disk_used"`     // Bytes
	DiskUsage    float64 `json:"disk_usage"`    // Percentage
	CPULoad      float64 `json:"cpu_load"`      // Percentage (last 1 min)
	CPUTemp      float64 `json:"cpu_temp"`      // Celsius
	Uptime       uint64  `json:"uptime"`        // Seconds
	HostName     string  `json:"host_name"`
	Platform     string  `json:"platform"`
	Kernel       string  `json:"kernel"`
	Containers   []DockerContainer `json:"containers"`
}

// handleGetSystemStats Get system statistics
func (s *Server) handleGetSystemStats(c *gin.Context) {
	host, err := sysinfo.Host()
	if err != nil {
		// Fallback to basic runtime stats if sysinfo fails
		var m runtime.MemStats
		runtime.ReadMemStats(&m)
		
		c.JSON(http.StatusOK, SystemStats{
			OS:          runtime.GOOS,
			Arch:        runtime.GOARCH,
			NumCPU:      runtime.NumCPU(),
			GoRoutines:  runtime.NumGoroutine(),
			MemoryUsed:  m.Alloc,
			MemoryTotal: m.Sys, // Not exactly total RAM, but total obtained from OS
			MemoryUsage: 0,     // Cannot determine without total RAM
		})
		return
	}

	info := host.Info()
	memory, err := host.Memory()
	
	var memUsed, memTotal uint64
	var memUsage float64
	
	if err == nil {
		memTotal = memory.Total
		memUsed = memory.Used
		if memTotal > 0 {
			memUsage = float64(memUsed) / float64(memTotal) * 100
		}
	} else {
		// Fallback for memory
		var m runtime.MemStats
		runtime.ReadMemStats(&m)
		memUsed = m.Alloc
	}

	// Get Disk Usage
	diskTotal, diskUsed := getDiskUsage()
	var diskUsage float64
	if diskTotal > 0 {
		diskUsage = float64(diskUsed) / float64(diskTotal) * 100
	}
	
	// Get CPU Load (Load Average)
	cpuLoad := 0.0
	percent, err := cpu.Percent(0, false)
	if err == nil && len(percent) > 0 {
		cpuLoad = percent[0]
	}

	// Get CPU Temperature
	// Note: Temperature readings vary significantly by OS and hardware support.
	// On macOS, it often requires specific tools or CGO/IOKit which might not work out of box with gopsutil pure go.
	// We'll try to get it, but expect 0 if not supported.
	cpuTemp := 0.0
	temps, err := gopsHost.SensorsTemperatures()
	if err == nil && len(temps) > 0 {
		// Try to find a CPU temperature sensor
		for _, t := range temps {
			// Common sensor keys often contain "cpu", "core", "package", or "k10temp" (amd), "coretemp" (intel)
			// For now just take the first one or the highest one found if we want to be simple
			if t.Temperature > cpuTemp {
				cpuTemp = t.Temperature
			}
		}
	}

	// Get Docker Containers
	var containers []DockerContainer = []DockerContainer{} // Initialize as empty array instead of nil
	
	// Create client with default options and environment variables
	cli, err := client.NewClientWithOpts(client.FromEnv, client.WithAPIVersionNegotiation())
	
	if err == nil {
		defer cli.Close()
		
		// List containers
		dockerContainers, err := cli.ContainerList(context.Background(), types.ContainerListOptions{All: true})
		if err == nil {
			for _, c := range dockerContainers {
				name := ""
				if len(c.Names) > 0 {
					name = c.Names[0]
					// Remove leading slash if present
					if len(name) > 0 && name[0] == '/' {
						name = name[1:]
					}
				}
				
				containers = append(containers, DockerContainer{
					ID:      c.ID[:12],
					Name:    name,
					Image:   c.Image,
					State:   c.State,
					Status:  c.Status,
					Created: c.Created,
				})
			}
		} else {
			// Try without options if environment variables fail or socket is different
			// This is a fallback
		}
	} else {
		// Log error
	}

	c.JSON(http.StatusOK, SystemStats{
		OS:          info.OS.Name,
		Arch:        info.Architecture,
		NumCPU:      runtime.NumCPU(),
		GoRoutines:  runtime.NumGoroutine(),
		MemoryUsed:  memUsed,
		MemoryTotal: memTotal,
		MemoryUsage: memUsage,
		DiskTotal:   diskTotal,
		DiskUsed:    diskUsed,
		DiskUsage:   diskUsage,
		CPULoad:     cpuLoad,
		CPUTemp:     cpuTemp,
		Uptime:      uint64(time.Since(info.BootTime).Seconds()),
		HostName:    info.Hostname,
		Platform:    info.OS.Platform,
		Kernel:      info.KernelVersion,
		Containers:  containers,
	})
}

func getDiskUsage() (uint64, uint64) {
	var stat syscall.Statfs_t
	wd, err := os.Getwd()
	if err != nil {
		return 0, 0
	}
	if err := syscall.Statfs(wd, &stat); err != nil {
		return 0, 0
	}
	
	// Convert to uint64 to be safe across platforms
	total := uint64(stat.Blocks) * uint64(stat.Bsize)
	free := uint64(stat.Bfree) * uint64(stat.Bsize)
	used := total - free
	
	return total, used
}
