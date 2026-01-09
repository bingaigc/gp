package filter

import (
	"context"
	"fmt"

	"github.com/bingaigc/gp/internal/domain/entity"
	"github.com/shopspring/decimal"
)

// Filter 过滤器接口
type Filter interface {
	// Filter 过滤股票，返回true表示通过
	Filter(ctx context.Context, stock *entity.Stock) (bool, string)
	// Name 过滤器名称
	Name() string
}

// Chain 过滤器链
type Chain struct {
	filters []Filter
}

// NewChain 创建过滤器链
func NewChain(filters ...Filter) *Chain {
	return &Chain{
		filters: filters,
	}
}

// AddFilter 添加过滤器
func (c *Chain) AddFilter(filter Filter) {
	c.filters = append(c.filters, filter)
}

// Filter 执行过滤链
func (c *Chain) Filter(ctx context.Context, stock *entity.Stock) (bool, string) {
	for _, filter := range c.filters {
		if pass, reason := filter.Filter(ctx, stock); !pass {
			return false, reason
		}
	}
	return true, ""
}

// BasicFilter 基础过滤器（一票否决条件）
type BasicFilter struct{}

// NewBasicFilter 创建基础过滤器
func NewBasicFilter() *BasicFilter {
	return &BasicFilter{}
}

// Filter 执行基础过滤
func (f *BasicFilter) Filter(ctx context.Context, stock *entity.Stock) (bool, string) {
	// 1. 低价股
	if stock.IsLowPriceStock() {
		return false, "低价股风险：股价<3元"
	}

	// 2. 量价背离
	if stock.HasFlowPriceDivergence() {
		return false, "量价背离：涨幅>3%但主力流出>1000万"
	}

	// 3. 小市值亏损股
	if stock.IsSmallCap() && stock.Valuation.PE.LessThan(decimal.Zero) {
		return false, "小市值亏损股：市值<50亿且PE<0"
	}

	// 4. 换手率过高
	if stock.Technical.TurnoverRate.GreaterThan(decimal.NewFromFloat(30)) {
		return false, "投机过度：换手率>30%"
	}

	return true, ""
}

// Name 过滤器名称
func (f *BasicFilter) Name() string {
	return "basic_filter"
}

// LiquidityFilter 流动性过滤器
type LiquidityFilter struct {
	minVolume int64           // 最小成交量
	minAmount decimal.Decimal // 最小成交额
}

// NewLiquidityFilter 创建流动性过滤器
func NewLiquidityFilter(minVolume int64, minAmount float64) *LiquidityFilter {
	return &LiquidityFilter{
		minVolume: minVolume,
		minAmount: decimal.NewFromFloat(minAmount),
	}
}

// Filter 执行流动性过滤
func (f *LiquidityFilter) Filter(ctx context.Context, stock *entity.Stock) (bool, string) {
	// 成交量检查
	if stock.Volume < f.minVolume {
		return false, fmt.Sprintf("成交量不足：%d < %d", stock.Volume, f.minVolume)
	}

	// 成交额检查
	if stock.Amount.LessThan(f.minAmount) {
		return false, fmt.Sprintf("成交额不足：%.2f < %.2f",
			stock.Amount.InexactFloat64(), f.minAmount.InexactFloat64())
	}

	return true, ""
}

// Name 过滤器名称
func (f *LiquidityFilter) Name() string {
	return "liquidity_filter"
}

// WhitelistFilter 白名单过滤器
type WhitelistFilter struct {
	codes map[string]bool
}

// NewWhitelistFilter 创建白名单过滤器
func NewWhitelistFilter(codes []string) *WhitelistFilter {
	codeMap := make(map[string]bool, len(codes))
	for _, code := range codes {
		codeMap[code] = true
	}
	return &WhitelistFilter{
		codes: codeMap,
	}
}

// Filter 执行白名单过滤
func (f *WhitelistFilter) Filter(ctx context.Context, stock *entity.Stock) (bool, string) {
	if len(f.codes) == 0 {
		return true, "" // 空白名单表示不限制
	}

	if !f.codes[stock.Code] {
		return false, "不在白名单中"
	}

	return true, ""
}

// Name 过滤器名称
func (f *WhitelistFilter) Name() string {
	return "whitelist_filter"
}

// BlacklistFilter 黑名单过滤器
type BlacklistFilter struct {
	codes map[string]bool
}

// NewBlacklistFilter 创建黑名单过滤器
func NewBlacklistFilter(codes []string) *BlacklistFilter {
	codeMap := make(map[string]bool, len(codes))
	for _, code := range codes {
		codeMap[code] = true
	}
	return &BlacklistFilter{
		codes: codeMap,
	}
}

// Filter 执行黑名单过滤
func (f *BlacklistFilter) Filter(ctx context.Context, stock *entity.Stock) (bool, string) {
	if f.codes[stock.Code] {
		return false, "在黑名单中"
	}

	return true, ""
}

// Name 过滤器名称
func (f *BlacklistFilter) Name() string {
	return "blacklist_filter"
}
