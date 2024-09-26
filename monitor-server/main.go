package main

//
//import (
//	"bytes"
//	"context"
//	"encoding/json"
//	"flag"
//	"fmt"
//	"github.com/influxdata/influxdb-client-go/v2"
//	"io/ioutil"
//	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
//	"k8s.io/client-go/kubernetes"
//	"k8s.io/client-go/tools/clientcmd"
//	metricsclientset "k8s.io/metrics/pkg/client/clientset/versioned"
//	"log"fv
//	"net/http"
//	"os"
//	"sync"
//	"time"
//)
//
//type HostData struct {
//	Hostname     string  `json:"hostname"`
//	IP           string  `json:"ip"`
//	NodeIP       string  `json:"node_ip"`
//	CPUUsage     float64 `json:"cpu_usage"`
//	MemoryUsage  float64 `json:"memory_usage"`
//	DiskUsage    float64 `json:"disk_usage"`
//	NetworkUsage float64 `json:"network_usage"`
//	Timestamp    int64   `json:"timestamp"`
//}
//
//type AlertConfig struct {
//	CPUThreshold    float64 `json:"cpu_threshold"`
//	MemoryThreshold float64 `json:"memory_threshold"`
//	DiskThreshold   float64 `json:"disk_threshold"`
//	CPUDuration     float64 `json:"cpu_duration"`
//	MemoryDuration  float64 `json:"memory_duration"`
//	Success         bool    `json:"success"`
//}
//
//var alertConfig = AlertConfig{
//	CPUThreshold:    90.0,
//	MemoryThreshold: 85.0,
//	DiskThreshold:   85.0,
//	CPUDuration:     20, // CPU 超过阈值持续时间
//	MemoryDuration:  10, // 内存超过阈值持续时间
//	Success:         false,
//}
//
//// InfluxDB 2.x 配置信息
//var InfluxURL = "10.10.18.116:8086"
//var InfluxToken = "wh56EgkTNCyt-oSz_4Uo8l_SYy9R57CnUFy2NZY4bxmjZ9bbBNiMvQ0kdo8W4cwdvP6JrgXY49uXpTI7d5mRtA=="
//var Org = "finchina"
//var Bucket = "finchina-dev"
//
//type AlertStatus struct {
//	LastAlertTime time.Time     // 最近一次发送警报的时间
//	Count         time.Duration // 超过阈值的累计时间
//}
//
//var (
//	dataStore      = make(map[string]HostData)
//	alertTimers    = make(map[string]map[string]*AlertStatus) // 存储每个主机的警报状态 // 存储每个主机的 CPU 和内存定时器
//	storeLock      = sync.RWMutex{}
//	alertLock      = sync.Mutex{}
//	podMetricsList []HostData
//)
//
//// 发送 HTTP POST 请求到接口
//func sendAlertToHttp(content string, receive string) {
//	//url := "http://222.73.12.12:8008/api/SendAppMsg"
//	//payload := map[string]interface{}{
//	//	"touser": receive,
//	//	"text": map[string]string{
//	//		"content": content,
//	//	},
//	//}
//
//	url := "http://222.73.12.12:8008/api/SendChatMsg"
//	payload := map[string]interface{}{
//		"chatid": 267,
//		"text": map[string]string{
//			"content": content,
//		},
//	}
//
//	jsonData, err := json.Marshal(payload)
//	if err != nil {
//		fmt.Println("Failed to encode JSON:", err)
//		return
//	}
//
//	req, err := http.NewRequest("POST", url, bytes.NewBuffer(jsonData))
//	if err != nil {
//		fmt.Println("Failed to create HTTP request:", err)
//		return
//	}
//
//	req.Header.Set("Content-Type", "application/json")
//
//	client := &http.Client{}
//	resp, err := client.Do(req)
//	if err != nil {
//		fmt.Println("Failed to send HTTP request:", err)
//		return
//	}
//	defer resp.Body.Close()
//
//	body, _ := ioutil.ReadAll(resp.Body)
//	fmt.Println("HTTP Response:", string(body))
//}
//
//func sendAlert(hostIP string, msg string) {
//
//	fmt.Println("发生告警：" + hostIP + ", 资源类型: " + msg)
//
//	sendAlertToHttp("中台服务器监控\n发生告警："+hostIP+", 资源类型: "+msg+" , 持续超过预警阈值", "kongxr")
//}
//func checkAlerts(host HostData) {
//	alertLock.Lock()
//	defer alertLock.Unlock()
//
//	if _, exists := alertTimers[host.IP]; !exists {
//		alertTimers[host.IP] = make(map[string]*AlertStatus)
//	}
//
//	// 检查 CPU 使用率
//	checkResource(host, "cpu", time.Duration(alertConfig.CPUDuration)*time.Second)
//
//	// 检查内存使用率
//	checkResource(host, "memory", time.Duration(alertConfig.MemoryDuration)*time.Second)
//
//	// 检查磁盘使用率
//	checkResource(host, "disk", 0) // 磁盘使用率不检查持续时间，直接发送警报
//}
//
//func checkResource(host HostData, resourceType string, threshold time.Duration) {
//	if _, exists := alertTimers[host.IP][resourceType]; !exists {
//		alertTimers[host.IP][resourceType] = &AlertStatus{LastAlertTime: time.Unix(0, 0), Count: 0}
//	}
//
//	alertStatus := alertTimers[host.IP][resourceType]
//
//	if threshold == 0 { // 特殊处理磁盘使用率
//		if host.DiskUsage > alertConfig.DiskThreshold {
//			if alertStatus.LastAlertTime.IsZero() || time.Since(alertStatus.LastAlertTime) >= time.Hour {
//				sendAlert(host.IP, "disk")
//				alertStatus.LastAlertTime = time.Now()
//			}
//		} else {
//			alertStatus.LastAlertTime = time.Unix(0, 0)
//		}
//	} else { // 处理 CPU 和内存使用率
//		if usage := getHostUsage(host, resourceType); usage > alertConfig.GetThreshold(resourceType) {
//			alertStatus.Count += 1 * time.Second
//			if alertStatus.Count >= threshold {
//				if alertStatus.LastAlertTime.IsZero() || time.Since(alertStatus.LastAlertTime) >= 30*time.Minute {
//					sendAlert(host.IP, resourceType)
//					alertStatus.LastAlertTime = time.Now()
//					alertStatus.Count = 0 // 重置计数器
//				}
//			}
//		} else {
//			alertStatus.Count = 0 // 重置计数器
//		}
//	}
//}
//
//func getHostUsage(host HostData, resourceType string) float64 {
//	switch resourceType {
//	case "cpu":
//		return host.CPUUsage
//	case "memory":
//		return host.MemoryUsage
//	case "disk":
//		return host.DiskUsage
//	default:
//		return 0
//	}
//}
//
//func (a *AlertConfig) GetThreshold(resourceType string) float64 {
//	switch resourceType {
//	case "cpu":
//		return a.CPUThreshold
//	case "memory":
//		return a.MemoryThreshold
//	case "disk":
//		return a.DiskThreshold
//	default:
//		return 0
//	}
//}
//
//func (a *AlertConfig) GetDuration(resourceType string) time.Duration {
//	switch resourceType {
//	case "cpu":
//		return time.Duration(a.CPUDuration) * time.Second
//	case "memory":
//		return time.Duration(a.MemoryDuration) * time.Second
//	default:
//		return 0
//	}
//}
//
//func handleHostData(w http.ResponseWriter, r *http.Request) {
//	if r.Method != "POST" {
//		http.Error(w, "Only POST allowed", http.StatusMethodNotAllowed)
//		return
//	}
//
//	var hostData HostData
//	err := json.NewDecoder(r.Body).Decode(&hostData)
//	if err != nil {
//		http.Error(w, "Invalid request payload", http.StatusBadRequest)
//		return
//	}
//
//	storeLock.Lock()
//	hostData.Timestamp = time.Now().Unix()
//	dataStore[hostData.IP] = hostData
//	storeLock.Unlock()
//
//	checkAlerts(hostData) // 检查警报
//
//	fmt.Fprintf(w, "Data received for host: %s", hostData.IP)
//}
//
//func handleDashboard(w http.ResponseWriter, r *http.Request) {
//	storeLock.RLock()
//	defer storeLock.RUnlock()
//
//	var hosts []HostData
//	for _, hostData := range dataStore {
//		hosts = append(hosts, hostData)
//	}
//
//	json.NewEncoder(w).Encode(hosts)
//}
//
//func handlePodDashboard(w http.ResponseWriter, r *http.Request) {
//	json.NewEncoder(w).Encode(podMetricsList)
//}
//
//func handleAlertConfig(w http.ResponseWriter, r *http.Request) {
//	if r.Method != "POST" {
//		http.Error(w, "Only POST allowed", http.StatusMethodNotAllowed)
//		return
//	}
//
//	var newConfig AlertConfig
//	err := json.NewDecoder(r.Body).Decode(&newConfig)
//	if err != nil {
//		http.Error(w, "Invalid request payload", http.StatusBadRequest)
//		return
//	}
//
//	alertConfig = newConfig // 更新警报配置
//
//	response := map[string]bool{"success": true}
//	w.Header().Set("Content-Type", "application/json")
//	w.WriteHeader(http.StatusOK)
//	json.NewEncoder(w).Encode(response)
//}
//
//func queryInfluxDB(hostname, start, end string) ([]map[string]interface{}, error) {
//	client := influxdb2.NewClient("http://"+InfluxURL, InfluxToken)
//	defer client.Close()
//
//	queryAPI := client.QueryAPI(Org)
//	query := fmt.Sprintf(`from(bucket: "%s") |> range(start: %s, stop: %s) |> filter(fn: (r) => r["_measurement"] == "host_metrics" and r["host"] == "%s")`, Bucket, start, end, hostname)
//
//	result, err := queryAPI.Query(context.Background(), query)
//	if err != nil {
//		return nil, err
//	}
//
//	var records []map[string]interface{}
//	for result.Next() {
//		records = append(records, result.Record().Values())
//	}
//
//	if result.Err() != nil {
//		return nil, result.Err()
//	}
//
//	return records, nil
//}
//
//func handleHostMetrics(w http.ResponseWriter, r *http.Request) {
//	host := r.URL.Query().Get("host")
//	start := r.URL.Query().Get("start")
//	end := r.URL.Query().Get("end")
//
//	if host == "" || start == "" || end == "" {
//		http.Error(w, "Host, start, and end parameters are required", http.StatusBadRequest)
//		return
//	}
//
//	results, err := queryInfluxDB(host, start, end)
//	if err != nil {
//		http.Error(w, "Failed to query InfluxDB", http.StatusInternalServerError)
//		return
//	}
//
//	json.NewEncoder(w).Encode(results)
//}
//
//func init() {
//	if influxTokenEnv := os.Getenv("INFLUX_TOKEN"); influxTokenEnv != "" {
//		InfluxToken = influxTokenEnv
//	}
//	if influxURLEnv := os.Getenv("INFLUX_URL"); influxURLEnv != "" {
//		InfluxURL = influxURLEnv
//	}
//	if influxOrgEnv := os.Getenv("INFLUX_ORG"); influxOrgEnv != "" {
//		Org = influxOrgEnv
//	}
//	if influxBucketEnv := os.Getenv("INFLUX_BUCKET"); influxBucketEnv != "" {
//		Bucket = influxBucketEnv
//	}
//}
//
//// 新增的处理函数，用于获取当前系统的警报指标
//func handleAlertMetrics(w http.ResponseWriter, r *http.Request) {
//	storeLock.RLock()
//	defer storeLock.RUnlock()
//	alertConfig.Success = true
//	// 返回当前的警报配置
//	json.NewEncoder(w).Encode(alertConfig)
//}
//
//func getPodMetrics(clientset *kubernetes.Clientset, metricsClientset *metricsclientset.Clientset) ([]HostData, error) {
//
//	pods, err := clientset.CoreV1().Pods("finchina-dev").List(context.TODO(), metav1.ListOptions{})
//	if err != nil {
//		return nil, err
//	}
//
//	for _, pod := range pods.Items {
//		quantity := pod.Spec.Containers[0].Resources.Limits["cpu"]
//		fmt.Println(quantity)
//	}
//
//	return nil, nil
//}
//
//// 获取所有Pod的实时资源占用
////func getPodMetrics(clientset *kubernetes.Clientset, metricsClientset *metricsclientset.Clientset) ([]HostData, error) {
////	pods, err := clientset.CoreV1().Pods("finchina-dev").List(context.TODO(), metav1.ListOptions{})
////	if err != nil {
////		return nil, err
////	}
////
////	// 预分配内存
////	temp := make([]HostData, 0, len(pods.Items))
////
////	var wg sync.WaitGroup
////	var mu sync.Mutex // 用于保护对 temp 的访问
////
////	for _, pod := range pods.Items {
////		wg.Add(1)
////		go func(pod v1.Pod) {
////			defer wg.Done()
////
////			node, err := clientset.CoreV1().Nodes().Get(context.TODO(), pod.Spec.NodeName, metav1.GetOptions{})
////			nodeIP := ""
////			if err == nil {
////				nodeIP = node.Status.Addresses[0].Address // 获取宿主机的 IP 地址
////			}
////
////			podMetrics, err := metricsClientset.MetricsV1beta1().PodMetricses(pod.Namespace).Get(context.TODO(), pod.Name, metav1.GetOptions{})
////			if err != nil {
////				// 记录错误并返回
////				log.Printf("未能获取 Pod %s 的指标: %v", pod.Name, err)
////				// 转换内存为 MiB
////				hostData := HostData{
////					Hostname:    pod.Name,
////					IP:          pod.Status.PodIP,
////					NodeIP:      nodeIP,
////					CPUUsage:    float64(-1),
////					MemoryUsage: float64(-1), // 转换为 MiB
////					Timestamp:   time.Now().Unix(),
////				}
////
////				mu.Lock()
////				temp = append(temp, hostData)
////				mu.Unlock()
////				return
////			}
////
////			// 只获取 Pod 的整体资源占用
////			var totalCPUUsage int64
////			var totalMemoryUsage int64
////
////			for _, container := range podMetrics.Containers {
////				totalCPUUsage += container.Usage.Cpu().MilliValue()
////				totalMemoryUsage += container.Usage.Memory().Value()
////			}
////
////			// 转换内存为 MiB
////			hostData := HostData{
////				Hostname:    pod.Name,
////				IP:          pod.Status.PodIP,
////				NodeIP:      nodeIP,
////				CPUUsage:    float64(totalCPUUsage),
////				MemoryUsage: float64(totalMemoryUsage) / (1024 * 1024), // 转换为 MiB
////				Timestamp:   time.Now().Unix(),
////			}
////
////			mu.Lock()
////			temp = append(temp, hostData)
////			mu.Unlock()
////		}(pod)
////	}
////
////	wg.Wait() // 等待所有 goroutines 完成
////	fmt.Println("pod 监测数据获取成功")
////	return temp, nil
////}
//
//func writePodMetricsToInflux(podMetrics []HostData) {
//	client := influxdb2.NewClient("http://"+InfluxURL, InfluxToken)
//	defer client.Close()
//
//	writeAPI := client.WriteAPIBlocking(Org, Bucket)
//
//	for _, host := range podMetrics {
//		p := influxdb2.NewPointWithMeasurement("pod_metrics").
//			AddTag("pod", host.Hostname).
//			AddField("cpu_usage", host.CPUUsage).
//			AddField("memory_usage", host.MemoryUsage).
//			SetTime(time.Unix(host.Timestamp, 0))
//
//		err := writeAPI.WritePoint(context.Background(), p)
//		if err != nil {
//			fmt.Println("Error writing to InfluxDB:", err)
//		}
//	}
//}
//
//func queryPodMetricsFromInflux(podName, start, end string) ([]map[string]interface{}, error) {
//	client := influxdb2.NewClient("http://"+InfluxURL, InfluxToken)
//	defer client.Close()
//
//	queryAPI := client.QueryAPI(Org)
//	query := fmt.Sprintf(`
//		from(bucket: "%s")
//		|> range(start: %s, stop: %s)
//		|> filter(fn: (r) => r["_measurement"] == "pod_metrics" and r["pod"] == "%s")
//		|> keep(columns: ["_time", "_value", "_field", "pod"])
//		`, Bucket, start, end, podName)
//
//	result, err := queryAPI.Query(context.Background(), query)
//	if err != nil {
//		return nil, err
//	}
//
//	var records []map[string]interface{}
//	for result.Next() {
//		records = append(records, result.Record().Values())
//	}
//
//	if result.Err() != nil {
//		return nil, result.Err()
//	}
//
//	return records, nil
//}
//
//func handlePodMetrics(w http.ResponseWriter, r *http.Request) {
//	podName := r.URL.Query().Get("pod")
//	start := r.URL.Query().Get("start")
//	end := r.URL.Query().Get("end")
//
//	if podName == "" || start == "" || end == "" {
//		http.Error(w, "Pod, start, and end parameters are required", http.StatusBadRequest)
//		return
//	}
//
//	results, err := queryPodMetricsFromInflux(podName, start, end)
//	if err != nil {
//		http.Error(w, "Failed to query InfluxDB", http.StatusInternalServerError)
//		return
//	}
//
//	json.NewEncoder(w).Encode(results)
//}
//
//func main() {
//
//	clientset, metricsClientset := initKubernetesClient()
//
//	// 定时获取Pod资源数据并写入InfluxDB
//	go func() {
//		for {
//			podMetrics, err := getPodMetrics(clientset, metricsClientset)
//			podMetricsList = podMetrics
//
//			if err == nil {
//				writePodMetricsToInflux(podMetrics)
//			}
//			time.Sleep(2 * time.Second) // 每2 获取一次数据
//		}
//	}()
//
//	http.HandleFunc("/api/host-data", handleHostData)
//	http.HandleFunc("/api/dashboard", handleDashboard)
//	http.HandleFunc("/api/pod-dashboard", handlePodDashboard)
//	http.HandleFunc("/api/host-metrics", handleHostMetrics)
//	http.HandleFunc("/api/pod-metrics", handlePodMetrics)
//	http.HandleFunc("/api/alert-config", handleAlertConfig) // 添加更新警报配置的接口
//	http.HandleFunc("/api/alert-metrics", handleAlertMetrics)
//
//	log.Println("Server running on :12800")
//
//	fs := http.FileServer(http.Dir("./templates"))
//	http.Handle("/", fs)
//
//	log.Fatal(http.ListenAndServe(":12800", nil))
//}
//
//func initKubernetesClient() (*kubernetes.Clientset, *metricsclientset.Clientset) {
//	// 初始化 Kubernetes 和 Metrics ClientSet
//	kubeconfig := flag.String("kubeconfig.yaml", "kubeconfig.yaml", "absolute path to the kubeconfig.yaml file")
//	flag.Parse()
//
//	config, err := clientcmd.BuildConfigFromFlags("", *kubeconfig)
//	if err != nil {
//		fmt.Printf("Error building kubeconfig.yaml: %s\n", err.Error())
//		os.Exit(1)
//	}
//
//	clientset, err := kubernetes.NewForConfig(config)
//	if err != nil {
//		fmt.Printf("Error creating Kubernetes client: %s\n", err.Error())
//		os.Exit(1)
//	}
//
//	metricsClientset, err := metricsclientset.NewForConfig(config)
//	if err != nil {
//		fmt.Printf("Error creating Metrics client: %s\n", err.Error())
//		os.Exit(1)
//	}
//
//	return clientset, metricsClientset
//}
