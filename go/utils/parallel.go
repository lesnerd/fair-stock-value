package utils

import (
	"context"
	"sync"
	"time"
)

// WorkerPool manages a pool of workers for parallel processing
type WorkerPool struct {
	maxWorkers int
	jobCh      chan func()
	wg         sync.WaitGroup
	ctx        context.Context
	cancel     context.CancelFunc
}

// NewWorkerPool creates a new worker pool
func NewWorkerPool(maxWorkers int) *WorkerPool {
	ctx, cancel := context.WithCancel(context.Background())
	
	wp := &WorkerPool{
		maxWorkers: maxWorkers,
		jobCh:      make(chan func(), maxWorkers*2), // Buffer for jobs
		ctx:        ctx,
		cancel:     cancel,
	}
	
	// Start workers
	for i := 0; i < maxWorkers; i++ {
		wp.wg.Add(1)
		go wp.worker()
	}
	
	return wp
}

// worker is the worker goroutine that processes jobs
func (wp *WorkerPool) worker() {
	defer wp.wg.Done()
	
	for {
		select {
		case job := <-wp.jobCh:
			if job != nil {
				job()
			}
		case <-wp.ctx.Done():
			return
		}
	}
}

// Submit submits a job to the worker pool
func (wp *WorkerPool) Submit(job func()) {
	select {
	case wp.jobCh <- job:
	case <-wp.ctx.Done():
		return
	}
}

// Close closes the worker pool and waits for all workers to finish
func (wp *WorkerPool) Close() {
	close(wp.jobCh)
	wp.cancel()
	wp.wg.Wait()
}

// ProcessWithTimeout processes a job with a timeout
func ProcessWithTimeout(job func() error, timeout time.Duration) error {
	done := make(chan error, 1)
	
	go func() {
		done <- job()
	}()
	
	select {
	case err := <-done:
		return err
	case <-time.After(timeout):
		return context.DeadlineExceeded
	}
}

// Batch processes items in batches
func Batch[T any](items []T, batchSize int, processor func([]T) error) error {
	for i := 0; i < len(items); i += batchSize {
		end := i + batchSize
		if end > len(items) {
			end = len(items)
		}
		
		batch := items[i:end]
		if err := processor(batch); err != nil {
			return err
		}
	}
	
	return nil
}

// RateLimiter limits the rate of operations
type RateLimiter struct {
	ticker   *time.Ticker
	requests chan struct{}
}

// NewRateLimiter creates a new rate limiter
func NewRateLimiter(requestsPerSecond int) *RateLimiter {
	if requestsPerSecond <= 0 {
		requestsPerSecond = 1
	}
	
	interval := time.Second / time.Duration(requestsPerSecond)
	ticker := time.NewTicker(interval)
	
	rl := &RateLimiter{
		ticker:   ticker,
		requests: make(chan struct{}, requestsPerSecond),
	}
	
	// Fill the initial bucket
	for i := 0; i < requestsPerSecond; i++ {
		rl.requests <- struct{}{}
	}
	
	// Start the ticker to refill the bucket
	go func() {
		for range ticker.C {
			select {
			case rl.requests <- struct{}{}:
			default:
				// Bucket is full, skip
			}
		}
	}()
	
	return rl
}

// Wait waits for permission to make a request
func (rl *RateLimiter) Wait(ctx context.Context) error {
	select {
	case <-rl.requests:
		return nil
	case <-ctx.Done():
		return ctx.Err()
	}
}

// Stop stops the rate limiter
func (rl *RateLimiter) Stop() {
	rl.ticker.Stop()
	close(rl.requests)
}