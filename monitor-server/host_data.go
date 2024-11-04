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
	ActuatorMetrics      map[string]interface{}
}

var dataStore = make(map[string]*HostData)
var storeLock = sync.RWMutex{}

func handleHostData(w http.ResponseWriter, r *http.Request) {
	if r.Method != "POST" {
		http.Error(w, "Only POST allowed", http.StatusMethodNotAllowed)
		return
	}

	var hostData HostData
	hostData.LastOfflineAlertTime = 0
	err := json.NewDecoder(r.Body).Decode(&hostData)
	if err != nil {
		http.Error(w, "Invalid request payload", http.StatusBadRequest)
		return
	}

	storeLock.Lock()
	hostData.Timestamp = time.Now().Unix()
	if _, exists := dataStore[hostData.IP]; exists {
		hostData.LastOfflineAlertTime = dataStore[hostData.IP].LastOfflineAlertTime
	}
	dataStore[hostData.IP] = &hostData
	storeLock.Unlock()

	checkAlerts(hostData) // 检查警报

	fmt.Fprintf(w, "Data received for host: %s", hostData.IP)
}

func handleDashboard(w http.ResponseWriter, r *http.Request) {
	storeLock.RLock()
	defer storeLock.RUnlock()

	var hosts []HostData
	for _, hostData := range dataStore {
		hostData.Label = HostManage[hostData.IP].Label
		hosts = append(hosts, *hostData)
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
