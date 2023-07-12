package provider

import (
	"context"

	"github.com/gnasnik/titan-container/api"
	"github.com/gnasnik/titan-container/api/types"
	"github.com/gnasnik/titan-container/node/common"
	"github.com/gnasnik/titan-container/provider"
	"github.com/google/uuid"
	"go.uber.org/fx"
)

var session = uuid.New()

// Provider represents a provider service in a cloud computing system.
type Provider struct {
	fx.In

	*common.CommonAPI
	ManagerAPI  api.Manager
	ProviderMgr provider.Manager
}

func (p *Provider) Session(ctx context.Context) (uuid.UUID, error) {
	return session, nil
}

func (p *Provider) Version(context.Context) (api.Version, error) {
	return api.ProviderAPIVersion0, nil
}

func (p *Provider) GetStatistics(ctx context.Context, id types.ProviderID) (*types.ResourcesStatistics, error) {
	return p.ProviderMgr.GetStatistics(ctx, id)
}

func (p *Provider) CreateDeployment(ctx context.Context, deployment *types.Deployment) error {
	return p.ProviderMgr.CreateDeployment(ctx, deployment)
}

func (p *Provider) UpdateDeployment(ctx context.Context, deployment *types.Deployment) error {
	return p.ProviderMgr.UpdateDeployment(ctx, deployment)
}

func (p *Provider) CloseDeployment(ctx context.Context, deployment *types.Deployment) error {
	return p.ProviderMgr.CloseDeployment(ctx, deployment)
}

var _ api.Provider = &Provider{}
