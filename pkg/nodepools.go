package main

import (
	"context"

	"github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/containerservice/armcontainerservice/v6"
	log "github.com/sirupsen/logrus"
)

type NodepoolStates struct {
	PowerState          *StateMetric
	ProvisioningState   *StateMetric
	ScalingState        *StateMetric
	ClusterScalingState *StateMetric
}

func NewNodepoolStates() *NodepoolStates {
	return &NodepoolStates{
		PowerState:          NewStateMetric([]string{}, "power", nodepoolPowerState),
		ProvisioningState:   NewStateMetric([]string{}, "provisioning", nodepoolProvisioningState),
		ScalingState:        NewStateMetric([]string{}, "scaling", nodepoolScalingState),
		ClusterScalingState: NewStateMetric([]string{}, "cluster_scaling", clusterScalingState),
	}
}

func (states *NodepoolStates) UpdateScalingStates(clusterName string, statuses *AutoScalerStatus) {
	// Update prometheus metrics for scaling state of each node group and cluster-wide
	states.ClusterScalingState.setState(clusterName, statuses.ClusterWide.Health.Status)
	for _, nodeGroup := range statuses.NodeGroupStatus {
		states.ScalingState.setState(nodeGroup.Name, nodeGroup.HealthStatus.Status)
	}
}

func (states *NodepoolStates) UpdateStatusStates(ctx context.Context, azureClient *AzureClient, clusterDetails *ClusterDetails) {
	// Get list of agent pools in cluster
	pager := azureClient.NodepoolClient.NewListPager(
		clusterDetails.ResourceGroup,
		clusterDetails.Name,
		&armcontainerservice.AgentPoolsClientListOptions{})

	for pager.More() {
		nextResult, err := pager.NextPage(ctx)
		if err != nil {
			log.Fatalf("failed to advance page: %v", err)
		}

		for _, v := range nextResult.Value {
			// Update power and provisioning states
			states.PowerState.setState(*v.Name, string(*v.Properties.PowerState.Code))
			states.ProvisioningState.setState(*v.Name, *v.Properties.ProvisioningState)
		}
	}
}
