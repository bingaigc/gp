package entity

import (
	"time"

	"github.com/shopspring/decimal"
)

// Stock 股票实体
type Stock struct {
	// 基础信息
	Code     string `json:"code"`     // 股票代码
	Name     string `json:"name"`     // 股票名称
	Market   string `json:"market"`   // SH/SZ
	Industry string `json:"industry"` // 所属行业

	// 行情数据
	Price  decimal.Decimal `json:"price"`
	Change decimal.Decimal `json:"change"`
	Volume int64           `json:"volume"`
	Amount decimal.Decimal `json:"amount"`

	// 聚合数据
	Capital   CapitalFlow `json:"capital"`
	Technical Technical   `json:"technical"`
	Valuation Valuation   `json:"valuation"`
	Sector    *Sector     `json:"sector,omitempty"`
	LHB       *LHBData    `json:"lhb,omitempty"`

	// 元数据
	Timestamp time.Time `json:"timestamp"`
	Source    string    `json:"source"`
}

// CapitalFlow 资金流向
type CapitalFlow struct {
	MainNetInflow   decimal.Decimal `json:"main_net_inflow"`
	SuperNetInflow  decimal.Decimal `json:"super_net_inflow"`
	BigNetInflow    decimal.Decimal `json:"big_net_inflow"`
	MiddleNetInflow decimal.Decimal `json:"middle_net_inflow"`
	SmallNetInflow  decimal.Decimal `json:"small_net_inflow"`
	ConsecutiveDays int             `json:"consecutive_days"`
}

// Technical 技术指标
type Technical struct {
	VolumeRatio  decimal.Decimal `json:"volume_ratio"`
	TurnoverRate decimal.Decimal `json:"turnover_rate"`
	Amplitude    decimal.Decimal `json:"amplitude"`
	MA5          decimal.Decimal `json:"ma5"`
	MA10         decimal.Decimal `json:"ma10"`
	MA20         decimal.Decimal `json:"ma20"`
}

// Valuation 估值指标
type Valuation struct {
	PE        decimal.Decimal `json:"pe"`         // 市盈率
	PB        decimal.Decimal `json:"pb"`         // 市净率
	MarketCap decimal.Decimal `json:"market_cap"` // 总市值
	CircCap   decimal.Decimal `json:"circ_cap"`   // 流通市值
}

// Sector 板块信息
type Sector struct {
	Name      string          `json:"name"`
	Rank      int             `json:"rank"`
	Change    decimal.Decimal `json:"change"`
	NetInflow decimal.Decimal `json:"net_inflow"`
}

// LHBData 龙虎榜数据
type LHBData struct {
	Reason       string          `json:"reason"`
	BuyAmount    decimal.Decimal `json:"buy_amount"`
	SellAmount   decimal.Decimal `json:"sell_amount"`
	NetAmount    decimal.Decimal `json:"net_amount"`
	BuyOrgs      int             `json:"buy_orgs"`
	SellOrgs     int             `json:"sell_orgs"`
	OrgNetInflow decimal.Decimal `json:"org_net_inflow"`
}

// HasFlowPriceDivergence 检测量价背离
func (s *Stock) HasFlowPriceDivergence() bool {
	// 涨幅>3% 但主力净流出>1000万
	return s.Change.GreaterThan(decimal.NewFromFloat(3)) &&
		s.Capital.MainNetInflow.LessThan(decimal.NewFromFloat(-10000000))
}

// IsLowPriceStock 是否低价股
func (s *Stock) IsLowPriceStock() bool {
	return s.Price.LessThan(decimal.NewFromFloat(3))
}

// IsSmallCap 是否小市值股票
func (s *Stock) IsSmallCap() bool {
	return s.Valuation.MarketCap.LessThan(decimal.NewFromFloat(5000000000)) // 50亿
}
