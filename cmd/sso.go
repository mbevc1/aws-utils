package cmd

import (
	"aws-utils/pkg/lz"

	"github.com/spf13/cobra"
)

var (
	ssoCmd = &cobra.Command{
		Use:     "ssoadmin",
		Aliases: []string{"sso", "iamic"},
		Short:   "AWS IAM IC (SSOadmin) commands",
	}

	ssoLsCmd = &cobra.Command{
		Use:     "list",
		Aliases: []string{"ls"},
		Short:   "IAM IC instances list",
		Run: func(cmd *cobra.Command, args []string) {
			checkDebug()
			lz.ListSSO(ctx, newCfg())
		},
	}

	ssoPsCmd = &cobra.Command{
		Use:     "permission-sets",
		Aliases: []string{"ps"},
		Short:   "IAM IC PermissionSet list",
		Run: func(cmd *cobra.Command, args []string) {
			checkDebug()
			lz.ListPermissionSets(ctx, newCfg())
		},
	}
)

func init() {
	rootCmd.AddCommand(ssoCmd)
	ssoCmd.AddCommand(ssoLsCmd)
	ssoCmd.AddCommand(ssoPsCmd)
}
