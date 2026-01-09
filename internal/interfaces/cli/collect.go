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

var collectCmd = &cobra.Command{
	Use:   "collect",
	Short: "收集历史数据",
	Long:  `收集股票历史数据用于模型训练和预测`,
	Run:   runCollect,
}

var (
	collectStockCodes string
	collectDays       int
	collectOutput     string
)

func init() {
	rootCmd.AddCommand(collectCmd)
	
	collectCmd.Flags().StringVar(&collectStockCodes, "stock", "600519", "股票代码，多个用逗号分隔")
	collectCmd.Flags().IntVar(&collectDays, "days", 365, "收集天数")
	collectCmd.Flags().StringVar(&collectOutput, "output", "", "输出目录（可选）")
}

func runCollect(cmd *cobra.Command, args []string) {
	fmt.Println("╔═══════════════════════════════════════════════════════════╗")
	fmt.Println("║   历史数据收集器 v1.0.0                                    ║")
	fmt.Println("╚═══════════════════════════════════════════════════════════╝")
	fmt.Println()
	
	// Parse stock codes
	codes := strings.Split(collectStockCodes, ",")
	for i := range codes {
		codes[i] = strings.TrimSpace(codes[i])
	}
	
	fmt.Printf("📊 收集配置:\n")
	fmt.Printf("   股票代码: %v\n", codes)
	fmt.Printf("   收集天数: %d\n", collectDays)
	fmt.Println()
	
	// Initialize components
	histStorage := storage.NewHistoricalStorage()
	collector := ml.NewHistoricalCollector(histStorage)
	
	// Collect data
	ctx := context.Background()
	results, err := collector.CollectBatch(ctx, codes, collectDays)
	if err != nil {
		log.Fatalf("❌ 收集失败: %v", err)
	}
	
	// Display results
	fmt.Println("\n📈 收集结果:")
	for code, series := range results {
		fmt.Printf("   %s: %d条记录 (%s 至 %s)\n",
			code, series.Count,
			series.StartDate.Format("2006-01-02"),
			series.EndDate.Format("2006-01-02"))
	}
	
	fmt.Println("\n✅ 收集完成！")
}
