package vault

import (
	"fmt"
	"time"

	vaultapi "github.com/hashicorp/vault/api"
)

// SecretInfo holds metadata about a Vault secret lease.
type SecretInfo struct {
	Path      string
	LeaseTTL  time.Duration
	ExpiresAt time.Time
}

// Client wraps the Vault API client.
type Client struct {
	vc *vaultapi.Client
}

// NewClient creates a new Vault client using the provided address and token.
func NewClient(address, token string) (*Client, error) {
	cfg := vaultapi.DefaultConfig()
	cfg.Address = address

	vc, err := vaultapi.NewClient(cfg)
	if err != nil {
		return nil, fmt.Errorf("creating vault client: %w", err)
	}

	vc.SetToken(token)

	return &Client{vc: vc}, nil
}

// GetSecretInfo retrieves lease TTL information for a secret at the given path.
func (c *Client) GetSecretInfo(path string) (*SecretInfo, error) {
	secret, err := c.vc.Logical().Read(path)
	if err != nil {
		return nil, fmt.Errorf("reading secret at %q: %w", path, err)
	}

	if secret == nil {
		return nil, fmt.Errorf("secret not found at path %q", path)
	}

	ttl := time.Duration(secret.LeaseDuration) * time.Second

	return &SecretInfo{
		Path:      path,
		LeaseTTL:  ttl,
		ExpiresAt: time.Now().Add(ttl),
	}, nil
}

// GetSecretsInfo retrieves lease information for multiple secret paths.
func (c *Client) GetSecretsInfo(paths []string) ([]*SecretInfo, []error) {
	var infos []*SecretInfo
	var errs []error

	for _, p := range paths {
		info, err := c.GetSecretInfo(p)
		if err != nil {
			errs = append(errs, err)
			continue
		}
		infos = append(infos, info)
	}

	return infos, errs
}
