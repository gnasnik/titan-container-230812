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
	kc          kube.Client
	providerCfg *config.ProviderCfg
}

var _ Manager = (*manager)(nil)

func NewManager(config *config.ProviderCfg) (Manager, error) {
	client, err := kube.NewClient(config.KubeConfigPath)
	if err != nil {
		return nil, err
	}
	return &manager{kc: client, providerCfg: config}, nil
}

func (m *manager) GetStatistics(ctx context.Context) (*types.ResourcesStatistics, error) {
	nodeResources, err := m.kc.FetchNodeResources(ctx)
	if err != nil {
		return nil, err
	}

	if nodeResources == nil {
		return nil, fmt.Errorf("nodes resources do not exist")
	}

	statistics := &types.ResourcesStatistics{}
	for _, node := range nodeResources {
		statistics.CPUCores.MaxCPUCores += node.CPU.Capacity.AsApproximateFloat64()
		statistics.CPUCores.Available += node.CPU.Allocatable.AsApproximateFloat64()
		statistics.CPUCores.Active += node.CPU.Allocated.AsApproximateFloat64()

		statistics.Memory.MaxMemory += uint64(node.Memory.Capacity.AsApproximateFloat64())
		statistics.Memory.Available += uint64(node.Memory.Allocatable.AsApproximateFloat64())
		statistics.Memory.Active += uint64(node.Memory.Allocated.AsApproximateFloat64())

		statistics.Storage.MaxStorage += uint64(node.EphemeralStorage.Capacity.AsApproximateFloat64())
		statistics.Storage.Available += uint64(node.EphemeralStorage.Allocatable.AsApproximateFloat64())
		statistics.Storage.Active += uint64(node.EphemeralStorage.Allocated.AsApproximateFloat64())
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

	// did := k8sDeployment.DeploymentID()
	// ns := builder.DidNS(did)

	// _, err := m.kc.GetNS(ctx, ns)
	// if err != nil {
	// 	return fmt.Errorf("deployment %s do not exist or %s", deployment.ID, err.Error())
	// }

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
		name := services[i].Name
		if port, ok := portMap[name]; ok {
			services[i].ExposePort = port
		}
	}

	return &types.Deployment{ID: id, Services: services, ProviderExposeIP: m.providerCfg.PublicIP}, nil
}
