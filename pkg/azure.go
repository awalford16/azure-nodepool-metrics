package main

import (
	"context"
	"fmt"
	"os"
	"regexp"
	"strings"

	"github.com/Azure/azure-sdk-for-go/sdk/azidentity"
	"github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/containerservice/armcontainerservice/v6"
	log "github.com/sirupsen/logrus"
)

type AzureClient struct {
	SubscriptionID       string
	NodepoolClient       *armcontainerservice.AgentPoolsClient
	ManagedClusterClient *armcontainerservice.ManagedClustersClient
}

type ClusterDetails struct {
	Name          string
	ResourceGroup string
	Fqdn          string
}

func NewClusterDetails() *ClusterDetails {
	return &ClusterDetails{}
}

func NewAzureClient() *AzureClient {
	subscriptionID := os.Getenv("AZURE_SUBSCRIPTION_ID")
	if subscriptionID == "" {
		log.Fatalf("AZURE_SUBSCRIPTION_ID environment variable not set")
	}

	cred, err := azidentity.NewDefaultAzureCredential(nil)
	if err != nil {
		log.Fatalf("failed to obtain a credential: %v", err)
	}
	clientFactory, err := armcontainerservice.NewClientFactory(subscriptionID, cred, nil)
	if err != nil {
		log.Fatalf("failed to create client: %v", err)
	}

	return &AzureClient{
		SubscriptionID:       subscriptionID,
		NodepoolClient:       clientFactory.NewAgentPoolsClient(),
		ManagedClusterClient: clientFactory.NewManagedClustersClient(),
	}
}

// Based on host, fill out the cluster details from Azure API
func (client *AzureClient) GetClusterDetailsFromContext(ctx context.Context, host string, clusterDetails *ClusterDetails) error {
	// List all clusters in subscription
	pager := client.ManagedClusterClient.NewListPager(&armcontainerservice.ManagedClustersClientListOptions{})
	for pager.More() {
		nextResult, err := pager.NextPage(ctx)
		if err != nil {
			log.Fatalf("failed to advance page: %v", err)
		}

		for _, v := range nextResult.Value {
			_ = v

			// Match the cluster Host field with the Azure FQDN to get cluster details
			// Or use the cluster_name env var if defined
			log.Debugf("Checking for cluster with name: %s", clusterDetails.Name)
			if strings.Contains(*v.Properties.Fqdn, clusterDetails.Name) && len(clusterDetails.Name) > 0 {
				err := parseAzureCluster(v, clusterDetails)
				if err != nil {
					return fmt.Errorf("failed to parse cluster: %v", err)
				}
				return nil
			} else if host == *v.Properties.Fqdn {
				log.Infof("%s not set, using host to find cluster", AzureClusterNameEnvVar)
				err := parseAzureCluster(v, clusterDetails)
				if err != nil {
					return fmt.Errorf("failed to parse cluster: %v", err)
				}
				return nil
			}
		}
	}

	return fmt.Errorf("could not find cluster with host: %s in subscription: %s", host, client.SubscriptionID)
}

// CREDIT: https://gist.github.com/vladbarosan/fb2528754cbd97df51ca11fe7be27d2f
// ParseResourceID parses a resource ID into a ResourceDetails struct
func ParseResourceID(resourceID string, clusterDetails *ClusterDetails) error {
	const resourceIDPatternText = `(?i)subscriptions/(.+)/resourceGroups/(.+)/providers/(.+?)/(.+?)/(.+)`
	resourceIDPattern := regexp.MustCompile(resourceIDPatternText)
	match := resourceIDPattern.FindStringSubmatch(resourceID)

	if len(match) == 0 {
		return fmt.Errorf("parsing failed for %s. Invalid resource Id format", resourceID)
	}

	v := strings.Split(match[5], "/")
	resourceName := v[len(v)-1]

	clusterDetails.Name = resourceName
	clusterDetails.ResourceGroup = match[2]

	return nil
}

func parseAzureCluster(cluster *armcontainerservice.ManagedCluster, clusterDetails *ClusterDetails) error {
	err := ParseResourceID(*cluster.ID, clusterDetails)
	if err != nil {
		return fmt.Errorf("failed to parse resource ID: %v", err)
	}
	log.Infof("Found cluster with Name: %s in RG: %s", clusterDetails.Name, clusterDetails.ResourceGroup)
	clusterDetails.Fqdn = *cluster.Properties.Fqdn

	return nil
}
