package provider

import (
	"context"
	"github.com/gnasnik/titan-container/api"
	"github.com/gnasnik/titan-container/api/types"
	"github.com/gnasnik/titan-container/node/common"
	"go.uber.org/fx"
)

// Provider represents a provider service in a cloud computing system.
type Provider struct {
	fx.In

	*common.CommonAPI
	ManagerAPI api.Manager
}

func (p *Provider) GetStatistics(ctx context.Context, id types.ProviderID) (*types.ResourcesStatistics, error) {
	//TODO implement me
	panic("implement me")
}

func (p *Provider) CreateDeployment(ctx context.Context, deployment *types.Deployment) error {
	//TODO implement me
	panic("implement me")
}

func (p *Provider) UpdateDeployment(ctx context.Context, deployment *types.Deployment) error {
	//TODO implement me
	panic("implement me")
}

func (p *Provider) CloseDeployment(ctx context.Context, deployment *types.Deployment) error {
	//TODO implement me
	panic("implement me")
}

var _ api.Provider = &Provider{}
