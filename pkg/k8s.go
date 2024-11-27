package main

import (
	"context"
	"fmt"
	"os"
	"regexp"

	"gopkg.in/yaml.v2"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"

	log "github.com/sirupsen/logrus"
)

var (
	CONFIGMAP_NAME      = "cluster-autoscaler-status"
	CONFIGMAP_NAMESPACE = "kube-system"
)

type StatusBlock struct {
	Status string `yaml:"status"`
}

type ScalingStatus struct {
	Name            string      `yaml:"name"`
	HealthStatus    StatusBlock `yaml:"health"`
	ScaleUpStatus   StatusBlock `yaml:"scaleUp"`
	ScaleDownStatus StatusBlock `yaml:"scaleDown"`
}

type ClusterWide struct {
	Health          StatusBlock `yaml:"health"`
	ScaleUpStatus   StatusBlock `yaml:"scaleUp"`
	ScaleDownStatus StatusBlock `yaml:"scaleDown"`
}

type AutoScalerStatus struct {
	ClusterWide     ClusterWide     `yaml:"clusterWide"`
	NodeGroupStatus []ScalingStatus `yaml:"nodeGroups"`
}

type K8sClient struct {
	Client      *kubernetes.Clientset
	IsInCluster bool
	Host        string
}

func NewK8sClient() (*K8sClient, error) {
	var config *rest.Config
	var err error
	isInCluster := os.Getenv("KUBERNETES_SERVICE_HOST") != ""

	// Setup auth to cluster
	if isInCluster {
		config, err = rest.InClusterConfig()
	} else {
		// Load kubeconfig from default location
		kubeConfigPath := getEnvOrDefault("KUBECONFIG_PATH", DefaultKubeConfigPath)
		config, err = clientcmd.BuildConfigFromFlags("", kubeConfigPath)
	}
	if err != nil {
		log.Fatalf("Error creating config: %v", err)
		return nil, err
	}

	// Create Kubernetes clientset
	clientset, err := kubernetes.NewForConfig(config)
	if err != nil {
		log.Fatalf("Error creating clientset: %v", err)
		return nil, err
	}

	return &K8sClient{Client: clientset, IsInCluster: isInCluster, Host: getCleanClusterHost(config.Host)}, nil
}

func (client *K8sClient) getConfigMapData() (*AutoScalerStatus, error) {
	var err error

	// Retrieve the ConfigMap data
	configMap, err := client.Client.CoreV1().ConfigMaps(CONFIGMAP_NAMESPACE).Get(context.TODO(), CONFIGMAP_NAME, metav1.GetOptions{})
	if err != nil {
		log.Warnf("Error getting ConfigMap: %v", err)
		return nil, err
	}

	// Process ConfigMap data
	var autoscalerStatus AutoScalerStatus
	err = yaml.Unmarshal([]byte(configMap.Data["status"]), &autoscalerStatus)
	if err != nil {
		fmt.Println(err)
		return nil, err
	}

	// Update metrics with scaling status
	return &autoscalerStatus, nil
}

func getCleanClusterHost(host string) string {
	// Remove protocol and port from host
	cleanHost := regexp.MustCompile(`.*://`).ReplaceAllString(host, "")
	cleanHost = regexp.MustCompile(`:\d+`).ReplaceAllString(cleanHost, "")
	log.Debugf("K8s host: %s", cleanHost)
	return cleanHost
}
