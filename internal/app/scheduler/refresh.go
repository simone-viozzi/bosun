package scheduler

import (
	"context"
	"log/slog"
	"time"

	"github.com/simone-viozzi/bosun/internal/domain/jobs"
)

// DiscoverFunc is a function that discovers the current set of enabled jobs.
// It is called periodically by the refresh loop to detect configuration changes.
type DiscoverFunc func(ctx context.Context) ([]jobs.Job, error)

// DiffResult describes the changes between the currently registered jobs
// and a freshly-discovered set.
type DiffResult struct {
	Added   []jobs.Job   // Jobs discovered but not currently registered.
	Removed []string     // Job names registered but no longer discovered.
	Changed []ChangedJob // Jobs whose schedule or overlap policy changed.
}

// ChangedJob holds the old and new state of a changed job.
type ChangedJob struct {
	OldSchedule      string
	OldOverlapPolicy jobs.OverlapPolicy
	NewJob           jobs.Job
}

// diffJobs compares a set of discovered jobs against the scheduler's current
// entries and returns the differences.
// Circuit-broken jobs that reappear in discovery are excluded from Added/Changed
// to honour FR-041: auto-disabled jobs MUST NOT be re-enabled by config refresh.
func (s *Scheduler) diffJobs(discovered []jobs.Job) DiffResult {
	s.mu.Lock()
	defer s.mu.Unlock()

	var result DiffResult

	// Build a set of discovered job names for fast lookup.
	discoveredMap := make(map[string]jobs.Job, len(discovered))
	for _, j := range discovered {
		discoveredMap[j.Name] = j
	}

	// Detect removed and changed jobs.
	for name, e := range s.entries {
		dj, found := discoveredMap[name]
		if !found {
			// Job no longer in discovery → removed (unless circuit-broken, still remove).
			result.Removed = append(result.Removed, name)
			continue
		}
		// Job exists in both — check for changes (skip if circuit-broken).
		if e.circuitBroken {
			continue
		}
		if e.job.Schedule != dj.Schedule || e.job.OverlapPolicy != dj.OverlapPolicy {
			result.Changed = append(result.Changed, ChangedJob{
				OldSchedule:      e.job.Schedule,
				OldOverlapPolicy: e.job.OverlapPolicy,
				NewJob:           dj,
			})
		}
	}

	// Detect new jobs.
	for name, dj := range discoveredMap {
		if _, exists := s.entries[name]; !exists {
			result.Added = append(result.Added, dj)
		}
	}

	return result
}

// ApplyRefresh discovers, diffs, and applies configuration changes.
// Returns the diff result for logging/testing purposes.
func (s *Scheduler) ApplyRefresh(ctx context.Context, discovered []jobs.Job) DiffResult {
	diff := s.diffJobs(discovered)

	// T057: Handle new jobs.
	for _, j := range diff.Added {
		if err := s.AddJob(ctx, j); err != nil {
			s.logger.ErrorContext(ctx, "refresh: failed to add job",
				slog.String("job", j.Name),
				slog.String("error", err.Error()),
			)
			continue
		}
		s.events.EmitJobAdded(ctx, j.Name)
	}

	// T058: Handle removed jobs (current run completes via cron removal).
	for _, name := range diff.Removed {
		if err := s.RemoveJob(ctx, name); err != nil {
			s.logger.ErrorContext(ctx, "refresh: failed to remove job",
				slog.String("job", name),
				slog.String("error", err.Error()),
			)
			continue
		}
		// NOTE: EmitJobRemoved is already called inside RemoveJob; no duplicate emit here.
	}

	// T059: Handle changed jobs (remove + re-add with new config).
	for _, ch := range diff.Changed {
		name := ch.NewJob.Name
		if err := s.RemoveJob(ctx, name); err != nil {
			s.logger.ErrorContext(ctx, "refresh: failed to remove changed job",
				slog.String("job", name),
				slog.String("error", err.Error()),
			)
			continue
		}
		if err := s.AddJob(ctx, ch.NewJob); err != nil {
			s.logger.ErrorContext(ctx, "refresh: failed to re-add changed job",
				slog.String("job", name),
				slog.String("error", err.Error()),
			)
			continue
		}
		s.events.EmitJobChanged(ctx, name, ch.OldSchedule, ch.NewJob.Schedule)
	}

	return diff
}

// startRefreshLoop runs a periodic refresh ticker that calls discoverFn,
// diffs the results against registered jobs, and applies changes.
// It blocks until ctx is cancelled.
func (s *Scheduler) startRefreshLoop(ctx context.Context, interval time.Duration, discoverFn DiscoverFunc) {
	if discoverFn == nil || interval <= 0 {
		return
	}

	ticker := time.NewTicker(interval)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			discovered, err := discoverFn(ctx)
			if err != nil {
				s.logger.ErrorContext(ctx, "refresh: discovery failed",
					slog.String("error", err.Error()),
				)
				continue
			}

			diff := s.ApplyRefresh(ctx, discovered)

			if len(diff.Added)+len(diff.Removed)+len(diff.Changed) > 0 {
				s.logger.InfoContext(ctx, "refresh: configuration updated",
					slog.Int("added", len(diff.Added)),
					slog.Int("removed", len(diff.Removed)),
					slog.Int("changed", len(diff.Changed)),
				)
			}
		}
	}
}
