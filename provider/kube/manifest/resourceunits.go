package manifest

type ResourceUnits struct {
	CPU       *CPU
	Memory    *Memory
	Storage   []*Storage
	GPU       *GPU
	Endpoints []*Endpoint
}

func NewResourceUnits(cpu, memory uint64) *ResourceUnits {
	return &ResourceUnits{CPU: NewCPU(cpu), Memory: NewMemory(memory)}
}
