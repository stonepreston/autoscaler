package juju

import (
	"context"
	"os"
	"os/exec"
	"strconv"
	"strings"

	"github.com/juju/juju/api"
	"github.com/juju/juju/api/application"
	"github.com/juju/juju/juju"
	"github.com/juju/juju/jujuclient"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/autoscaler/cluster-autoscaler/cloudprovider"
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
	units      map[string]*Unit
	config     config.AutoscalingOptions
}

func (m *Manager) getRoot() api.Connection {

	params := juju.NewAPIConnectionParams{
		ControllerName: "test",
		Store: jujuclient.NewFileClientStore(),
		OpenAPI: nil,
		DialOpts: api.DialOpts{},
		AccountDetails: &jujuclient.AccountDetails{},
		ModelUUID: "",
	}

	conn, err := juju.NewAPIConnection(params)
	if err != nil {
		panic("Error getting Juju API")
	}

	return conn

}

func (m *Manager) init() error {
	var status []byte
	var hostname string
    // TODO: Application root fetch.
	root := m.getRoot()
	client := application.NewClient(root)
	rootClient := root.Client()




	if client != nil {
		panic("Created juju client")
	}

	for _, line := range strings.Split(string(status), "\n") {
		if strings.Contains(line, "kubernetes-worker/") {
			info := strings.Fields(line)
			unitName := strings.Replace(info[0],"*", "", -1)
			patterns := make([]string, 1)
			patterns[0] = "kubernetes-master"
			status, err := rootClient.Status(patterns)
			klog.Infof("Applications %s", status.Applications)
			nodeExec, _ := exec.Command("juju", "exec", "-u", unitName, "hostname").Output()
			hostname = strings.Fields(string(nodeExec))[0]
            // POINT 1..,

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
            // TODO: Test
            _, err = clientset.NodeV1().RuntimeClasses().Patch(context.TODO(), hostname, types.StrategicMergePatchType, []byte(`{"spec":{"providerID":"` + hostname + `"}}`), v1.PatchOptions{})

            //			exec.Command("kubectl", "patch", "node", hostname, "-p", `{"spec":{"providerID":"` + hostname + `"}}`).Output()
			m.units[unitName] = &Unit{
				state:      cloudprovider.InstanceRunning,
				jujuName:   unitName,
				kubeName:   hostname,
				registered: true,
			}
		}
	}

	return nil
}

func (m *Manager) addUnits(delta int) error {
	juju, _ := exec.LookPath("juju")

	prevStatus := m.getStatus()

	cmd := exec.Cmd{
		Path: juju,
		Args: []string {juju, "add-unit", "-n", strconv.Itoa(delta), "kubernetes-worker"},
		Stderr: os.Stdout,
	}
	cmd.Run()

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

func (m *Manager) removeUnit(name string) error {
	juju, _ := exec.LookPath("juju")
	unit := m.getUnit(name)
	unit.state = cloudprovider.InstanceDeleting

	cmd := exec.Cmd{
		Path: juju,
		Args: []string {juju, "run-action", unit.jujuName, "pause", "--wait"},
		Stderr: os.Stdout,
	}
	cmd.Run()

	cmd = exec.Cmd{
		Path: juju,
		Args: []string {juju, "remove-unit", unit.jujuName},
		Stderr: os.Stdout,
	}
	cmd.Run()

	return nil
}

func (m *Manager) refresh() error {
	for key, val := range m.getStatus() {
		if _, ok := m.units[key]; ok {
			m.units[key].agent = val[0]
			m.units[key].workload = val[1]
		}
	}

	for _, unit := range m.units {
		if unit.state == cloudprovider.InstanceCreating {
			if unit.kubeName == "" {
				nodeExec, _ := exec.Command("juju", "exec", "-u", unit.jujuName, "hostname").Output()
				if len(strings.Fields(string(nodeExec))) > 0 {
					unit.kubeName = strings.Fields(string(nodeExec))[0]
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

func (m *Manager) getStatus() map[string][]string {
	var status []byte
	units := make(map[string][]string)

	status, _ = exec.Command("juju", "status", "kubernetes-worker").Output()
	for _, line := range strings.Split(string(status), "\n") {
		if strings.Contains(line, "kubernetes-worker/") {
			info := strings.Fields(line)
			unitName := strings.Replace(info[0],"*", "", -1)
			if (info[1] == "terminated") {
				continue
			} else {
				units[unitName] = info[0:]
			}
		}
	}
	return units
}
