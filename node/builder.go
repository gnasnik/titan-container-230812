package node

import (
	"context"
	"errors"
	"github.com/gnasnik/titan-container/journal"
	"github.com/gnasnik/titan-container/journal/alerting"
	"github.com/gnasnik/titan-container/node/common"
	"github.com/gnasnik/titan-container/node/config"
	"github.com/gnasnik/titan-container/node/modules"
	"github.com/gnasnik/titan-container/node/modules/dtypes"
	"github.com/gnasnik/titan-container/node/modules/helpers"
	"github.com/gnasnik/titan-container/node/repo"
	"github.com/gnasnik/titan-container/node/types"
	logging "github.com/ipfs/go-log/v2"
	metricsi "github.com/ipfs/go-metrics-interface"
	"go.uber.org/fx"
	"golang.org/x/xerrors"
)

var log = logging.Logger("builder")

type invoke int

// Invokes are called in the order they are defined.
//
//nolint:golint
const (
	// InitJournal at position 0 initializes the journal global var as soon as
	// the system starts, so that it's available for all other components.
	InitJournalKey = invoke(iota)

	ExtractAPIKey

	CheckFDLimit

	SetApiEndpointKey

	_nInvokes // keep this last
)

type Settings struct {
	// modules is a map of constructors for DI
	//
	// In most cases the index will be a reflect. Type of element returned by
	// the constructor, but for some 'constructors' it's hard to specify what's
	// the return type should be (or the constructor returns fx group)
	modules map[interface{}]fx.Option

	// invokes are separate from modules as they can't be referenced by return
	// type, and must be applied in correct order
	invokes []fx.Option

	nodeType repo.RepoType

	Base   bool // Base option applied
	Config bool // Config option applied
}

// Basic services
func defaults() []Option {
	return []Option{
		// global system journal.
		Override(new(journal.DisabledEvents), journal.EnvDisabledEvents),
		Override(new(*alerting.Alerting), alerting.NewAlertingSystem),
	}
}

// Config sets up constructors based on the provided Config
func ConfigCommon(cfg *config.Common) Option {
	return Options(
		func(s *Settings) error { s.Config = true; return nil },
		//Override(new(dtypes.APIEndpoint), func() (dtypes.APIEndpoint, error) {
		//	return multiaddr.NewMultiaddr(cfg.API.ListenAddress)
		//}),
		//Override(SetApiEndpointKey, func(lr repo.LockedRepo, e dtypes.APIEndpoint) error {
		//	return lr.SetAPIEndpoint(e)
		//}),
		ApplyIf(func(s *Settings) bool { return s.Base }), // apply only if Base has already been applied
	)
}

func Repo(r repo.Repo) Option {
	return func(settings *Settings) error {
		lr, err := r.Lock(settings.nodeType)
		if err != nil {
			return err
		}
		c, err := lr.Config()
		if err != nil {
			return err
		}
		return Options(
			Override(CheckFDLimit, modules.CheckFdLimit),
			Override(new(repo.LockedRepo), modules.LockedRepo(lr)), // module handles closing
			Override(new(*common.CommonAPI), common.NewCommonAPI),
			Override(new(types.KeyStore), modules.KeyStore),
			Override(new(*dtypes.APIAlg), modules.APISecret),
			Override(new(helpers.MetricsCtx), func() context.Context {
				return metricsi.CtxScope(context.Background(), "titan")
			}),

			ApplyIf(IsType(repo.Manager), ConfigManager(c)),
			ApplyIf(IsType(repo.Provider), ConfigProvider(c)),
		)(settings)
	}
}

type StopFunc func(context.Context) error

// New builds and starts new Titan node
func New(ctx context.Context, opts ...Option) (StopFunc, error) {
	settings := Settings{
		modules: map[interface{}]fx.Option{},
		invokes: make([]fx.Option, _nInvokes),
	}

	// apply module options in the right order
	if err := Options(Options(defaults()...), Options(opts...))(&settings); err != nil {
		return nil, xerrors.Errorf("applying node options failed: %w", err)
	}

	// gather constructors for fx.Options
	ctors := make([]fx.Option, 0, len(settings.modules))
	for _, opt := range settings.modules {
		ctors = append(ctors, opt)
	}

	// fill holes in invokes for use in fx.Options
	for i, opt := range settings.invokes {
		if opt == nil {
			settings.invokes[i] = fx.Options()
		}
	}

	app := fx.New(
		fx.Options(ctors...),
		fx.Options(settings.invokes...),

		fx.NopLogger,
	)

	// TODO: we probably should have a 'firewall' for Closing signal
	//  on this context, and implement closing logic through lifecycles
	//  correctly
	if err := app.Start(ctx); err != nil {
		// comment fx.NopLogger few lines above for easier debugging
		return nil, xerrors.Errorf("starting node: %w", err)
	}

	return app.Stop, nil
}

func IsType(t repo.RepoType) func(s *Settings) bool {
	return func(s *Settings) bool { return s.nodeType == t }
}

func Base() Option {
	return Options(
		func(s *Settings) error { s.Base = true; return nil }, // mark Base as applied
		ApplyIf(func(s *Settings) bool { return s.Config },
			Error(errors.New("the Base() option must be set before Config option")),
		),
	)
}
