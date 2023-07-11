package modules

import (
	"context"
	"github.com/gnasnik/titan-container/journal"
	"github.com/gnasnik/titan-container/journal/fsjournal"
	"github.com/gnasnik/titan-container/node/repo"
	"go.uber.org/fx"
)

func OpenFilesystemJournal(lr repo.LockedRepo, lc fx.Lifecycle, disabled journal.DisabledEvents) (journal.Journal, error) {
	jrnl, err := fsjournal.OpenFSJournal(lr, disabled)
	if err != nil {
		return nil, err
	}

	lc.Append(fx.Hook{
		OnStop: func(_ context.Context) error { return jrnl.Close() },
	})

	return jrnl, err
}
