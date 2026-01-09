package cli

import (
	"context"
	"fmt"
	"log"
	"log/slog"
	"os"
	"time"

	"github.com/bingaigc/gp/internal/application/pipeline"
	"github.com/bingaigc/gp/internal/application/risk"
	"github.com/bingaigc/gp/internal/application/scorer"
	"github.com/bingaigc/gp/internal/config"
	"github.com/bingaigc/gp/internal/infrastructure/gateway/eastmoney"
	"github.com/bingaigc/gp/internal/infrastructure/gateway/kimi"
	"github.com/bingaigc/gp/internal/interfaces/sink"
	"github.com/spf13/cobra"
)

var (
	useRealData bool
	useAI       bool
)

var scanRealCmd = &cobra.Command{
	Use:   "scan-real",
	Short: "使用真实数据源和AI分析扫描市场",
	Long:  "连接东方财富API获取实时数据，使用Kimi AI进行深度分析，生成交易信号",
	Run:   runScanReal,
}

func init() {
	rootCmd.AddCommand(scanRealCmd)
	scanRealCmd.Flags().BoolVar(&useRealData, "real-data", true, "使用真实数据源")
	scanRealCmd.Flags().BoolVar(&useAI, "ai", true, "使用AI分析")
}

func runScanReal(cmd *cobra.Command, args []string) {
	ctx := context.Background()

	// 加载配置
	cfg, err := config.Load(GetConfigFile())
	if err != nil {
		log.Fatalf("加载配置失败: %v", err)
	}

	if err := cfg.Validate(); err != nil {
		log.Fatalf("配置验证失败: %v", err)
	}

	// 创建logger
	logger := slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{
		Level: slog.LevelInfo,
	}))

	fmt.Println("╔═══════════════════════════════════════════════════════════╗")
	fmt.Println("║   CAMP-I Distributed Alpha Signal Detector v2.0.0        ║")
	fmt.Println("║   企业级分布式Alpha信号探测系统 - 实时数据+AI分析         ║")
	fmt.Println("╚═══════════════════════════════════════════════════════════╝")
	fmt.Println()
	fmt.Printf("📋 配置: %s (%s)\n", cfg.System.Name, cfg.System.Environment)
	fmt.Printf("🔧 数据源: %s (%s)\n", cfg.DataSource.Primary, getDataSourceStatus(useRealData))
	fmt.Printf("🤖 AI模型: %s (%s) - %s\n", cfg.AI.Provider, cfg.AI.Model, getAIStatus(useAI))
	fmt.Printf("⚙️  并发数: %d workers\n", cfg.Pipeline.Workers)
	fmt.Printf("📊 扫描数量: %d stocks\n", cfg.Pipeline.Limit)
	fmt.Println()

	// 初始化组件
	dataFetcher := eastmoney.NewClient("", cfg.DataSource.Timeout)
	scorer := scorer.NewCAMPIScorer()
	analyzer := kimi.NewClient(cfg.AI.BaseURL, cfg.AI.APIKey, cfg.AI.Model, cfg.AI.Timeout)
	riskCtrl := risk.NewController(
		cfg.Risk.DailyMaxLoss,
		cfg.Risk.MaxSignalsPerDay,
		cfg.Risk.MaxConsecutiveLosses,
	)
	signalSink := sink.NewConsoleSink(true)

	// 创建管道编排器
	orchestrator := pipeline.NewOrchestrator(
		dataFetcher,
		scorer,
		analyzer,
		riskCtrl,
		signalSink,
		cfg.Pipeline.Workers,
		logger,
	)

	fmt.Println("🚀 开始扫描市场...")
	fmt.Println()

	startTime := time.Now()

	// 运行管道
	if err := orchestrator.Run(ctx, cfg.Pipeline.Limit); err != nil {
		log.Fatalf("管道执行失败: %v", err)
	}

	elapsed := time.Since(startTime)

	fmt.Println()
	fmt.Printf("✅ 扫描完成！耗时: %.2f秒\n", elapsed.Seconds())
	fmt.Println()

	// 显示风控统计
	stats := riskCtrl.GetStats()
	fmt.Println("📊 风控统计:")
	fmt.Printf("  日内信号数: %d/%d\n", stats["daily_signal_count"], cfg.Risk.MaxSignalsPerDay)
	fmt.Printf("  日内亏损: ¥%.2f/¥%.2f\n", stats["daily_loss"], cfg.Risk.DailyMaxLoss)
	fmt.Printf("  连续亏损: %d/%d\n", stats["consecutive_losses"], cfg.Risk.MaxConsecutiveLosses)

	// 显示AI成本
	if useAI {
		totalCost := analyzer.GetTotalCost()
		fmt.Println()
		fmt.Println("💰 AI成本统计:")
		fmt.Printf("  本次总成本: ¥%.2f\n", totalCost)
		fmt.Printf("  平均成本/信号: ¥%.2f\n", totalCost/float64(stats["daily_signal_count"].(int)))
	}
}

func getDataSourceStatus(useReal bool) string {
	if useReal {
		return "✅ 实时数据"
	}
	return "⚠️ 演示数据"
}

func getAIStatus(useAI bool) string {
	if useAI {
		return "✅ AI启用"
	}
	return "⚠️ AI禁用"
}
