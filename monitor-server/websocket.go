package main

import (
	"encoding/json"
	"github.com/sasha-s/go-deadlock"
	"log"
	"net/http"
	"time"

	"github.com/gorilla/websocket"
)

// WebSocket相关配置
var (
	// WebSocket升级器
	upgrader = websocket.Upgrader{
		CheckOrigin: func(r *http.Request) bool {
			return true // 允许所有来源的WebSocket连接
		},
	}

	// 客户端连接管理
	clients    = make(map[*websocket.Conn]bool)
	clientsMux deadlock.Mutex
)

// WebSocket连接处理
func handleWebSocket(w http.ResponseWriter, r *http.Request) {
	// 升级HTTP连接为WebSocket连接
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
		log.Println("准备发送广播数据")
		// 不加锁直接读取数据
		var hosts []HostData
		for _, hostData := range DataStore {
			hostCopy := *hostData // 创建副本
			if label, exists := HostManage[hostCopy.IP]; exists {
				hostCopy.Label = label.Label
			}
			hosts = append(hosts, hostCopy)
		}

		data := map[string]interface{}{
			"hosts":   hosts,
			"loading": false,
			"error":   nil,
		}

		jsonData, err := json.Marshal(data)
		if err != nil {
			log.Printf("JSON marshal error: %v", err)
			continue
		}

		sendBroadcast(jsonData)
		time.Sleep(time.Second)
	}
}

// 发送广播消息
func sendBroadcast(message []byte) {
	clientsMux.Lock()
	defer clientsMux.Unlock()

	log.Printf("本次广播数据的客户端数量：%d，广播数据量：%d\n", len(clients), len(message))
	for client := range clients {
		err := client.WriteMessage(websocket.TextMessage, message)
		if err != nil {
			log.Printf("Write error: %v", err)
			client.Close()
			delete(clients, client)
			continue
		}
	}
}
