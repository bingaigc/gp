package cli

import (
	"context"
	"fmt"
	"time"

	"github.com/bingaigc/gp/internal/config"
	"github.com/bingaigc/gp/internal/infrastructure/gateway/eastmoney"
	"github.com/bingaigc/gp/internal/infrastructure/gateway/kimi"
	"github.com/spf13/cobra"
)

var healthCmd = &cobra.Command{
	Use:   "health",
	Short: "检查系统健康状态",
	Long:  "检查配置、数据源和AI服务的连接状态",
	Run:   runHealth,
}

func init() {
	rootCmd.AddCommand(healthCmd)
}

func runHealth(cmd *cobra.Command, args []string) {
	fmt.Println("╔═══════════════════════════════════════════════════════════╗")
	fmt.Println("║   CAMP-I System Health Check                             ║")
	fmt.Println("╚═══════════════════════════════════════════════════════════╝")
	fmt.Println()

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	// 1. 检查配置
	fmt.Print("🔧 检查配置文件... ")
	cfg, err := config.Load(GetConfigFile())
	if err != nil {
		fmt.Printf("❌ 失败: %v\n", err)
		return
	}
	fmt.Println("✅ 成功")

	// 2. 检查数据源
	fmt.Print("📡 检查数据源 (EastMoney)... ")
	dataClient := eastmoney.NewClient("", cfg.DataSource.Timeout)
	if err := dataClient.HealthCheck(ctx); err != nil {
		fmt.Printf("❌ 失败: %v\n", err)
	} else {
		fmt.Println("✅ 成功")
	}

	// 3. 检查AI服务
	if cfg.AI.APIKey != "" {
		fmt.Print("🤖 检查AI服务 (Kimi)... ")
		aiClient := kimi.NewClient(cfg.AI.BaseURL, cfg.AI.APIKey, cfg.AI.Model, cfg.AI.Timeout)
		if err := aiClient.HealthCheck(ctx); err != nil {
			fmt.Printf("❌ 失败: %v\n", err)
		} else {
			fmt.Println("✅ 成功")
		}
	} else {
		fmt.Println("⚠️  AI服务未配置 (需要设置 ALPHA_KIMI_API_KEY)")
	}

	// 4. 显示配置摘要
	fmt.Println()
	fmt.Println("📊 配置摘要:")
	fmt.Printf("  环境: %s\n", cfg.System.Environment)
	fmt.Printf("  日志级别: %s\n", cfg.System.LogLevel)
	fmt.Printf("  并发Workers: %d\n", cfg.Pipeline.Workers)
	fmt.Printf("  扫描限制: %d stocks\n", cfg.Pipeline.Limit)
	fmt.Printf("  日最大亏损: ¥%.0f\n", cfg.Risk.DailyMaxLoss)
	fmt.Printf("  日最大信号: %d\n", cfg.Risk.MaxSignalsPerDay)

	fmt.Println()
	fmt.Println("✅ 健康检查完成")
}
