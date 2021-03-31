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
	"fmt"

	"github.com/juju/juju/api"
	"github.com/juju/juju/api/application"
	apiv1 "k8s.io/api/core/v1"

	"k8s.io/autoscaler/cluster-autoscaler/cloudprovider"
	"k8s.io/autoscaler/cluster-autoscaler/config"
	schedulerframework "k8s.io/kubernetes/pkg/scheduler/framework"
)

const (
	nodeIDLabel        = "juju/node-id"
)

type NodeGroup struct {
	id         string
	minSize    int
	maxSize    int
	target     int
	connection *api.Connection
}

// MaxSize returns maximum size of the node group.
func (n *NodeGroup) MaxSize() int {
	return n.maxSize
}

// MinSize returns minimum size of the node group.
func (n *NodeGroup) MinSize() int {
	return n.minSize
}

// TargetSize returns the current target size of the node group. It is possible that the
// number of nodes in Kubernetes is different at the moment but should be equal
// to Size() once everything stabilizes (new nodes finish startup and registration or
// removed nodes are deleted completely). Implementation required.
func (n *NodeGroup) TargetSize() (int, error) {
	return n.target, nil
}

// IncreaseSize increases the size of the node group. To delete a node you need
// to explicitly name it and use DeleteNode. This function should wait until
// node group size is updated. Implementation required.
func (n *NodeGroup) IncreaseSize(delta int) error {
	if delta <= 0 {
		return fmt.Errorf("delta must be positive, have: %d", delta)
	}

	targetSize := n.target + delta

	if targetSize > n.MaxSize() {
		return fmt.Errorf("size increase is too large. current: %d desired: %d max: %d",
			n.target, targetSize, n.MaxSize())
	}

	appClient := application.NewClient(*n.connection)
	client := api.Client(*n.connection)

	appClient.AddUnits(application.AddUnitsParams{
		ApplicationName: "kubernetes-worker",
		NumUnits:        delta,
	})
	//loop through and delay
	client.Status([]string{"kubernetes-worker"})
	if updatedCount != targetSize { // TODO check to see if the node group size actually increased
		return fmt.Errorf("couldn't increase size to %d (delta: %d). Current size is: %d",
			targetSize, delta, updatedCount)
	}

	// update internal cache
	n.target = targetSize
	return nil
}

// DeleteNodes deletes nodes from this node group. Error is returned either on
// failure or if the given node doesn't belong to this node group. This function
// should wait until node group size is updated. Implementation required.
func (n *NodeGroup) DeleteNodes(nodes []*apiv1.Node) error {
	// ctx := context.Background()
	for _, node := range nodes {
		nodeID, ok := node.Labels[nodeIDLabel] //TODO create nodeID label
		if !ok {
			// CA creates fake node objects to represent upcoming VMs that
			// haven't registered as nodes yet. We cannot delete the node at
			// this point.
			return fmt.Errorf("cannot delete node %q on node pool %q: node ID label %q is missing", node.Name, n.id, nodeIDLabel)
		}

		appClient := application.NewClient(*n.connection)
		appClient.DestroyUnits(application.DestroyUnitsParams{
			Units: []string{node}
		})
		if err != nil {
			return fmt.Errorf("deleting node failed for node pool: %q node: %q: %s",
				n.id, nodeID, err)
		}

		// decrement the count by one  after a successful delete
		n.target--
	}

	return nil
}

// DecreaseTargetSize decreases the target size of the node group. This function
// doesn't permit to delete any existing node and can be used only to reduce the
// request for new nodes that have not been yet fulfilled. Delta should be negative.
// It is assumed that cloud provider will not delete the existing nodes when there
// is an option to just decrease the target. Implementation required.
func (n *NodeGroup) DecreaseTargetSize(delta int) error {
	if delta >= 0 {
		return fmt.Errorf("delta must be negative, have: %d", delta)
	}

	targetSize := n.target + delta
	if targetSize < n.MinSize() {
		return fmt.Errorf("size decrease is too small. current: %d desired: %d min: %d",
			n.target, targetSize, n.MinSize())
	}

	// update internal cache
	n.target = targetSize
	return nil
}

// Id returns an unique identifier of the node group.
func (n *NodeGroup) Id() string {
	return n.id
}

// Debug returns a string containing all information regarding this node group.
func (n *NodeGroup) Debug() string {
	return fmt.Sprintf("cluster ID: %s (min:%d max:%d)", n.id, n.MinSize(), n.MaxSize())
}

// Nodes returns a list of all nodes that belong to this node group.
// It is required that Instance objects returned by this method have Id field set.
// Other fields are optional.
// This list should include also instances that might have not become a kubernetes node yet.
func (n *NodeGroup) Nodes() ([]cloudprovider.Instance, error) {
	// TODO RETURN NODES
	nodes := make([]cloudprovider.Instance, 1)
	// n.client.Status([]string{"kubernetes-worker"})

	// for len of nodes
	// append instance to nodes

	return nodes, nil
}

// TemplateNodeInfo returns a schedulerframework.NodeInfo structure of an empty
// (as if just started) node. This will be used in scale-up simulations to
// predict what would a new node look like if a node group was expanded. The returned
// NodeInfo is expected to have a fully populated Node object, with all of the labels,
// capacity and allocatable information as well as all pods that are started on
// the node by default, using manifest (most likely only kube-proxy). Implementation optional.
func (n *NodeGroup) TemplateNodeInfo() (*schedulerframework.NodeInfo, error) {
	return nil, cloudprovider.ErrNotImplemented
}

// Exist checks if the node group really exists on the cloud provider side. Allows to tell the
// theoretical node group from the real one. Implementation required.
func (n *NodeGroup) Exist() bool {
	return true //TODO IMPLEMENT LOGIC
}

// Create creates the node group on the cloud provider side. Implementation optional.
func (n *NodeGroup) Create() (cloudprovider.NodeGroup, error) {
	return nil, cloudprovider.ErrNotImplemented
}

// Delete deletes the node group on the cloud provider side.
// This will be executed only for autoprovisioned node groups, once their size drops to 0.
// Implementation optional.
func (n *NodeGroup) Delete() error {
	return cloudprovider.ErrNotImplemented
}

// Autoprovisioned returns true if the node group is autoprovisioned. An autoprovisioned group
// was created by CA and can be deleted when scaled to 0.
func (n *NodeGroup) Autoprovisioned() bool {
	return false
}

// GetOptions returns NodeGroupAutoscalingOptions that should be used for this particular
// NodeGroup. Returning a nil will result in using default options.
// Implementation optional.
func (n *NodeGroup) GetOptions(defaults config.NodeGroupAutoscalingOptions) (*config.NodeGroupAutoscalingOptions, error) {
	return nil, cloudprovider.ErrNotImplemented
}
