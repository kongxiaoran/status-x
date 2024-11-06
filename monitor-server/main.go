package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"sync"
	"time"

	"github.com/gorilla/handlers"
	"github.com/gorilla/websocket"
)

var PodCollectFrequency = 3
var ActuatorFrequency = 2

var upgrader = websocket.Upgrader{
	CheckOrigin: func(r *http.Request) bool {
		return true // 允许所有来源的WebSocket连接
	},
}

// 添加WebSocket连接管理
var (
	clients    = make(map[*websocket.Conn]bool)
	clientsMux sync.Mutex
)

func init() {
	if influxTokenEnv := os.Getenv("INFLUX_TOKEN"); influxTokenEnv != "" {
		InfluxToken = influxTokenEnv
	}
	if influxURLEnv := os.Getenv("INFLUX_URL"); influxURLEnv != "" {
		InfluxURL = influxURLEnv
	}
	if influxOrgEnv := os.Getenv("INFLUX_ORG"); influxOrgEnv != "" {
		Org = influxOrgEnv
	}
	if influxBucketEnv := os.Getenv("INFLUX_BUCKET"); influxBucketEnv != "" {
		Bucket = influxBucketEnv
	}
	if podCollectFrequency := os.Getenv("POD_COLLECT_FREQUENCY"); podCollectFrequency != "" {
		tempEnv, err := strconv.Atoi(podCollectFrequency)
		if err != nil {
			fmt.Println("POD_COLLECT_FREQUENCY 转换错误:", err)
			return
		}
		PodCollectFrequency = tempEnv
	}

	if actuatorFrequency := os.Getenv("ACTUATOR_FREQUENCY"); actuatorFrequency != "" {
		tempEnv, err := strconv.Atoi(actuatorFrequency)
		if err != nil {
			fmt.Println("ACTUATOR_FREQUENCY 转换错误:", err)
			return
		}
		ActuatorFrequency = tempEnv
	}
}

func main() {

	// 初始化 MySQL 数据库
	initDB()
	// 从数据库加载报警配置
	loadAlertConfigFromDB()
	// 从数据库加载主机信息
	loadHostsFromDB()

	clientset, metricsClientset := initKubernetesClient()

	// 定时获取Pod资源数据并写入InfluxDB
	go func() {
		for {
			podMetrics, err := getPodMetrics(clientset, metricsClientset)
			podMetricsListMu.Lock()
			podMetricsList = podMetrics
			podMetricsListMu.Unlock()

			if err == nil {
				writePodMetricsToInflux(podMetrics)
			}
			time.Sleep(time.Duration(PodCollectFrequency) * time.Second) // 每 POD_COLLECT_FREQUENCY 秒 获取一次数据
		}
	}()

	go func() {
		for {
			hostLock.RLock()  // 添加 HostManage 的读锁
			storeLock.RLock() // 添加 DataStore 的读锁
			for _, host := range HostManage {
				currentTime := time.Now().Unix()
				if _, exists := DataStore[host.IPAddress]; !exists {
					continue
				}
				latestTime := DataStore[host.IPAddress].Timestamp
				lastOfflineTime := DataStore[host.IPAddress].LastOfflineAlertTime
				// 触发离线
				if currentTime-latestTime > 60 && host.AlertEnabled {
					// 判断是否重复报警
					if currentTime-lastOfflineTime > 600 || lastOfflineTime == 0 {
						storeLock.RUnlock() // 临时释放读锁
						storeLock.Lock()    // 获取写锁
						SendAlert(host.IPAddress, "offline")
						DataStore[host.IPAddress].LastOfflineAlertTime = currentTime
						storeLock.Unlock() // 释放写锁
						storeLock.RLock()  // 重新获取读锁
					}
				}
			}
			storeLock.RUnlock() // 释放 DataStore 的读锁
			hostLock.RUnlock()  // 释放 HostManage 的读锁
			time.Sleep(time.Duration(ActuatorFrequency) * time.Second)
		}
	}()

	//go func() {
	//	for {
	//		if podMetricsList != nil {
	//			var actuatorHostTemp []HostData
	//			for _, pod := range podMetricsList {
	//				if strings.Contains(pod.Hostname, "information") {
	//					actuator := getPodActuator(pod.IP)
	//					podMetricsListMu.Lock()
	//					pod.ActuatorMetrics = actuator
	//					actuatorHostTemp = append(actuatorHostTemp, pod)
	//					podMetricsListMu.Unlock()
	//				}
	//			}
	//			actuatorList = actuatorHostTemp
	//			writeApplicationMetricsToInflux(actuatorList)
	//		}
	//		time.Sleep(time.Duration(ActuatorFrequency) * time.Second) // 每 ActuatorFrequency 获取一次数据
	//	}
	//}()

	http.HandleFunc("/api/host-data", handleHostData)
	http.HandleFunc("/api/dashboard", handleDashboard)
	http.HandleFunc("/api/pod-dashboard", handlePodDashboard)
	http.HandleFunc("/api/actuator-dashboard", handleActuatorDashboard)
	http.HandleFunc("/api/host-metrics", handleHostMetrics)
	http.HandleFunc("/api/pod-metrics", handlePodMetrics)
	http.HandleFunc("/api/alert-config", handleAlertConfig) // 添加更新警报配置的接口
	http.HandleFunc("/api/alert-metrics", handleAlertMetrics)
	http.HandleFunc("/api/host-management", handleHostManagement)

	// 添加WebSocket路由
	http.HandleFunc("/ws/dashboard", handleWebSocket)

	// 启动广播协程
	go broadcastMetrics()

	log.Println("Server running on :12800")

	corsHeaders := handlers.CORS(
		handlers.AllowedOrigins([]string{"*"}),                                       // 允许所有来源
		handlers.AllowedMethods([]string{"GET", "POST", "OPTIONS", "PUT", "DELETE"}), // 允许的方法
		handlers.AllowedHeaders([]string{"Content-Type", "Authorization"}),           // 允许的请求头
	)

	http.HandleFunc("/vue/", VueHandler)

	htmlFS := http.FileServer(http.Dir("./templates"))
	http.Handle("/", htmlFS)

	log.Fatal(http.ListenAndServe(":12800", corsHeaders(http.DefaultServeMux)))
}

func VueHandler(w http.ResponseWriter, r *http.Request) {
	path := r.URL.Path[len("/vue/"):]
	fullPath := filepath.Join("./frontend", path)

	_, err := os.Stat(fullPath)
	if os.IsNotExist(err) {
		http.ServeFile(w, r, "./frontend/index.html")
		return
	}
	http.StripPrefix("/vue/", http.FileServer(http.Dir("./frontend"))).ServeHTTP(w, r)
}

// 添加WebSocket处理函数
func handleWebSocket(w http.ResponseWriter, r *http.Request) {
	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Printf("WebSocket upgrade failed: %v", err)
		return
	}
	defer conn.Close()

	// 注册新客户端
	clientsMux.Lock()
	clients[conn] = true
	clientsMux.Unlock()

	// 清理断开的客户端
	defer func() {
		clientsMux.Lock()
		delete(clients, conn)
		clientsMux.Unlock()
	}()

	// 保持连接并处理消息
	for {
		_, _, err := conn.ReadMessage()
		if err != nil {
			break
		}
	}
}

// 广播数据给所有连接的客户端
func broadcastMetrics() {
	for {
		// 获取数据时加锁
		storeLock.RLock()
		hostLock.RLock() // 添加 HostManage 的读锁
		var hosts []HostData
		for _, hostData := range DataStore {
			hostData.Label = HostManage[hostData.IP].Label // 这里访问 HostManage 需要加锁
			hosts = append(hosts, *hostData)
		}
		hostLock.RUnlock() // 释放 HostManage 的读锁
		storeLock.RUnlock()

		// 准备要发送的数据
		data := map[string]interface{}{
			"hosts":   hosts,
			"loading": false,
			"error":   nil,
		}

		// 序列化数据
		jsonData, err := json.Marshal(data)
		if err != nil {
			log.Printf("JSON marshal error: %v", err)
			continue
		}

		// 广播给所有客户端
		clientsMux.Lock()
		for client := range clients {
			err := client.WriteMessage(websocket.TextMessage, jsonData)
			if err != nil {
				log.Printf("Write error: %v", err)
				client.Close()
				delete(clients, client)
			}
		}
		clientsMux.Unlock()

		time.Sleep(time.Second) // 每秒更新一次
	}
}
