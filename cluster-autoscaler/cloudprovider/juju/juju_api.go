package juju

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"strings"

	"github.com/juju/juju/apiserver/params"

	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/autoscaler/cluster-autoscaler/cloudprovider"
	"k8s.io/autoscaler/cluster-autoscaler/cloudprovider/juju/api"
	"k8s.io/autoscaler/cluster-autoscaler/cloudprovider/juju/client"
	"k8s.io/autoscaler/cluster-autoscaler/config"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/clientcmd"
	klog "k8s.io/klog/v2"
)

type Unit struct {
	state      cloudprovider.InstanceState
	jujuName   string
	kubeName   string
	workload   string
	agent      string
	registered bool
}

type Manager struct {
	units  map[string]*Unit
	config config.AutoscalingOptions
}

// FOR TESTING
func dump(name string, value interface{}) {
	output, _ := json.MarshalIndent(value, "", "	")
	fmt.Println("=== ", name)
	fmt.Println(string(output))
	fmt.Println()
}

func (m *Manager) init() error {

	// rootClient := root.Client()
	client, err := client.NewClient()
	if err != nil {
		log.Fatal(err)
	}

	jujuStatus := m.getStatus()

	// panic("dumped models")
	keys := make([]string, 0, len(jujuStatus["applications"]["kubernetes-worker"]["units"]))
	for k := range jujuStatus["applications"]["kubernetes-worker"]["units"] {
		keys = append(keys, k)
	}
	for k8sworker := range keys {
		machine := jujuStatus["applications"]["kubernetes-worker"]["units"][k8sworker]["machine"]
		hostname := jujuStatus["machines"][machine]["hostname"]
		m.units[keys[k8sworker]] = &Unit{
			state:      cloudprovider.InstanceRunning,
			jujuName:   keys[k8sworker],
			kubeName:   hostname,
			registered: true,
		}

	}

	return nil
}

func (m *Manager) scaleUnits(name string, delta int) error {

	prevStatus := m.getStatus()
	client, err := client.NewClient()
	if err != nil {
		return err
	}

	applicationAPI := api.NewApplicationAPI(client)

	applicationAPI.ScaleApplication(name, delta)

	for key, _ := range m.getStatus() {
		if _, ok := prevStatus[key]; !ok {
			m.units[key] = &Unit{
				state:    cloudprovider.InstanceCreating,
				jujuName: key,
			}
		}
	}

	return nil
}

// func (m *Manager) removeUnit(name string, target int) error {
// 	unit := m.getUnit(name)
// 	unit.state = cloudprovider.InstanceDeleting
// 	client, err := client.NewClient()
// 	if err != nil {
// 		return err
// 	}
// 	// cmd = exec.Cmd{
// 	// 	Path:   juju,
// 	// 	Args:   []string{juju, "remove-unit", unit.jujuName},
// 	// 	Stderr: os.Stdout,
// 	// }
// 	applicationAPI := api.NewApplicationAPI(client)

// 	applicationAPI.ScaleApplication(name, target)
// 	// TODO: is this required?
// 	// cmd := exec.Cmd{
// 	// 	Path:   juju,
// 	// 	Args:   []string{juju, "run-action", unit.jujuName, "pause", "--wait"},
// 	// 	Stderr: os.Stdout,
// 	// }
// 	// cmd.Run()

// 	return nil
// }

func (m *Manager) refresh() error {
	for key, val := range m.getStatus() {
		if _, ok := m.units[key]; ok {
			m.units[key].agent = val[0]
			m.units[key].workload = val[1]
		}
	}
	jujuStatus := m.getStatus()
	// panic("dumped models")
	keys := make([]string, 0, len(jujuStatus["applications"]["kubernetes-worker"]["units"]))
	for k := range jujuStatus["applications"]["kubernetes-worker"]["units"] {
		keys = append(keys, k)
	}

	for _, unit := range m.units {
		if unit.state == cloudprovider.InstanceCreating {
			if unit.kubeName == "" {
				machine := jujuStatus["applications"]["kubernetes-worker"]["units"][unit.jujuName]["machine"]
				hostname := jujuStatus["machines"][machine]["hostname"]
				if len(strings.Fields(string(hostname))) > 0 {
					unit.kubeName = strings.Fields(string(hostname))[0]
				}
			}

			// use the current context in kubeconfig
			config, err := clientcmd.BuildConfigFromFlags("", m.config.KubeConfigPath)
			if err != nil {
				panic(err.Error())
			}

			// create the clientset
			clientset, err := kubernetes.NewForConfig(config)
			if err != nil {
				panic(err.Error())
			}
			if unit.workload == "active" && !unit.registered {

				// TODO: Test
				_, err := clientset.CoreV1().Nodes().Patch(context.TODO(), "pjds-focal-kvm", types.StrategicMergePatchType, []byte(`{"metadata":{"labels":{"test": "true"}}}`), v1.PatchOptions{})
				// :output, err = clientset.NodeV1()..Get(context.TODO(), "pjds-focal-kvm", metav1.GetOptions{})
				// patch := []byte(`{"metadata":{"labels":{"test":"go-two"}}}`)
				// output, err := clientset.CoreV1().Pods(namespace).Patch(context.TODO(), pod, types.StrategicMergePatchType, patch, v1.PatchOptions{})

				if err == nil {
					unit.registered = true
					unit.state = cloudprovider.InstanceRunning
					klog.Warningf(unit.kubeName + " registered.")
				}
			}
		} else if unit.state == cloudprovider.InstanceDeleting {
			delete(m.units, unit.jujuName)
		}
	}

	return nil
}

func (m *Manager) getUnit(name string) *Unit {
	for _, unit := range m.units {
		if unit.kubeName == name {
			return unit
		}
	}
	return nil
}

func (m *Manager) getStatus() *params.FullStatus {

	// rootClient := root.Client()
	client, err := client.NewClient()
	if err != nil {
		log.Fatal(err)
	}

	statusAPI := api.NewStatusAPI(client)
	jujuStatus, err := statusAPI.FullStatus(nil)
	if err != nil {
		log.Fatal(err)
	}

	return jujuStatus
}
