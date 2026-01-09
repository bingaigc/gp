package backtest

import (
	"context"
	"fmt"
	"time"

	"github.com/bingaigc/gp/internal/domain/entity"
)

// BacktestEngine performs historical backtesting of trading strategies
type BacktestEngine struct {
	historicalData map[string][]*entity.HistoricalData
	signals        []*entity.Signal
	positions      map[string]*Position
	equity         []float64
	trades         []*Trade
	startDate      time.Time
	endDate        time.Time
	initialCapital float64
	currentCapital float64
	commission     float64 // Commission rate (e.g., 0.0003 for 0.03%)
	slippage       float64 // Slippage rate (e.g., 0.001 for 0.1%)
}

// Position represents a trading position
type Position struct {
	StockCode  string
	Quantity   int
	AvgPrice   float64
	OpenTime   time.Time
	CurrentVal float64
	UnrealPnL  float64
}

// Trade represents a completed trade
type Trade struct {
	StockCode   string
	Direction   string // "BUY" or "SELL"
	Quantity    int
	Price       float64
	Time        time.Time
	Commission  float64
	PnL         float64
	HoldingDays int
	ReturnPct   float64
}

// BacktestResult contains backtest performance metrics
type BacktestResult struct {
	// Returns
	TotalReturn      float64
	AnnualizedReturn float64
	MaxDrawdown      float64
	SharpeRatio      float64
	SortinoRatio     float64
	CalmarRatio      float64

	// Trade Statistics
	TotalTrades    int
	WinningTrades  int
	LosingTrades   int
	WinRate        float64
	AvgWin         float64
	AvgLoss        float64
	ProfitFactor   float64
	AvgHoldingDays float64

	// Risk Metrics
	Volatility         float64
	DownsideVolatility float64
	MaxConsecutiveLoss int
	VaR95              float64 // Value at Risk at 95% confidence
	CVaR95             float64 // Conditional VaR

	// Equity Curve
	EquityCurve   []float64
	DrawdownCurve []float64

	// Trade Details
	Trades []*Trade

	// Period
	StartDate   time.Time
	EndDate     time.Time
	TradingDays int
}

// NewBacktestEngine creates a new backtest engine
func NewBacktestEngine(initialCapital float64, commission, slippage float64) *BacktestEngine {
	return &BacktestEngine{
		historicalData: make(map[string][]*entity.HistoricalData),
		positions:      make(map[string]*Position),
		initialCapital: initialCapital,
		currentCapital: initialCapital,
		commission:     commission,
		slippage:       slippage,
		equity:         []float64{initialCapital},
	}
}

// LoadHistoricalData loads historical data for backtesting
func (e *BacktestEngine) LoadHistoricalData(stockCode string, data []*entity.HistoricalData) {
	e.historicalData[stockCode] = data
	if len(data) > 0 {
		if e.startDate.IsZero() || data[0].Date.Before(e.startDate) {
			e.startDate = data[0].Date
		}
		if e.endDate.IsZero() || data[len(data)-1].Date.After(e.endDate) {
			e.endDate = data[len(data)-1].Date
		}
	}
}

// Run executes the backtest
func (e *BacktestEngine) Run(ctx context.Context, signalGenerator func(ctx context.Context, date time.Time, stock string) (*entity.Signal, error)) (*BacktestResult, error) {
	// Iterate through each trading day
	for date := e.startDate; !date.After(e.endDate); date = date.AddDate(0, 0, 1) {
		// Check context
		select {
		case <-ctx.Done():
			return nil, ctx.Err()
		default:
		}

		// Update positions with current market prices
		e.updatePositions(date)

		// Generate signals for all stocks
		for stockCode := range e.historicalData {
			signal, err := signalGenerator(ctx, date, stockCode)
			if err != nil {
				continue
			}
			if signal != nil {
				e.processSignal(signal, date)
			}
		}

		// Record equity
		equity := e.calculateEquity(date)
		e.equity = append(e.equity, equity)
	}

	// Close all positions at end
	e.closeAllPositions(e.endDate)

	// Calculate performance metrics
	return e.calculateResults(), nil
}

// updatePositions updates position values based on current market price
func (e *BacktestEngine) updatePositions(date time.Time) {
	for stockCode, pos := range e.positions {
		data := e.getHistoricalDataForDate(stockCode, date)
		if data != nil {
			pos.CurrentVal = float64(pos.Quantity) * data.Close
			pos.UnrealPnL = pos.CurrentVal - float64(pos.Quantity)*pos.AvgPrice
		}
	}
}

// processSignal processes a trading signal
func (e *BacktestEngine) processSignal(signal *entity.Signal, date time.Time) {
	data := e.getHistoricalDataForDate(signal.StockCode, date)
	if data == nil {
		return
	}

	// Determine action based on signal type
	if signal.SignalType.IsBullish() {
		e.openPosition(signal, data, date)
	} else if signal.SignalType.IsBearish() {
		e.closePosition(signal.StockCode, data, date)
	}
}

// openPosition opens a new position or adds to existing
func (e *BacktestEngine) openPosition(signal *entity.Signal, data *entity.HistoricalData, date time.Time) {
	// Calculate position size based on position advice
	positionSizePct, _ := signal.PositionAdvice.SizePct.Float64()
	positionSize := e.currentCapital * (positionSizePct / 100.0)

	// Apply slippage
	buyPrice := data.Close * (1 + e.slippage)

	// Calculate quantity (round down to 100 shares - standard lot)
	quantity := int(positionSize/buyPrice/100) * 100
	if quantity <= 0 {
		return
	}

	// Calculate cost
	cost := float64(quantity) * buyPrice
	commission := cost * e.commission
	totalCost := cost + commission

	// Check if enough capital
	if totalCost > e.currentCapital {
		return
	}

	// Update capital
	e.currentCapital -= totalCost

	// Create or update position
	if pos, exists := e.positions[signal.StockCode]; exists {
		// Add to existing position (average price)
		totalQty := pos.Quantity + quantity
		pos.AvgPrice = (pos.AvgPrice*float64(pos.Quantity) + buyPrice*float64(quantity)) / float64(totalQty)
		pos.Quantity = totalQty
	} else {
		// Create new position
		e.positions[signal.StockCode] = &Position{
			StockCode: signal.StockCode,
			Quantity:  quantity,
			AvgPrice:  buyPrice,
			OpenTime:  date,
		}
	}

	// Record trade
	e.trades = append(e.trades, &Trade{
		StockCode:  signal.StockCode,
		Direction:  "BUY",
		Quantity:   quantity,
		Price:      buyPrice,
		Time:       date,
		Commission: commission,
	})
}

// closePosition closes an existing position
func (e *BacktestEngine) closePosition(stockCode string, data *entity.HistoricalData, date time.Time) {
	pos, exists := e.positions[stockCode]
	if !exists {
		return
	}

	// Apply slippage
	sellPrice := data.Close * (1 - e.slippage)

	// Calculate proceeds
	proceeds := float64(pos.Quantity) * sellPrice
	commission := proceeds * e.commission
	netProceeds := proceeds - commission

	// Calculate P&L
	cost := float64(pos.Quantity) * pos.AvgPrice
	pnl := netProceeds - cost
	returnPct := (pnl / cost) * 100

	// Update capital
	e.currentCapital += netProceeds

	// Calculate holding days
	holdingDays := int(date.Sub(pos.OpenTime).Hours() / 24)

	// Record trade
	e.trades = append(e.trades, &Trade{
		StockCode:   stockCode,
		Direction:   "SELL",
		Quantity:    pos.Quantity,
		Price:       sellPrice,
		Time:        date,
		Commission:  commission,
		PnL:         pnl,
		HoldingDays: holdingDays,
		ReturnPct:   returnPct,
	})

	// Remove position
	delete(e.positions, stockCode)
}

// closeAllPositions closes all open positions
func (e *BacktestEngine) closeAllPositions(date time.Time) {
	for stockCode := range e.positions {
		data := e.getHistoricalDataForDate(stockCode, date)
		if data != nil {
			e.closePosition(stockCode, data, date)
		}
	}
}

// calculateEquity calculates total equity (capital + positions)
func (e *BacktestEngine) calculateEquity(date time.Time) float64 {
	equity := e.currentCapital
	for stockCode, pos := range e.positions {
		data := e.getHistoricalDataForDate(stockCode, date)
		if data != nil {
			equity += float64(pos.Quantity) * data.Close
		}
	}
	return equity
}

// getHistoricalDataForDate retrieves historical data for a specific date
func (e *BacktestEngine) getHistoricalDataForDate(stockCode string, date time.Time) *entity.HistoricalData {
	data, exists := e.historicalData[stockCode]
	if !exists {
		return nil
	}

	// Binary search for date
	for _, d := range data {
		if d.Date.Equal(date) || (d.Date.After(date) && d.Date.Sub(date) < 24*time.Hour) {
			return d
		}
	}
	return nil
}

// calculateResults calculates backtest performance metrics
func (e *BacktestEngine) calculateResults() *BacktestResult {
	result := &BacktestResult{
		StartDate: e.startDate,
		EndDate:   e.endDate,
		Trades:    e.trades,
	}

	// Calculate returns
	finalEquity := e.equity[len(e.equity)-1]
	result.TotalReturn = ((finalEquity - e.initialCapital) / e.initialCapital) * 100

	// Calculate trading days
	result.TradingDays = len(e.equity) - 1
	yearsTrading := float64(result.TradingDays) / 252.0
	result.AnnualizedReturn = result.TotalReturn / yearsTrading

	// Calculate maximum drawdown
	result.MaxDrawdown, result.DrawdownCurve = e.calculateMaxDrawdown()

	// Calculate Sharpe ratio
	returns := e.calculateDailyReturns()
	result.Volatility = e.calculateStdDev(returns) * 15.874 // sqrt(252)
	avgReturn := e.calculateMean(returns) * 252
	riskFreeRate := 0.02 // Assume 2% risk-free rate
	if result.Volatility > 0 {
		result.SharpeRatio = (avgReturn - riskFreeRate) / result.Volatility
	}

	// Calculate Sortino ratio
	downside := e.calculateDownsideReturns(returns)
	result.DownsideVolatility = e.calculateStdDev(downside) * 15.874
	if result.DownsideVolatility > 0 {
		result.SortinoRatio = (avgReturn - riskFreeRate) / result.DownsideVolatility
	}

	// Calculate Calmar ratio
	if result.MaxDrawdown != 0 {
		result.CalmarRatio = result.AnnualizedReturn / abs(result.MaxDrawdown)
	}

	// Calculate trade statistics
	e.calculateTradeStats(result)

	// Calculate VaR and CVaR
	result.VaR95, result.CVaR95 = e.calculateRiskMetrics(returns)

	result.EquityCurve = e.equity

	return result
}

// calculateMaxDrawdown calculates maximum drawdown
func (e *BacktestEngine) calculateMaxDrawdown() (float64, []float64) {
	maxDrawdown := 0.0
	peak := e.equity[0]
	drawdownCurve := make([]float64, len(e.equity))

	for i, equity := range e.equity {
		if equity > peak {
			peak = equity
		}
		drawdown := ((equity - peak) / peak) * 100
		drawdownCurve[i] = drawdown
		if drawdown < maxDrawdown {
			maxDrawdown = drawdown
		}
	}

	return maxDrawdown, drawdownCurve
}

// calculateDailyReturns calculates daily returns
func (e *BacktestEngine) calculateDailyReturns() []float64 {
	returns := make([]float64, len(e.equity)-1)
	for i := 1; i < len(e.equity); i++ {
		returns[i-1] = (e.equity[i] - e.equity[i-1]) / e.equity[i-1]
	}
	return returns
}

// calculateDownsideReturns extracts negative returns for downside volatility
func (e *BacktestEngine) calculateDownsideReturns(returns []float64) []float64 {
	downside := []float64{}
	for _, r := range returns {
		if r < 0 {
			downside = append(downside, r)
		}
	}
	return downside
}

// calculateMean calculates mean of a slice
func (e *BacktestEngine) calculateMean(values []float64) float64 {
	if len(values) == 0 {
		return 0
	}
	sum := 0.0
	for _, v := range values {
		sum += v
	}
	return sum / float64(len(values))
}

// calculateStdDev calculates standard deviation
func (e *BacktestEngine) calculateStdDev(values []float64) float64 {
	if len(values) == 0 {
		return 0
	}
	mean := e.calculateMean(values)
	sumSq := 0.0
	for _, v := range values {
		diff := v - mean
		sumSq += diff * diff
	}
	return sqrt(sumSq / float64(len(values)))
}

// calculateTradeStats calculates trade statistics
func (e *BacktestEngine) calculateTradeStats(result *BacktestResult) {
	result.TotalTrades = len(e.trades)

	totalWin := 0.0
	totalLoss := 0.0
	totalHoldingDays := 0
	consecutiveLoss := 0
	maxConsecutiveLoss := 0

	for _, trade := range e.trades {
		if trade.Direction == "SELL" {
			totalHoldingDays += trade.HoldingDays

			if trade.PnL > 0 {
				result.WinningTrades++
				totalWin += trade.PnL
				consecutiveLoss = 0
			} else if trade.PnL < 0 {
				result.LosingTrades++
				totalLoss += abs(trade.PnL)
				consecutiveLoss++
				if consecutiveLoss > maxConsecutiveLoss {
					maxConsecutiveLoss = consecutiveLoss
				}
			}
		}
	}

	totalCompletedTrades := result.WinningTrades + result.LosingTrades
	if totalCompletedTrades > 0 {
		result.WinRate = (float64(result.WinningTrades) / float64(totalCompletedTrades)) * 100
		result.AvgHoldingDays = float64(totalHoldingDays) / float64(totalCompletedTrades)
	}

	if result.WinningTrades > 0 {
		result.AvgWin = totalWin / float64(result.WinningTrades)
	}
	if result.LosingTrades > 0 {
		result.AvgLoss = totalLoss / float64(result.LosingTrades)
	}
	if totalLoss > 0 {
		result.ProfitFactor = totalWin / totalLoss
	}

	result.MaxConsecutiveLoss = maxConsecutiveLoss
}

// calculateRiskMetrics calculates VaR and CVaR
func (e *BacktestEngine) calculateRiskMetrics(returns []float64) (float64, float64) {
	if len(returns) == 0 {
		return 0, 0
	}

	// Sort returns
	sorted := make([]float64, len(returns))
	copy(sorted, returns)
	for i := 0; i < len(sorted); i++ {
		for j := i + 1; j < len(sorted); j++ {
			if sorted[i] > sorted[j] {
				sorted[i], sorted[j] = sorted[j], sorted[i]
			}
		}
	}

	// VaR at 95% confidence (5th percentile)
	varIndex := int(float64(len(sorted)) * 0.05)
	var95 := sorted[varIndex] * 100

	// CVaR (average of returns below VaR)
	cvarSum := 0.0
	for i := 0; i <= varIndex; i++ {
		cvarSum += sorted[i]
	}
	cvar95 := (cvarSum / float64(varIndex+1)) * 100

	return var95, cvar95
}

// Helper functions
func sqrt(x float64) float64 {
	if x == 0 {
		return 0
	}
	z := x
	for i := 0; i < 10; i++ {
		z = z - (z*z-x)/(2*z)
	}
	return z
}

func abs(x float64) float64 {
	if x < 0 {
		return -x
	}
	return x
}

// String returns a formatted summary of backtest results
func (r *BacktestResult) String() string {
	return fmt.Sprintf(`
╔══════════════════════════════════════════════════════════════════╗
║                    BACKTEST RESULTS                              ║
╚══════════════════════════════════════════════════════════════════╝

📊 收益指标:
   总收益率: %.2f%%
   年化收益率: %.2f%%
   最大回撤: %.2f%%
   夏普比率: %.2f
   索提诺比率: %.2f
   卡玛比率: %.2f

📈 交易统计:
   总交易次数: %d
   盈利次数: %d
   亏损次数: %d
   胜率: %.2f%%
   平均盈利: ¥%.2f
   平均亏损: ¥%.2f
   盈亏比: %.2f
   平均持仓天数: %.1f

⚠️  风险指标:
   波动率 (年化): %.2f%%
   下行波动率: %.2f%%
   最大连续亏损: %d
   VaR (95%%): %.2f%%
   CVaR (95%%): %.2f%%

📅 回测周期:
   开始日期: %s
   结束日期: %s
   交易天数: %d
`,
		r.TotalReturn, r.AnnualizedReturn, r.MaxDrawdown,
		r.SharpeRatio, r.SortinoRatio, r.CalmarRatio,
		r.TotalTrades, r.WinningTrades, r.LosingTrades,
		r.WinRate, r.AvgWin, r.AvgLoss, r.ProfitFactor,
		r.AvgHoldingDays,
		r.Volatility, r.DownsideVolatility, r.MaxConsecutiveLoss,
		r.VaR95, r.CVaR95,
		r.StartDate.Format("2006-01-02"),
		r.EndDate.Format("2006-01-02"),
		r.TradingDays,
	)
}
