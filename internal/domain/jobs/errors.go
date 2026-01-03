// Package jobs contains error types for job execution.
package jobs

import (
	"errors"
	"fmt"
	"time"
)

// Sentinel errors for error checking.
var (
	ErrJobNotFound       = errors.New("job not found")
	ErrStackNotFound     = errors.New("stack not found")
	ErrImageNotFound     = errors.New("worker image not found")
	ErrStackPartialState = errors.New("stack in partial state")
	ErrExecutionTimeout  = errors.New("execution timeout")
)

// StopError indicates a failure during stack stop.
type StopError struct {
	StackName     string
	ContainerName string
	ContainerID   string
	Cause         error
}

func (e *StopError) Error() string {
	return fmt.Sprintf("failed to stop container %s (%s) in stack %s: %v",
		e.ContainerName, e.ContainerID, e.StackName, e.Cause)
}

func (e *StopError) Unwrap() error {
	return e.Cause
}

// StartError indicates a failure during stack start.
type StartError struct {
	StackName     string
	ContainerName string
	ContainerID   string
	Cause         error
}

func (e *StartError) Error() string {
	return fmt.Sprintf("failed to start container %s (%s) in stack %s: %v",
		e.ContainerName, e.ContainerID, e.StackName, e.Cause)
}

func (e *StartError) Unwrap() error {
	return e.Cause
}

// TimeoutError indicates an operation exceeded its timeout.
type TimeoutError struct {
	Operation string        // "stop", "start", "worker"
	Target    string        // Stack or container name
	Duration  time.Duration // Timeout that was exceeded
}

func (e *TimeoutError) Error() string {
	return fmt.Sprintf("%s operation on %s timed out after %s",
		e.Operation, e.Target, e.Duration)
}

func (e *TimeoutError) Is(target error) bool {
	return target == ErrExecutionTimeout
}

// WorkerError indicates worker container failure.
type WorkerError struct {
	ExitCode int
	Logs     string
}

func (e *WorkerError) Error() string {
	return fmt.Sprintf("worker failed with exit code %d", e.ExitCode)
}

// ValidationError for pre-execution validation failures.
type ValidationError struct {
	JobName string
	Field   string
	Message string
}

func (e *ValidationError) Error() string {
	return fmt.Sprintf("job %s validation failed: %s - %s",
		e.JobName, e.Field, e.Message)
}
