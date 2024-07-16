package main

import (
	"context"
	"fmt"
	"github.com/spf13/cobra"
	"os"
	"tun/internal/client"
	"tun/internal/pkg/log"
	"tun/pkg/version"
)

func main() {
	Execute()
}

var (
	showVersion bool
	token       string
)

func init() {
	rootCmd.PersistentFlags().BoolVarP(&showVersion, "version", "v", false, "show version")
	rootCmd.PersistentFlags().StringVarP(&token, "token", "t", "", "tunnel token")
}

var rootCmd = &cobra.Command{
	Use:   "tunc",
	Short: "tunc is the client of tun (https://github.com/WuYulong/tun)",
	RunE: func(cmd *cobra.Command, args []string) error {
		if showVersion {
			fmt.Println(version.Full())
			return nil
		}

		if token == "" {
			fmt.Println("请输入 token")
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
	log.InitLogger("console", "info", 3, false)
	tc := client.NewClient(token)
	tc.Run(context.Background())
	return nil
}
