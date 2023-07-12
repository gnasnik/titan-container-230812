package db

import (
	"context"
	"github.com/gnasnik/titan-container/api/types"
	_ "github.com/go-sql-driver/mysql"
	"github.com/jmoiron/sqlx"
)

type ManagerDB struct {
	db *sqlx.DB
}

func NewManagerDB(db *sqlx.DB) *ManagerDB {
	return &ManagerDB{
		db: db,
	}
}

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

	return addNewServices(ctx, tx, deployment.Services)
}

func (m *ManagerDB) AddNewProvider(ctx context.Context, provider *types.Provider) error {
	qry := `INSERT INTO providers (id, owner, host_uri, ip, state, created_at, updated_at) 
		        VALUES (:id, :owner, :host_uri, :ip, :state, :created_at, :updated_at) ON DUPLICATE KEY UPDATE  owner=:owner, host_uri=:host_uri, 
		            ip=:ip, state=:state, updated_at=:updated_at`
	_, err := m.db.NamedExecContext(ctx, qry, provider)

	return err
}

func (m *ManagerDB) GetAllProviders(ctx context.Context) ([]*types.Provider, error) {
	var out []*types.Provider
	qry := `SELECT * from providers`
	err := m.db.SelectContext(ctx, &out, qry)
	if err != nil {
		return nil, err
	}
	return out, nil
}
