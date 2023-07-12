package api

import (
	"context"

	"github.com/gnasnik/titan-container/api/types"
	"github.com/google/uuid"
)

type Provider interface {
	GetStatistics(ctx context.Context) (*types.ResourcesStatistics, error)    //perm:read
	CreateDeployment(ctx context.Context, deployment *types.Deployment) error //perm:admin
	UpdateDeployment(ctx context.Context, deployment *types.Deployment) error //perm:admin
	CloseDeployment(ctx context.Context, deployment *types.Deployment) error  //perm:admin

	Version(context.Context) (Version, error)   //perm:admin
	Session(context.Context) (uuid.UUID, error) //perm:admin
}
