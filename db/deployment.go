package db

import (
	"context"

	"github.com/gnasnik/titan-container/api/types"
	"github.com/jmoiron/sqlx"
)

func addNewDeployment(ctx context.Context, tx *sqlx.Tx, deployment *types.Deployment) error {
	qry := `INSERT INTO deployments (id, name, owner, state, type, version, balance, cost, expiration, provider_id, created_at, updated_at) 
		        VALUES (:id, :name, :owner, :state, :type, :version, :balance, :cost, :expiration, :provider_id, :created_at, :updated_at)
		         ON DUPLICATE KEY UPDATE  state=:state, version=:version, balance=:balance, cost=:cost, expiration=:expiration, updated_at=:updated_at`
	_, err := tx.NamedExecContext(ctx, qry, deployment)

	return err
}

func addNewServices(ctx context.Context, tx *sqlx.Tx, services []*types.Service) error {
	qry := `INSERT INTO deployments (id, image, port, cpu, memory, storage, created_at, updated_at) 
		        VALUES (:id, :image, :port, :memory, :storage, :created_at, :updated_at)`
	_, err := tx.NamedExecContext(ctx, qry, services)

	return err
}

func (m *ManagerDB) GetDeployments(ctx context.Context, option *types.GetDeploymentOption) ([]*types.Deployment, error) {
	var out []*types.Deployment
	qry := `SELECT * from deployments`
	err := m.db.SelectContext(ctx, &out, qry)
	if err != nil {
		return nil, err
	}
	return out, nil
}
