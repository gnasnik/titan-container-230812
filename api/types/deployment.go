package types

import (
	"database/sql/driver"
	"encoding/json"
	"errors"
	"strings"
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
	ID       DeploymentID    `db:"id"`
	Name     string          `db:"name"`
	Owner    string          `db:"owner"`
	State    DeploymentState `db:"state"`
	Version  []byte          `db:"version"`
	Services []*Service

	// Internal
	Type             DeploymentType `db:"type"`
	Balance          float64        `db:"balance"`
	Cost             float64        `db:"cost"`
	ProviderID       ProviderID     `db:"provider_id"`
	Expiration       time.Time      `db:"expiration"`
	CreatedAt        time.Time      `db:"created_at"`
	UpdatedAt        time.Time      `db:"updated_at"`
	ProviderExposeIP string         `db:"provider_expose_ip"`
}

type ReplicasStatus struct {
	TotalReplicas     int
	ReadyReplicas     int
	AvailableReplicas int
}

type Service struct {
	Image        string         `db:"image"`
	Name         string         `db:"name"`
	Port         int            `db:"port"`
	ExposePort   int            `db:"expose_port"`
	Env          Env            `db:"env"`
	Status       ReplicasStatus `db:"status"`
	ErrorMessage string         `db:"error_message"`
	Arguments    Arguments      `db:"arguments"`
	ComputeResources

	// Internal
	ID           int64        `db:"id"`
	DeploymentID DeploymentID `db:"deployment_id"`
	CreatedAt    time.Time    `db:"created_at"`
	UpdatedAt    time.Time    `db:"updated_at"`
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

type Arguments []string

func (a Arguments) Value() (driver.Value, error) {
	return strings.Join(a, ","), nil
}

func (a Arguments) Scan(value interface{}) error {
	b, ok := value.([]byte)
	if !ok {
		return errors.New("type assertion to []byte failed")
	}
	copy(a, strings.Split(string(b), ","))
	return nil
}

func (s *Service) Apply(in *Service) {
	if in.Name != "" {
		s.Name = in.Name
	}
	if in.Port > 0 {
		s.Port = in.Port
	}
	if in.ExposePort > 0 {
		s.ExposePort = in.ExposePort
	}
	if in.Status != (ReplicasStatus{}) {
		s.Status = in.Status
	}
	if in.ComputeResources != (ComputeResources{}) {
		s.ComputeResources = in.ComputeResources
	}
}

type GetDeploymentOption struct {
	Owner        string
	DeploymentID DeploymentID
	State        []DeploymentState
	Page         int
	Size         int
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
