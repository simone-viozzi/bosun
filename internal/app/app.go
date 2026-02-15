package app

import (
	"context"
	"fmt"
	"log/slog"
	"time"

	"github.com/docker/docker/client"
	"golang.org/x/sync/semaphore"

	compose "github.com/simone-viozzi/bosun/internal/adapters/docker/compose"
	worker "github.com/simone-viozzi/bosun/internal/adapters/docker/worker"
	"github.com/simone-viozzi/bosun/internal/adapters/dockerlabels"
	"github.com/simone-viozzi/bosun/internal/adapters/events"
	"github.com/simone-viozzi/bosun/internal/adapters/joblabels"
	"github.com/simone-viozzi/bosun/internal/adapters/state"
	"github.com/simone-viozzi/bosun/internal/app/executor"
	"github.com/simone-viozzi/bosun/internal/app/planner"
	"github.com/simone-viozzi/bosun/internal/ports"
)

// App is the legacy entrypoint (kept for backward compatibility).
type App struct{}

// New creates an App. Deprecated: use Bootstrap instead.
func New() *App { return &App{} }

// Run is a no-op placeholder. Deprecated: use Bootstrap + Services.
func (a *App) Run(_ context.Context, _ []string) error { return nil }

// Services provides dependency-injected application services.
// Created via Bootstrap() to resolve #141 (CLI→adapter coupling).
type Services struct {
	// LabelSource takes Docker label snapshots for job discovery.
	LabelSource ports.LabelSource

	// Discoverer parses label snapshots into Job definitions.
	Discoverer ports.JobDiscoverer

	// Planner generates execution plans for jobs.
	Planner ports.JobPlanner

	// Executor runs jobs through the plan-driven pipeline.
	Executor ports.JobExecutor

	// EventEmitter emits job lifecycle events for observability.
	EventEmitter ports.EventEmitter

	// StateStore persists per-job state (M4: in-memory, #177: durable).
	StateStore ports.JobStateStore

	// GlobalSem limits total concurrent job execution.
	GlobalSem *semaphore.Weighted

	// DockerClient is the underlying Docker SDK client.
	// Caller is responsible for closing it.
	DockerClient *client.Client
}

// BootstrapOptions configures service initialization.
type BootstrapOptions struct {
	// Parallelism is the global concurrency limit (default 1).
	Parallelism int

	// RefreshInterval is how often to refresh config from Docker labels.
	// Only used in daemon mode.
	RefreshInterval time.Duration
}

// Bootstrap creates and wires all application services.
// This is the single place where adapters are instantiated, resolving #141.
func Bootstrap(_ context.Context, opts BootstrapOptions) (*Services, error) {
	if opts.Parallelism < 1 {
		opts.Parallelism = 1
	}
	if opts.RefreshInterval == 0 {
		opts.RefreshInterval = 5 * time.Minute
	}

	// 1. Create Docker client
	dockerClient, err := client.NewClientWithOpts(client.FromEnv, client.WithAPIVersionNegotiation())
	if err != nil {
		return nil, fmt.Errorf("failed to create Docker client: %w", err)
	}

	// 2. Instantiate adapters
	labelSource, err := dockerlabels.NewFromEnv()
	if err != nil {
		_ = dockerClient.Close()
		return nil, fmt.Errorf("failed to create label source: %w", err)
	}

	discoverer := joblabels.NewDiscoverer()
	composeCtrl := compose.NewController(dockerClient)
	workerRunner := worker.NewRunner(dockerClient)
	eventEmitter := events.NewLogEmitter(slog.Default())
	stateStore := state.NewInMemoryStateStore()

	// 3. Create app-layer services
	jobPlanner := planner.New()
	exec := executor.New(jobPlanner, composeCtrl, workerRunner, dockerClient, eventEmitter)

	// 4. Create global semaphore
	globalSem := semaphore.NewWeighted(int64(opts.Parallelism))

	return &Services{
		LabelSource:  labelSource,
		Discoverer:   discoverer,
		Planner:      jobPlanner,
		Executor:     exec,
		EventEmitter: eventEmitter,
		StateStore:   stateStore,
		GlobalSem:    globalSem,
		DockerClient: dockerClient,
	}, nil
}

// Close releases resources held by Services.
func (s *Services) Close() error {
	if s.DockerClient != nil {
		return s.DockerClient.Close()
	}
	return nil
}
