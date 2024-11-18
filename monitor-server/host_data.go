package main

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"sync"
	"time"
)

type HostData struct {
	Hostname             string  `json:"hostname"`
	IP                   string  `json:"ip"`
	NodeIP               string  `json:"node_ip"`
	Label                string  `json:"label"`
	NameSpace            string  `json:"name_space"`
	CPUUsage             float64 `json:"cpu_usage"`
	MemoryUsage          float64 `json:"memory_usage"`
	DiskUsage            float64 `json:"disk_usage"`
	CPUCores             int     `json:"cpu_cores"`    // CPU 核心数
	TotalMemory          uint64  `json:"total_memory"` // 总内存
	TotalDisk            uint64  `json:"total_disk"`   // 总磁盘大小
	NetworkIO            float64 `json:"network_io"`
	ReadWriteIO          float64 `json:"read_write_io"`
	NetConnCount         int     `json:"net_conn_count"` // 网络连接数
	Timestamp            int64   `json:"timestamp"`
	LastOfflineAlertTime int64   `json:"last_offline_alert_time"` // 最新离线报警时间
	Status               string  `json:"status"`
	ActuatorMetrics      map[string]interface{}
}

var DataStore = make(map[string]*HostData)
var storeLock = sync.RWMutex{}

// 更新主机数据
func updateHostData(newData *HostData) {
	storeLock.Lock()
	defer storeLock.Unlock()

	// 获取现有数据
	existingData, exists := DataStore[newData.IP]
	if !exists {
		// 如果是新主机，创建新记录
		DataStore[newData.IP] = &HostData{
			IP:                   newData.IP,
			Hostname:             newData.Hostname,
			Status:               "online",
			LastOfflineAlertTime: 0,
		}
		existingData = DataStore[newData.IP]
	}

	// 只更新监控指标相关字段
	existingData.Timestamp = newData.Timestamp
	existingData.CPUUsage = newData.CPUUsage
	existingData.MemoryUsage = newData.MemoryUsage
	existingData.DiskUsage = newData.DiskUsage
	existingData.CPUCores = newData.CPUCores
	existingData.TotalMemory = newData.TotalMemory
	existingData.TotalDisk = newData.TotalDisk
	existingData.NetworkIO = newData.NetworkIO
	existingData.ReadWriteIO = newData.ReadWriteIO
	existingData.NetConnCount = newData.NetConnCount
}

// 处理主机数据上报
func handleHostData(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var hostData HostData
	if err := json.NewDecoder(r.Body).Decode(&hostData); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	// 设置时间戳
	hostData.Timestamp = time.Now().Unix()

	// 更新主机数据
	updateHostData(&hostData)

	storeLock.RLock()
	checkAlerts(hostData) // 检查警报
	storeLock.RUnlock()

	fmt.Fprintf(w, "Data received for host: %s", hostData.IP)
}

// 获取仪表盘数据
func handleDashboard(w http.ResponseWriter, r *http.Request) {
	storeLock.RLock()
	hostLock.RLock()
	defer storeLock.RUnlock()
	defer hostLock.RUnlock()

	var hosts []HostData
	for _, hostData := range DataStore {
		// 复制一份数据，避免直接暴露内部数据
		host := *hostData
		if label, exists := HostManage[host.IP]; exists {
			host.Label = label.Label
		}
		hosts = append(hosts, host)
	}

	json.NewEncoder(w).Encode(hosts)
}

func handlePodDashboard(w http.ResponseWriter, r *http.Request) {

	// 解析请求参数
	nodeIP := r.URL.Query().Get("host_ip")
	nameSpace := r.URL.Query().Get("namespace")

	var filteredList []HostData
	for _, item := range podMetricsList {
		if (nodeIP == "" || strings.Contains(item.NodeIP, nodeIP)) &&
			(nameSpace == "" || strings.Contains(item.NameSpace, nameSpace)) {
			filteredList = append(filteredList, item)
		}
	}
	json.NewEncoder(w).Encode(filteredList)
}

func handleActuatorDashboard(w http.ResponseWriter, r *http.Request) {

	// 解析请求参数
	nodeIP := r.URL.Query().Get("host_ip")
	nameSpace := r.URL.Query().Get("namespace")

	var filteredList []HostData
	for _, item := range actuatorList {
		if (nodeIP == "" || strings.Contains(item.NodeIP, nodeIP)) &&
			(nameSpace == "" || strings.Contains(item.NameSpace, nameSpace)) {
			filteredList = append(filteredList, item)
		}
	}
	json.NewEncoder(w).Encode(filteredList)
}
