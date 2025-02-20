/*
Copyright © 2025 NAME HERE <EMAIL ADDRESS>
*/
package cmd

import (
	"fmt"
	"os"

	"github.com/mproffitt/bmx/pkg/config"
	"github.com/spf13/cobra"
)

// configCmd represents the config command
var configCmd = &cobra.Command{
	Use:   "config",
	Short: "configure " + executable,
	Run: func(cmd *cobra.Command, args []string) {
		m := config.NewConfigModel(tmsConfig)
		if m == nil {
			fmt.Println("Failed to set up config model")
			os.Exit(1)
		}
		run(m)
	},
}

func init() {
	rootCmd.AddCommand(configCmd)
}
