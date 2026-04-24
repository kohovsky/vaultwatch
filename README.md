# vaultwatch

A CLI tool that monitors HashiCorp Vault secret expiration and sends configurable alerts before leases expire.

---

## Installation

```bash
go install github.com/yourusername/vaultwatch@latest
```

Or build from source:

```bash
git clone https://github.com/yourusername/vaultwatch.git
cd vaultwatch && go build -o vaultwatch .
```

---

## Usage

Set your Vault address and token, then run vaultwatch with a config file:

```bash
export VAULT_ADDR="https://vault.example.com"
export VAULT_TOKEN="s.yourtoken"

vaultwatch --config config.yaml
```

**Example `config.yaml`:**

```yaml
alert_threshold: 72h
notify:
  slack:
    webhook_url: "https://hooks.slack.com/services/..."
secrets:
  - path: secret/data/my-app/db-credentials
  - path: aws/creds/my-role
```

vaultwatch will poll the specified secret paths and send alerts when a lease is within the configured threshold of expiring.

### Flags

| Flag | Description | Default |
|------|-------------|---------|
| `--config` | Path to config file | `config.yaml` |
| `--interval` | Poll interval | `10m` |
| `--dry-run` | Print alerts without sending | `false` |

---

## License

MIT © 2024 yourusername