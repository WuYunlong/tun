package main

import (
	"context"
	"fmt"
	"github.com/spf13/cobra"
	"github.com/wuyunlong/tun/pkg/version"
	"os"
	"tun/internal/config"
	"tun/internal/pkg/common"
	"tun/internal/pkg/log"
	"tun/internal/server"
)

func main() {
	Execute()
}

var (
	showVersion bool
	configFile  string
)

func init() {
	rootCmd.PersistentFlags().BoolVarP(&showVersion, "version", "v", false, "show version")
	rootCmd.PersistentFlags().StringVarP(&configFile, "config", "c", "conf/tuns.yaml", "config file path")
}

var rootCmd = &cobra.Command{
	Use:   "tuns",
	Short: "tuns is the server of tun (https://github.com/WuYulong/tun)",
	RunE: func(cmd *cobra.Command, args []string) error {
		if showVersion {
			fmt.Println(version.Full())
			return nil
		}

		if configFile == "" {
			configFile = "conf/tuns.yaml"
		}

		if !common.FileExists(configFile) {
			fmt.Println("config file can not fount")
			return nil
		}

		if err := runServer(); err != nil {
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

func runServer() error {
	// 初始化配置
	cfg := config.LoadServerConfig(configFile)
	fmt.Println(cfg.Log.Level)
	// 初始化日志
	log.InitLogger(cfg.Log.To, cfg.Log.Level, cfg.Log.MaxDays, cfg.Log.DisableLogColor)
	// 初始化服务
	ts, err := server.NewServer(cfg)
	if err != nil {
		return err
	}
	// 运行服务
	ts.Run(context.Background())
	return nil
}
