package db

import (
	"context"
	"github.com/gnasnik/titan-container/api/types"
	"github.com/jmoiron/sqlx"
)

func (m *ManagerDB) CreateDeployment(ctx context.Context, deployment *types.Deployment) error {
	tx, err := m.db.Beginx()
	if err != nil {
		return err
	}
	defer tx.Rollback()

	err = addNewDeployment(ctx, tx, deployment)
	if err != nil {
		return err
	}

	err = addNewServices(ctx, tx, deployment.Services)
	if err != nil {
		return err
	}

	return tx.Commit()
}

func addNewDeployment(ctx context.Context, tx *sqlx.Tx, deployment *types.Deployment) error {
	qry := `INSERT INTO deployments (id, name, owner, image, state, type, version, balance, cost, expiration, provider_id, env, created_at, updated_at) 
		        VALUES (:id, :name, :owner, :image, :state, :type, :version, :balance, :cost, :expiration, :provider_id, :env, :created_at, :updated_at)
		         ON DUPLICATE KEY UPDATE  state=:state, version=:version, balance=:balance, cost=:cost, expiration=:expiration, env=:env, updated_at=:updated_at`
	_, err := tx.NamedExecContext(ctx, qry, deployment)

	return err
}

func addNewServices(ctx context.Context, tx *sqlx.Tx, services []*types.Service) error {
	qry := `INSERT INTO services (id, image, port, cpu, memory, storage, deployment_id, created_at, updated_at) 
		        VALUES (:id, :image, :port, :cpu, :memory, :storage, :deployment_id, :created_at, :updated_at)`
	_, err := tx.NamedExecContext(ctx, qry, services)

	return err
}

type DeploymentService struct {
	types.Deployment
	types.Service
}

func (m *ManagerDB) GetDeployments(ctx context.Context, option *types.GetDeploymentOption) ([]*types.Deployment, error) {
	var ds []*DeploymentService
	qry := `SELECT d.*, s.cpu, s.memory,s.storage, s.port, p.host_uri AS provider_expose_ip FROM deployments d LEFT JOIN services s ON d.id = s.deployment_id LEFT JOIN providers p ON d.provider_id = p.id`
	err := m.db.SelectContext(ctx, &ds, qry)
	if err != nil {
		return nil, err
	}

	var out []*types.Deployment
	deploymentToServices := make(map[types.DeploymentID]*types.Deployment)
	for _, d := range ds {
		_, ok := deploymentToServices[d.Deployment.ID]
		if !ok {
			deploymentToServices[d.Deployment.ID] = &d.Deployment
			deploymentToServices[d.Deployment.ID].Services = make([]*types.Service, 0)
			out = append(out, deploymentToServices[d.Deployment.ID])
		}
		deploymentToServices[d.Deployment.ID].Services = append(deploymentToServices[d.Deployment.ID].Services, &d.Service)
	}

	return out, nil
}
