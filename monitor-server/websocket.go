package main

import (
	"encoding/json"
	"log"
	"net/http"
	"sync"
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
	clientsMux sync.Mutex
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
		// 获取数据时加锁
		storeLock.RLock()
		hostLock.RLock()
		var hosts []HostData
		for _, hostData := range DataStore {
			hostData.Label = HostManage[hostData.IP].Label
			hosts = append(hosts, *hostData)
		}
		hostLock.RUnlock()
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
		sendBroadcast(jsonData)

		time.Sleep(time.Second) // 每秒更新一次
	}
}

// 发送广播消息
func sendBroadcast(message []byte) {
	// 获取客户端列表前加锁
	clientsMux.Lock()
	clientList := make([]*websocket.Conn, 0, len(clients))
	for client := range clients {
		clientList = append(clientList, client)
	}
	clientsMux.Unlock()

	log.Printf("本次广播数据的客户端数量：%d，广播数据量：%d\n", len(clientList), len(message))
	for _, client := range clientList {
		err := client.WriteMessage(websocket.TextMessage, message)
		if err != nil {
			log.Printf("Write error: %v", err)
			client.Close()
			delete(clients, client)
		}
	}
}
