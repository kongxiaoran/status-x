package main

//
//import (
//	"crypto/tls"
//	"flag"
//	"fmt"
//	"io/ioutil"
//	"k8s.io/client-go/rest"
//	"k8s.io/client-go/tools/clientcmd"
//	"k8s.io/klog"
//	"net/http"
//)
//
//func main() {
//	// 使用 Kubeconfig 文件加载配置
//	// 初始化 Kubernetes 和 Metrics ClientSet
//	kubeconfig := flag.String("kubeconfig.yaml", "kubeconfig.yaml", "absolute path to the kubeconfig.yaml file")
//	flag.Parse()
//
//	config, err := clientcmd.BuildConfigFromFlags("", *kubeconfig)
//
//	// 使用配置创建一个 HTTP 客户端
//	client, err := restClientForKubelet(config)
//	if err != nil {
//		klog.Fatalf("Failed to create HTTP client: %v", err)
//	}
//
//	// 发送 HTTP 请求到 Kubelet 的 /stats/summary API
//	nodeIP := "NODE_IP_ADDRESS" // 替换为你实际的节点 IP
//	kubeletURL := fmt.Sprintf("https://%s:10250/stats/summary", nodeIP)
//
//	resp, err := client.Get(kubeletURL)
//	if err != nil {
//		klog.Fatalf("Failed to send request to Kubelet: %v", err)
//	}
//	defer resp.Body.Close()
//
//	body, err := ioutil.ReadAll(resp.Body)
//	if err != nil {
//		klog.Fatalf("Failed to read response: %v", err)
//	}
//
//	fmt.Println("Kubelet Metrics Response: ", string(body))
//}
//
//// 创建一个与 Kubelet API 交互的 HTTP 客户端
//func restClientForKubelet(config *rest.Config) (*http.Client, error) {
//	// 生成 TLS 配置，跳过 Kubelet 自签名证书验证
//	tlsConfig := &tls.Config{InsecureSkipVerify: true}
//
//	// 使用 Kubernetes 配置中的 Transport
//	//transportConfig, err := rest.TransportFor(config)
//	//if err != nil {
//	//	return nil, err
//	//}
//
//	client := &http.Client{
//		Transport: &http.Transport{
//			TLSClientConfig: tlsConfig,
//			Proxy:           http.ProxyFromEnvironment,
//		},
//	}
//	return client, nil
//}
