package main

import (
	"flag"
	"fmt"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/clientcmd"
	metricsclientset "k8s.io/metrics/pkg/client/clientset/versioned"
	"os"
)

func initKubernetesClient() (*kubernetes.Clientset, *metricsclientset.Clientset) {
	kubeconfig := flag.String("kubeconfig.yaml", "kubeconfig.yaml", "absolute path to the kubeconfig.yaml file")
	flag.Parse()

	config, err := clientcmd.BuildConfigFromFlags("", *kubeconfig)
	if err != nil {
		fmt.Printf("Error building kubeconfig.yaml: %s\n", err.Error())
		os.Exit(1)
	}

	clientset, err := kubernetes.NewForConfig(config)
	if err != nil {
		fmt.Printf("Error creating Kubernetes client: %s\n", err.Error())
		os.Exit(1)
	}

	metricsClientset, err := metricsclientset.NewForConfig(config)
	if err != nil {
		fmt.Printf("Error creating Metrics client: %s\n", err.Error())
		os.Exit(1)
	}

	return clientset, metricsClientset
}
