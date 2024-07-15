package main

import (
	"fmt"
	"github.com/spf13/cobra"
	"github.com/wuyunlong/tun/pkg/version"
	"os"
)

func main() {
	Execute()
}

var (
	showVersion bool
)

func init() {
	rootCmd.PersistentFlags().BoolVarP(&showVersion, "version", "v", false, "show version")
}

var rootCmd = &cobra.Command{
	Use:   "tunc",
	Short: "tunc is the client of tun (https://github.com/WuYulong/tun)",
	RunE: func(cmd *cobra.Command, args []string) error {
		if showVersion {
			fmt.Println(version.Full())
			return nil
		}
		if err := runClient(); err != nil {
			fmt.Println(err)
			os.Exit(1)
		}
		return nil
	},
}

func Execute() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}

func runClient() error {
	fmt.Println("run client ...")
	return nil
}
