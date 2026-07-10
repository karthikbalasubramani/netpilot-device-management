package health

import (
	"fmt"
	"time"

	"github.com/shirou/gopsutil/v4/cpu"
	"github.com/shirou/gopsutil/v4/disk"
	"github.com/shirou/gopsutil/v4/host"
	"github.com/shirou/gopsutil/v4/mem"
)

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

func GetSystemInfoHealth() (*SystemState, error) {
	// CPU Percentage
	cpuPercentages, err := cpu.Percent(500*time.Millisecond, false)
	if err != nil {
		return nil, fmt.Errorf("Failed to get CPU usage: %w", err)
	}
	// Memory Stats
	memoryStats, err := mem.VirtualMemory()
	if err != nil {
		return nil, fmt.Errorf("failed to get memory usage: %w", err)
	}
	// Disk Stats
	diskStats, err := disk.Usage("/")
	if err != nil {
		return nil, fmt.Errorf("failed to get disk usage: %w", err)
	}
	// Host stats
	hostStats, err := host.Info()
	if err != nil {
		return nil, fmt.Errorf("failed to get host info: %w", err)
	}
	// CPU Usage
	cpuUsage := 0.0
	if len(cpuPercentages) > 0 {
		cpuUsage = cpuPercentages[0]
	}
	if cpuUsage >= 60 {
		return nil, fmt.Errorf("CPU Usage Percentage is greater than 60%%")
	}
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

func bytesToMB(bytes uint64) uint64 {
	return bytes / 1024 / 1024
}

func bytesToGB(bytes uint64) uint64 {
	return bytes / 1024 / 1024 / 1024
}

func round(value float64) float64 {
	return float64(int(value*100)) / 100
}
