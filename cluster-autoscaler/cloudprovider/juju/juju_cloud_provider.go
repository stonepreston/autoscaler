/*
Copyright 2019 The Kubernetes Authors.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package juju

import (
	"io"
	"os"

	apiv1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
	"k8s.io/autoscaler/cluster-autoscaler/cloudprovider"
	"k8s.io/autoscaler/cluster-autoscaler/config"
	"k8s.io/autoscaler/cluster-autoscaler/utils/errors"
	klog "k8s.io/klog/v2"
)

var _ cloudprovider.CloudProvider = (*jujuCloudProvider)(nil)

const (
	// GPULabel is the label added to nodes with GPU resource.
	GPULabel = "juju/gpu-node"
)

// jujuCloudProvider implements CloudProvider interface.
type jujuCloudProvider struct {
	resourceLimiter *cloudprovider.ResourceLimiter
	nodeGroup *NodeGroup
}

func newJujuCloudProvider(rl *cloudprovider.ResourceLimiter, nodeGroup *NodeGroup) (*jujuCloudProvider, error) { //TODO
	return &jujuCloudProvider{
		resourceLimiter: rl,
		nodeGroup: nodeGroup,
	}, nil
}

// Name returns name of the cloud provider.
func (j *jujuCloudProvider) Name() string {
	return cloudprovider.JujuProviderName
}

// NodeGroups returns all node groups configured for this cloud provider.
func (j *jujuCloudProvider) NodeGroups() []cloudprovider.NodeGroup {
	return []cloudprovider.NodeGroup{j.nodeGroup}
}

// NodeGroupForNode returns the node group for the given node, nil if the node
// should not be processed by cluster autoscaler, or non-nil error if such
// occurred. Must be implemented.
func (j *jujuCloudProvider) NodeGroupForNode(node *apiv1.Node) (cloudprovider.NodeGroup, error) {
	return j.nodeGroup, nil
}

// Pricing returns pricing model for this cloud provider or error if not
// available. Implementation optional.
func (j *jujuCloudProvider) Pricing() (cloudprovider.PricingModel, errors.AutoscalerError) {
	return nil, cloudprovider.ErrNotImplemented
}

// GetAvailableMachineTypes get all machine types that can be requested from
// the cloud provider. Implementation optional.
func (j *jujuCloudProvider) GetAvailableMachineTypes() ([]string, error) {
	return []string{}, nil
}

// NewNodeGroup builds a theoretical node group based on the node definition
// provided. The node group is not automatically created on the cloud provider
// side. The node group is not returned by NodeGroups() until it is created.
// Implementation optional.
func (j *jujuCloudProvider) NewNodeGroup(
	machineType string,
	labels map[string]string,
	systemLabels map[string]string,
	taints []apiv1.Taint,
	extraResources map[string]resource.Quantity,
) (cloudprovider.NodeGroup, error) {
	return nil, cloudprovider.ErrNotImplemented
}

// GetResourceLimiter returns struct containing limits (max, min) for
// resources (cores, memory etc.).
func (j *jujuCloudProvider) GetResourceLimiter() (*cloudprovider.ResourceLimiter, error) {
	return j.resourceLimiter, nil
}

// GPULabel returns the label added to nodes with GPU resource.
func (j *jujuCloudProvider) GPULabel() string {
	return GPULabel
}

// GetAvailableGPUTypes return all available GPU types cloud provider supports.
func (j *jujuCloudProvider) GetAvailableGPUTypes() map[string]struct{} {
	return nil
}

// Cleanup cleans up open resources before the cloud provider is destroyed,
// i.e. go routines etc.
func (j *jujuCloudProvider) Cleanup() error {
	return nil
}

// Refresh is called before every main loop and can be used to dynamically
// update cloud provider state. In particular the list of node groups returned
// by NodeGroups() can change as a result of CloudProvider.Refresh().
func (j *jujuCloudProvider) Refresh() error {
	// klog.V(4).Info("Refreshing node group cache")
	// return d.manager.Refresh() //TODO
	return nil
}

// BuildJuju builds the Juju cloud provider.
func BuildJuju(
	opts config.AutoscalingOptions,
	do cloudprovider.NodeGroupDiscoveryOptions,
	rl *cloudprovider.ResourceLimiter,
) cloudprovider.CloudProvider {
	var configFile io.ReadCloser
	if opts.CloudConfig != "" {
		var err error
		configFile, err = os.Open(opts.CloudConfig)
		if err != nil {
			klog.Fatalf("Couldn't open cloud provider configuration %s: %#v", opts.CloudConfig, err)
		}
		defer configFile.Close()
	}

	man := &Manager{
		units:    make(map[string]*Unit),
	}
	man.init()

	ng := &NodeGroup{
		id:         "juju",
		minSize:    3,
		maxSize:    10,
		target:     len(man.units),
		manager:    man,
	}
	provider, err := newJujuCloudProvider(rl, ng)
	if err != nil {
		klog.Fatalf("Failed to create Juju cloud provider: %v", err)
	}

	return provider
}
