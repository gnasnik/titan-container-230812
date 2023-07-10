package node

import (
	"errors"
	"github.com/gnasnik/titan-container/db"
	"github.com/gnasnik/titan-container/node/impl/manager"
	"github.com/gnasnik/titan-container/node/modules"

	"github.com/gnasnik/titan-container/api"
	"github.com/gnasnik/titan-container/node/config"
	"github.com/gnasnik/titan-container/node/repo"
	"github.com/jmoiron/sqlx"
	"go.uber.org/fx"
	"golang.org/x/xerrors"
)

func Manager(out *api.Manager) Option {
	return Options(
		ApplyIf(func(s *Settings) bool { return s.Config },
			Error(errors.New("the Manager option must be set before Config option")),
		),

		func(s *Settings) error {
			s.nodeType = repo.Manager
			return nil
		},

		func(s *Settings) error {
			resAPI := &manager.Manager{}
			s.invokes[ExtractAPIKey] = fx.Populate(resAPI)
			*out = resAPI
			return nil
		},
	)
}

func ConfigManager(c interface{}) Option {
	cfg, ok := c.(*config.ManagerCfg)
	if !ok {
		return Error(xerrors.Errorf("invalid config from repo, got: %T", c))
	}

	return Options(
		Override(new(*config.ManagerCfg), cfg),
		ConfigCommon(&cfg.Common),
		Override(new(*sqlx.DB), modules.NewManagerDB(cfg.DatabaseAddress)),
		Override(new(*db.ManagerDB), db.NewManagerDB),
	)
}
