// Package redis provides a Redis implementation of the WorkerRegistry interface.
//
// This implementation uses Redis to store worker registration information with
// automatic expiration through TTL. Workers must send periodic heartbeats to
// remain registered in the system.
//
// Key Design:
//   - Worker data is stored as JSON under key: dago:workers:{worker_id}
//   - Each key has a TTL (default 30 seconds) that is renewed on heartbeat
//   - Pending task counts are retrieved from Redis Streams consumer info
//
// Usage:
//
//	client := redis.NewClient(&redis.Options{Addr: "localhost:6379"})
//	registry := redis.NewRegistry(client, logger)
//
//	// Register a worker
//	worker := ports.WorkerInfo{
//	    ID:   "executor-1",
//	    Type: ports.WorkerTypeExecutor,
//	    Status: ports.WorkerStatusIdle,
//	    RegisteredAt: time.Now(),
//	    LastHeartbeat: time.Now(),
//	}
//	registry.Register(ctx, worker)
//
//	// Send heartbeat every 10 seconds
//	registry.Heartbeat(ctx, "executor-1", ports.WorkerStatusBusy, "task-123")
//
//	// List all healthy workers
//	workers, _ := registry.ListWorkers(ctx, ports.WorkerFilter{HealthyOnly: true})
package redis
