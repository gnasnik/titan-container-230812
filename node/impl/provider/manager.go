package provider

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/gnasnik/titan-container/api/types"
	"github.com/gnasnik/titan-container/node/config"
	"github.com/gnasnik/titan-container/node/impl/provider/kube"
	"github.com/gnasnik/titan-container/node/impl/provider/kube/builder"
	logging "github.com/ipfs/go-log/v2"
	corev1 "k8s.io/api/core/v1"
)

var log = logging.Logger("provider-manager")

type Manager interface {
	GetStatistics(ctx context.Context) (*types.ResourcesStatistics, error)
	CreateDeployment(ctx context.Context, deployment *types.Deployment) error
	UpdateDeployment(ctx context.Context, deployment *types.Deployment) error
	CloseDeployment(ctx context.Context, deployment *types.Deployment) error
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
	buf, _ := json.Marshal(statistics)
	log.Debugf("statistics:%s", string(buf))
	return statistics, nil
}

func (m *manager) CreateDeployment(ctx context.Context, deployment *types.Deployment) error {
	buf, _ := json.Marshal(deployment)
	fmt.Printf("deployment:%#v\n", string(buf))
	// for _, service := range deployment.Services {
	// 	fmt.Printf("service:%#v\n", *service)
	// }

	k8sDeployment := ClusterDeploymentFromDeployment(deployment)
	if k8sDeployment.ManifestGroup() == nil || len(k8sDeployment.ManifestGroup().Services) == 0 {
		return fmt.Errorf("deployment service can not empty")
	}

	ctx = context.WithValue(ctx, builder.SettingsKey, builder.NewDefaultSettings())
	return m.kc.Deploy(ctx, k8sDeployment)
}

func (m *manager) UpdateDeployment(ctx context.Context, deployment *types.Deployment) error {
	k8sDeployment := ClusterDeploymentFromDeployment(deployment)
	if k8sDeployment.ManifestGroup() == nil || len(k8sDeployment.ManifestGroup().Services) == 0 {
		return fmt.Errorf("deployment service can not empty")
	}

	ctx = context.WithValue(ctx, builder.SettingsKey, builder.NewDefaultSettings())
	return m.kc.Deploy(ctx, k8sDeployment)
}

func (m *manager) CloseDeployment(ctx context.Context, deployment *types.Deployment) error {
	k8sDeployment := ClusterDeploymentFromDeployment(deployment)
	did := k8sDeployment.DeploymentID()
	ns := builder.DidNS(did)
	if len(ns) == 0 {
		return fmt.Errorf("can not get ns from deployment id %s and owner %s", deployment.ID, deployment.Owner)
	}
	return m.kc.DeleteNS(ctx, ns)
}
