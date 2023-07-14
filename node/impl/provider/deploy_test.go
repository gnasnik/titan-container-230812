package provider

import (
	"context"
	"encoding/json"
	"testing"

	"github.com/gnasnik/titan-container/api/types"
	"github.com/gnasnik/titan-container/node/config"
	"github.com/gnasnik/titan-container/node/impl/provider/kube"
	"github.com/gnasnik/titan-container/node/impl/provider/kube/builder"
	"github.com/gnasnik/titan-container/node/impl/provider/kube/manifest"
	"github.com/stretchr/testify/require"
	"k8s.io/apimachinery/pkg/api/resource"
)

func TestDeploy(t *testing.T) {
	kubeconfig := "./test/config"
	client, err := kube.NewClient(kubeconfig)
	require.NoError(t, err)

	service := types.Service{Image: "redis:latest", Port: 6379, ComputeResources: types.ComputeResources{CPU: 0.1, Memory: 100, Storage: 100}}
	deploy := types.Deployment{
		ID:       types.DeploymentID("123"),
		Owner:    "test",
		Services: []*types.Service{&service},
	}

	k8sDeploy, err := ClusterDeploymentFromDeployment(&deploy)
	require.NoError(t, err)

	ctx := context.WithValue(context.Background(), builder.SettingsKey, builder.NewDefaultSettings())
	err = client.Deploy(ctx, k8sDeploy)
	require.NoError(t, err)
}

func TestDeleteDeploy(t *testing.T) {
	kubeconfig := "./test/config"
	client, err := kube.NewClient(kubeconfig)
	require.NoError(t, err)

	deploy := types.Deployment{
		ID:       types.DeploymentID("123"),
		Owner:    "test",
		Services: []*types.Service{},
	}

	ns := builder.DidNS(manifest.DeploymentID{ID: string(deploy.ID)})
	err = client.DeleteNS(context.Background(), ns)
	require.NoError(t, err)
}

func TestCPUUnixt(t *testing.T) {
	resources := manifest.NewResourceUnits(1000, 100, 100)
	t.Logf("cpu:%d", resources.CPU.Units.Val.Uint64())
}

func TestMemory(t *testing.T) {
	quantity := resource.NewQuantity(512000000, resource.DecimalSI)
	buf, _ := json.Marshal(*quantity)
	t.Logf("memory:%s", string(buf))
}

func TestResourcesStatistics(t *testing.T) {
	config := &config.ProviderCfg{KubeConfigPath: "./test/config", PublicIP: "192.168.0.132"}
	manager, err := NewManager(config)
	require.NoError(t, err)

	statistics, err := manager.GetStatistics(context.Background())
	require.NoError(t, err)

	t.Logf("nodeResources %#v", *statistics)

}

func TestGetDeployment(t *testing.T) {
	config := &config.ProviderCfg{KubeConfigPath: "./test/config", PublicIP: "192.168.0.132"}
	manager, err := NewManager(config)
	require.NoError(t, err)

	deployment, err := manager.GetDeployment(context.Background(), types.DeploymentID("123"))
	require.NoError(t, err)

	for _, service := range deployment.Services {
		t.Logf("deployment:%#v", *service)
	}

	t.Logf("deployment:%#v", *deployment)

}

func TestListDeployment(t *testing.T) {
	kubeconfig := "./test/config"
	client, err := kube.NewClient(kubeconfig)
	require.NoError(t, err)

	deploymentList, err := client.ListDeployments(context.Background(), "ld795irmdp488enid0r1rtb7bknfg1t8qtn0scua6org4")
	require.NoError(t, err)

	for _, deployment := range deploymentList.Items {
		buf, _ := json.Marshal(deployment.Status.Conditions)
		t.Logf("deployment:%s", string(buf))
	}
}
