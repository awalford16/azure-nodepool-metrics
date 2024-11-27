# Nodepool Metrics

Provides a way of monitoring Azure nodepool states via Prometheus metrics:

```
# TYPE nodepool_power_state gauge
nodepool_power_state{nodepool="Running",state="mynodepool"} 1
nodepool_power_state{nodepool="Running",state="default"} 1
# HELP nodepool_provisioning_state The provisioning status of a nodepool
# TYPE nodepool_provisioning_state gauge
nodepool_provisioning_state{nodepool="Succeeded",state="mynodepool"} 1
nodepool_provisioning_state{nodepool="Succeeded",state="default"} 1
# HELP nodepool_scaling_state The health state of nodepool autoscaling
# TYPE nodepool_scaling_state gauge
nodepool_scaling_state{nodepool="Healthy",state="aks-mynodepool-10354089-vmss"} 1
```

The service will continuously poll the Azure Nodepool API for nodepool updates and set a gauge based on the state. If the state changes it will reset all previously reported states to 0 and set the current state to 1, so there should be no nodepool reporting multiple states.


## Configuration

The service requires only one environment variable: `AZURE_SUBSCRIPTION_ID`.

It will require credentials to authenticate with this subscription which can be provided with `AZURE_CLIENT_ID` and `AZURE_CLIENT_SECRET`, or use local Azure credentials.

The service also requires access to a K8s cluster, either via a `~/.kube/config` file or using in-cluster config.

If you are using in-cluster config, you will need to set `AZURE_CLUSTER_NAME` environment variable. For `kubeconfig` file setup, the service will read from the `host` field for the cluster name, however this will not work from within an Azure cluster and so the cluster name will need to be set explicitly to call the Azure API to gather nodepool metrics.


## Deployment

Azure nodepool-metrics can be deployed via the helm chart found under the deploy directory.

Since this deployment will be running in-cluster, you will need to pass the `AZURE_CLUSTER_NAME` as well as the Azure credentials in the helm chart values:

```yaml
# values.yaml
cluster: YOUR_CLUSTER

azureCredentials:
    data:
        # -- Base64 encoded Azure Subscription ID
        azure_subscription_id: B64_AZURE_SUBSCRIPTION_ID

        # -- Base64 encoded Azure Tenant ID
        azure_tenant_id: B64_AZURE_TENANT_ID

        # -- Base64 encoded Azure Client ID
        azure_client_id: B64_AZURE_CLIENT_ID

        # -- Base64 encoded Azure Client Secret
        azure_client_secret: B64_AZURE_CLIENT_SECRET
```

Install with the commands below:

```bash
cd deploy/azure-nodepool-metrics

helm install azure-nodepool-metrics . -f values.yaml
```
