package api

import (
	"context"

	"github.com/gnasnik/titan-container/api/types"
)

// Manager is an interface for manager
type Manager interface {
	Common

	// ProviderConnect provider registration
	ProviderConnect(ctx context.Context, url string, provider *types.Provider) error //perm:admin
	UpdateProvider(ctx context.Context, provider *types.Provider) error              //perm:admin

	GetProviderList(ctx context.Context) ([]*types.Provider, error) //perm:read
	GetTemplateList(ctx context.Context) ([]*types.Template, error) //perm:read

	GetDeploymentList(ctx context.Context, opt *types.GetDeploymentOption) ([]*types.Deployment, error) //perm:read
	CreateDeployment(ctx context.Context, id types.ProviderID, deployment *types.Deployment) error      //perm:admin
	UpdateDeployment(ctx context.Context, id types.ProviderID, deployment *types.Deployment) error      //perm:admin
	CloseDeployment(ctx context.Context, id types.ProviderID, deployment *types.Deployment) error       //perm:admin
}
