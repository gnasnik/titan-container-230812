package types

import "time"

type DeploymentID string

type DeploymentState int

const (
	DeploymentStateActive DeploymentState = iota + 1
	DeploymentStateInActive
	DeploymentStateClose
)

type DeploymentType int

const (
	DeploymentTypeWeb DeploymentType = iota + 1
)

type Deployment struct {
	ID         DeploymentID    `db:"id"`
	Name       string          `db:"name"`
	Owner      string          `db:"owner"`
	Image      string          `db:"image"`
	State      DeploymentState `db:"state"`
	Type       DeploymentType  `db:"type"`
	Version    []byte          `db:"version"`
	Balance    float64         `db:"balance"`
	Cost       float64         `db:"cost"`
	ProviderID ProviderID      `db:"provider_id"`
	Expiration time.Time       `db:"expiration"`
	CreatedAt  time.Time       `db:"created_at"`
	UpdatedAt  time.Time       `db:"updated_at"`

	Services []*Service
}

type Service struct {
	ID        int64     `db:"id"`
	Image     string    `db:"image"`
	Port      int       `db:"port"`
	CreatedAt time.Time `db:"created_at"`
	UpdatedAt time.Time `db:"updated_at"`
	ComputeResources
}

type GetDeploymentOption struct {
	Owner        string
	DeploymentID DeploymentID
}

type ComputeResources struct {
	CPU     float64 `db:"cpu"`
	Memory  int64   `db:"memory"`
	Storage int64   `db:"storage"`
}

type OrderID string

type Order struct {
	ID OrderID
}
