package juju

import (
	"encoding/json"
	"fmt"
	"log"
	"strconv"
	"strings"
	"sync"

	"github.com/juju/errors"
	"github.com/juju/juju/apiserver/params"

	apiv1 "k8s.io/api/core/v1"
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
	client *client.Client
	mu     sync.Mutex
}

// FOR TESTING
func dump(name string, value interface{}) {
	output, _ := json.MarshalIndent(value, "", "	")
	fmt.Println("=== ", name)
	fmt.Println(string(output))
	fmt.Println()
}

func (m *Manager) init() error {

	jujuStatus := m.getStatus()

	// panic("dumped models")
	keys := make([]string, 0, len(jujuStatus.Applications["kubernetes-worker"].Units))
	for k := range jujuStatus.Applications["kubernetes-worker"].Units {
		keys = append(keys, k)
	}

	for k8sworker := range keys {
		machine := jujuStatus.Applications["kubernetes-worker"].Units[strconv.Itoa(k8sworker)].Machine
		hostname := jujuStatus.Machines[machine].Hostname
		m.units[keys[k8sworker]] = &Unit{
			state:      cloudprovider.InstanceRunning,
			jujuName:   keys[k8sworker],
			kubeName:   hostname,
			registered: true,
		}

	}

	return nil
}

func (m *Manager) getApplicationAPI() (*api.ApplicationAPI, error) {
	if m.client != nil {
		var err error
		m.client, err = client.NewClient()
		if err != nil {
			log.Fatal(err)
		}

	}

	return api.NewApplicationAPI(m.client), nil

}

func (m *Manager) removeUnits(nodeHostnames []*apiv1.Node) error {
	prevStatus := m.getStatus()

	kubernetesWorkerMachine := make([]string, len(nodeHostnames))
	// find unit by hostname
	for index := range nodeHostnames {
		for key := range prevStatus.Machines {
			klog.Warningf("Finding Hostname -> machine. Comparison: %s with %s", nodeHostnames[index].ObjectMeta.Name, prevStatus.Machines[key].Hostname)
			if nodeHostnames[index].ObjectMeta.Name == prevStatus.Machines[key].Hostname {
				kubernetesWorkerMachine = append(kubernetesWorkerMachine, key)
				break
			}

		}
	}
	// map machines to units.
	units := make([]string, len(nodeHostnames))
	for machine := range kubernetesWorkerMachine {
		for key, _ := range prevStatus.Applications["kubernetes-worker"].Units {
			klog.Warningf("Finding machine -> Unit. Comparison: %s with %s", prevStatus.Applications["kubernetes-worker"].Units[key].Machine, kubernetesWorkerMachine[machine])
			if prevStatus.Applications["kubernetes-worker"].Units[key].Machine == kubernetesWorkerMachine[machine] {
				units = append(units, key)
				unit := m.getUnit(prevStatus.Machines[kubernetesWorkerMachine[machine]].Hostname)
				if unit != nil {
					unit.state = cloudprovider.InstanceDeleting
				}
			}
		}
	}

	applicationAPI, err := m.getApplicationAPI()
	if err != nil {
		log.Fatal(err)
	}
	klog.Warningf("Removing units: %s", units)
	_, err = applicationAPI.RemoveUnits(units)
	if err != nil {
		log.Fatal(err)
	}
	return nil
}

func (m *Manager) addUnits(name string, delta int) error {
	prevStatus := m.getStatus()
	applicationAPI, err := m.getApplicationAPI()

	if err != nil {
		return err
	}

	_, err = applicationAPI.AddUnit(name, delta)
	if err != nil {
		panic(errors.Trace(err))
	}
	jujuStatus := m.getStatus()

	for key, _ := range jujuStatus.Applications["kubernetes-worker"].Units {
		if _, ok := prevStatus.Applications["kubernetes-worker"].Units[key]; !ok {
			m.units[key] = &Unit{
				state:    cloudprovider.InstanceCreating,
				jujuName: key,
			}
		}
	}

	return nil
}

func (m *Manager) refresh() error {
	jujuStatus := m.getStatus()
	for key, val := range jujuStatus.Applications["kubernetes-worker"].Units {
		if _, ok := m.units[key]; ok {
			m.units[key].agent = val.AgentStatus.Info
			m.units[key].workload = val.WorkloadStatus.Info
		}
	}

	for _, unit := range m.units {
		if unit.state == cloudprovider.InstanceCreating {
			if unit.kubeName == "" {
				machine := jujuStatus.Applications["kubernetes-worker"].Units[unit.jujuName].Machine
				hostname := jujuStatus.Machines[machine].Hostname
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
			_, err = kubernetes.NewForConfig(config)
			if err != nil {
				panic(err.Error())
			}
			if unit.workload == "active" && !unit.registered {
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
	m.mu.Lock()
	if m.client == nil {
		var err error
		m.client, err = client.NewClient()
		if err != nil {
			log.Fatal(err)
		}
	}

	statusAPI := api.NewStatusAPI(m.client)

	jujuStatus, err := statusAPI.FullStatus(nil)
	if err != nil {
		log.Fatal(err)
	}
	m.mu.Unlock()
	return jujuStatus
}
