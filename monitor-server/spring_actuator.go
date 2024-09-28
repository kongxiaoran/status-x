package main

import (
	"fmt"
	"io/ioutil"
	"net/http"
	"strings"

	"encoding/json"
)

// MetricMap 用于表示整个 JSON 数据
type MetricMap map[string][]StatisticData

// StatisticData 用于表示单个统计数据
type StatisticData struct {
	Statistic string  `json:"statistic"`
	Value     float64 `json:"value"`
}

func getPodActuator(address string) map[string]interface{} {

	url := fmt.Sprintf("http://%s:8081/actuator/batch-metrics", address)
	// 发送 HTTP 请求获取指标数据
	resp, err := http.Get(url)
	if err != nil {
		fmt.Printf("Error fetching metrics: %v\n", err)
		return nil
	}
	defer resp.Body.Close()

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		fmt.Printf("Error reading response body: %v\n", err)
		return nil
	}

	var metrics MetricMap
	if err := json.Unmarshal(body, &metrics); err != nil {
		fmt.Printf("Error unmarshalling JSON: %v\n", err)
		return nil
	}

	// 存储和处理数据
	localMetrics := make(map[string]interface{})
	for name, stats := range metrics {
		for _, stat := range stats {
			tempName := strings.Replace(name, ".", "_", -1)
			switch name {
			case "jvm.memory.max", "jvm.memory.used":
				valueMB := bytesToMB(stat.Value)
				stat.Value = valueMB
				localMetrics[tempName+"_"+stat.Statistic] = valueMB
			default:
				localMetrics[tempName+"_"+stat.Statistic] = stat.Value
			}
		}
	}
	return localMetrics
}

func bytesToMB(bytes float64) float64 {
	return bytes / (1024 * 1024)
}
