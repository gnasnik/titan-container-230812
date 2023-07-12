package provider

import (
	"context"
	"encoding/json"
	"testing"

	"github.com/gnasnik/titan-container/api/types"
	"github.com/gnasnik/titan-container/provider/kube"
	"github.com/gnasnik/titan-container/provider/kube/builder"
	"github.com/gnasnik/titan-container/provider/kube/manifest"
	"github.com/stretchr/testify/require"
	"k8s.io/apimachinery/pkg/api/resource"
)

func TestDeploy(t *testing.T) {
	kubeconfig := "./test/config"
	client, err := kube.NewClient(kubeconfig)
	require.NoError(t, err)

	service := types.Service{Image: "nginx:1.14.2", Port: 80, ComputeResources: types.ComputeResources{CPU: 1, Memory: 100}}
	deploy := types.Deployment{
		ID:       types.DeploymentID("123"),
		Owner:    "test",
		Services: []types.Service{service},
	}

	k8sDeploy := ClusterDeploymentFromDeployment(&deploy)

	ctx := context.WithValue(context.Background(), builder.SettingsKey, builder.NewDefaultSettings())
	err = client.Deploy(ctx, k8sDeploy)
	require.NoError(t, err)
}

func TestDeleteDeploy(t *testing.T) {
	kubeconfig := "./test/config"
	client, err := kube.NewClient(kubeconfig)
	require.NoError(t, err)

	service := types.Service{Image: "nginx:1.14.2", Port: 80, ComputeResources: types.ComputeResources{CPU: 1, Memory: 100}}
	deploy := types.Deployment{
		ID:       types.DeploymentID("123"),
		Owner:    "test",
		Services: []types.Service{service},
	}

	k8sDeploy := ClusterDeploymentFromDeployment(&deploy)
	ns := builder.DidNS(k8sDeploy.DeploymentID())

	err = client.DeleteNS(context.Background(), ns)
	require.NoError(t, err)
}

func TestCPUUnixt(t *testing.T) {
	resources := manifest.NewResourceUnits(1000, 512)
	t.Logf("cpu:%d", resources.CPU.Units.Val.Uint64())
}

func TestMemory(t *testing.T) {
	quantity := resource.NewQuantity(512000000, resource.DecimalSI)
	buf, _ := json.Marshal(*quantity)
	t.Logf("memory:%s", string(buf))
}

func TestListNode(t *testing.T) {
	kubeconfig := "./test/config"
	client, err := kube.NewClient(kubeconfig)
	require.NoError(t, err)

	nodeList, err := client.ListNodes(context.Background())
	require.NoError(t, err)

	statistics := &types.ResourcesStatistics{}
	for _, node := range nodeList.Items {
		cpu, memory, storage := getResources(node.Status.Capacity)
		statistics.CPUCores.MaxCPUCores += cpu
		statistics.Memory.MaxMemory += memory
		statistics.Storage.MaxStorage += storage
		t.Logf("max cpu %d, memory %d storage %d", cpu, memory, storage)
		cpu, memory, storage = getResources(node.Status.Allocatable)
		statistics.CPUCores.Available += cpu
		statistics.Memory.Available += memory
		statistics.Storage.Available += storage
		t.Logf("Available cpu %d, memory %d storage %d", cpu, memory, storage)

	}

	t.Logf("statistics %#v", *statistics)

}
