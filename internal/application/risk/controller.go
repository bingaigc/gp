package risk

import (
	"context"
	"fmt"
	"sync"

	"github.com/bingaigc/gp/internal/domain/entity"
	"github.com/bingaigc/gp/pkg/errors"
	"github.com/shopspring/decimal"
)

// Controller 风控控制器
type Controller struct {
	// 日内限制
	dailyMaxLoss     float64
	dailyLoss        float64
	dailySignalCount int
	maxSignalsPerDay int

	// 连续亏损限制
	consecutiveLosses    int
	maxConsecutiveLosses int

	// 市场环境
	marketDowngrade bool
	downgradeMu     sync.RWMutex

	mu sync.Mutex
}

// NewController 创建风控控制器
func NewController(dailyMaxLoss float64, maxSignalsPerDay, maxConsecutiveLosses int) *Controller {
	return &Controller{
		dailyMaxLoss:         dailyMaxLoss,
		maxSignalsPerDay:     maxSignalsPerDay,
		maxConsecutiveLosses: maxConsecutiveLosses,
	}
}

// Validate 验证信号
func (c *Controller) Validate(ctx context.Context, signal *entity.Signal) error {
	c.mu.Lock()
	defer c.mu.Unlock()

	// 检查日内信号数量
	if c.dailySignalCount >= c.maxSignalsPerDay {
		return errors.New(errors.ErrorTypeRiskControl, "达到日内最大信号数量限制")
	}

	// 检查日内亏损
	if c.dailyLoss >= c.dailyMaxLoss {
		return errors.New(errors.ErrorTypeRiskControl, "达到日内最大亏损限制")
	}

	// 检查连续亏损
	if c.consecutiveLosses >= c.maxConsecutiveLosses {
		return errors.New(errors.ErrorTypeRiskControl, "连续亏损次数过多，进入冷却期")
	}

	// 一票否决条件检查
	if err := c.checkVetoConditions(signal); err != nil {
		return err
	}

	c.dailySignalCount++
	return nil
}

// checkVetoConditions 检查一票否决条件
func (c *Controller) checkVetoConditions(signal *entity.Signal) error {
	// 低置信度
	if signal.Confidence.LessThan(decimal.NewFromFloat(0.6)) {
		return errors.New(errors.ErrorTypeRiskControl, "置信度低于60%")
	}

	// 高风险信号
	if signal.RiskScore > 70 {
		return errors.New(errors.ErrorTypeRiskControl, fmt.Sprintf("风险评分过高: %d", signal.RiskScore))
	}

	return nil
}

// ShouldDowngrade 是否需要降级
func (c *Controller) ShouldDowngrade(ctx context.Context) (bool, string) {
	c.downgradeMu.RLock()
	defer c.downgradeMu.RUnlock()

	if c.marketDowngrade {
		return true, "市场环境恶劣，信号降级"
	}

	return false, ""
}

// SetMarketDowngrade 设置市场降级
func (c *Controller) SetMarketDowngrade(downgrade bool) {
	c.downgradeMu.Lock()
	defer c.downgradeMu.Unlock()
	c.marketDowngrade = downgrade
}

// RecordLoss 记录亏损
func (c *Controller) RecordLoss(amount float64) {
	c.mu.Lock()
	defer c.mu.Unlock()

	c.dailyLoss += amount
	c.consecutiveLosses++
}

// RecordProfit 记录盈利
func (c *Controller) RecordProfit(amount float64) {
	c.mu.Lock()
	defer c.mu.Unlock()

	c.consecutiveLosses = 0 // 重置连续亏损
}

// RecordSignal 记录信号
func (c *Controller) RecordSignal() {
	c.mu.Lock()
	defer c.mu.Unlock()

	c.dailySignalCount++
}

// ResetDaily 重置日内统计（每日调用）
func (c *Controller) ResetDaily() {
	c.mu.Lock()
	defer c.mu.Unlock()

	c.dailyLoss = 0
	c.dailySignalCount = 0
}

// GetStats 获取统计信息
func (c *Controller) GetStats() map[string]interface{} {
	c.mu.Lock()
	defer c.mu.Unlock()

	return map[string]interface{}{
		"daily_loss":          c.dailyLoss,
		"daily_signal_count":  c.dailySignalCount,
		"consecutive_losses":  c.consecutiveLosses,
		"market_downgrade":    c.marketDowngrade,
	}
}
