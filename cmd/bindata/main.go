package main

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
)

const (
	version = "0.0.1"
)

func main() {
	rootCmd := &cobra.Command{
		Use:          "bindata",
		Short:        "bind data into go source code",
		SilenceUsage: true,
	}

	rootCmd.AddCommand(&cobra.Command{
		Use:     "version",
		Short:   "show version",
		Aliases: []string{"ver"},
		RunE: func(cmd *cobra.Command, args []string) error {
			fmt.Println(version)
			return nil
		},
	})

	rootCmd.AddCommand(NewGenerateCommand())

	if err := rootCmd.Execute(); err != nil {
		os.Exit(1)
	}
}
