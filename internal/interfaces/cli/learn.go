package cli

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/bingaigc/gp/internal/application/ml"
	"github.com/bingaigc/gp/internal/infrastructure/storage"
	"github.com/spf13/cobra"
)

var learnCmd = &cobra.Command{
	Use:   "learn",
	Short: "持续学习优化模型",
	Long:  `通过比较预测值和实际值持续优化模型`,
	Run:   runLearn,
}

var (
	learnStockCode  string
	learnIndicator  string
	learnAnalyze    bool
	learnContinuous bool
	learnInterval   int
)

func init() {
	rootCmd.AddCommand(learnCmd)
	
	learnCmd.Flags().StringVar(&learnStockCode, "stock", "600519", "股票代码")
	learnCmd.Flags().StringVar(&learnIndicator, "indicator", "MA5", "指标类型")
	learnCmd.Flags().BoolVar(&learnAnalyze, "analyze", false, "是否进行详细分析")
	learnCmd.Flags().BoolVar(&learnContinuous, "continuous", false, "是否持续学习")
	learnCmd.Flags().IntVar(&learnInterval, "interval", 3600, "持续学习间隔（秒）")
}

func runLearn(cmd *cobra.Command, args []string) {
	fmt.Println("╔═══════════════════════════════════════════════════════════╗")
	fmt.Println("║   机器学习引擎 v1.0.0                                      ║")
	fmt.Println("╚═══════════════════════════════════════════════════════════╝")
	fmt.Println()
	
	fmt.Printf("🎓 学习配置:\n")
	fmt.Printf("   股票代码: %s\n", learnStockCode)
	fmt.Printf("   指标类型: %s\n", learnIndicator)
	fmt.Printf("   详细分析: %v\n", learnAnalyze)
	fmt.Printf("   持续学习: %v\n", learnContinuous)
	if learnContinuous {
		fmt.Printf("   学习间隔: %d秒\n", learnInterval)
	}
	fmt.Println()
	
	// Initialize components
	histStorage := storage.NewHistoricalStorage()
	modelStorage := storage.NewModelStorage()
	
	// Setup
	collector := ml.NewHistoricalCollector(histStorage)
	predictor := ml.NewIndicatorPredictor(modelStorage, histStorage)
	learner := ml.NewLearningEngine(modelStorage, histStorage, predictor)
	
	ctx := context.Background()
	
	// Collect initial data
	fmt.Println("📊 收集历史数据...")
	_, err := collector.Collect(ctx, learnStockCode, 60)
	if err != nil {
		log.Fatalf("❌ 收集失败: %v", err)
	}
	
	if learnContinuous {
		fmt.Printf("\n🔄 开始持续学习 (每%d秒一次)...\n", learnInterval)
		fmt.Println("按 Ctrl+C 停止")
		
		interval := time.Duration(learnInterval) * time.Second
		learner.ContinuousLearning(ctx, learnStockCode, interval)
	} else {
		// Single learning cycle
		fmt.Println("🎓 执行学习周期...")
		if err := learner.Learn(ctx, learnStockCode, learnIndicator); err != nil {
			log.Fatalf("❌ 学习失败: %v", err)
		}
		
		// Analyze convergence if requested
		if learnAnalyze {
			fmt.Println("\n📊 收敛分析:")
			analysis, err := learner.AnalyzeConvergence(learnStockCode, learnIndicator)
			if err != nil {
				log.Printf("❌ 分析失败: %v", err)
			} else {
				fmt.Printf("   初始MAPE: %.2f%%\n", analysis.InitialMAPE)
				fmt.Printf("   当前MAPE: %.2f%%\n", analysis.CurrentMAPE)
				fmt.Printf("   改进幅度: %.2f%%\n", analysis.Improvement)
				fmt.Printf("   收敛率: %.4f%%/次\n", analysis.ConvergenceRate)
				fmt.Printf("   是否收敛: %v\n", analysis.IsConverged)
				if !analysis.IsConverged {
					fmt.Printf("   预计迭代次数: %d次\n", analysis.IterationsToTarget)
				}
			}
		}
		
		fmt.Println("\n✅ 学习完成！")
	}
}
