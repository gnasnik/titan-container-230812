package types

import "time"

type ProviderID string

type Provider struct {
	ID        ProviderID `db:"id"`
	Owner     string     `db:"owner"`
	HostURI   string     `db:"host_uri"`
	IP        string     `db:"ip"`
	CreatedAt time.Time  `db:"created_at"`
	UpdatedAt time.Time  `db:"updated_at"`
}

type ResourcesStatistics struct {
	Memory   Memory
	CPUCores CPUCores
	Storage  Storage
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
