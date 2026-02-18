package cmd

import (
	"context"
	"fmt"
	"log/slog"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/spf13/cobra"

	"github.com/simone-viozzi/bosun/internal/app"
	"github.com/simone-viozzi/bosun/internal/app/scheduler"
	"github.com/simone-viozzi/bosun/internal/domain/jobs"
	dlabels "github.com/simone-viozzi/bosun/internal/domain/labels"
	"github.com/simone-viozzi/bosun/internal/ports"
)

// NewDaemonCmd creates the `bosun daemon` command.
// This is a long-lived process that schedules and executes jobs based on
// Docker label configuration with periodic refresh.
func NewDaemonCmd() *cobra.Command {
	var (
		parallelism     int
		refreshInterval string
	)

	cmd := &cobra.Command{
		Use:   "daemon",
		Short: "Run as a long-lived scheduling daemon",
		Long: `Starts Bosun as a long-lived daemon that continuously discovers jobs
from Docker labels and schedules them using cron expressions.

The daemon:
  1. Discovers all enabled jobs from Docker labels
  2. Registers each job with its cron schedule
  3. Executes jobs through the standard execution pipeline
  4. Periodically refreshes configuration to detect changes
  5. Shuts down gracefully on SIGTERM/SIGINT

Concurrency is controlled by three layers:
  - Per-job overlap policies (queue or skip)
  - Global parallelism limit (--parallelism flag)
  - Automatic per-stack mutual exclusion`,
		Example: `  # Start daemon with default settings (serial execution)
  bosun daemon

  # Allow up to 3 concurrent jobs
  bosun daemon --parallelism 3

  # Custom refresh interval
  bosun daemon --refresh-interval 1m`,
		RunE: func(cmd *cobra.Command, _ []string) error {
			return runDaemon(cmd.Context(), parallelism, refreshInterval)
		},
	}

	cmd.Flags().IntVar(&parallelism, "parallelism", 1, "Maximum number of concurrent jobs (default: 1)")
	cmd.Flags().StringVar(&refreshInterval, "refresh-interval", "5m", "How often to refresh job configuration from Docker labels")

	return cmd
}

// shutdownTimeout is the maximum time to wait for graceful shutdown.
const shutdownTimeout = 60 * time.Second

// runDaemon implements the daemon lifecycle:
// Bootstrap → discover jobs → Scheduler.Start → block on signals.
func runDaemon(ctx context.Context, parallelism int, refreshIntervalStr string) error {
	logger := slog.Default()

	refreshInterval, err := time.ParseDuration(refreshIntervalStr)
	if err != nil {
		return fmt.Errorf("invalid --refresh-interval %q: %w", refreshIntervalStr, err)
	}

	logger.InfoContext(ctx, "daemon starting",
		slog.Int("parallelism", parallelism),
		slog.Duration("refresh_interval", refreshInterval),
	)

	// 1. Bootstrap services.
	svc, err := app.Bootstrap(ctx, app.BootstrapOptions{
		Parallelism:     parallelism,
		RefreshInterval: refreshInterval,
	})
	if err != nil {
		return fmt.Errorf("bootstrap failed: %w", err)
	}
	defer func() {
		if err := svc.Close(); err != nil {
			logger.ErrorContext(ctx, "error closing services", slog.String("error", err.Error()))
		}
	}()

	// 2. Create scheduler with refresh loop configuration.
	sched := scheduler.New(
		svc.Executor,
		svc.EventEmitter,
		svc.StateStore,
		scheduler.Options{
			Parallelism:     int64(parallelism),
			RefreshInterval: refreshInterval,
			DiscoverFn:      makeDiscoverFunc(svc, logger),
		},
		logger,
	)

	// 3. Discover and register initial jobs.
	if err := discoverAndRegisterJobs(ctx, svc, sched, logger); err != nil {
		return fmt.Errorf("initial job discovery failed: %w", err)
	}

	// 4. Run scheduler with signal handling.
	return runWithSignalHandling(ctx, sched, logger)
}

// runnableScheduler abstracts the scheduler's Start/Stop lifecycle for testability.
type runnableScheduler interface {
	Start(ctx context.Context) error
	Stop(ctx context.Context) error
}

// runWithSignalHandling starts the scheduler and blocks until a SIGTERM/SIGINT
// triggers graceful shutdown. A second signal cancels the shutdown context for
// immediate exit.
func runWithSignalHandling(ctx context.Context, sched runnableScheduler, logger *slog.Logger) error {
	// Set up signal handling.
	sigCtx, stop := signal.NotifyContext(ctx, syscall.SIGTERM, syscall.SIGINT)
	defer stop()

	// Start scheduler in a goroutine.
	schedErr := make(chan error, 1)
	go func() {
		schedErr <- sched.Start(sigCtx)
	}()

	logger.InfoContext(ctx, "daemon started: scheduler running")

	// Wait for first signal (graceful shutdown).
	<-sigCtx.Done()
	logger.InfoContext(ctx, "shutdown initiated: received signal")
	stop() // Reset signal handling for double-signal behavior.

	// Double-signal handler: second SIGTERM/SIGINT forces immediate exit.
	forceSig := make(chan os.Signal, 1)
	signal.Notify(forceSig, syscall.SIGTERM, syscall.SIGINT)
	defer signal.Stop(forceSig)

	// Wait for scheduler to stop with timeout.
	shutdownCtx, shutdownCancel := context.WithTimeout(context.Background(), shutdownTimeout)
	defer shutdownCancel()

	go func() {
		select {
		case <-forceSig:
			logger.WarnContext(ctx, "force shutdown: received second signal")
			shutdownCancel()
		case <-shutdownCtx.Done():
		}
	}()

	// Wait for scheduler to stop.
	if err := sched.Stop(shutdownCtx); err != nil {
		logger.ErrorContext(ctx, "scheduler stop error", slog.String("error", err.Error()))
	}

	logger.InfoContext(ctx, "daemon stopped: shutdown complete")

	return nil
}

// discoverAndRegisterJobs discovers jobs from Docker labels and registers them with the scheduler.
func discoverAndRegisterJobs(ctx context.Context, svc *app.Services, sched *scheduler.Scheduler, logger *slog.Logger) error {
	// 1. Take a label snapshot.
	snapshot, err := svc.LabelSource.Snapshot(ctx, ports.Selector{
		Prefixes: []string{dlabels.DefaultLabelPrefix},
	})
	if err != nil {
		return fmt.Errorf("label snapshot failed: %w", err)
	}

	// 2. Discover jobs from labels.
	discovered, validationErrs, err := svc.Discoverer.DiscoverJobs(ctx, snapshot)
	if err != nil {
		return fmt.Errorf("job discovery failed: %w", err)
	}
	for _, ve := range validationErrs {
		logger.WarnContext(ctx, "validation error during discovery",
			slog.String("entity", ve.EntityName),
			slog.String("field", ve.Field),
			slog.String("message", ve.Message),
		)
	}

	// 3. Register each enabled job.
	registered := 0
	for _, job := range discovered {
		if !job.Enabled {
			logger.InfoContext(ctx, "skipping disabled job", slog.String("job", job.Name))
			continue
		}
		if job.Schedule == "" {
			logger.WarnContext(ctx, "skipping job without schedule", slog.String("job", job.Name))
			continue
		}
		if err := sched.AddJob(ctx, job); err != nil {
			logger.ErrorContext(ctx, "failed to register job",
				slog.String("job", job.Name),
				slog.String("error", err.Error()),
			)
			continue
		}
		registered++
	}

	logger.InfoContext(ctx, "jobs discovered and registered",
		slog.Int("total", len(discovered)),
		slog.Int("registered", registered),
	)

	return nil
}

// makeDiscoverFunc creates a DiscoverFunc that uses the app services to
// discover enabled jobs from Docker labels. Used by the scheduler's refresh loop.
func makeDiscoverFunc(svc *app.Services, logger *slog.Logger) scheduler.DiscoverFunc {
	return func(ctx context.Context) ([]jobs.Job, error) {
		snapshot, err := svc.LabelSource.Snapshot(ctx, ports.Selector{
			Prefixes: []string{dlabels.DefaultLabelPrefix},
		})
		if err != nil {
			return nil, fmt.Errorf("label snapshot failed: %w", err)
		}

		discovered, validationErrs, err := svc.Discoverer.DiscoverJobs(ctx, snapshot)
		if err != nil {
			return nil, fmt.Errorf("job discovery failed: %w", err)
		}
		for _, ve := range validationErrs {
			logger.WarnContext(ctx, "validation error during refresh",
				slog.String("entity", ve.EntityName),
				slog.String("field", ve.Field),
				slog.String("message", ve.Message),
			)
		}

		// Filter to only enabled jobs with schedules.
		enabled := make([]jobs.Job, 0, len(discovered))
		for _, j := range discovered {
			if j.Enabled && j.Schedule != "" {
				enabled = append(enabled, j)
			}
		}
		return enabled, nil
	}
}
