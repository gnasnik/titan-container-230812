package db

import (
	"context"
	_ "embed"
	logging "github.com/ipfs/go-log/v2"
	"github.com/jmoiron/sqlx"
	"github.com/pkg/errors"
)

var log = logging.Logger("db")

func SqlDB(dsn string) (*sqlx.DB, error) {
	client, err := sqlx.Open("mysql", dsn)
	if err != nil {
		return nil, err
	}

	if err = client.Ping(); err != nil {
		return nil, err
	}

	// initialize the database
	log.Info("db: creating tables")
	err = createAllTables(context.Background(), client)
	if err != nil {
		return nil, errors.Errorf("failed to init db: %w", err)
	}

	return client, nil
}

//go:embed create_main_db.sql
var createMainDBSQL string

func createAllTables(ctx context.Context, mainDB *sqlx.DB) error {
	if _, err := mainDB.ExecContext(ctx, createMainDBSQL); err != nil {
		return errors.Errorf("failed to create tables in main DB: %w", err)
	}
	return nil
}
