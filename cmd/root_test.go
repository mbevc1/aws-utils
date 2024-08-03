package cmd

import (
	"fmt"
	"testing"
)

func TestRootCmd(t *testing.T) {
	// Test that the root command is set up correctly
	if rootCmd.Use != Name {
		t.Errorf("Expected Use to be %s, got %s", Name, rootCmd.Use)
	}
	if rootCmd.Short != fmt.Sprintf("%s is a simple CLI to manage AWS Landing Zone", Name) {
		t.Errorf("Expected Short to be %s, got %s", fmt.Sprintf("%s is a simple CLI to manage AWS Landing Zone", Name), rootCmd.Short)
	}
}
