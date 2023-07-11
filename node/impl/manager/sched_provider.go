package manager

import (
	"context"
	"github.com/gnasnik/titan-container/api"
	"github.com/gnasnik/titan-container/api/types"
	"sync"
	"time"
)

var HeartbeatInterval = 10 * time.Second

var ProviderTTL = 30 * time.Second

type ProviderScheduler struct {
	lk        sync.RWMutex
	providers map[types.ProviderID]*providerLife
}

type providerLife struct {
	api.Provider
	LastSeen time.Time
}

func (p *providerLife) Update() {
	p.LastSeen = time.Now()
}

func (p *providerLife) Expired() bool {
	if p.LastSeen.Add(ProviderTTL).Before(time.Now()) {
		return true
	}
	return false
}

func NewProviderScheduler() *ProviderScheduler {
	s := &ProviderScheduler{
		providers: make(map[types.ProviderID]*providerLife),
	}

	go s.watch()
	return s
}

func (p *ProviderScheduler) AddProvider(id types.ProviderID, providerApi api.Provider) error {
	p.lk.Lock()
	p.lk.Unlock()

	_, exist := p.providers[id]
	if exist {
		return nil
	}

	p.providers[id] = &providerLife{
		Provider: providerApi,
		LastSeen: time.Now(),
	}
	return nil
}

func (p *ProviderScheduler) delProvider(id types.ProviderID) {
	p.lk.Lock()
	defer p.lk.Unlock()
	if _, ok := p.providers[id]; ok {
		delete(p.providers, id)
	}
	return
}

func (p *ProviderScheduler) watch() {
	heartbeatTimer := time.NewTicker(HeartbeatInterval)
	defer heartbeatTimer.Stop()

	ctx, cancel := context.WithCancel(context.TODO())
	defer cancel()

	for {
		select {
		case <-heartbeatTimer.C:
		}

		sctx, scancel := context.WithTimeout(ctx, HeartbeatInterval/2)

		p.lk.Lock()
		for id, provider := range p.providers {
			_, err := provider.Session(sctx)
			scancel()
			if err != nil {
				if !provider.Expired() {
					// Likely temporary error
					log.Warnw("failed to check provider session", "error", err)
					continue
				}

				log.Warnw("Provider closing", "ProviderID", id)
				delete(p.providers, id)
				continue
			}
			provider.Update()
		}
		p.lk.Unlock()

	}
}
