package pst

import (
	"sync"

	"github.com/navid/pst2jmap-migration/internal/model"
)

func StartWorker(wg *sync.WaitGroup, jobs <-chan model.Job, process func(string) error) {
	defer wg.Done()

	for job := range jobs {
		_ = process(job.Path)
	}
}
