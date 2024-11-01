package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"sync"
)

type Host struct {
	ID           int    `json:"id"`
	IPAddress    string `json:"ip_address"`
	Label        string `json:"label"`
	AlertEnabled bool   `json:"alert_enabled"`
}

var HostManage = make(map[string]Host) // 内存中的主机信息
var hostLock = sync.RWMutex{}          // 读写锁，保证并发安全

// 从数据库加载所有主机信息
func loadHostsFromDB() {
	query := `SELECT id, ip_address,IFNULL(label, ''), alert_enabled FROM hosts`
	rows, err := db.Query(query)
	if err != nil {
		log.Fatalf("Failed to load hosts: %v", err)
	}
	defer rows.Close()

	hostLock.Lock()
	defer hostLock.Unlock()

	for rows.Next() {
		var h Host
		err := rows.Scan(&h.ID, &h.IPAddress, &h.Label, &h.AlertEnabled)
		if err != nil {
			log.Printf("Error scanning host row: %v", err)
			continue
		}
		HostManage[h.IPAddress] = h
	}

	fmt.Println("Successfully loaded hosts from database")
}

// 增加或更新主机信息
func addHostInDB(host Host) error {
	query := `REPLACE INTO hosts (ip_address,label, alert_enabled) VALUES (?, ?, ?)`
	_, err := db.Exec(query, host.IPAddress, host.Label, host.AlertEnabled)
	if err == nil {
		hostLock.Lock()
		HostManage[host.IPAddress] = host
		hostLock.Unlock()
	}
	return err
}

// 增加或更新主机信息
func updateHostInDB(host Host) error {
	query := `UPDATE hosts SET alert_enabled = ? WHERE ip_address = ?;`
	_, err := db.Exec(query, host.AlertEnabled, host.IPAddress)
	if err == nil {
		hostLock.Lock()
		HostManage[host.IPAddress] = host
		hostLock.Unlock()
	}
	return err
}

// 删除主机信息
func deleteHostFromDB(ip string) error {
	query := `DELETE FROM hosts WHERE ip_address = ?`
	_, err := db.Exec(query, ip)
	if err == nil {
		hostLock.Lock()
		delete(HostManage, ip)
		hostLock.Unlock()
	}
	return err
}

// 提供增删改查接口
func handleHostManagement(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case "POST":
		// 添加或更新主机信息
		var host Host
		err := json.NewDecoder(r.Body).Decode(&host)
		if err != nil {
			http.Error(w, "Invalid request payload", http.StatusBadRequest)
			return
		}
		err = addHostInDB(host)
		if err != nil {
			http.Error(w, "Failed to add or update host", http.StatusInternalServerError)
			return
		}

		response := map[string]bool{"success": true}
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(response)

	case "DELETE":
		// 删除主机信息
		ip := r.URL.Query().Get("ip")
		if ip == "" {
			http.Error(w, "IP address is required", http.StatusBadRequest)
			return
		}
		err := deleteHostFromDB(ip)
		if err != nil {
			http.Error(w, "Failed to delete host", http.StatusInternalServerError)
			return
		}
		w.WriteHeader(http.StatusOK)
		response := map[string]bool{"success": true}
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(response)

	case "GET":
		// 返回所有主机信息
		hostLock.RLock()
		defer hostLock.RUnlock()

		var tempHosts []Host
		for _, host := range HostManage {
			tempHosts = append(tempHosts, host)
		}
		json.NewEncoder(w).Encode(tempHosts)
	case "PUT":

		var host Host
		err := json.NewDecoder(r.Body).Decode(&host)
		if err != nil {
			http.Error(w, "Invalid request payload", http.StatusBadRequest)
			return
		}
		err = updateHostInDB(host)
		if err != nil {
			http.Error(w, "Failed to add or update host", http.StatusInternalServerError)
			return
		}

		response := map[string]bool{"success": true}
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(response)

	}
}
