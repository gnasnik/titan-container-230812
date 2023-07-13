package manager

import (
	"context"
	"strings"
	"time"

	"github.com/gnasnik/titan-container/api"
	"github.com/gnasnik/titan-container/api/types"
	"github.com/gnasnik/titan-container/db"
	"github.com/gnasnik/titan-container/node/handler"
	"github.com/gnasnik/titan-container/node/modules/dtypes"
	"github.com/google/uuid"
	logging "github.com/ipfs/go-log/v2"
	"github.com/pkg/errors"
	"go.uber.org/fx"
)

var log = logging.Logger("manager")

// Manager represents a manager service in a cloud computing system.
type Manager struct {
	fx.In

	api.Common
	DB *db.ManagerDB

	ProviderScheduler *ProviderScheduler

	SetManagerConfigFunc dtypes.SetManagerConfigFunc
	GetManagerConfigFunc dtypes.GetManagerConfigFunc
}

func (m *Manager) ProviderConnect(ctx context.Context, url string, provider *types.Provider) error {
	remoteAddr := handler.GetRemoteAddr(ctx)

	p, err := connectRemoteProvider(ctx, m, url)
	if err != nil {
		return errors.Errorf("connecting remote provider failed: %v", err)
	}

	log.Infof("Connected to a remote provider at %s, provider id %s", remoteAddr, provider.ID)

	err = m.ProviderScheduler.AddProvider(provider.ID, p)
	if err != nil {
		return err
	}

	if provider.IP == "" {
		provider.IP = strings.Split(remoteAddr, ":")[0]
	}

	provider.State = types.ProviderStateOnline
	provider.CreatedAt = time.Now()
	provider.UpdatedAt = time.Now()
	return m.DB.AddNewProvider(ctx, provider)
}

func (m *Manager) UpdateProvider(ctx context.Context, provider *types.Provider) error {
	//TODO implement me
	panic("implement me")
}

func (m *Manager) GetProviderList(ctx context.Context) ([]*types.Provider, error) {
	return m.DB.GetAllProviders(ctx)
}

func (m *Manager) GetTemplateList(ctx context.Context) ([]*types.Template, error) {
	//TODO implement me
	panic("implement me")
}

func (m *Manager) CreateOrder(ctx context.Context) error {
	//TODO implement me
	panic("implement me")
}

func (m *Manager) GetDeploymentList(ctx context.Context, opt *types.GetDeploymentOption) ([]*types.Deployment, error) {
	return m.DB.GetDeployments(ctx, opt)
}

func (m *Manager) CreateDeployment(ctx context.Context, deployment *types.Deployment) error {
	providerApi, err := m.ProviderScheduler.Get(deployment.ProviderID)
	if err != nil {
		return err
	}

	deployment.ID = types.DeploymentID(uuid.New().String())
	for i := 0; i < len(deployment.Services); i++ {
		deployment.Services[i].DeploymentID = deployment.ID
	}
	deployment.State = types.DeploymentStateActive
	deployment.CreatedAt = time.Now()
	deployment.UpdatedAt = time.Now()

	err = providerApi.CreateDeployment(ctx, deployment)
	if err != nil {
		return err
	}

	//successDeployment, err := providerApi.GetDeployment(ctx, deployment.ID)
	//if err != nil {
	//	return err
	//}

	err = m.DB.CreateDeployment(ctx, deployment)
	if err != nil {
		return err
	}

	return nil
}

func (m *Manager) UpdateDeployment(ctx context.Context, deployment *types.Deployment) error {
	providerApi, err := m.ProviderScheduler.Get(deployment.ProviderID)
	if err != nil {
		return err
	}

	err = providerApi.CreateDeployment(ctx, deployment)
	if err != nil {
		return err
	}

	err = m.DB.CreateDeployment(ctx, deployment)
	if err != nil {
		return err
	}

	return nil
}

func (m *Manager) CloseDeployment(ctx context.Context, deployment *types.Deployment) error {
	providerApi, err := m.ProviderScheduler.Get(deployment.ProviderID)
	if err != nil {
		return err
	}

	err = providerApi.CloseDeployment(ctx, deployment)
	if err != nil {
		return err
	}

	return m.DB.UpdateDeploymentState(ctx, deployment.ID, types.DeploymentStateClose)
}

var _ api.Manager = &Manager{}
