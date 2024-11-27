package main

import (
	"context"
	"net/http"
	"time"

	"github.com/prometheus/client_golang/prometheus/promhttp"

	log "github.com/sirupsen/logrus"
)

func main() {
	ctx := context.Background()

	// Set up logger
	configureLogging()

	// Create Azure client
	azureClient := NewAzureClient()
	clusterDetails := NewClusterDetails()
	clusterDetails.Name = getEnvOrDefault(AzureClusterNameEnvVar, "")

	// Objects to store states of nodepools
	states := NewNodepoolStates()
	k8sClient, err := NewK8sClient()
	if err != nil {
		log.Fatalf("failed to create k8s client: %v", err)
	}

	// Require cluster name to be provided when running in-cluster
	if len(clusterDetails.Name) == 0 && k8sClient.IsInCluster {
		log.Fatalf("Cluster name must be provided via environment variable %s", AzureClusterNameEnvVar)
	}

	// Use in-cluster details
	err = azureClient.GetClusterDetailsFromContext(ctx, k8sClient.Host, clusterDetails)
	if err != nil {
		log.Fatalf("Failed to get cluster details: %v", err)
	}

	// Check autoscaler configmap exists and can be parsed
	_, err = k8sClient.getConfigMapData()
	if err != nil {
		log.Warnf("Disabling scaling metrics. Error parsing configmap: %v", err)
		EnableScalingMetrics = false
	}

	// Register prometheus metrics
	RegisterMetrics(EnableScalingMetrics)

	// Start metric collection
	go func() {
		for {
			// Update provisioning/power states based on agentpool data from API
			states.UpdateStatusStates(ctx, azureClient, clusterDetails)

			// Only update scaling states if the configmap exists
			if EnableScalingMetrics {
				// Get autoscaler status data from configmap
				autoscaleStatus, err := k8sClient.getConfigMapData()
				if err != nil {
					log.Fatalf("failed to get autoscaler status: %v", err)
				}

				// Update prometheus metrics with configmap data
				states.UpdateScalingStates(clusterDetails.Name, autoscaleStatus)
			}
			time.Sleep(1 * time.Minute)
		}
	}()

	// Set up HTTP handler for metrics endpoint
	http.Handle("/metrics", promhttp.Handler())

	// Start the server
	log.Fatal(http.ListenAndServe(":8002", nil))
}
