package retrier

import (
	"log/slog"
	"time"
)

func RunWithRetryGetOperationStatus(logger *slog.Logger, retries int, fn func(string, time.Duration) (string, int, error), name string, timeout time.Duration) (out string, code int, err error) {
	for i := 0; i < retries; i++ {
		out, code, err = fn(name, timeout)
		if err == nil {
			return out, code, nil
		}
		logger.Warn("Retrying...", "retries", retries, "attempt", i+1, "error", err)
	}
	return
}

func RunWithRetryOperation(logger *slog.Logger, retries int, fn func(string) ([]byte, int, error), arg string) (out []byte, code int, err error) {
	for i := 0; i < retries; i++ {
		out, code, err = fn(arg)
		if err == nil {
			return out, code, nil
		}
		logger.Warn("Retrying...", "retries", retries, "attempt", i+1, "error", err)
	}
	return
}

func RunWithRetryError(logger *slog.Logger, retries int, fn func(string) error, arg string) (err error) {
	for i := 0; i < retries; i++ {
		err = fn(arg)
		if err == nil {
			return nil
		}
		logger.Warn("Retrying...", "retries", retries, "attempt", i+1, "error", err)
	}
	return
}
