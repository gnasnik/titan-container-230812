package db

import (
	"context"
	"fmt"
	"github.com/gnasnik/titan-container/api/types"
	"github.com/jmoiron/sqlx"
	"strconv"
	"strings"
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
	qry := `INSERT INTO deployments (id, name, owner, state, type, version, balance, cost, expiration, provider_id, env, created_at, updated_at) 
		        VALUES (:id, :name, :owner, :state, :type, :version, :balance, :cost, :expiration, :provider_id, :env, :created_at, :updated_at)
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
	qry := `SELECT d.*, s.image, s.cpu, s.memory,s.storage, s.port, p.host_uri AS provider_expose_ip FROM deployments d LEFT JOIN services s ON d.id = s.deployment_id LEFT JOIN providers p ON d.provider_id = p.id`

	var condition []string
	if option.DeploymentID != "" {
		condition = append(condition, fmt.Sprintf(`d.id = '%s'`, option.DeploymentID))
	}

	if option.Owner != "" {
		condition = append(condition, fmt.Sprintf(`d.owner = '%s'`, option.Owner))
	}

	if len(option.State) > 0 {
		var states []string
		for _, s := range option.State {
			states = append(states, strconv.Itoa(int(s)))
		}
		condition = append(condition, fmt.Sprintf(`d.state in (%s)`, strings.Join(states, ",")))
	}

	if len(condition) > 0 {
		qry += ` WHERE `
		qry += strings.Join(condition, ` AND `)
	}

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

func (m *ManagerDB) UpdateDeploymentState(ctx context.Context, id types.DeploymentID, state types.DeploymentState) error {
	qry := `Update deployments set state = ? where id = ?`
	_, err := m.db.ExecContext(ctx, qry, state, id)
	return err
}
