package cli

import (
	"context"
	"fmt"
	"log"
	"strings"

	"github.com/bingaigc/gp/internal/application/ml"
	"github.com/bingaigc/gp/internal/infrastructure/storage"
	"github.com/spf13/cobra"
)

var predictCmd = &cobra.Command{
	Use:   "predict",
	Short: "预测技术指标",
	Long:  `使用机器学习模型预测未来技术指标`,
	Run:   runPredict,
}

var (
	predictStockCodes   string
	predictIndicator    string
	predictModel        string
)

func init() {
	rootCmd.AddCommand(predictCmd)
	
	predictCmd.Flags().StringVar(&predictStockCodes, "stock", "600519", "股票代码，多个用逗号分隔")
	predictCmd.Flags().StringVar(&predictIndicator, "indicator", "MA5", "指标类型 (MA5, MA10, MA20, RSI, MACD)")
	predictCmd.Flags().StringVar(&predictModel, "model", "linear", "模型类型 (linear, arima, lstm)")
}

func runPredict(cmd *cobra.Command, args []string) {
	fmt.Println("╔═══════════════════════════════════════════════════════════╗")
	fmt.Println("║   技术指标预测器 v1.0.0                                    ║")
	fmt.Println("╚═══════════════════════════════════════════════════════════╝")
	fmt.Println()
	
	// Parse stock codes
	codes := strings.Split(predictStockCodes, ",")
	for i := range codes {
		codes[i] = strings.TrimSpace(codes[i])
	}
	
	fmt.Printf("🔮 预测配置:\n")
	fmt.Printf("   股票代码: %v\n", codes)
	fmt.Printf("   指标类型: %s\n", predictIndicator)
	fmt.Printf("   模型类型: %s\n", predictModel)
	fmt.Println()
	
	// Initialize components
	histStorage := storage.NewHistoricalStorage()
	modelStorage := storage.NewModelStorage()
	
	// Generate sample historical data first
	collector := ml.NewHistoricalCollector(histStorage)
	ctx := context.Background()
	for _, code := range codes {
		_, err := collector.Collect(ctx, code, 30)
		if err != nil {
			log.Printf("❌ 收集历史数据失败: %v", err)
			continue
		}
	}
	
	predictor := ml.NewIndicatorPredictor(modelStorage, histStorage)
	
	// Make predictions
	fmt.Println("📊 预测结果:")
	for _, code := range codes {
		indicator, err := predictor.Predict(ctx, code, predictIndicator)
		if err != nil {
			log.Printf("❌ 预测失败 %s: %v", code, err)
			continue
		}
		
		fmt.Printf("\n   %s - %s:\n", code, predictIndicator)
		fmt.Printf("      预测值: %.4f\n", indicator.PredictedValue)
		fmt.Printf("      置信度: %.2f%%\n", indicator.Confidence*100)
		fmt.Printf("      模型: %s (%s)\n", indicator.ModelType, indicator.ModelVersion)
	}
	
	fmt.Println("\n✅ 预测完成！")
}
