package cli

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
)

var (
	configFile string
	rootCmd    = &cobra.Command{
		Use:   "alpha-detector",
		Short: "CAMP-I Distributed Alpha Signal Detector",
		Long: `CAMP-I (Capital + Action + Margin + Plate + Institution) 
企业级分布式Alpha信号探测系统
构建持续盈利的量化交易机器`,
	}
)

// Execute 执行CLI
func Execute() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

func init() {
	rootCmd.PersistentFlags().StringVarP(&configFile, "config", "c", "configs/config.yaml", "配置文件路径")
}

// GetConfigFile 获取配置文件路径
func GetConfigFile() string {
	return configFile
}
