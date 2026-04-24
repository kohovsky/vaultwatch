package cmd

import (
	"fmt"
	"os"
	"time"

	"github.com/spf13/cobra"

	"github.com/yourorg/vaultwatch/internal/alert"
	"github.com/yourorg/vaultwatch/internal/config"
	"github.com/yourorg/vaultwatch/internal/monitor"
	"github.com/yourorg/vaultwatch/internal/vault"
)

var cfgFile string

var rootCmd = &cobra.Command{
	Use:   "vaultwatch",
	Short: "Monitor HashiCorp Vault secret expiration and send alerts",
	Long: `vaultwatch connects to a Vault instance, checks secret lease
expirations against configured thresholds, and dispatches alerts
via stdout, file, or webhook.`,
	RunE: runWatch,
}

func Execute() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

func init() {
	rootCmd.PersistentFlags().StringVarP(&cfgFile, "config", "c", "vaultwatch.yaml", "path to config file")
}

func runWatch(cmd *cobra.Command, args []string) error {
	cfg, err := config.Load(cfgFile)
	if err != nil {
		return fmt.Errorf("loading config: %w", err)
	}

	client, err := vault.NewClient(cfg.VaultAddress, cfg.VaultToken)
	if err != nil {
		return fmt.Errorf("creating vault client: %w", err)
	}

	notifier, err := alert.NewNotifier(cfg)
	if err != nil {
		return fmt.Errorf("creating notifier: %w", err)
	}

	thresholds, err := cfg.ParsedThresholds()
	if err != nil {
		return fmt.Errorf("parsing thresholds: %w", err)
	}

	secrets, errs := client.GetSecretsInfo(cmd.Context(), cfg.Paths)
	for path, fetchErr := range errs {
		fmt.Fprintf(os.Stderr, "warn: could not fetch %s: %v\n", path, fetchErr)
	}

	statuses := monitor.CheckAll(secrets, thresholds)
	for _, status := range statuses {
		if alertErr := notifier.Notify(status); alertErr != nil {
			fmt.Fprintf(os.Stderr, "alert error for %s: %v\n", status.Path, alertErr)
		}
	}

	fmt.Printf("vaultwatch: checked %d secret(s) at %s\n", len(statuses), time.Now().Format(time.RFC3339))
	return nil
}
