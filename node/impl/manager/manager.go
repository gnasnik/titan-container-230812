package manager

import (
	"context"
	"github.com/gnasnik/titan-container/api"
	"github.com/gnasnik/titan-container/api/types"
	"github.com/gnasnik/titan-container/node/common"
	"github.com/gnasnik/titan-container/node/handler"
	logging "github.com/ipfs/go-log/v2"
	"go.uber.org/fx"
)

var log = logging.Logger("manager")

// Manager represents a manager service in a cloud computing system.
type Manager struct {
	fx.In

	*common.CommonAPI
}

func (m *Manager) ProviderConnect(ctx context.Context) error {
	remoteAddr := handler.GetRemoteAddr(ctx)
	log.Infof("provider connected address: %s", remoteAddr)
	return nil
}

func (m *Manager) CreateProvider(ctx context.Context, provider *types.Provider) error {
	//TODO implement me
	panic("implement me")
}

func (m *Manager) UpdateProvider(ctx context.Context, provider *types.Provider) error {
	//TODO implement me
	panic("implement me")
}

func (m *Manager) DeleteProvider(ctx context.Context, provider *types.Provider) error {
	//TODO implement me
	panic("implement me")
}

func (m *Manager) GetProviderList(ctx context.Context) ([]*types.Provider, error) {
	//TODO implement me
	panic("implement me")
}

func (m *Manager) GetTemplateList(ctx context.Context) ([]*types.Template, error) {
	//TODO implement me
	panic("implement me")
}

func (m *Manager) CreateOrder(ctx context.Context) error {
	//TODO implement me
	panic("implement me")
}

func (m *Manager) CreateDeployment(ctx context.Context, id types.ProviderID, deployment *types.Deployment) error {
	//TODO implement me
	panic("implement me")
}

func (m *Manager) UpdateDeployment(ctx context.Context, id types.ProviderID, deployment *types.Deployment) error {
	//TODO implement me
	panic("implement me")
}

func (m *Manager) CloseDeployment(ctx context.Context, id types.ProviderID, deployment *types.Deployment) error {
	//TODO implement me
	panic("implement me")
}

var _ api.Manager = &Manager{}
