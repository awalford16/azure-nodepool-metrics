package main

import (
	"os"

	log "github.com/sirupsen/logrus"
	"k8s.io/client-go/util/homedir"
)

var (
	DefaultKubeConfigPath  = homedir.HomeDir() + "/.kube/config"
	DefaultLogLevel        = "INFO"
	AzureClusterNameEnvVar = "AZURE_CLUSTER_NAME"
	EnableScalingMetrics   = true
)

func getEnvOrDefault(envVar string, defaultValue string) string {
	value := os.Getenv(envVar)
	if value == "" {
		return defaultValue
	}
	return value
}

func configureLogging() {
	log.SetFormatter(&log.JSONFormatter{
		TimestampFormat: "2006-01-02T15:04:05.000",
	})
	log.SetOutput(os.Stdout)

	level := getEnvOrDefault("LOG_LEVEL", DefaultLogLevel)
	logLevel, err := log.ParseLevel(level)
	if err == nil {
		log.SetLevel(logLevel)
	} else {
		log.Infof("Unable to parse log Level %v. Setting level as default", level)
	}
}
