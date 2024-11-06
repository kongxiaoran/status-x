package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"sync"
	"time"
)

type AlertConfig struct {
	CPUThreshold    float64 `json:"cpu_threshold"`
	MemoryThreshold float64 `json:"memory_threshold"`
	DiskThreshold   float64 `json:"disk_threshold"`
	CPUDuration     int     `json:"cpu_duration"`
	MemoryDuration  int     `json:"memory_duration"`
	Success         bool    `json:"success"`
}

var alertConfig = AlertConfig{
	CPUThreshold:    90.0,
	MemoryThreshold: 85.0,
	DiskThreshold:   85.0,
	CPUDuration:     20, // CPU 超过阈值持续时间
	MemoryDuration:  10, // 内存超过阈值持续时间
	Success:         false,
}

type AlertStatus struct {
	LastAlertTime time.Time     // 最近一次发送警报的时间
	Count         time.Duration // 超过阈值的累计时间
}

var alertTimers = make(map[string]map[string]*AlertStatus) // 存储每个主机的警报状态 // 存储每个主机的 CPU 和内存定时器
var alertLock = sync.Mutex{}

// 新增的处理函数，用于获取当前系统的警报指标
func handleAlertMetrics(w http.ResponseWriter, r *http.Request) {
	storeLock.RLock()
	defer storeLock.RUnlock()
	alertConfig.Success = true
	// 返回当前的警报配置
	json.NewEncoder(w).Encode(alertConfig)
}

func handleAlertConfig(w http.ResponseWriter, r *http.Request) {
	if r.Method != "POST" {
		http.Error(w, "Only POST allowed", http.StatusMethodNotAllowed)
		return
	}

	var newConfig AlertConfig
	err := json.NewDecoder(r.Body).Decode(&newConfig)
	if err != nil {
		http.Error(w, "Invalid request payload", http.StatusBadRequest)
		return
	}

	// 更新报警配置到数据库
	err = saveAlertConfigToDB(newConfig)
	if err != nil {
		http.Error(w, "Failed to update alert config", http.StatusInternalServerError)
		return
	}

	alertConfig = newConfig // 更新警报配置

	response := map[string]bool{"success": true}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(response)
}

// 发送 HTTP POST 请求到接口
func SendAlertToHttp(content string) {
	//url := "http://222.73.12.12:8008/api/SendAppMsg"
	//payload := map[string]interface{}{
	//	"touser": receive,
	//	"text": map[string]string{
	//		"content": content,
	//	},
	//}

	//url := "http://222.73.12.12:8008/api/SendChatMsg"
	//payload := map[string]interface{}{
	//	"chatid": 267,
	//	"text": map[string]string{
	//		"content": content,
	//	},
	//}

	url := "https://qyapi.weixin.qq.com/cgi-bin/webhook/send?key=b28f4f07-bc17-4a5a-815b-89f8ca485ba2"

	payload := map[string]interface{}{
		"msgtype": "markdown",
		"markdown": map[string]interface{}{
			"content": content,
		},
	}

	jsonData, err := json.Marshal(payload)
	if err != nil {
		fmt.Println("Failed to encode JSON:", err)
		return
	}

	req, err := http.NewRequest("POST", url, bytes.NewBuffer(jsonData))
	if err != nil {
		fmt.Println("Failed to create HTTP request:", err)
		return
	}

	req.Header.Set("Content-Type", "application/json")

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		fmt.Println("Failed to send HTTP request:", err)
		return
	}
	defer resp.Body.Close()

	body, _ := ioutil.ReadAll(resp.Body)
	fmt.Println("HTTP Response:", string(body))
}

func SendAlert(hostIP string, resourceType string) {

	var msg = ""
	switch resourceType {
	case "cpu":
		msg = fmt.Sprintf("cpu 持续 %d分钟 占用超过 %0.1f%%, 触发预警阈值", alertConfig.CPUDuration, alertConfig.CPUThreshold)
	case "memory":
		msg = fmt.Sprintf("内存 持续 %d分钟 占用超过 %0.1f%%, 触发预警阈值", alertConfig.MemoryDuration, alertConfig.MemoryThreshold)
	case "disk":
		msg = fmt.Sprintf("磁盘 占用超过 %0.1f%%, 触发预警阈值", alertConfig.DiskThreshold)
	case "offline":
		msg = fmt.Sprintf("离线/关机 请注意检查")
	default:

	}
	name := hostIP
	currentHost, exists := HostManage[hostIP]
	if exists && currentHost.Label != "" {
		name = name + " | " + currentHost.Label
	}
	owner := "all"
	if exists && currentHost.Owner != "" {
		owner = currentHost.Owner
	}
	alertMessage := fmt.Sprintf("[中台服务器监控](http://10.15.97.66:42800/)\n发生告警：主机 [%s](http://10.15.97.66:42800/), %s\n<@%s>", name, msg, owner)
	SendAlertToHttp(alertMessage)
	log.Println(alertMessage)
}

func checkAlerts(host HostData) {
	alertLock.Lock()
	defer alertLock.Unlock()

	if _, exists := alertTimers[host.IP]; !exists {
		alertTimers[host.IP] = make(map[string]*AlertStatus)
	}

	// 检查 CPU 使用率
	checkResource(host, "cpu", time.Duration(alertConfig.CPUDuration)*time.Minute)

	// 检查内存使用率
	checkResource(host, "memory", time.Duration(alertConfig.MemoryDuration)*time.Minute)

	// 检查磁盘使用率
	checkResource(host, "disk", 0) // 磁盘使用率不检查持续时间，直接发送警报
}

func checkResource(host HostData, resourceType string, threshold time.Duration) {
	if _, exists := alertTimers[host.IP][resourceType]; !exists {
		alertTimers[host.IP][resourceType] = &AlertStatus{LastAlertTime: time.Unix(0, 0), Count: 0}
	}

	alertStatus := alertTimers[host.IP][resourceType]

	if threshold == 0 { // 特殊处理磁盘使用率
		if host.DiskUsage > alertConfig.DiskThreshold {
			if alertStatus.LastAlertTime.IsZero() || time.Since(alertStatus.LastAlertTime) >= time.Hour {
				SendAlert(host.IP, "disk")
				alertStatus.LastAlertTime = time.Now()
			}
		} else {
			alertStatus.LastAlertTime = time.Unix(0, 0)
		}
	} else { // 处理 CPU 和内存使用率
		if usage := getHostUsage(host, resourceType); usage > alertConfig.GetThreshold(resourceType) {
			alertStatus.Count += 1 * time.Second
			if alertStatus.Count >= threshold {
				if alertStatus.LastAlertTime.IsZero() || time.Since(alertStatus.LastAlertTime) >= 30*time.Minute {
					SendAlert(host.IP, resourceType)
					alertStatus.LastAlertTime = time.Now()
					alertStatus.Count = 0 // 重置计数器
				}
			}
		} else {
			alertStatus.Count = 0 // 重置计数器
		}
	}
}

func getHostUsage(host HostData, resourceType string) float64 {
	switch resourceType {
	case "cpu":
		return host.CPUUsage
	case "memory":
		return host.MemoryUsage
	case "disk":
		return host.DiskUsage
	default:
		return 0
	}
}

func (a *AlertConfig) GetThreshold(resourceType string) float64 {
	switch resourceType {
	case "cpu":
		return a.CPUThreshold
	case "memory":
		return a.MemoryThreshold
	case "disk":
		return a.DiskThreshold
	default:
		return 0
	}
}

func (a *AlertConfig) GetDuration(resourceType string) time.Duration {
	switch resourceType {
	case "cpu":
		return time.Duration(a.CPUDuration) * time.Second
	case "memory":
		return time.Duration(a.MemoryDuration) * time.Second
	default:
		return 0
	}
}
