package cmd

import (
	"context"
	"fmt"
	"log/slog"
	"os"

	"github.com/spf13/cobra"
)

// Name & Version of the app
var (
	Name    = "aws-utils"
	Version string
)

var (
	regionFlag string
	debugFlag  bool
	ctx        context.Context
)

var rootCmd = &cobra.Command{
	Use:           fmt.Sprintf("%s", Name),
	Version:       Version,
	SilenceUsage:  true,
	SilenceErrors: true,
	Short:         fmt.Sprintf("%s is a simple CLI to manage AWS Landing Zone", Name),
	Long: `     ___        ______        _   _ _   _ _     
    / \ \      / / ___|      | | | | |_(_) |___ 
   / _ \ \ /\ / /\___ \ _____| | | | __| | / __|
  / ___ \ V  V /  ___) |_____| |_| | |_| | \__ \
 /_/   \_\_/\_/  |____/       \___/ \__|_|_|___/
                                                
 A simple CLI to better manage AWS environment.`,
}

func init() {
	rootCmd.PersistentFlags().StringVarP(&regionFlag, "region", "r", "", "A global region setting, defaults to profile settings")
	rootCmd.PersistentFlags().BoolVarP(&debugFlag, "debug", "d", false, "Enable debugging logging")

	// start with empty Context
	ctx = context.Background()
}

func Execute(version string) {
	Version = version
	// fmt.Println()
	// defer fmt.Println()

	rootCmd.Version = version

	if err := rootCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}

	if debugFlag {
		slog.SetLogLoggerLevel(slog.LevelDebug)
	}
	slog.Debug(fmt.Sprintf("App version: %s", Version))
}
