package main

import (
	"log"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"time"

	"k8s.io/client-go/kubernetes"
	metricsclientset "k8s.io/metrics/pkg/client/clientset/versioned"

	"github.com/gorilla/handlers"
)

// 全局配置
var (
	PodCollectFrequency = 3
	ActuatorFrequency   = 2
)

func init() {
	loadEnvConfig()
}

// 加载环境变量配置
func loadEnvConfig() {
	// InfluxDB配置
	if token := os.Getenv("INFLUX_TOKEN"); token != "" {
		InfluxToken = token
	}
	if url := os.Getenv("INFLUX_URL"); url != "" {
		InfluxURL = url
	}
	if org := os.Getenv("INFLUX_ORG"); org != "" {
		Org = org
	}
	if bucket := os.Getenv("INFLUX_BUCKET"); bucket != "" {
		Bucket = bucket
	}

	// 采集频率配置
	if freq := os.Getenv("POD_COLLECT_FREQUENCY"); freq != "" {
		if val, err := strconv.Atoi(freq); err == nil {
			PodCollectFrequency = val
		}
	}
	if freq := os.Getenv("ACTUATOR_FREQUENCY"); freq != "" {
		if val, err := strconv.Atoi(freq); err == nil {
			ActuatorFrequency = val
		}
	}
}

func main() {
	// 初始化服务
	initDB()
	loadAlertConfigFromDB()
	loadHostsFromDB()

	// 初始化Kubernetes客户端
	clientset, metricsClientset := initKubernetesClient()

	// 启动Pod资源监控
	go monitorPodResources(clientset, metricsClientset)

	// 启动主机状态监控
	go monitorHostStatus()

	// 设置HTTP路由
	setupRoutes()

	// 启动WebSocket广播
	go broadcastMetrics()

	// 启动HTTP服务
	startHTTPServer()
}

// 监控Pod资源
func monitorPodResources(clientset *kubernetes.Clientset, metricsClientset *metricsclientset.Clientset) {
	for {
		podMetrics, err := getPodMetrics(clientset, metricsClientset)
		podMetricsListMu.Lock()
		podMetricsList = podMetrics
		podMetricsListMu.Unlock()

		if err == nil {
			writePodMetricsToInflux(podMetrics)
		}
		time.Sleep(time.Duration(PodCollectFrequency) * time.Second)
	}
}

// 监控主机状态
func monitorHostStatus() {
	for {
		checkHosts()
		time.Sleep(time.Duration(ActuatorFrequency) * time.Second)
	}
}

// 检查主机状态
func checkHosts() {
	storeLock.RLock()
	hostLock.RLock()
	defer hostLock.RUnlock()
	defer storeLock.RUnlock()

	currentTime := time.Now().Unix()
	for _, host := range HostManage {
		if hostData, exists := DataStore[host.IPAddress]; exists {
			checkHostOffline(&host, hostData, currentTime)
		}
	}
}

// 检查主机离线状态
func checkHostOffline(host *Host, hostData *HostData, currentTime int64) {
	latestTime := hostData.Timestamp
	lastOfflineTime := hostData.LastOfflineAlertTime

	if currentTime-latestTime > 60 {
		if host.AlertEnabled && (currentTime-lastOfflineTime > 3600*2 || lastOfflineTime == 0) {
			if CheckServerNetwork() {
				SendAlert(host.IPAddress, "offline")
				hostData.LastOfflineAlertTime = time.Now().Unix()
			} else {
				log.Printf("Server network check failed, skip offline alert for host: %s", host.IPAddress)
			}
		}
		hostData.Status = "offline"
	} else {
		hostData.Status = "online"
	}
}

// 设置HTTP路由
func setupRoutes() {
	// API路由
	http.HandleFunc("/api/host-data", handleHostData)
	http.HandleFunc("/api/dashboard", handleDashboard)
	http.HandleFunc("/api/pod-dashboard", handlePodDashboard)
	http.HandleFunc("/api/actuator-dashboard", handleActuatorDashboard)
	http.HandleFunc("/api/host-metrics", handleHostMetrics)
	http.HandleFunc("/api/pod-metrics", handlePodMetrics)
	http.HandleFunc("/api/alert-config", handleAlertConfig)
	http.HandleFunc("/api/alert-metrics", handleAlertMetrics)
	http.HandleFunc("/api/host-management", handleHostManagement)

	// WebSocket路由
	http.HandleFunc("/ws/dashboard", handleWebSocket)

	// 静态文件路由
	http.HandleFunc("/vue/", handleVueFiles)
	http.Handle("/", http.FileServer(http.Dir("./templates")))
}

// 处理Vue文件
func handleVueFiles(w http.ResponseWriter, r *http.Request) {
	path := r.URL.Path[len("/vue/"):]
	fullPath := filepath.Join("./frontend", path)

	if _, err := os.Stat(fullPath); os.IsNotExist(err) {
		http.ServeFile(w, r, "./frontend/index.html")
		return
	}
	http.StripPrefix("/vue/", http.FileServer(http.Dir("./frontend"))).ServeHTTP(w, r)
}

// 启动HTTP服务
func startHTTPServer() {
	corsHeaders := handlers.CORS(
		handlers.AllowedOrigins([]string{"*"}),
		handlers.AllowedMethods([]string{"GET", "POST", "OPTIONS", "PUT", "DELETE"}),
		handlers.AllowedHeaders([]string{"Content-Type", "Authorization"}),
	)

	log.Println("Server running on :12800")
	log.Fatal(http.ListenAndServe(":12800", corsHeaders(http.DefaultServeMux)))
}
