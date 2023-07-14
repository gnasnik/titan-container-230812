package provider

import (
	"context"
	"fmt"

	"github.com/gnasnik/titan-container/api/types"
	"github.com/gnasnik/titan-container/node/config"
	"github.com/gnasnik/titan-container/node/impl/provider/kube"
	"github.com/gnasnik/titan-container/node/impl/provider/kube/builder"
	"github.com/gnasnik/titan-container/node/impl/provider/kube/manifest"
	logging "github.com/ipfs/go-log/v2"
	corev1 "k8s.io/api/core/v1"
)

var log = logging.Logger("provider")

type Manager interface {
	GetStatistics(ctx context.Context) (*types.ResourcesStatistics, error)
	CreateDeployment(ctx context.Context, deployment *types.Deployment) error
	UpdateDeployment(ctx context.Context, deployment *types.Deployment) error
	CloseDeployment(ctx context.Context, deployment *types.Deployment) error
	GetDeployment(ctx context.Context, id types.DeploymentID) (*types.Deployment, error)
}

type manager struct {
	kc kube.Client
}

var _ Manager = (*manager)(nil)

func NewManager(config *config.ProviderCfg) (Manager, error) {
	client, err := kube.NewClient(config.KubeConfigPath)
	if err != nil {
		return nil, err
	}
	return &manager{kc: client}, nil
}

func getResources(resources corev1.ResourceList) (uint64, uint64, uint64) {
	cpu := resources[corev1.ResourceCPU]
	memory := resources[corev1.ResourceMemory]
	storage := resources[corev1.ResourceEphemeralStorage]
	return uint64(cpu.Value()), uint64(memory.Value()), uint64(storage.Value())
}

func (m *manager) GetStatistics(ctx context.Context) (*types.ResourcesStatistics, error) {
	nodeList, err := m.kc.ListNodes(ctx)
	if err != nil {
		return nil, err
	}

	statistics := &types.ResourcesStatistics{}
	for _, node := range nodeList.Items {
		cpu, memory, storage := getResources(node.Status.Capacity)
		statistics.CPUCores.MaxCPUCores += cpu
		statistics.Memory.MaxMemory += memory
		statistics.Storage.MaxStorage += storage

		cpu, memory, storage = getResources(node.Status.Allocatable)
		statistics.CPUCores.Available += cpu
		statistics.Memory.Available += memory
		statistics.Storage.Available += storage
	}
	return statistics, nil
}

func (m *manager) CreateDeployment(ctx context.Context, deployment *types.Deployment) error {
	k8sDeployment, err := ClusterDeploymentFromDeployment(deployment)
	if err != nil {
		log.Errorf("CreateDeployment %s", err.Error())
		return err
	}

	ctx = context.WithValue(ctx, builder.SettingsKey, builder.NewDefaultSettings())
	return m.kc.Deploy(ctx, k8sDeployment)
}

func (m *manager) UpdateDeployment(ctx context.Context, deployment *types.Deployment) error {
	k8sDeployment, err := ClusterDeploymentFromDeployment(deployment)
	if err != nil {
		log.Errorf("UpdateDeployment %s", err.Error())
		return err
	}

	ctx = context.WithValue(ctx, builder.SettingsKey, builder.NewDefaultSettings())
	return m.kc.Deploy(ctx, k8sDeployment)
}

func (m *manager) CloseDeployment(ctx context.Context, deployment *types.Deployment) error {
	k8sDeployment, err := ClusterDeploymentFromDeployment(deployment)
	if err != nil {
		log.Errorf("CloseDeployment %s", err.Error())
		return err
	}

	did := k8sDeployment.DeploymentID()
	ns := builder.DidNS(did)
	if len(ns) == 0 {
		return fmt.Errorf("can not get ns from deployment id %s and owner %s", deployment.ID, deployment.Owner)
	}
	return m.kc.DeleteNS(ctx, ns)
}

func (m *manager) GetDeployment(ctx context.Context, id types.DeploymentID) (*types.Deployment, error) {
	deploymentID := manifest.DeploymentID{ID: string(id)}
	ns := builder.DidNS(deploymentID)

	deploymentList, err := m.kc.ListDeployments(ctx, ns)
	if err != nil {
		return nil, err
	}

	services, err := k8sDeploymentsToServices(deploymentList)
	if err != nil {
		return nil, err
	}

	serviceList, err := m.kc.ListServices(ctx, ns)
	if err != nil {
		return nil, err
	}

	portMap, err := k8sServiceToPortMap(serviceList)
	if err != nil {
		return nil, err
	}

	for i := range services {
		image := services[i].Image
		key := imageToServiceName(image)
		if port, ok := portMap[key]; ok {
			services[i].ExposePort = port
		}
	}

	return &types.Deployment{ID: id, Services: services}, nil
}
