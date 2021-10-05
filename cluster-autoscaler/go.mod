module k8s.io/autoscaler/cluster-autoscaler

go 1.16

require (
	cloud.google.com/go v0.81.0
	github.com/Azure/azure-sdk-for-go v55.0.0+incompatible
	github.com/Azure/go-autorest/autorest v0.11.18
	github.com/Azure/go-autorest/autorest/adal v0.9.13
	github.com/Azure/go-autorest/autorest/date v0.3.0
	github.com/Azure/go-autorest/autorest/to v0.4.0
	github.com/aws/aws-sdk-go v1.40.46
	github.com/digitalocean/godo v1.27.0
	github.com/ghodss/yaml v1.0.0
	github.com/go-macaroon-bakery/macaroon-bakery/v3 v3.0.0-20210309064400-d73aa8f92aa2
	github.com/golang/mock v1.5.0
	github.com/google/uuid v1.2.0
	github.com/jmespath/go-jmespath v0.4.0
	github.com/json-iterator/go v1.1.11
	github.com/juju/errors v0.0.0-20210818161939-5560c4c073ff
	github.com/juju/idmclient/v2 v2.0.0-20210309081103-6b4a5212f851
	github.com/juju/juju v0.0.0-20211001142224-cd69c141f7c8
	github.com/juju/names/v4 v4.0.0-20200929085019-be23e191fee0
	github.com/pkg/errors v0.9.1
	github.com/prometheus/client_golang v1.11.0
	github.com/satori/go.uuid v1.2.0
	github.com/spf13/pflag v1.0.5
	github.com/stretchr/testify v1.7.0
	golang.org/x/crypto v0.0.0-20210921155107-089bfa567519
	golang.org/x/oauth2 v0.0.0-20210514164344-f6687ab2804c
	google.golang.org/api v0.46.0
	gopkg.in/gcfg.v1 v1.2.0
	gopkg.in/juju/environschema.v1 v1.0.1-0.20201027142642-c89a4490670a
	gopkg.in/yaml.v2 v2.4.0
	k8s.io/api v0.23.0-alpha.1
	k8s.io/apimachinery v0.23.0-alpha.1
	k8s.io/apiserver v0.23.0-alpha.1
	k8s.io/client-go v0.23.0-alpha.1
	k8s.io/cloud-provider v0.23.0-alpha.1
	k8s.io/component-base v0.23.0-alpha.1
	k8s.io/component-helpers v0.23.0-alpha.1
	k8s.io/klog/v2 v2.9.0
	k8s.io/kubelet v0.0.0
	k8s.io/kubernetes v1.23.0-alpha.1
	k8s.io/legacy-cloud-providers v0.0.0
	k8s.io/utils v0.0.0-20210802155522-efc7438f0176
)

replace github.com/digitalocean/godo => github.com/digitalocean/godo v1.27.0

replace github.com/rancher/go-rancher => github.com/rancher/go-rancher v0.1.0

replace k8s.io/api => k8s.io/api v0.23.0-alpha.0

replace k8s.io/apiextensions-apiserver => k8s.io/apiextensions-apiserver v0.23.0-alpha.0

replace k8s.io/apimachinery => k8s.io/apimachinery v0.23.0-alpha.0

replace k8s.io/apiserver => k8s.io/apiserver v0.23.0-alpha.0

replace k8s.io/cli-runtime => k8s.io/cli-runtime v0.23.0-alpha.0

replace k8s.io/client-go => k8s.io/client-go v0.23.0-alpha.0

replace k8s.io/cloud-provider => k8s.io/cloud-provider v0.23.0-alpha.0

replace k8s.io/cluster-bootstrap => k8s.io/cluster-bootstrap v0.23.0-alpha.0

replace k8s.io/code-generator => k8s.io/code-generator v0.23.0-alpha.0

replace k8s.io/component-base => k8s.io/component-base v0.23.0-alpha.0

replace k8s.io/component-helpers => k8s.io/component-helpers v0.23.0-alpha.0

replace k8s.io/controller-manager => k8s.io/controller-manager v0.23.0-alpha.0

replace k8s.io/cri-api => k8s.io/cri-api v0.23.0-alpha.0

replace k8s.io/csi-translation-lib => k8s.io/csi-translation-lib v0.23.0-alpha.0

replace k8s.io/kube-aggregator => k8s.io/kube-aggregator v0.23.0-alpha.0

replace k8s.io/kube-controller-manager => k8s.io/kube-controller-manager v0.23.0-alpha.0

replace k8s.io/kube-proxy => k8s.io/kube-proxy v0.23.0-alpha.0

replace k8s.io/kube-scheduler => k8s.io/kube-scheduler v0.23.0-alpha.0

replace k8s.io/kubectl => k8s.io/kubectl v0.23.0-alpha.0

replace k8s.io/kubelet => k8s.io/kubelet v0.23.0-alpha.0

replace k8s.io/legacy-cloud-providers => k8s.io/legacy-cloud-providers v0.23.0-alpha.0

replace k8s.io/metrics => k8s.io/metrics v0.23.0-alpha.0

replace k8s.io/mount-utils => k8s.io/mount-utils v0.23.0-alpha.0

replace k8s.io/sample-apiserver => k8s.io/sample-apiserver v0.23.0-alpha.0

replace k8s.io/sample-cli-plugin => k8s.io/sample-cli-plugin v0.23.0-alpha.0

replace k8s.io/sample-controller => k8s.io/sample-controller v0.23.0-alpha.0

replace k8s.io/pod-security-admission => k8s.io/pod-security-admission v0.23.0-alpha.0

replace github.com/Azure/azure-sdk-for-go => github.com/Azure/azure-sdk-for-go v46.4.0+incompatible

replace k8s.io/autoscaler => /Users/pete/canonical/deployments/dsv/autoscaler/autoscaler
