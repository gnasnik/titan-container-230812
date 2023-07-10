package types

type DeploymentID string

type DeploymentState int

const (
	DeploymentStateActive DeploymentState = iota + 1
	DeploymentStateInActive
	DeploymentStateClose
)

type ServiceType int

const (
	ServiceTypeWeb ServiceType = iota + 1
)

type Deployment struct {
	ID        DeploymentID
	Owner     string
	State     DeploymentState
	Services  []Service
	Version   []byte
	CreatedAt int64
}

type Service struct {
	Image            string
	Port             int
	Type             ServiceType
	ComputeResources ComputeResources
}

type ComputeResources struct {
	CPU     float64
	Memory  int64
	Storage int64
}

type OrderID string

type Order struct {
	ID OrderID
}
