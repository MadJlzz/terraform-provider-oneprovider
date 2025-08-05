package common

import (
	"context"
	"fmt"
	"time"

	"github.com/hashicorp/terraform-plugin-log/tflog"
)

// RetryConfig defines configuration for retry behavior.
type RetryConfig struct {
	MaxRetries    int
	InitialDelay  time.Duration
	MaxDelay      time.Duration
	BackoffFactor float64
}

// DefaultRetryConfig returns sensible defaults for OneProvider's eventual consistency.
func DefaultRetryConfig() RetryConfig {
	return RetryConfig{
		MaxRetries:    8,
		InitialDelay:  500 * time.Millisecond,
		MaxDelay:      30 * time.Second,
		BackoffFactor: 2.0,
	}
}

// RetryableFunc represents a function that can be retried
// It should return (result, shouldRetry, error)
// - result: the result of the operation (only used if shouldRetry is false and error is nil)
// - shouldRetry: true if the operation should be retried, false if it succeeded or failed permanently
// - error: any error that occurred.
type RetryableFunc[T any] func(ctx context.Context, attempt int) (result T, shouldRetry bool, err error)

// WithRetry executes a function with exponential backoff retry logic
// The function will be retried until it succeeds, fails permanently, or timeout is reached.
func WithRetry[T any](ctx context.Context, config RetryConfig, operation RetryableFunc[T]) (T, error) {
	var zero T
	delay := config.InitialDelay

	for attempt := 0; attempt < config.MaxRetries; attempt++ {
		// Check context cancellation before each attempt
		select {
		case <-ctx.Done():
			return zero, ctx.Err()
		default:
		}

		// Apply delay before retry attempts (not on first attempt)
		if attempt > 0 {
			tflog.Debug(ctx, fmt.Sprintf("Retrying operation after %v (attempt %d/%d)", delay, attempt+1, config.MaxRetries))
			time.Sleep(delay)
		}

		// Execute the operation
		result, shouldRetry, err := operation(ctx, attempt)

		// If operation succeeded, return the result
		if err == nil && !shouldRetry {
			if attempt > 0 {
				tflog.Debug(ctx, fmt.Sprintf("Operation succeeded after %d attempts", attempt+1))
			}
			return result, nil
		}

		// If operation failed permanently (error with shouldRetry=false), return error
		if err != nil && !shouldRetry {
			tflog.Debug(ctx, fmt.Sprintf("Operation failed permanently: %v", err))
			return zero, err
		}

		// Log retry reason
		if err != nil {
			tflog.Debug(ctx, fmt.Sprintf("Operation failed, will retry: %v", err))
		} else {
			tflog.Debug(ctx, "Operation needs retry (validation failed)")
		}

		// Calculate next delay with exponential backoff
		delay = time.Duration(float64(delay) * config.BackoffFactor)
		if delay > config.MaxDelay {
			delay = config.MaxDelay
		}
	}

	return zero, fmt.Errorf("operation failed after %d attempts", config.MaxRetries)
}

// WithRetryUntilValid retries an operation until validation passes or MaxRetries is reached
// It will retry up to config.MaxRetries times with exponential backoff.
func WithRetryUntilValid[T any](
	ctx context.Context,
	config RetryConfig,
	operation func(ctx context.Context) (T, error),
	isValid func(result T) bool,
) (T, error) {
	return WithRetry(ctx, config, func(ctx context.Context, attempt int) (T, bool, error) {
		result, err := operation(ctx)
		if err != nil {
			// Retry on errors (assuming they might be transient)
			return result, true, err
		}

		// Check validation
		if isValid(result) {
			return result, false, nil
		}

		// Validation failed, retry
		var zero T
		return zero, true, nil
	})
}
