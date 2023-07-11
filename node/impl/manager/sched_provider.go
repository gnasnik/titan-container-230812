package manager

import (
	"github.com/gnasnik/titan-container/api"
	"github.com/gnasnik/titan-container/api/types"
	"sync"
)

type ProviderScheduler struct {
	lk        sync.RWMutex
	providers map[types.ProviderID]api.Provider
}

func NewProviderScheduler() *ProviderScheduler {
	return &ProviderScheduler{
		providers: make(map[types.ProviderID]api.Provider),
	}
}

func (p *ProviderScheduler) AddProvider(id types.ProviderID, provider api.Provider) error {
	p.lk.Lock()
	p.lk.Unlock()

	_, exist := p.providers[id]
	if exist {
		return nil
	}

	p.providers[id] = provider
	return nil
}
