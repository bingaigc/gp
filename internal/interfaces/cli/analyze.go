package cli

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"os"

	"github.com/bingaigc/gp/internal/application/ml"
	"github.com/bingaigc/gp/internal/domain/entity"
	"github.com/bingaigc/gp/internal/infrastructure/storage"
	"github.com/spf13/cobra"
)

var analyzeCmd = &cobra.Command{
	Use:   "analyze",
	Short: "分析预测误差",
	Long:  `深度分析预测误差的原因和模式`,
	Run:   runAnalyze,
}

var (
	analyzeStockCode  string
	analyzeIndicator  string
	analyzeDetailed   bool
	analyzeOutputFile string
)

func init() {
	rootCmd.AddCommand(analyzeCmd)

	analyzeCmd.Flags().StringVar(&analyzeStockCode, "stock", "600519", "股票代码")
	analyzeCmd.Flags().StringVar(&analyzeIndicator, "indicator", "MA5", "指标类型")
	analyzeCmd.Flags().BoolVar(&analyzeDetailed, "detailed", false, "输出详细分析")
	analyzeCmd.Flags().StringVar(&analyzeOutputFile, "output", "", "输出文件（JSON格式）")
}

func runAnalyze(cmd *cobra.Command, args []string) {
	fmt.Println("╔═══════════════════════════════════════════════════════════╗")
	fmt.Println("║   误差分析器 v1.0.0                                        ║")
	fmt.Println("╚═══════════════════════════════════════════════════════════╝")
	fmt.Println()

	fmt.Printf("🔍 分析配置:\n")
	fmt.Printf("   股票代码: %s\n", analyzeStockCode)
	fmt.Printf("   指标类型: %s\n", analyzeIndicator)
	fmt.Printf("   详细模式: %v\n", analyzeDetailed)
	fmt.Println()

	// Initialize components
	histStorage := storage.NewHistoricalStorage()
	modelStorage := storage.NewModelStorage()

	// Setup and generate sample predictions
	collector := ml.NewHistoricalCollector(histStorage)
	predictor := ml.NewIndicatorPredictor(modelStorage, histStorage)
	learner := ml.NewLearningEngine(modelStorage, histStorage, predictor)
	analyzer := ml.NewIndicatorAnalyzer()

	ctx := context.Background()

	// Collect and generate predictions
	fmt.Println("📊 准备数据...")
	_, err := collector.Collect(ctx, analyzeStockCode, 30)
	if err != nil {
		log.Fatalf("❌ 数据收集失败: %v", err)
	}

	// Run some learning cycles to generate data
	fmt.Println("🎓 执行学习周期...")
	for i := 0; i < 10; i++ {
		if err := learner.Learn(ctx, analyzeStockCode, analyzeIndicator); err != nil {
			log.Printf("❌ 学习失败 #%d: %v", i+1, err)
		}
	}

	// Generate sample indicators for analysis
	var indicators []*entity.TechnicalIndicator
	for i := 0; i < 20; i++ {
		ind, err := predictor.Predict(ctx, analyzeStockCode, analyzeIndicator)
		if err != nil {
			continue
		}
		// Simulate actual value
		actualValue := ind.PredictedValue * (1 + float64(i%5-2)*0.01)
		ind.UpdateActual(actualValue)
		indicators = append(indicators, ind)
	}

	// Perform analysis
	fmt.Println("\n🔬 执行误差分析...")
	analysis, err := analyzer.Analyze(ctx, indicators)
	if err != nil {
		log.Fatalf("❌ 分析失败: %v", err)
	}

	// Display results
	fmt.Println("\n📈 误差统计:")
	fmt.Printf("   MAE: %.4f\n", analysis.ErrorMetrics.MAE)
	fmt.Printf("   RMSE: %.4f\n", analysis.ErrorMetrics.RMSE)
	fmt.Printf("   MAPE: %.2f%%\n", analysis.ErrorMetrics.MAPE)
	fmt.Printf("   最大误差: %.4f\n", analysis.ErrorMetrics.MaxError)
	fmt.Printf("   最小误差: %.4f\n", analysis.ErrorMetrics.MinError)
	fmt.Printf("   标准差: %.4f\n", analysis.ErrorMetrics.StdDev)

	if analyzeDetailed {
		fmt.Println("\n🔍 根本原因:")
		for i, cause := range analysis.RootCauses {
			fmt.Printf("   %d. [%s] %s\n", i+1, cause.Type, cause.Description)
			fmt.Printf("      影响程度: %.1f%% | 置信度: %.1f%%\n",
				cause.Impact*100, cause.Confidence*100)
		}

		fmt.Println("\n📊 错误模式:")
		for i, pattern := range analysis.Patterns {
			fmt.Printf("   %d. %s\n", i+1, pattern.Name)
			fmt.Printf("      %s\n", pattern.Description)
			fmt.Printf("      频率: %.1f%% | 严重程度: %.1f%%\n",
				pattern.Frequency*100, pattern.Severity*100)
		}
	}

	fmt.Println("\n💡 优化建议:")
	for i, rec := range analysis.Recommendations {
		fmt.Printf("   %d. %s\n", i+1, rec)
	}

	// Export to file if requested
	if analyzeOutputFile != "" {
		data, err := json.MarshalIndent(analysis, "", "  ")
		if err != nil {
			log.Printf("❌ JSON序列化失败: %v", err)
		} else {
			if err := os.WriteFile(analyzeOutputFile, data, 0644); err != nil {
				log.Printf("❌ 写入文件失败: %v", err)
			} else {
				fmt.Printf("\n💾 分析报告已保存: %s\n", analyzeOutputFile)
			}
		}
	}

	fmt.Println("\n✅ 分析完成！")
}
