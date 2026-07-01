package services

import (
	"context"
	"errors"
	"log/slog"
	"sync"
)

type WorkerPool struct {
	size    int
	queue   chan ImageJob
	service *ImageService
	logger  *slog.Logger
	wg      sync.WaitGroup
}

func NewWorkerPool(size int, queueSize int, service *ImageService, logger *slog.Logger) *WorkerPool {
	return &WorkerPool{
		size:    size,
		queue:   make(chan ImageJob, queueSize),
		service: service,
		logger:  logger,
	}
}

func (p *WorkerPool) Start() {
	for i := 0; i < p.size; i++ {
		workerID := i + 1
		p.wg.Add(1)
		go func() {
			defer p.wg.Done()
			for job := range p.queue {
				result := p.service.Generate(context.Background(), job)
				job.Result <- result
				close(job.Result)
				p.logger.Info("image job finished", "worker_id", workerID, "user_id", job.UserID)
			}
		}()
	}
}

func (p *WorkerPool) Stop() {
	close(p.queue)
	p.wg.Wait()
}

func (p *WorkerPool) Submit(job ImageJob) error {
	select {
	case p.queue <- job:
		return nil
	default:
		return errors.New("image worker queue is full")
	}
}
