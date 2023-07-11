package api

import (
	"context"

	"github.com/gnasnik/titan-container/api/types"
)

type Provider interface {
	Version(context.Context) (Version, error)                                                   //perm:admin
	GetStatistics(ctx context.Context, id types.ProviderID) (*types.ResourcesStatistics, error) //perm:read
	CreateDeployment(ctx context.Context, deployment *types.Deployment) error                   //perm:admin
	UpdateDeployment(ctx context.Context, deployment *types.Deployment) error                   //perm:admin
	CloseDeployment(ctx context.Context, deployment *types.Deployment) error                    //perm:admin
}
