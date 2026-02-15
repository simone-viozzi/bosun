// Package events provides EventEmitter adapter implementations.
//
// The default adapter is LogEmitter, which writes structured log output
// via log/slog for all job lifecycle events (scheduled, started, completed,
// failed, skipped, circuit-broken, etc.).
//
// Additional adapters (Prometheus, Slack, email) are planned for M5+.
package events
