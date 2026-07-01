package pst

import (
	"context"
	"sync"

	"github.com/navid/pst2jmap-migration/internal/model"
)

func StartWorker(
	ctx context.Context,
	wg *sync.WaitGroup,
	jobs <-chan model.Job,
	process func(context.Context, string) error,
) {
	defer wg.Done()

	for {
		select {
		case <-ctx.Done():
			return

		case job, ok := <-jobs:
			if !ok {
				return
			}

			_ = process(ctx, job.PSTFile)
		}
	}
}
