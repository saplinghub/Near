package main

import (
	"os/exec"
	"strconv"
	"strings"

	"github.com/shirou/gopsutil/v3/cpu"
	"github.com/shirou/gopsutil/v3/mem"
)

// SystemStats 系统统计信息
type SystemStats struct {
	CPU         float64 `json:"cpu"`
	Memory      float64 `json:"memory"`
	Temperature string  `json:"temperature"`
	TotalMemory uint64  `json:"total_memory"`
	UsedMemory  uint64  `json:"used_memory"`
}

// SystemService 系统服务
type SystemService struct{}

// NewSystemService 创建系统服务
func NewSystemService() *SystemService {
	return &SystemService{}
}

// GetSystemStats 获取系统统计信息
func (a *App) GetSystemStats() (*SystemStats, error) {
	// CPU 使用率
	cpuPercent, err := cpu.Percent(0, false)
	cpuUsage := 0.0
	if err == nil && len(cpuPercent) > 0 {
		cpuUsage = cpuPercent[0]
	}

	// 内存使用
	vmStat, err := mem.VirtualMemory()
	memoryUsage := 0.0
	totalMemory := uint64(0)
	usedMemory := uint64(0)
	if err == nil {
		memoryUsage = vmStat.UsedPercent
		totalMemory = vmStat.Total
		usedMemory = vmStat.Used
	}

	// 温度（macOS）
	temperature := a.systemService.getMacOSTemperature()

	return &SystemStats{
		CPU:         cpuUsage,
		Memory:      memoryUsage,
		Temperature: temperature,
		TotalMemory: totalMemory,
		UsedMemory:  usedMemory,
	}, nil
}

// getMacOSTemperature 获取 macOS 温度
func (s *SystemService) getMacOSTemperature() string {
	// 尝试从 ioreg 获取温度
	cmd := exec.Command("sh", "-c", "ioreg -l | grep '\"TC0C\"' | awk '{print $4}' | sed 's/[^0-9]//g'")
	output, err := cmd.Output()
	if err != nil {
		return "N/A"
	}

	tempStr := strings.TrimSpace(string(output))
	if tempStr == "" {
		return "N/A"
	}

	// 转换温度值
	temp, err := strconv.ParseFloat(tempStr, 64)
	if err != nil {
		return "N/A"
	}

	// 转换为摄氏度
	celsius := (temp - 273.15) / 10.0

	return strconv.FormatFloat(celsius, 'f', 1, 64) + "°C"
}

// GetCPUUsage 获取 CPU 使用率（用于动画）
func (s *SystemService) GetCPUUsage() float64 {
	cpuPercent, err := cpu.Percent(0, false)
	if err != nil || len(cpuPercent) == 0 {
		return 0.0
	}
	return cpuPercent[0]
}
