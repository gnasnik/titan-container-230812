package modules

import (
	"github.com/gnasnik/titan-container/node/config"
	"github.com/gnasnik/titan-container/node/repo"
	logging "github.com/ipfs/go-log/v2"
)

var log = logging.Logger("modules")

// NewSetManagerConfigFunc creates a function to set the manager config
func NewSetManagerConfigFunc(r repo.LockedRepo) func(cfg config.ManagerCfg) error {
	return func(cfg config.ManagerCfg) (err error) {
		return r.SetConfig(func(raw interface{}) {
			_, ok := raw.(*config.ManagerCfg)
			if !ok {
				return
			}
		})
	}
}

// NewGetManagerConfigFunc creates a function to get the manager config
func NewGetManagerConfigFunc(r repo.LockedRepo) func() (config.ManagerCfg, error) {
	return func() (out config.ManagerCfg, err error) {
		raw, err := r.Config()
		if err != nil {
			return
		}

		scfg, ok := raw.(*config.ManagerCfg)
		if !ok {
			return
		}

		out = *scfg
		return
	}
}
