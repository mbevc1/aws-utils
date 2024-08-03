package cmd

import (
	"aws-utils/pkg/lz"
	"aws-utils/pkg/util"

	"github.com/spf13/cobra"
)

var (
	lzCmd = &cobra.Command{
		Use:     "landingzone",
		Aliases: []string{"lz"},
		Short:   "Landing Zone (LZ) commands",
	}

	lzDescCmd = &cobra.Command{
		Use:     "describe",
		Aliases: []string{"de", "desc"},
		Short:   "Landing Zone describe - get information about LZ",
		Run: func(cmd *cobra.Command, args []string) {
			checkDebug()
			lz.ListLandingZone(ctx, newCfg())
		},
	}

	lzProdsCmd = &cobra.Command{
		Use:     "products",
		Aliases: []string{"pr", "prods", "pl"},
		Short:   "List CT account factory deployed products",
		Run: func(cmd *cobra.Command, args []string) {
			checkDebug()

			// List provisioned products
			lz.ListProvisionedProducts(ctx, newCfg())
		},
	}

	lzVendCmd = &cobra.Command{
		Use:     "vend-account",
		Aliases: []string{"va", "ve", "vend"},
		Short:   "Vend CT account factory product",
		Run: func(cmd *cobra.Command, args []string) {
			checkDebug()

			// Create a provisioned product
			lz.VendAccountProduct(ctx, newCfg(), ap)
		},
	}

	lzTermCmd = &cobra.Command{
		Use:     "terminate",
		Aliases: []string{"te", "term", "tp"},
		Short:   "Terminate CT account factory product",
		Run: func(cmd *cobra.Command, args []string) {
			checkDebug()

			// Terminate provisioned product
			lz.TerminateProvisionedProduct(ctx, newCfg(), ap)
		},
	}

	lzRstCmd = &cobra.Command{
		Use:     "reset",
		Aliases: []string{"re", "res"},
		Short:   "Reset LZ to the parameters specified in the original configuration",
		Run: func(cmd *cobra.Command, args []string) {
			checkDebug()

			// Reset LZ
			lz.ResetLandingZone(ctx, newCfg())
		},
	}

	lzUpdateCmd = &cobra.Command{
		Use:     "update",
		Aliases: []string{"up", "upd"},
		Short:   "Update LZ to latest available version",
		Run: func(cmd *cobra.Command, args []string) {
			checkDebug()

			// Update LZ
			lz.UpdateLandingZone(ctx, newCfg())
		},
	}

	lzLsOpsCmd = &cobra.Command{
		Use:     "list-operations",
		Aliases: []string{"lo", "ls-ops"},
		Short:   "List LZ operations from past 90 days",
		Run: func(cmd *cobra.Command, args []string) {
			checkDebug()
			// List LZ operations
			lz.ListOpsLandingZone(ctx, newCfg())
		},
	}
)

var (
	ap util.AccountParams
)

func init() {
	rootCmd.AddCommand(lzCmd)
	lzCmd.AddCommand(lzDescCmd)
	lzCmd.AddCommand(lzProdsCmd)
	lzCmd.AddCommand(lzVendCmd)
	lzCmd.AddCommand(lzTermCmd)
	lzCmd.AddCommand(lzRstCmd)
	lzCmd.AddCommand(lzUpdateCmd)
	lzCmd.AddCommand(lzLsOpsCmd)

	lzTermCmd.Flags().StringVarP(&ap.ID, "account", "a", "", "Set value for AccountID")
	lzTermCmd.Flags().StringVarP(&ap.Name, "name", "n", "", "Set value for AccountName|ProductName")
	lzTermCmd.MarkFlagRequired("account")
	lzTermCmd.MarkFlagRequired("name")

	lzVendCmd.Flags().StringVarP(&ap.Email, "email", "e", "", "Set value for AccountEmail")
	lzVendCmd.Flags().StringVarP(&ap.Name, "name", "n", "", "Set value for AccountName|ProductName")
	lzVendCmd.Flags().StringVarP(&ap.OU, "ou", "o", "", "Set value for ManagedOrganizationalUnit: OU (ID)")
	lzVendCmd.Flags().StringVarP(&ap.SSOemail, "ssoemail", "", "", "Set value for SSOUserEmail")
	lzVendCmd.Flags().StringVarP(&ap.SSOfirstname, "ssofirstname", "", "AWS Control Tower", "Set value for SSOUserFirstName")
	lzVendCmd.Flags().StringVarP(&ap.SSOlastname, "ssolastname", "", "Admin", "Set value for SSOUserLastName")
	lzVendCmd.MarkFlagRequired("email")
	lzVendCmd.MarkFlagRequired("name")
	lzVendCmd.MarkFlagRequired("ou")
}
