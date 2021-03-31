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

	"github.com/juju/juju/api"
	names "github.com/juju/names/v4"

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
	connection      *api.Connection
	resourceLimiter *cloudprovider.ResourceLimiter
}

func newJujuCloudProvider(conn *api.Connection, rl *cloudprovider.ResourceLimiter) (*jujuCloudProvider, error) { //TODO
	return &jujuCloudProvider{
		connection:      conn,
		resourceLimiter: rl,
	}, nil
}

// Name returns name of the cloud provider.
func (j *jujuCloudProvider) Name() string {
	return cloudprovider.JujuProviderName
}

// NodeGroups returns all node groups configured for this cloud provider.
func (j *jujuCloudProvider) NodeGroups() []cloudprovider.NodeGroup {
	var groups []cloudprovider.NodeGroup
	groups = append(groups, cloudprovider.NodeGroup{
		id:         "juju",
		minSize:    3,
		maxSize:    10,
		target:     5,
		connection: &j.connection,
	})
	return groups
}

// NodeGroupForNode returns the node group for the given node, nil if the node
// should not be processed by cluster autoscaler, or non-nil error if such
// occurred. Must be implemented.
func (j *jujuCloudProvider) NodeGroupForNode(node *apiv1.Node) (cloudprovider.NodeGroup, error) {
	return cloudprovider.NodeGroup{
		id:         "juju",
		minSize:    3,
		maxSize:    10,
		target:     5,
		connection: &j.connection,
	}, nil
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

	type Controller struct {
		APIEndpoints  []string
		CACert        string
		PublicDNSName string
	}

	controller := Controller{
		APIEndpoints: []string{"10.245.164.22:17070", "10.5.0.8:17070", "252.0.8.1:17070", "10.245.164.12:17070",
			"10.5.0.7:17070", "252.0.7.1:17070", "10.245.164.198:17070", "10.5.0.18:17070",
			"252.0.18.1:17070"},
		CACert: "-----BEGIN CERTIFICATE-----" +
			"MIID8zCCAlugAwIBAgIUUcUGXoAQImjr9i5aXndLR+kUAfgwDQYJKoZIhvcNAQEL" +
			"BQAwITENMAsGA1UEChMESnVqdTEQMA4GA1UEAxMHanVqdS1jYTAeFw0yMTAzMDMy" +
			"MDE1NDZaFw0zMTAzMDMyMDIwNDZaMCExDTALBgNVBAoTBEp1anUxEDAOBgNVBAMT" +
			"B2p1anUtY2EwggGiMA0GCSqGSIb3DQEBAQUAA4IBjwAwggGKAoIBgQC8Zm2gP1q/" +
			"Y6AhJQWFAppHZgL2CJ0XwWj7TnKO4B/W7w0dSsj0KobdLIpihwZAEypUcUv9FjqS" +
			"MAQ55/syMzMoGsMQa/xIbQwH5JqhQKfsFyRX9yAJ6TYNfBlvo8tG8XsyqtE6eZgg" +
			"k+0nRm7XRsvf0ky5+grG6aAAn1PJLrF1SAUfW1KFvAFtM/k8OkuUwtowiJyzW4jL" +
			"4UXZk9VGlgsKAF99N+CICG3ySx8EY+NLtnWMmHfrmpdFlCFqd10xRowfvirv6rIy" +
			"nOTx1QBWJO36jQ3gECgBDHml6u2lQjJ+VaSRGakTYxO5R//Pmw2EHRbR90U/CJI7" +
			"SEMrTUzDzwfSCO1/UZ0ZsZnvFAj6LSjiLpAeWjXnh3jyAh5ltup9esDwD6QvP5hF" +
			"vkYRA3LZ40Y2yVANRrgLkHjl7w5LjbIHKJjpwpkfpXxlpLXzpGpMiMkxwKfccw44" +
			"Q04Ek8mnVp0+uqs8ak1WWtbi1tKDFMcrArq/0D5xhqq9w4iddxP385cCAwEAAaMj" +
			"MCEwDgYDVR0PAQH/BAQDAgKkMA8GA1UdEwEB/wQFMAMBAf8wDQYJKoZIhvcNAQEL" +
			"BQADggGBAGutPcN/cLPbwYtgWpeEftEn5ZErTq7gE4lVb1QSOjAg3L489A8eaw4i" +
			"Ko0sabhLQXASCe9J8NVT3VBYnEaNAm9Nb6GOwrtvq1H/4y06s4BYuSu2TKOqsxKK" +
			"tvc1Z7ARU0Dp13VVQm4xtX46Td29hYrHWtlm69shPLUe0gFsBqeOLu2jlMjk3vde" +
			"qB6j3EAUFg4uR9sy/CXKiDx0LewwrWm/dXs+GQ7Tr9atH8Wr6/Kwpu52s1mkcnaW" +
			"mM3JRXnohEjmggUiposPJfNzFmfvCo8iwm1rkt2UUHsnYQb5Kw+0sCsGbNyufGKh" +
			"TtjdnjRK+V9OjYEb9wS+aXL0sUn39MCcMJt1OxXRi+5nSQFBG7/G1B1KDqjtPLM9" +
			"NN15JkIidoRDcCjstxnr3a7oVlKzNPt4fukT+LEEH45+unfkD6i/FeMHm8aX3Xz0" +
			"mo3RpFFJb5Uhdwrpz214V2B3mI1bUxa0tE71uAyfLJA1CbpWIBs33yfb3Ky4PEv3" +
			"HOQEnSY0yw==" +
			"-----END CERTIFICATE-----",
		PublicDNSName: "",
	}

	model := map[string]string{
		"ModelUUID": "7c63f91a-706a-4da9-868b-c67d1f25ea4b",
	}

	account := map[string]string{
		"User":     "admin",
		"Password": "084aa8cffeea398788ed8df4747c6f2b",
	}

	conn, err := api.Open(&api.Info{
		Addrs:  controller.APIEndpoints,
		CACert: controller.CACert,
		// SNIHostName: controller["PublicDNSName"], // optional
		ModelTag: names.NewModelTag(model["ModelUUID"]),
		Tag:      names.NewUserTag(account["User"]),
		Password: account["Password"],
	}, api.DefaultDialOpts())

	provider, err := newJujuCloudProvider(&conn, rl)
	if err != nil {
		klog.Fatalf("Failed to create Juju cloud provider: %v", err)
	}

	return provider
}
