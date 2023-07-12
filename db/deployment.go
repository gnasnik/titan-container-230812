package db

import (
	"context"
	"github.com/gnasnik/titan-container/api/types"
	"github.com/jmoiron/sqlx"
)

func addNewDeployment(ctx context.Context, tx *sqlx.Tx, deployment *types.Deployment) error {
	qry := `INSERT INTO deployments.sql (id, name, owner, state, type, version, balance, cost, expiration, provider_id, created_at, updated_at) 
		        VALUES (:id, :name, :owner, :state, :type, :version, :balance, :cost, :expiration, :provider_id, :created_at, :updated_at)
		         ON DUPLICATE KEY UPDATE  state=:state, version=:version, balance=:balance, cost=:cost, expiration=:expiration, updated_at=:updated_at`
	_, err := tx.NamedExecContext(ctx, qry, deployment)

	return err
}

func addNewServices(ctx context.Context, tx *sqlx.Tx, services []*types.Service) error {
	qry := `INSERT INTO deployments.sql (id, image, port, cpu, memory, storage, created_at, updated_at) 
		        VALUES (:id, :image, :port, :memory, :storage, :created_at, :updated_at)`
	_, err := tx.NamedExecContext(ctx, qry, services)

	return err
}
