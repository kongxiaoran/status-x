package main

import (
	"fmt"
	"log"
	"net/http"
	"os"
	"time"
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
}
func main() {

	// 初始化 MySQL 数据库
	//initDB()
	// 从数据库加载报警配置
	//loadAlertConfigFromDB()
	// 从数据库加载主机信息
	//loadHostsFromDB()

	clientset, metricsClientset := initKubernetesClient()

	// 定时获取Pod资源数据并写入InfluxDB
	go func() {
		for {
			podMetrics, err := getPodMetrics(clientset, metricsClientset)
			podMetricsList = podMetrics

			if err == nil {
				writePodMetricsToInflux(podMetrics)
			}
			time.Sleep(5 * time.Second) // 每2 获取一次数据
		}
	}()

	hostData := HostData{
		Hostname: "test",
		IP:       "localhost",
	}

	podMetricsList = append(podMetricsList, hostData)

	go func() {
		for {
			if podMetricsList != nil {
				for _, pod := range podMetricsList {
					actuator := getPodActuator(pod.IP)
					fmt.Print(actuator)
				}
			}
			time.Sleep(2 * time.Second) // 每2 获取一次数据
		}
	}()

	http.HandleFunc("/api/host-data", handleHostData)
	http.HandleFunc("/api/dashboard", handleDashboard)
	http.HandleFunc("/api/pod-dashboard", handlePodDashboard)
	http.HandleFunc("/api/host-metrics", handleHostMetrics)
	http.HandleFunc("/api/pod-metrics", handlePodMetrics)
	http.HandleFunc("/api/alert-config", handleAlertConfig) // 添加更新警报配置的接口
	http.HandleFunc("/api/alert-metrics", handleAlertMetrics)
	http.HandleFunc("/api/host-management", handleHostManagement)

	log.Println("Server running on :12800")

	fs := http.FileServer(http.Dir("./templates"))
	http.Handle("/", fs)

	log.Fatal(http.ListenAndServe(":12800", nil))
}
