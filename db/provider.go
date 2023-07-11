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

func (m *ManagerDB) AddNewProvider(ctx context.Context, provider *types.Provider) error {
	qry := `INSERT INTO providers (id, owner, host_uri, ip, created_at, updated_at) 
		        VALUES (:id, :owner, :host_uri, :ip, :created_at, :updated_at) ON DUPLICATE KEY UPDATE  owner=:owner, host_uri=:host_uri, 
		            ip=:ip, updated_at=:updated_at`
	_, err := m.db.NamedExecContext(ctx, qry, provider)

	return err
}
