package main

import (
	"context"
	"encoding/json"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	metricsclientset "k8s.io/metrics/pkg/client/clientset/versioned"
	"log"
	"net/http"
	"sync"
	"time"
)

var podMetricsList []HostData
var podMetricsListMu sync.Mutex

func handlePodMetrics(w http.ResponseWriter, r *http.Request) {
	podName := r.URL.Query().Get("pod")
	start := r.URL.Query().Get("start")
	end := r.URL.Query().Get("end")

	if podName == "" || start == "" || end == "" {
		http.Error(w, "Pod, start, and end parameters are required", http.StatusBadRequest)
		return
	}

	results, err := queryPodMetricsFromInflux(podName, start, end)
	if err != nil {
		http.Error(w, "Failed to query InfluxDB", http.StatusInternalServerError)
		return
	}

	json.NewEncoder(w).Encode(results)
}

// 获取所有Pod的实时资源占用
func getPodMetrics(clientset *kubernetes.Clientset, metricsClientset *metricsclientset.Clientset) ([]HostData, error) {

	// 指定要查询的命名空间
	namespaces := []string{"finchina-dev", "finchina-test2", "finchina-api", "finchina-userservice-dev"} // 命名空间

	var wg sync.WaitGroup
	var mu sync.Mutex // 用于保护对 temp 的访问

	var temp []HostData
	for _, ns := range namespaces {
		wg.Add(1)
		go func() {
			defer wg.Done()
			pods, err := clientset.CoreV1().Pods(ns).List(context.TODO(), metav1.ListOptions{})
			log.Printf("%s 命名空间存活pod数量：%d", ns, len(pods.Items))

			if err != nil {
				return
			}
			for _, pod := range pods.Items {
				//fmt.Println(pod.Namespace, "  ", pod.Name)
				node, err := clientset.CoreV1().Nodes().Get(context.TODO(), pod.Spec.NodeName, metav1.GetOptions{})
				nodeIP := ""
				if err == nil {
					nodeIP = node.Status.Addresses[0].Address // 获取宿主机的 IP 地址
				}

				podMetrics, err := metricsClientset.MetricsV1beta1().PodMetricses(pod.Namespace).Get(context.TODO(), pod.Name, metav1.GetOptions{})
				if err != nil {
					// 记录错误并返回
					//log.Printf("未能获取 Pod %s 的指标: %v", pod.Name, err)
					// 转换内存为 MiB
					hostData := HostData{
						Hostname:    pod.Name,
						IP:          pod.Status.PodIP,
						NameSpace:   pod.Namespace,
						NodeIP:      nodeIP,
						CPUUsage:    float64(-1),
						MemoryUsage: float64(-1), // 转换为 MiB
						Timestamp:   time.Now().Unix(),
					}

					mu.Lock()
					temp = append(temp, hostData)
					mu.Unlock()
					continue
				}

				// 只获取 Pod 的整体资源占用
				var totalCPUUsage int64
				var totalMemoryUsage int64

				for _, container := range podMetrics.Containers {
					totalCPUUsage += container.Usage.Cpu().MilliValue()
					totalMemoryUsage += container.Usage.Memory().Value()
				}

				// 转换内存为 MiB
				hostData := HostData{
					Hostname:    pod.Name,
					IP:          pod.Status.PodIP,
					NodeIP:      nodeIP,
					NameSpace:   pod.Namespace,
					CPUUsage:    float64(totalCPUUsage),
					MemoryUsage: float64(totalMemoryUsage) / (1024 * 1024), // 转换为 MiB
					Timestamp:   time.Now().Unix(),
				}

				mu.Lock()
				temp = append(temp, hostData)
				mu.Unlock()
			}

		}()
	}
	wg.Wait() // 等待所有 goroutines 完成
	//fmt.Println("pod 监测数据获取成功")
	return temp, nil
}
