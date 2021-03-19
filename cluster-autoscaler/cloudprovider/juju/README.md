# Cluster Autoscaler for Juju

The cluster autoscaler for Juju scales worker nodes within any
specified charmed-kubernetes cluster's node pool. 

# Configuration

The `cluster-autoscaler` dynamically runs based on tags associated with node
pools. These are the current valid tags:

```
k8s-cluster-autoscaler-enabled:true
k8s-cluster-autoscaler-min:3
k8s-cluster-autoscaler-max:10
```

The syntax is in form of `key:value`.

* If `k8s-cluster-autoscaler-enabled:true` is absent or
  `k8s-cluster-autoscaler-enabled` is **not** set to `true`, the
  `cluster-autoscaler` will not process the node pool by default.
* To set the minimum number of nodes to use `k8s-cluster-autoscaler-min`
* To set the maximum number of nodes to use `k8s-cluster-autoscaler-max`


If you don't set the minimum and maximum tags, node pools will have the
following default limits:

```
minimum number of nodes: 1
maximum number of nodes: 200
```

# Development

Make sure you're inside the root path of the [autoscaler
repository](https://github.com/kubernetes/autoscaler)

1.) Build the `cluster-autoscaler` binary:


```
make build-in-docker
```

