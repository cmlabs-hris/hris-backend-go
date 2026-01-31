package cron

import (
	"context"
	"log/slog"
	"sync"
	"time"
)

// Job represents a scheduled job
type Job struct {
	Name     string
	Interval time.Duration
	Fn       func(ctx context.Context) error
}

// Scheduler manages scheduled jobs
type Scheduler struct {
	jobs   []Job
	ctx    context.Context
	cancel context.CancelFunc
	wg     sync.WaitGroup
	mu     sync.Mutex
}

// NewScheduler creates a new cron scheduler
func NewScheduler() *Scheduler {
	ctx, cancel := context.WithCancel(context.Background())
	return &Scheduler{
		jobs:   make([]Job, 0),
		ctx:    ctx,
		cancel: cancel,
	}
}

// AddJob adds a job to the scheduler
func (s *Scheduler) AddJob(name string, interval time.Duration, fn func(ctx context.Context) error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	s.jobs = append(s.jobs, Job{
		Name:     name,
		Interval: interval,
		Fn:       fn,
	})
	slog.Info("Cron job registered", "name", name, "interval", interval)
}

// Start begins running all scheduled jobs
func (s *Scheduler) Start() {
	s.mu.Lock()
	defer s.mu.Unlock()

	for _, job := range s.jobs {
		s.wg.Add(1)
		go s.runJob(job)
	}

	slog.Info("Cron scheduler started", "job_count", len(s.jobs))
}

// Stop gracefully stops all scheduled jobs
func (s *Scheduler) Stop() {
	slog.Info("Stopping cron scheduler...")
	s.cancel()
	s.wg.Wait()
	slog.Info("Cron scheduler stopped")
}

// runJob runs a single job on its schedule
func (s *Scheduler) runJob(job Job) {
	defer s.wg.Done()

	ticker := time.NewTicker(job.Interval)
	defer ticker.Stop()

	// Run immediately on start
	s.executeJob(job)

	for {
		select {
		case <-s.ctx.Done():
			slog.Info("Cron job stopping", "name", job.Name)
			return
		case <-ticker.C:
			s.executeJob(job)
		}
	}
}

// executeJob executes a job and logs results
func (s *Scheduler) executeJob(job Job) {
	start := time.Now()
	slog.Debug("Cron job starting", "name", job.Name)

	if err := job.Fn(s.ctx); err != nil {
		slog.Error("Cron job failed", "name", job.Name, "error", err, "duration", time.Since(start))
	} else {
		slog.Debug("Cron job completed", "name", job.Name, "duration", time.Since(start))
	}
}

// RunOnce runs all jobs once (useful for testing)
func (s *Scheduler) RunOnce(ctx context.Context) {
	s.mu.Lock()
	defer s.mu.Unlock()

	for _, job := range s.jobs {
		if err := job.Fn(ctx); err != nil {
			slog.Error("Cron job failed", "name", job.Name, "error", err)
		}
	}
}
