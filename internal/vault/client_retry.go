package vault

import (
	"fmt"

	"github.com/czeslavo/vaultwatch/internal/monitor"
)

// RetryingClient wraps a Client and retries transient Vault API errors.
type RetryingClient struct {
	inner   *Client
	retrier *monitor.Retrier
}

// NewRetryingClient creates a RetryingClient with the given retry config.
func NewRetryingClient(inner *Client, cfg monitor.RetryConfig) *RetryingClient {
	return &RetryingClient{
		inner:   inner,
		retrier: monitor.NewRetrier(cfg),
	}
}

// GetSecretInfo fetches secret metadata, retrying on transient errors.
func (rc *RetryingClient) GetSecretInfo(path string) (*SecretInfo, error) {
	var info *SecretInfo
	err := rc.retrier.Do(func() error {
		var fetchErr error
		info, fetchErr = rc.inner.GetSecretInfo(path)
		if fetchErr != nil {
			if monitor.IsNonRetryable(fetchErr) {
				return fetchErr
			}
			return fetchErr
		}
		return nil
	})
	if err != nil {
		return nil, fmt.Errorf("vault get %q (with retries): %w", path, err)
	}
	return info, nil
}

// GetSecretsInfo fetches multiple secrets, retrying each individually.
func (rc *RetryingClient) GetSecretsInfo(paths []string) ([]*SecretInfo, []error) {
	results := make([]*SecretInfo, 0, len(paths))
	var errs []error

	for _, path := range paths {
		info, err := rc.GetSecretInfo(path)
		if err != nil {
			errs = append(errs, err)
			continue
		}
		results = append(results, info)
	}

	return results, errs
}
