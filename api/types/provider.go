package types

type ProviderID string

type Provider struct {
	ID      ProviderID
	Owner   string
	HostURI string
}

type ResourcesStatistics struct {
	Memory
	CPUCores
	Storage
}

type Memory struct {
	MaxMemory uint64
	Available uint64
	Active    uint64
	Pending   uint64
}

type CPUCores struct {
	MaxCPUCores uint64
	Available   uint64
	Active      uint64
	Pending     uint64
}

type Storage struct {
	MaxStorage uint64
	Available  uint64
	Active     uint64
	Pending    uint64
}
