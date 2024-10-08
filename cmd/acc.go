package cmd

import (
	"aws-utils/pkg/lz"

	"github.com/spf13/cobra"
)

var (
	filterFlag string
)

var (
	accCmd = &cobra.Command{
		Use:     "accounts",
		Aliases: []string{"acc"},
		Short:   "AWS accounts commands",
	}

	accFiCmd = &cobra.Command{
		Use:     "find",
		Aliases: []string{"fi"},
		Short:   "Find AWS accounts, you can filter using ID parameter or additional filter argument",
		Args:    cobra.MatchAll(cobra.MaximumNArgs(1), cobra.ArbitraryArgs),
		Run: func(cmd *cobra.Command, args []string) {
			var Id string = ""

			if len(args) == 1 {
				Id = args[0]
			}

			checkDebug()
			lz.FindAccounts(ctx, newCfg(), Id, filterFlag)
		},
	}

	accLsCmd = &cobra.Command{
		Use:     "list",
		Aliases: []string{"ls"},
		Short:   "Organizations accounts list",
		Run: func(cmd *cobra.Command, args []string) {
			checkDebug()
			lz.ListAccounts(ctx, newCfg())
		},
	}
)

func init() {
	rootCmd.AddCommand(accCmd)
	accCmd.AddCommand(accFiCmd)
	accCmd.AddCommand(accLsCmd)

	accFiCmd.Flags().StringVarP(&filterFlag, "filter", "f", "", "A search filter for account names")
}
