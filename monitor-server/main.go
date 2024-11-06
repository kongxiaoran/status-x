package main

import (
	"fmt"
	"github.com/gorilla/handlers"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"time"
)

var PodCollectFrequency = 3
var ActuatorFrequency = 2

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
			for _, host := range HostManage {
				currentTime := time.Now().Unix()
				// 获取该主机最新的上传记录的时间戳
				if _, exists := dataStore[host.IPAddress]; !exists {
					continue
				}
				latestTime := dataStore[host.IPAddress].Timestamp
				lastOfflineTime := dataStore[host.IPAddress].LastOfflineAlertTime
				// 触发离线
				if currentTime-latestTime > 60 && host.AlertEnabled {
					// 判断是否重复报警
					if currentTime-lastOfflineTime > 600 || lastOfflineTime == 0 {
						SendAlert(host.IPAddress, "offline")
						storeLock.Lock()
						dataStore[host.IPAddress].LastOfflineAlertTime = currentTime
						storeLock.Unlock()
					}
				}
			}
			time.Sleep(time.Duration(ActuatorFrequency) * time.Second) // 每 ActuatorFrequency 检查一次
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
