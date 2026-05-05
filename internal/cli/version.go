package cli

import (
	"fmt"

	"github.com/spf13/cobra"

	"github.com/marstid/nuc/pkg/version"
)

func newVersionCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "version",
		Short: "Print the version of nuc",
		Long:  "Print the version, commit, and build date of the nuc binary.",
		Args:  cobra.NoArgs,
		Run: func(_ *cobra.Command, _ []string) {
			fmt.Printf("nuc version %s\n", version.String())
		},
	}
}
