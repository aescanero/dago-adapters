package redis

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"github.com/aescanero/dago-libs/pkg/ports"
	"github.com/redis/go-redis/v9"
	"go.uber.org/zap"
)

const (
	// Default TTL for worker heartbeats (30 seconds)
	defaultWorkerTTL = 30 * time.Second

	// Key prefix for worker data
	workerKeyPrefix = "dago:workers:"

	// Stream keys for executor and router workers
	executorStreamKey = "executor.work"
	routerStreamKey   = "router.work"

	// Consumer group names
	executorConsumerGroup = "executor-workers"
	routerConsumerGroup   = "router-workers"
)

// Registry implements ports.WorkerRegistry using Redis
type Registry struct {
	client *redis.Client
	logger *zap.Logger
	ttl    time.Duration
}

// NewRegistry creates a new Redis worker registry
func NewRegistry(client *redis.Client, logger *zap.Logger) *Registry {
	return &Registry{
		client: client,
		logger: logger,
		ttl:    defaultWorkerTTL,
	}
}

// NewRegistryWithTTL creates a new Redis worker registry with custom TTL
func NewRegistryWithTTL(client *redis.Client, ttl time.Duration, logger *zap.Logger) *Registry {
	return &Registry{
		client: client,
		logger: logger,
		ttl:    ttl,
	}
}

// Register registers a new worker in the system
func (r *Registry) Register(ctx context.Context, worker ports.WorkerInfo) error {
	key := r.getWorkerKey(worker.ID)

	// Serialize worker info to JSON
	data, err := json.Marshal(worker)
	if err != nil {
		return fmt.Errorf("failed to marshal worker info: %w", err)
	}

	// Store in Redis with TTL
	if err := r.client.Set(ctx, key, data, r.ttl).Err(); err != nil {
		return fmt.Errorf("failed to register worker: %w", err)
	}

	r.logger.Info("worker registered",
		zap.String("worker_id", worker.ID),
		zap.String("type", string(worker.Type)),
		zap.Duration("ttl", r.ttl))

	return nil
}

// Unregister removes a worker from the registry
func (r *Registry) Unregister(ctx context.Context, workerID string) error {
	key := r.getWorkerKey(workerID)

	if err := r.client.Del(ctx, key).Err(); err != nil {
		return fmt.Errorf("failed to unregister worker: %w", err)
	}

	r.logger.Info("worker unregistered", zap.String("worker_id", workerID))
	return nil
}

// Heartbeat updates the last heartbeat timestamp for a worker
func (r *Registry) Heartbeat(ctx context.Context, workerID string, status ports.WorkerStatus, currentTask string) error {
	key := r.getWorkerKey(workerID)

	// Get existing worker info
	worker, err := r.GetWorker(ctx, workerID)
	if err != nil {
		// Worker not found, this shouldn't happen but we can recover
		r.logger.Warn("heartbeat for unregistered worker, auto-registering",
			zap.String("worker_id", workerID))

		// Try to determine worker type from ID
		workerType := r.inferWorkerType(workerID)

		worker = &ports.WorkerInfo{
			ID:            workerID,
			Type:          workerType,
			Status:        status,
			RegisteredAt:  time.Now(),
			LastHeartbeat: time.Now(),
			CurrentTask:   currentTask,
		}
	} else {
		// Update existing worker info
		worker.Status = status
		worker.LastHeartbeat = time.Now()
		worker.CurrentTask = currentTask
	}

	// Get pending tasks from Redis Streams consumer info
	pendingTasks, err := r.getPendingTasksForWorker(ctx, workerID, worker.Type)
	if err != nil {
		r.logger.Warn("failed to get pending tasks",
			zap.String("worker_id", workerID),
			zap.Error(err))
	} else {
		worker.PendingTasks = pendingTasks
	}

	// Serialize and save with renewed TTL
	data, err := json.Marshal(worker)
	if err != nil {
		return fmt.Errorf("failed to marshal worker info: %w", err)
	}

	if err := r.client.Set(ctx, key, data, r.ttl).Err(); err != nil {
		return fmt.Errorf("failed to update heartbeat: %w", err)
	}

	return nil
}

// GetWorker retrieves information about a specific worker
func (r *Registry) GetWorker(ctx context.Context, workerID string) (*ports.WorkerInfo, error) {
	key := r.getWorkerKey(workerID)

	data, err := r.client.Get(ctx, key).Bytes()
	if err != nil {
		if err == redis.Nil {
			return nil, fmt.Errorf("worker not found: %s", workerID)
		}
		return nil, fmt.Errorf("failed to get worker: %w", err)
	}

	var worker ports.WorkerInfo
	if err := json.Unmarshal(data, &worker); err != nil {
		return nil, fmt.Errorf("failed to unmarshal worker info: %w", err)
	}

	// Check if worker is healthy based on last heartbeat
	if time.Since(worker.LastHeartbeat) > r.ttl {
		worker.Status = ports.WorkerStatusUnhealthy
	}

	return &worker, nil
}

// ListWorkers retrieves all workers matching the filter criteria
func (r *Registry) ListWorkers(ctx context.Context, filter ports.WorkerFilter) ([]ports.WorkerInfo, error) {
	// Scan for all worker keys
	pattern := workerKeyPrefix + "*"
	keys, err := r.scanKeys(ctx, pattern)
	if err != nil {
		return nil, fmt.Errorf("failed to scan worker keys: %w", err)
	}

	var workers []ports.WorkerInfo

	for _, key := range keys {
		data, err := r.client.Get(ctx, key).Bytes()
		if err != nil {
			if err == redis.Nil {
				continue // Key expired between scan and get
			}
			r.logger.Warn("failed to get worker",
				zap.String("key", key),
				zap.Error(err))
			continue
		}

		var worker ports.WorkerInfo
		if err := json.Unmarshal(data, &worker); err != nil {
			r.logger.Warn("failed to unmarshal worker",
				zap.String("key", key),
				zap.Error(err))
			continue
		}

		// Check if worker is healthy
		isHealthy := time.Since(worker.LastHeartbeat) <= r.ttl
		if !isHealthy {
			worker.Status = ports.WorkerStatusUnhealthy
		}

		// Apply filters
		if !r.matchesFilter(worker, filter, isHealthy) {
			continue
		}

		workers = append(workers, worker)
	}

	return workers, nil
}

// GetWorkerStats returns aggregate statistics about workers
func (r *Registry) GetWorkerStats(ctx context.Context, workerType ports.WorkerType) (*ports.WorkerStats, error) {
	filter := ports.WorkerFilter{
		Types: []ports.WorkerType{workerType},
	}

	workers, err := r.ListWorkers(ctx, filter)
	if err != nil {
		return nil, err
	}

	stats := &ports.WorkerStats{
		Type:         workerType,
		TotalWorkers: len(workers),
	}

	for _, worker := range workers {
		switch worker.Status {
		case ports.WorkerStatusIdle:
			stats.IdleWorkers++
		case ports.WorkerStatusBusy:
			stats.BusyWorkers++
		case ports.WorkerStatusUnhealthy:
			stats.UnhealthyWorkers++
		}
		stats.TotalPendingTasks += worker.PendingTasks
	}

	return stats, nil
}

// CleanupStaleWorkers removes workers that haven't sent a heartbeat within the timeout
func (r *Registry) CleanupStaleWorkers(ctx context.Context, timeout time.Duration) (int, error) {
	// Scan for all worker keys
	pattern := workerKeyPrefix + "*"
	keys, err := r.scanKeys(ctx, pattern)
	if err != nil {
		return 0, fmt.Errorf("failed to scan worker keys: %w", err)
	}

	cleaned := 0

	for _, key := range keys {
		data, err := r.client.Get(ctx, key).Bytes()
		if err != nil {
			if err == redis.Nil {
				continue // Already expired
			}
			continue
		}

		var worker ports.WorkerInfo
		if err := json.Unmarshal(data, &worker); err != nil {
			continue
		}

		// Check if worker is stale
		if time.Since(worker.LastHeartbeat) > timeout {
			if err := r.client.Del(ctx, key).Err(); err != nil {
				r.logger.Warn("failed to delete stale worker",
					zap.String("worker_id", worker.ID),
					zap.Error(err))
			} else {
				r.logger.Info("cleaned up stale worker",
					zap.String("worker_id", worker.ID),
					zap.Duration("idle_time", time.Since(worker.LastHeartbeat)))
				cleaned++
			}
		}
	}

	return cleaned, nil
}

// Helper methods

func (r *Registry) getWorkerKey(workerID string) string {
	return workerKeyPrefix + workerID
}

func (r *Registry) scanKeys(ctx context.Context, pattern string) ([]string, error) {
	var keys []string
	var cursor uint64

	for {
		var scanKeys []string
		var err error

		scanKeys, cursor, err = r.client.Scan(ctx, cursor, pattern, 100).Result()
		if err != nil {
			return nil, err
		}

		keys = append(keys, scanKeys...)

		if cursor == 0 {
			break
		}
	}

	return keys, nil
}

func (r *Registry) matchesFilter(worker ports.WorkerInfo, filter ports.WorkerFilter, isHealthy bool) bool {
	// Filter by type
	if len(filter.Types) > 0 {
		typeMatch := false
		for _, t := range filter.Types {
			if worker.Type == t {
				typeMatch = true
				break
			}
		}
		if !typeMatch {
			return false
		}
	}

	// Filter by status
	if len(filter.Statuses) > 0 {
		statusMatch := false
		for _, s := range filter.Statuses {
			if worker.Status == s {
				statusMatch = true
				break
			}
		}
		if !statusMatch {
			return false
		}
	}

	// Filter by health
	if filter.HealthyOnly && !isHealthy {
		return false
	}

	return true
}

func (r *Registry) inferWorkerType(workerID string) ports.WorkerType {
	// Infer worker type from worker ID
	if strings.Contains(workerID, "executor") {
		return ports.WorkerTypeExecutor
	}
	if strings.Contains(workerID, "router") {
		return ports.WorkerTypeRouter
	}
	// Default to executor if can't determine
	return ports.WorkerTypeExecutor
}

func (r *Registry) getPendingTasksForWorker(ctx context.Context, workerID string, workerType ports.WorkerType) (int, error) {
	// Determine stream and consumer group based on worker type
	var streamKey, consumerGroup string
	switch workerType {
	case ports.WorkerTypeExecutor:
		streamKey = executorStreamKey
		consumerGroup = executorConsumerGroup
	case ports.WorkerTypeRouter:
		streamKey = routerStreamKey
		consumerGroup = routerConsumerGroup
	default:
		return 0, nil
	}

	// Get consumer info using XINFO CONSUMERS
	consumers, err := r.client.XInfoConsumers(ctx, streamKey, consumerGroup).Result()
	if err != nil {
		// Stream or consumer group might not exist yet
		if strings.Contains(err.Error(), "NOGROUP") {
			return 0, nil
		}
		return 0, err
	}

	// Find this worker in the consumers list
	for _, consumer := range consumers {
		if consumer.Name == workerID {
			// Return pending count
			return int(consumer.Pending), nil
		}
	}

	return 0, nil
}
