package scheduler

import (
	"context"
	"log/slog"
	"time"
)

type SyncFunc func() error

type SyncScheduler struct {
	interval time.Duration
	syncFunc SyncFunc

	innerctx context.Context
	cancel   context.CancelFunc
}

func NewSyncScheduler(interval time.Duration, fn SyncFunc) *SyncScheduler {
	ctx, cancel := context.WithCancel(context.Background())
	return &SyncScheduler{
		interval: interval,
		syncFunc: fn,
		innerctx: ctx,
		cancel:   cancel,
	}
}

func (s *SyncScheduler) Run(ctx context.Context) {
	t := time.NewTicker(s.interval)
	defer t.Stop()

	for {
		select {
		case <-s.innerctx.Done():
			return
		case <-ctx.Done():
			return
		case <-t.C:
			if err := s.syncFunc(); err != nil {
				slog.Error("sync job failed", "error", err)
			} else {
				slog.Debug("sync job finished")
			}
		}
	}
}

func (s *SyncScheduler) Stop() {
	s.cancel()
}
