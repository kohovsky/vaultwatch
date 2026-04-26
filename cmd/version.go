package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
)

// Build-time variables injected via -ldflags.
var (
	Version   = "dev"
	Commit    = "none"
	BuildDate = "unknown"
)

// versionCmd prints the current build version, commit hash, and build date.
var versionCmd = &cobra.Command{
	Use:   "version",
	Short: "Print the version information for vaultwatch",
	Long: `Displays the semantic version, git commit hash, and build date
for this vaultwatch binary. These values are embedded at build time
using -ldflags.`,
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Printf("vaultwatch %s\n", Version)
		fmt.Printf("  commit:     %s\n", Commit)
		fmt.Printf("  build date: %s\n", BuildDate)
	},
}

func init() {
	rootCmd.AddCommand(versionCmd)
}
