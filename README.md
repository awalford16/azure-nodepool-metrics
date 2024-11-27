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
