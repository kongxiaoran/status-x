package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"github.com/shirou/gopsutil/cpu"
	"github.com/shirou/gopsutil/disk"
	"github.com/shirou/gopsutil/mem"
	gonet "github.com/shirou/gopsutil/net"
	"math"
	"net"
	"net/http"
	"os"
	"strconv"
	"time"
)

// 主机监控数据结构
type HostData struct {
	Hostname     string  `json:"hostname"`
	IP           string  `json:"ip"`
	CPUUsage     float64 `json:"cpu_usage"`
	MemoryUsage  float64 `json:"memory_usage"`
	DiskUsage    float64 `json:"disk_usage"`
	NetworkIO    float64 `json:"network_io"`     // MB
	ReadWriteIO  float64 `json:"read_write_io"`  // MB
	NetConnCount int     `json:"net_conn_count"` // 网络连接数
	Timestamp    int64   `json:"timestamp"`
}

var IpAdress = ""
var InfluxURL = "10.10.18.116:8086"
var ServerURL = "localhost:12800"
var CollectFrequency = 1

// name:admin pass:adminadmin
var InfluxToken = "wh56EgkTNCyt-oSz_4Uo8l_SYy9R57CnUFy2NZY4bxmjZ9bbBNiMvQ0kdo8W4cwdvP6JrgXY49uXpTI7d5mRtA=="

var latestReadAndWriteIO = 0.0
var latestNetIO = 0.0

func getHostData() (HostData, error) {
	cpuUsage, _ := cpu.Percent(0, false)
	memStat, _ := mem.VirtualMemory()
	diskStat, _ := disk.Usage("/")
	netIOCounters, _ := gonet.IOCounters(false)
	connStats, _ := gonet.Connections("all")

	// 计算网络 I/O
	var totalNetworkIO float64
	for _, stat := range netIOCounters {
		totalNetworkIO += float64(stat.BytesSent + stat.BytesRecv)
	}

	netWorkIO := totalNetworkIO - latestNetIO
	latestNetIO = totalNetworkIO

	// 将字节转换为 MB
	netWorkIO /= 1024 * 1024 // 转换为 MB
	netWorkIO = round(netWorkIO, 2)

	// 获取读写 I/O（假设你要获取的为 disk I/O）
	diskReadWriteIO, _ := disk.IOCounters()
	var totalReadAndWriteIO float64
	for _, stat := range diskReadWriteIO {
		totalReadAndWriteIO += float64(stat.ReadBytes + stat.WriteBytes)
	}

	readAndWriteIO := totalReadAndWriteIO - latestReadAndWriteIO
	latestReadAndWriteIO = totalReadAndWriteIO

	readAndWriteIO /= 1024 * 1024 // 转换为 MB
	readAndWriteIO = round(readAndWriteIO, 2)

	return HostData{
		IP:           IpAdress,
		CPUUsage:     cpuUsage[0],
		MemoryUsage:  memStat.UsedPercent,
		DiskUsage:    diskStat.UsedPercent,
		NetworkIO:    netWorkIO,
		ReadWriteIO:  readAndWriteIO,
		NetConnCount: len(connStats),
	}, nil
}

func round(value float64, precision int) float64 {
	pow := math.Pow(10, float64(precision))
	return math.Round(value*pow) / pow
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
	//增加请求延迟打印延迟
	now := time.Now()
	resp, err := client.Do(req)
	cost := time.Since(now).Milliseconds()
	if err != nil {
		fmt.Println("Error sending data to InfluxDB:", err)
		return
	}
	defer resp.Body.Close()
	fmt.Println("Data sent to InfluxDB, response status: ", resp.Status, " cost:", cost, "ms")
}

// 将监控数据发送到服务端
func sendDataToServer(data HostData) {
	jsonData, err := json.Marshal(data)
	if err != nil {
		fmt.Println("Error marshaling host data:", err)
		return
	}
	now := time.Now()
	resp, err := http.Post("http://"+ServerURL+"/api/host-data", "application/json", bytes.NewBuffer(jsonData))
	cost := time.Since(now).Milliseconds()
	if err != nil {
		fmt.Println("Error sending data to server:", err)
		return
	}
	defer resp.Body.Close()
	fmt.Println("Data sent to server, response status:", resp.Status, " cost:", cost, "ms")
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
	if collectFrequencyEnv := os.Getenv("COLLECT_FREQUENCY"); collectFrequencyEnv != "" {
		tempEnv, err := strconv.Atoi(collectFrequencyEnv)
		if err != nil {
			fmt.Println("COLLECT_FREQUENCY 转换错误:", err)
			return
		}
		CollectFrequency = tempEnv
	}
}

func main() {
	IpAdress, _ = getIPv4Address()

	diskReadWriteIO, _ := disk.IOCounters()
	var totalReadAndWriteIO float64
	for _, stat := range diskReadWriteIO {
		totalReadAndWriteIO += float64(stat.ReadBytes)
		totalReadAndWriteIO += float64(stat.WriteBytes)
	}
	latestReadAndWriteIO = totalReadAndWriteIO

	// 计算网络 I/O
	var totalNetworkIO float64
	netIOCounters, _ := gonet.IOCounters(false)
	for _, stat := range netIOCounters {
		totalNetworkIO += float64(stat.BytesSent + stat.BytesRecv)
	}
	latestNetIO = totalNetworkIO

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

		time.Sleep(time.Duration(CollectFrequency) * time.Second) // 每 CollectFrequency 秒采集一次数据
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
