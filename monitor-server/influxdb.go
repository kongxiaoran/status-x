package main

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/influxdata/influxdb-client-go/v2"
	"net/http"
	"time"
)

var InfluxURL = "10.10.18.116:8086"
var InfluxToken = "wh56EgkTNCyt-oSz_4Uo8l_SYy9R57CnUFy2NZY4bxmjZ9bbBNiMvQ0kdo8W4cwdvP6JrgXY49uXpTI7d5mRtA=="
var Org = "finchina"
var Bucket = "finchina-dev"

func queryInfluxDB(hostname, start, end string) ([]map[string]interface{}, error) {
	client := influxdb2.NewClient("http://"+InfluxURL, InfluxToken)
	defer client.Close()

	queryAPI := client.QueryAPI(Org)
	query := fmt.Sprintf(`from(bucket: "%s") |> range(start: %s, stop: %s) |> filter(fn: (r) => r["_measurement"] == "host_metrics" and r["host"] == "%s")`, Bucket, start, end, hostname)

	result, err := queryAPI.Query(context.Background(), query)
	if err != nil {
		return nil, err
	}

	var records []map[string]interface{}
	for result.Next() {
		records = append(records, result.Record().Values())
	}

	if result.Err() != nil {
		return nil, result.Err()
	}

	return records, nil
}

func handleHostMetrics(w http.ResponseWriter, r *http.Request) {
	host := r.URL.Query().Get("host")
	start := r.URL.Query().Get("start")
	end := r.URL.Query().Get("end")

	results, err := queryInfluxDB(host, start, end)
	if err != nil {
		http.Error(w, "Failed to query InfluxDB", http.StatusInternalServerError)
		return
	}

	json.NewEncoder(w).Encode(results)
}

func writePodMetricsToInflux(podMetrics []HostData) {
	client := influxdb2.NewClient("http://"+InfluxURL, InfluxToken)
	defer client.Close()

	writeAPI := client.WriteAPIBlocking(Org, Bucket)

	for _, host := range podMetrics {
		p := influxdb2.NewPointWithMeasurement("pod_metrics").
			AddTag("pod", host.Hostname).
			AddField("cpu_usage", host.CPUUsage).
			AddField("memory_usage", host.MemoryUsage).
			SetTime(time.Unix(host.Timestamp, 0))

		err := writeAPI.WritePoint(context.Background(), p)
		if err != nil {
			fmt.Println("Error writing to InfluxDB:", err)
		}
	}
}
func queryPodMetricsFromInflux(podName, start, end string) ([]map[string]interface{}, error) {
	client := influxdb2.NewClient("http://"+InfluxURL, InfluxToken)
	defer client.Close()

	queryAPI := client.QueryAPI(Org)
	query := fmt.Sprintf(`
		from(bucket: "%s")
		|> range(start: %s, stop: %s)
		|> filter(fn: (r) => r["_measurement"] == "pod_metrics" and r["pod"] == "%s")
		|> keep(columns: ["_time", "_value", "_field", "pod"])
		`, Bucket, start, end, podName)

	result, err := queryAPI.Query(context.Background(), query)
	if err != nil {
		return nil, err
	}

	var records []map[string]interface{}
	for result.Next() {
		records = append(records, result.Record().Values())
	}

	if result.Err() != nil {
		return nil, result.Err()
	}

	return records, nil
}
