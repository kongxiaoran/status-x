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
	Hostname     string  `json:"hostname"`
	IP           string  `json:"ip"`
	NodeIP       string  `json:"node_ip"`
	NameSpace    string  `json:"name_space"`
	CPUUsage     float64 `json:"cpu_usage"`
	MemoryUsage  float64 `json:"memory_usage"`
	DiskUsage    float64 `json:"disk_usage"`
	NetworkIO    float64 `json:"network_io"`
	ReadWriteIO  float64 `json:"read_write_io"`
	NetConnCount int     `json:"net_conn_count"` // 网络连接数
	Timestamp    int64   `json:"timestamp"`
}

var dataStore = make(map[string]HostData)
var storeLock = sync.RWMutex{}

func handleHostData(w http.ResponseWriter, r *http.Request) {
	if r.Method != "POST" {
		http.Error(w, "Only POST allowed", http.StatusMethodNotAllowed)
		return
	}

	var hostData HostData
	err := json.NewDecoder(r.Body).Decode(&hostData)
	if err != nil {
		http.Error(w, "Invalid request payload", http.StatusBadRequest)
		return
	}

	storeLock.Lock()
	hostData.Timestamp = time.Now().Unix()
	dataStore[hostData.IP] = hostData
	storeLock.Unlock()

	checkAlerts(hostData) // 检查警报

	fmt.Fprintf(w, "Data received for host: %s", hostData.IP)
}

func handleDashboard(w http.ResponseWriter, r *http.Request) {
	storeLock.RLock()
	defer storeLock.RUnlock()

	var hosts []HostData
	for _, hostData := range dataStore {
		hosts = append(hosts, hostData)
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
