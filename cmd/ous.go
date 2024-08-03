package cmd

import (
	"aws-utils/pkg/lz"

	"github.com/spf13/cobra"
)

var (
	ousCmd = &cobra.Command{
		Use:     "ous",
		Aliases: []string{"ou"},
		Short:   "AWS Organizations OU commands",
	}

	ousFiCmd = &cobra.Command{
		Use:     "find",
		Aliases: []string{"fi"},
		Short:   "Find OUs by name as an argument (case sensitive)",
		Args:    cobra.MatchAll(cobra.ExactArgs(1), cobra.ArbitraryArgs),
		Run: func(cmd *cobra.Command, args []string) {
			var name string = args[0]

			checkDebug()
			lz.FindOUs(ctx, newCfg(), name)
		},
	}

	ousLsCmd = &cobra.Command{
		Use:     "list",
		Aliases: []string{"ls"},
		Short:   "Landing Zone OUs list",
		Run: func(cmd *cobra.Command, args []string) {
			checkDebug()
			lz.ListOUs(ctx, newCfg())
		},
	}

	ousDescCmd = &cobra.Command{
		Use:     "describe",
		Aliases: []string{"desc", "de"},
		Short:   "Describe Landing Zone OU",
		Args:    cobra.MatchAll(cobra.ExactArgs(1), cobra.ArbitraryArgs),
		Run: func(cmd *cobra.Command, args []string) {
			var Id string = args[0]

			checkDebug()
			lz.DescribeOU(ctx, newCfg(), Id)
		},
	}
)

func init() {
	rootCmd.AddCommand(ousCmd)
	ousCmd.AddCommand(ousFiCmd)
	ousCmd.AddCommand(ousLsCmd)
	ousCmd.AddCommand(ousDescCmd)
}
