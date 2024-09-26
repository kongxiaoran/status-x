package main

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log"

	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/client"
)

// Stats 是用来解析 Docker API 返回的容器统计数据
type Stats struct {
	CPUStats    CPUStats                `json:"cpu_stats"`
	MemoryStats MemoryStats             `json:"memory_stats"`
	Networks    map[string]NetworkStats `json:"networks"`
	BlkioStats  BlkioStats              `json:"blkio_stats"`
}

// CPUStats 包含 CPU 使用的统计信息
type CPUStats struct {
	CPUUsage struct {
		TotalUsage  uint64   `json:"total_usage"`
		PercpuUsage []uint64 `json:"percpu_usage"`
	} `json:"cpu_usage"`
	SystemCPUUsage uint64 `json:"system_cpu_usage"`
}

// MemoryStats 包含内存使用的统计信息
type MemoryStats struct {
	Usage uint64 `json:"usage"`
}

// NetworkStats 包含网络流量的统计信息
type NetworkStats struct {
	RxBytes uint64 `json:"rx_bytes"`
	TxBytes uint64 `json:"tx_bytes"`
}

// BlkioStats 包含 I/O 使用的统计信息
type BlkioStats struct {
	IoServiceBytesRecursive []BlkioEntry `json:"io_service_bytes_recursive"`
}

// BlkioEntry 包含每次 I/O 操作的详细信息
type BlkioEntry struct {
	Op    string `json:"op"`
	Value uint64 `json:"value"`
}

func main() {
	apiClient, err := client.NewClientWithOpts(client.FromEnv, client.WithVersion("1.40"))
	if err != nil {
		panic(err)
	}
	defer apiClient.Close()

	containers, err := apiClient.ContainerList(context.Background(), container.ListOptions{All: false})
	if err != nil {
		panic(err)
	}

	for _, ctr := range containers {
		fmt.Printf("%s %s (status: %s)\n", ctr.ID, ctr.Image, ctr.Status)
		stats, err := apiClient.ContainerStats(context.Background(), ctr.ID, false)
		if err != nil {
			log.Fatalf("Error getting container stats: %v", err)
		}
		defer stats.Body.Close()

		// 解析资源使用信息
		if err := processStats(stats.Body); err != nil {
			log.Fatalf("Error processing stats: %v", err)
		}
	}
}

// processStats 解析并打印容器的资源使用信息
func processStats(statsBody io.ReadCloser) error {
	var stats Stats
	if err := json.NewDecoder(statsBody).Decode(&stats); err != nil {
		return err
	}

	// 打印 CPU 使用情况，转换为微核
	cpuUsageMicroCores := calculateCPUUsageMicroCores(stats.CPUStats)
	fmt.Printf("CPU Usage: %.2f μCores\n", cpuUsageMicroCores)

	// 打印内存使用情况，转换为 MB
	memoryUsageMB := float64(stats.MemoryStats.Usage) / (1024 * 1024)
	fmt.Printf("Memory Usage: %.2f MB\n", memoryUsageMB)

	// 打印网络使用情况，转换为 MB
	var rxMB, txMB float64
	for _, network := range stats.Networks {
		rxMB += float64(network.RxBytes) / (1024 * 1024)
		txMB += float64(network.TxBytes) / (1024 * 1024)
	}
	fmt.Printf("Network: Received %.2f MB, Transmitted %.2f MB\n", rxMB, txMB)

	// 打印 I/O 使用情况，转换为 MB
	var ioReadMB, ioWriteMB float64
	for _, entry := range stats.BlkioStats.IoServiceBytesRecursive {
		if entry.Op == "Read" {
			ioReadMB += float64(entry.Value) / (1024 * 1024)
		} else if entry.Op == "Write" {
			ioWriteMB += float64(entry.Value) / (1024 * 1024)
		}
	}
	fmt.Printf("I/O: Read %.2f MB, Write %.2f MB\n", ioReadMB, ioWriteMB)

	return nil
}

// calculateCPUUsageMicroCores 将 CPU 使用量转换为微核 (μCores)
func calculateCPUUsageMicroCores(cpuStats CPUStats) float64 {
	// 获取总的 CPU 使用量（单位是纳秒），并将其转换为微核
	// μCores = (Total CPU Usage / System CPU Usage) * 1,000,000
	if cpuStats.SystemCPUUsage == 0 {
		return 0
	}
	cpuUsageRatio := float64(cpuStats.CPUUsage.TotalUsage) / float64(cpuStats.SystemCPUUsage)
	return cpuUsageRatio * 1_000_000
}
