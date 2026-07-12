package health

import (
	"fmt"
	"time"

	"github.com/shirou/gopsutil/v4/cpu"
	"github.com/shirou/gopsutil/v4/disk"
	"github.com/shirou/gopsutil/v4/host"
	"github.com/shirou/gopsutil/v4/mem"
)

// SystemState represents the current system resource usage returned by the health API.
type SystemState struct {
	CPUUsagePercent    float64 `json:"cpu_usage_percent"`
	MemoryUsagePercent float64 `json:"memory_usage_percent"`
	MemoryTotalMB      uint64  `json:"memory_total_mb"`
	MemoryUsedMB       uint64  `json:"memory_used_mb"`
	DiskUsagePercent   float64 `json:"disk_usage_percent"`
	DiskTotalGB        uint64  `json:"disk_total_gb"`
	DiskUsedGB         uint64  `json:"disk_used_gb"`
	UptimeSeconds      uint64  `json:"uptime_seconds"`
}

// GetSystemInfoHealth collects CPU, memory, disk, and uptime details of the host system.
// It returns an error if any system metric cannot be collected or if CPU usage crosses the threshold.
func GetSystemInfoHealth(cpuThresholdPercent float64) (*SystemState, error) {
	// Collect CPU usage percentage with a small sampling interval.
	cpuPercentages, err := cpu.Percent(500*time.Millisecond, false)
	if err != nil {
		return nil, fmt.Errorf("failed to get CPU usage: %w", err)
	}

	// Collect virtual memory statistics.
	memoryStats, err := mem.VirtualMemory()
	if err != nil {
		return nil, fmt.Errorf("failed to get memory usage: %w", err)
	}

	// Collect disk usage statistics for the root disk path.
	diskStats, err := disk.Usage("/")
	if err != nil {
		return nil, fmt.Errorf("failed to get disk usage: %w", err)
	}

	// Collect host information, including system uptime.
	hostStats, err := host.Info()
	if err != nil {
		return nil, fmt.Errorf("failed to get host info: %w", err)
	}

	// Read CPU usage from the collected CPU percentage list.
	cpuUsage := 0.0
	if len(cpuPercentages) > 0 {
		cpuUsage = cpuPercentages[0]
	}

	// Mark health check as failed if CPU usage is above the configured threshold.
	if cpuUsage >= cpuThresholdPercent {
		return nil, fmt.Errorf("CPU usage percentage %.2f%% is greater than threshold %.2f%%", cpuUsage, cpuThresholdPercent)
	}

	// Return final system state after collecting and validating all metrics.
	return &SystemState{
		CPUUsagePercent:    round(cpuUsage),
		MemoryUsagePercent: round(memoryStats.UsedPercent),
		MemoryTotalMB:      bytesToMB(memoryStats.Total),
		MemoryUsedMB:       bytesToMB(memoryStats.Used),
		DiskUsagePercent:   round(diskStats.UsedPercent),
		DiskTotalGB:        bytesToGB(diskStats.Total),
		DiskUsedGB:         bytesToGB(diskStats.Used),
		UptimeSeconds:      hostStats.Uptime,
	}, nil
}

// bytesToMB converts bytes into megabytes.
func bytesToMB(bytes uint64) uint64 {
	return bytes / 1024 / 1024
}

// bytesToGB converts bytes into gigabytes.
func bytesToGB(bytes uint64) uint64 {
	return bytes / 1024 / 1024 / 1024
}

// round rounds a float value to two decimal places.
func round(value float64) float64 {
	return float64(int(value*100)) / 100
}
