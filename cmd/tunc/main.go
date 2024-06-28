package main

import (
	"fmt"
	"github.com/spf13/cobra"
	"os"
)

var (
	showVersion bool
)

func init() {
	rootCmd.PersistentFlags().BoolVarP(&showVersion, "version", "v", false, "version of tunc")
}

var rootCmd = &cobra.Command{
	Use:   "tunc",
	Short: "tunc is client of tun ...",
	RunE: func(cmd *cobra.Command, args []string) error {
		if showVersion {
			fmt.Println("1.0.0")
			return nil
		}
		if err := runClient(); err != nil {
			fmt.Println("run client error")
			os.Exit(1)
		}
		return nil
	},
}

func runClient() error {
	fmt.Println("Hello tunc ...")
	return nil
}

func main() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}
