package types

import (
	"database/sql/driver"
	"encoding/json"
	"errors"
	"time"
)

type DeploymentID string

type DeploymentState int

const (
	DeploymentStateActive DeploymentState = iota + 1
	DeploymentStateInActive
	DeploymentStateClose
)

func DeploymentStateString(state DeploymentState) string {
	switch state {
	case DeploymentStateActive:
		return "Active"
	case DeploymentStateInActive:
		return "InActive"
	case DeploymentStateClose:
		return "Deleted"
	default:
		return "Unknown"
	}
}

var AllDeploymentStates = []DeploymentState{DeploymentStateActive, DeploymentStateInActive, DeploymentStateClose}

type DeploymentType int

const (
	DeploymentTypeWeb DeploymentType = iota + 1
)

type Deployment struct {
	ID               DeploymentID    `db:"id"`
	Name             string          `db:"name"`
	Owner            string          `db:"owner"`
	State            DeploymentState `db:"state"`
	Type             DeploymentType  `db:"type"`
	Version          []byte          `db:"version"`
	Balance          float64         `db:"balance"`
	Cost             float64         `db:"cost"`
	ProviderID       ProviderID      `db:"provider_id"`
	Expiration       time.Time       `db:"expiration"`
	CreatedAt        time.Time       `db:"created_at"`
	UpdatedAt        time.Time       `db:"updated_at"`
	ProviderExposeIP string          `db:"provider_expose_ip"`
	Services         []*Service
}

type Service struct {
	ID           int64        `db:"id"`
	Image        string       `db:"image"`
	Port         int          `db:"port"`
	ExposePort   int          `db:"expose_port"`
	DeploymentID DeploymentID `db:"deployment_id"`
	Env          Env          `db:"env"`
	Arguments    string       `db:"arguments"`
	CreatedAt    time.Time    `db:"created_at"`
	UpdatedAt    time.Time    `db:"updated_at"`
	ComputeResources
}

type Env map[string]string

func (e Env) Value() (driver.Value, error) {
	x := make(map[string]string)
	for k, v := range e {
		x[k] = v
	}
	return json.Marshal(x)
}

func (e Env) Scan(value interface{}) error {
	b, ok := value.([]byte)
	if !ok {
		return errors.New("type assertion to []byte failed")
	}
	if err := json.Unmarshal(b, &e); err != nil {
		return err
	}
	return nil
}

type GetDeploymentOption struct {
	Owner        string
	DeploymentID DeploymentID
	State        []DeploymentState
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
