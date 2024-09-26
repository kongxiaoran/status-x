package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net"
	"net/http"
	"os"
	"time"

	"github.com/shirou/gopsutil/cpu"
	"github.com/shirou/gopsutil/disk"
	"github.com/shirou/gopsutil/mem"
)

// 主机监控数据结构
type HostData struct {
	Hostname     string  `json:"hostname"`
	IP           string  `json:"ip"`
	CPUUsage     float64 `json:"cpu_usage"`
	MemoryUsage  float64 `json:"memory_usage"`
	DiskUsage    float64 `json:"disk_usage"`
	NetworkUsage float64 `json:"network_usage"`
	Timestamp    int64   `json:"timestamp"`
}

var IpAdress = ""
var InfluxURL = "10.10.18.116:8086"
var ServerURL = "localhost:8080"
var InfluxToken = "wh56EgkTNCyt-oSz_4Uo8l_SYy9R57CnUFy2NZY4bxmjZ9bbBNiMvQ0kdo8W4cwdvP6JrgXY49uXpTI7d5mRtA=="

// 获取监控数据
func getHostData() (HostData, error) {
	cpuUsage, _ := cpu.Percent(0, false)
	memStat, _ := mem.VirtualMemory()
	diskStat, _ := disk.Usage("/")
	ipAddress := IpAdress

	return HostData{
		IP:          ipAddress, // 可以使用函数获取真实 IP
		CPUUsage:    cpuUsage[0],
		MemoryUsage: memStat.UsedPercent,
		DiskUsage:   diskStat.UsedPercent,
	}, nil
}

// 将监控数据发送到 InfluxDB
func sendDataToInfluxDB(data HostData) {
	influxDBURL := "http://" + InfluxURL + "/write?db=finchina-dev"
	influxData := fmt.Sprintf(
		"host_metrics,host=%s cpu_usage=%.2f,memory_usage=%.2f,disk_usage=%.2f",
		data.IP, data.CPUUsage, data.MemoryUsage, data.DiskUsage)

	req, err := http.NewRequest("POST", influxDBURL, bytes.NewBuffer([]byte(influxData)))
	if err != nil {
		fmt.Println("Failed to create request:", err)
		return
	}
	req.Header.Set("Authorization", "Token "+InfluxToken) // 使用 Token 进行认证
	req.Header.Set("Content-Type", "text/plain")          // InfluxDB 2.x 使用 "text/plain" 类型
	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		fmt.Println("Error sending data to InfluxDB:", err)
		return
	}
	defer resp.Body.Close()
	fmt.Println("Data sent to InfluxDB, response status:", resp.Status)
}

// 将监控数据发送到服务端
func sendDataToServer(data HostData) {
	jsonData, err := json.Marshal(data)
	if err != nil {
		fmt.Println("Error marshaling host data:", err)
		return
	}

	resp, err := http.Post("http://"+ServerURL+"/api/host-data", "application/json", bytes.NewBuffer(jsonData))
	if err != nil {
		fmt.Println("Error sending data to server:", err)
		return
	}
	defer resp.Body.Close()
	fmt.Println("Data sent to server, response status:", resp.Status)
}

// 初始化，加载环境变量
func init() {
	if influxTokenEnv := os.Getenv("INFLUX_TOKEN"); influxTokenEnv != "" {
		InfluxToken = influxTokenEnv
	}
	if influxURLEnv := os.Getenv("INFLUX_URL"); influxURLEnv != "" {
		InfluxURL = influxURLEnv
	}
	if serverURLEnv := os.Getenv("SERVER_URL"); serverURLEnv != "" {
		ServerURL = serverURLEnv
	}
}

func main() {
	IpAdress, _ = getIPv4Address()
	for {
		hostData, err := getHostData()
		if err != nil {
			fmt.Println("Error getting host data:", err)
			continue
		}

		// 将数据发送到 InfluxDB
		sendDataToInfluxDB(hostData)

		// 将数据发送到服务端
		sendDataToServer(hostData)

		time.Sleep(1 * time.Second) // 每 10 秒采集一次数据
	}
}

func getIPv4Address() (string, error) {
	interfaces, err := net.Interfaces()
	if err != nil {
		return "", err
	}

	for _, i := range interfaces {
		addrs, err := i.Addrs()
		if err != nil {
			return "", err
		}

		for _, addr := range addrs {
			var ip net.IP
			switch v := addr.(type) {
			case *net.IPNet:
				ip = v.IP
			case *net.IPAddr:
				ip = v.IP
			}

			// 过滤出 IPv4 地址，并排除回环地址 (127.0.0.1)
			if ip != nil && ip.To4() != nil && !ip.IsLoopback() {
				return ip.String(), nil
			}
		}
	}
	return "", fmt.Errorf("no IPv4 address found")
}
