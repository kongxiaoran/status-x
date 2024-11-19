package main

import (
	"net/http"
	"time"
)

// 检查服务器网络连通性
func CheckServerNetwork() bool {
	successCount := 0
	testUrls := []string{
		"https://www.baidu.com",
		"https://www.qq.com",
		"https://www.taobao.com",
		"https://www.jd.com",
	}

	client := &http.Client{
		Timeout: 5 * time.Second,
	}

	for _, url := range testUrls {
		resp, err := client.Get(url)
		if err == nil {
			resp.Body.Close()
			successCount++
		}
	}

	// 成功连接2个以上网站则认为网络正常
	return successCount >= 2
}
