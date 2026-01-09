package cli

import (
	"context"
	"fmt"
	"time"

	"github.com/bingaigc/gp/internal/application/backtest"
	"github.com/bingaigc/gp/internal/domain/entity"
	"github.com/bingaigc/gp/internal/domain/valueobject"
	"github.com/shopspring/decimal"
	"github.com/spf13/cobra"
)

var backtestCmd = &cobra.Command{
	Use:   "backtest",
	Short: "运行回测验证策略表现",
	Long: `回测命令用于验证交易策略的历史表现
	
使用历史数据模拟交易，评估策略的收益和风险指标，包括：
- 总收益率、年化收益率
- 最大回撤、波动率
- 夏普比率、索提诺比率、卡玛比率
- VaR/CVaR风险指标
- 交易统计（胜率、盈亏比等）

示例:
  # 回测指定股票
  alpha-detector backtest --stock 600519 --days 365 --capital 1000000
  
  # 回测多只股票
  alpha-detector backtest --stock 600519,000858 --days 365 --capital 1000000
  
  # 指定佣金和滑点
  alpha-detector backtest --stock 600519 --days 365 --commission 0.0003 --slippage 0.001`,
	Run: runBacktest,
}

var (
	backtestStocks     string
	backtestDays       int
	backtestCapital    float64
	backtestCommission float64
	backtestSlippage   float64
)

func init() {
	backtestCmd.Flags().StringVar(&backtestStocks, "stock", "600519", "股票代码，多个用逗号分隔")
	backtestCmd.Flags().IntVar(&backtestDays, "days", 365, "回测天数")
	backtestCmd.Flags().Float64Var(&backtestCapital, "capital", 1000000, "初始资金 (元)")
	backtestCmd.Flags().Float64Var(&backtestCommission, "commission", 0.0003, "佣金率 (0.0003 = 0.03%)")
	backtestCmd.Flags().Float64Var(&backtestSlippage, "slippage", 0.001, "滑点率 (0.001 = 0.1%)")
	rootCmd.AddCommand(backtestCmd)
}

func runBacktest(cmd *cobra.Command, args []string) {
	ctx := context.Background()

	fmt.Println(`
╔══════════════════════════════════════════════════════════════════╗
║               CAMP-I 策略回测系统 v1.0.0                         ║
╚══════════════════════════════════════════════════════════════════╝`)

	fmt.Printf("\n📊 回测参数:\n")
	fmt.Printf("   股票代码: %s\n", backtestStocks)
	fmt.Printf("   回测天数: %d\n", backtestDays)
	fmt.Printf("   初始资金: ¥%.2f\n", backtestCapital)
	fmt.Printf("   佣金率: %.4f%%\n", backtestCommission*100)
	fmt.Printf("   滑点率: %.4f%%\n", backtestSlippage*100)
	fmt.Println()

	// Create backtest engine
	engine := backtest.NewBacktestEngine(backtestCapital, backtestCommission, backtestSlippage)

	// Load historical data (Demo: use sample data)
	fmt.Println("📥 加载历史数据...")

	// In production, load real historical data
	// For demo, create sample data
	startDate := time.Now().AddDate(0, 0, -backtestDays)
	sampleData := generateSampleHistoricalData("600519", startDate, backtestDays)
	engine.LoadHistoricalData("600519", sampleData)

	fmt.Printf("✅ 加载完成: %d个交易日\n", len(sampleData))
	fmt.Println()

	// Run backtest
	fmt.Println("🔄 运行回测...")

	// Demo signal generator (in production, use real CAMP-I strategy)
	signalGenerator := func(ctx context.Context, date time.Time, stock string) (*entity.Signal, error) {
		data := findDataForDate(sampleData, date)
		if data == nil {
			return nil, nil
		}

		// Simple demo strategy: Buy when MA5 > MA20, Sell when MA5 < MA20
		if data.MA5 > data.MA20 {
			return &entity.Signal{
				StockCode:  stock,
				SignalType: valueobject.SignalBuy,
				TotalScore: 75,
				Confidence: decimal.NewFromFloat(0.75),
				PositionAdvice: entity.PositionAdvice{
					SizePct:       decimal.NewFromInt(30), // 30% position
					StopLossPct:   decimal.NewFromInt(5),
					TakeProfitPct: decimal.NewFromInt(15),
				},
				GeneratedAt: date,
			}, nil
		} else if data.MA5 < data.MA20 {
			return &entity.Signal{
				StockCode:  stock,
				SignalType: valueobject.SignalAvoid,
				TotalScore: 40,
				Confidence: decimal.NewFromFloat(0.7),
				PositionAdvice: entity.PositionAdvice{
					SizePct: decimal.NewFromInt(0),
				},
				GeneratedAt: date,
			}, nil
		}

		return nil, nil
	}

	result, err := engine.Run(ctx, signalGenerator)
	if err != nil {
		fmt.Printf("❌ 回测失败: %v\n", err)
		return
	}

	fmt.Println("✅ 回测完成!")
	fmt.Println()

	// Display results
	fmt.Print(result.String())

	// Additional insights
	fmt.Println("\n💡 策略分析:")
	if result.SharpeRatio > 2.0 {
		fmt.Println("   ⭐ 优秀: 夏普比率 >2.0，风险调整后收益优异")
	} else if result.SharpeRatio > 1.0 {
		fmt.Println("   ✅ 良好: 夏普比率 >1.0，风险调整后收益合理")
	} else {
		fmt.Println("   ⚠️  一般: 夏普比率 <1.0，建议优化策略")
	}

	if result.MaxDrawdown < -20 {
		fmt.Println("   ⚠️  警告: 最大回撤超过20%，风险控制需要加强")
	} else if result.MaxDrawdown < -10 {
		fmt.Println("   ⚠️  注意: 最大回撤10-20%，风险中等")
	} else {
		fmt.Println("   ✅ 优秀: 最大回撤 <10%，风险控制良好")
	}

	if result.WinRate > 60 {
		fmt.Println("   ✅ 优秀: 胜率 >60%，策略稳定性高")
	} else if result.WinRate > 50 {
		fmt.Println("   ✅ 良好: 胜率 >50%，策略有效")
	} else {
		fmt.Println("   ⚠️  警告: 胜率 <50%，建议优化入场条件")
	}

	if result.ProfitFactor > 2.0 {
		fmt.Println("   ⭐ 优秀: 盈亏比 >2.0，盈利质量高")
	} else if result.ProfitFactor > 1.5 {
		fmt.Println("   ✅ 良好: 盈亏比 >1.5，盈利能力合理")
	} else {
		fmt.Println("   ⚠️  一般: 盈亏比 <1.5，需要提高盈利质量")
	}

	fmt.Println("\n📈 下一步建议:")
	fmt.Println("   1. 在不同市场环境下测试策略鲁棒性")
	fmt.Println("   2. 优化参数以提高夏普比率")
	fmt.Println("   3. 考虑加入止损止盈逻辑")
	fmt.Println("   4. 分析亏损交易，找出改进点")
	fmt.Println()
}

// generateSampleHistoricalData generates sample historical data for demo
func generateSampleHistoricalData(stockCode string, startDate time.Time, days int) []*entity.HistoricalData {
	data := make([]*entity.HistoricalData, days)
	basePrice := 2000.0

	for i := 0; i < days; i++ {
		date := startDate.AddDate(0, 0, i)

		// Generate price with trend and noise
		trend := float64(i) * 0.5
		noise := (float64(i%10) - 5) * 10
		price := basePrice + trend + noise

		// Calculate MA5 and MA20
		ma5 := price
		ma20 := price - 20
		if i >= 5 {
			ma5 = (price*5 + basePrice*2) / 7
		}
		if i >= 20 {
			ma20 = (price*20 + basePrice*10) / 30
		}

		data[i] = &entity.HistoricalData{
			StockCode: stockCode,
			Date:      date,
			Open:      price * 0.99,
			High:      price * 1.02,
			Low:       price * 0.98,
			Close:     price,
			Volume:    float64(1000000 + i*10000),
			Amount:    price * float64(1000000+i*10000),
			MA5:       ma5,
			MA20:      ma20,
		}
	}

	return data
}

// findDataForDate finds historical data for a specific date
func findDataForDate(data []*entity.HistoricalData, date time.Time) *entity.HistoricalData {
	for _, d := range data {
		if d.Date.Equal(date) || (d.Date.After(date) && d.Date.Sub(date) < 24*time.Hour) {
			return d
		}
	}
	return nil
}
