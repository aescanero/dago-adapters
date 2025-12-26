// Package worker_registry provides adapters for the WorkerRegistry interface.
//
// The WorkerRegistry interface is defined in dago-libs/pkg/ports and provides
// a transport-agnostic way to manage worker registration and heartbeats.
//
// Available implementations:
//   - redis: Uses Redis for storage with automatic expiration via TTL
//
// Future implementations could include:
//   - kafka: Using Kafka topics for worker state
//   - websocket: Using WebSocket connections for real-time updates
//   - database: Using PostgreSQL or other databases
package worker_registry
