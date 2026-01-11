package scorer

import (
	"context"

	"github.com/bingaigc/gp/internal/domain/entity"
	"github.com/shopspring/decimal"
)

// CAMPIScorer CAMP-I五维度评分器
type CAMPIScorer struct {
	// 权重配置
	capitalWeight       float64
	technicalWeight     float64
	valuationWeight     float64
	sectorWeight        float64
	institutionalWeight float64
}

// NewCAMPIScorer 创建评分器
func NewCAMPIScorer() *CAMPIScorer {
	return &CAMPIScorer{
		capitalWeight:       0.30, // 资金面 30%
		technicalWeight:     0.25, // 技术面 25%
		valuationWeight:     0.20, // 估值面 20%
		sectorWeight:        0.15, // 板块轮动 15%
		institutionalWeight: 0.10, // 机构行为 10%
	}
}

// Score 计算综合评分
func (s *CAMPIScorer) Score(ctx context.Context, stock *entity.Stock) (int, []entity.FactorScore, error) {
	factors := make([]entity.FactorScore, 0, 5)

	// 1. Capital Flow (资金面)
	capitalScore, capitalReason := s.scoreCapitalFlow(stock)
	factors = append(factors, entity.FactorScore{
		Name:   "资金面",
		Weight: decimal.NewFromFloat(s.capitalWeight),
		Score:  capitalScore,
		Reason: capitalReason,
	})

	// 2. Technical Action (技术面)
	techScore, techReason := s.scoreTechnical(stock)
	factors = append(factors, entity.FactorScore{
		Name:   "技术面",
		Weight: decimal.NewFromFloat(s.technicalWeight),
		Score:  techScore,
		Reason: techReason,
	})

	// 3. Valuation Margin (估值面)
	valuationScore, valuationReason := s.scoreValuation(stock)
	factors = append(factors, entity.FactorScore{
		Name:   "估值面",
		Weight: decimal.NewFromFloat(s.valuationWeight),
		Score:  valuationScore,
		Reason: valuationReason,
	})

	// 4. Sector Rotation (板块轮动)
	sectorScore, sectorReason := s.scoreSector(stock)
	factors = append(factors, entity.FactorScore{
		Name:   "板块轮动",
		Weight: decimal.NewFromFloat(s.sectorWeight),
		Score:  sectorScore,
		Reason: sectorReason,
	})

	// 5. Institutional (机构行为)
	instScore, instReason := s.scoreInstitutional(stock)
	factors = append(factors, entity.FactorScore{
		Name:   "机构行为",
		Weight: decimal.NewFromFloat(s.institutionalWeight),
		Score:  instScore,
		Reason: instReason,
	})

	// 加权综合评分
	totalScore := int(
		float64(capitalScore)*s.capitalWeight +
			float64(techScore)*s.technicalWeight +
			float64(valuationScore)*s.valuationWeight +
			float64(sectorScore)*s.sectorWeight +
			float64(instScore)*s.institutionalWeight,
	)

	return totalScore, factors, nil
}

// scoreCapitalFlow 资金面评分
func (s *CAMPIScorer) scoreCapitalFlow(stock *entity.Stock) (int, string) {
	mainFlow := stock.Capital.MainNetInflow.InexactFloat64()
	consecutiveDays := stock.Capital.ConsecutiveDays

	var score int
	var reason string

	// 检查量价背离 (一票否决)
	if stock.HasFlowPriceDivergence() {
		return 0, "⚠️ 量价背离！涨幅>3%但主力净流出>1000万"
	}

	// 主力净流入评分
	switch {
	case mainFlow > 100000000 && consecutiveDays >= 3:
		score = 100
		reason = "主力连续3日以上流入，累计>1亿"
	case mainFlow > 50000000 && consecutiveDays >= 2:
		score = 80
		reason = "主力连续流入>5000万"
	case mainFlow > 20000000:
		score = 60
		reason = "主力流入>2000万"
	case mainFlow > 0:
		score = 30
		reason = "主力小幅流入"
	default:
		score = 0
		reason = "主力资金流出"
	}

	return score, reason
}

// scoreTechnical 技术面评分
func (s *CAMPIScorer) scoreTechnical(stock *entity.Stock) (int, string) {
	turnover := stock.Technical.TurnoverRate.InexactFloat64()
	volumeRatio := stock.Technical.VolumeRatio.InexactFloat64()
	change := stock.Change.InexactFloat64()

	var score int
	var reason string

	// 换手率检查
	if turnover > 30 {
		return 0, "⚠️ 换手率过高>30%，投机过度"
	}

	// 放量上涨
	if change > 1 && volumeRatio > 1.5 && turnover >= 3 && turnover <= 15 {
		score = 100
		reason = "放量上涨，换手率健康"
	} else if volumeRatio > 1.2 && turnover >= 5 && turnover <= 15 {
		score = 80
		reason = "温和放量，换手率适中"
	} else if change < 0 && volumeRatio < 0.8 {
		score = 60
		reason = "缩量回调，可能为洗盘"
	} else if change < -2 && volumeRatio > 2 {
		score = 0
		reason = "⚠️ 放量下跌"
	} else {
		score = 40
		reason = "量价关系一般"
	}

	return score, reason
}

// scoreValuation 估值面评分
func (s *CAMPIScorer) scoreValuation(stock *entity.Stock) (int, string) {
	pe := stock.Valuation.PE.InexactFloat64()
	marketCap := stock.Valuation.MarketCap.InexactFloat64()
	price := stock.Price.InexactFloat64()

	var score int
	var reason string

	// 低价股风险
	if price < 3 {
		return 0, "⚠️ 股价<3元，低价股风险"
	}

	// PE为负且小市值
	if pe < 0 && marketCap < 10000000000 {
		return 0, "⚠️ PE为负且市值<100亿，亏损小盘股"
	}

	// PE估值评分
	switch {
	case pe > 0 && pe < 20 && marketCap > 100000000000:
		score = 100
		reason = "PE低估，超大盘蓝筹"
	case pe >= 20 && pe < 40 && marketCap > 50000000000:
		score = 80
		reason = "PE合理，大盘股"
	case pe >= 40 && pe < 80:
		score = 60
		reason = "PE偏高但可接受"
	case pe >= 80:
		score = 20
		reason = "PE高估"
	case pe < 0:
		score = 10
		reason = "公司亏损"
	default:
		score = 50
		reason = "估值中等"
	}

	return score, reason
}

// scoreSector 板块轮动评分
func (s *CAMPIScorer) scoreSector(stock *entity.Stock) (int, string) {
	if stock.Sector == nil {
		return 50, "无板块数据"
	}

	rank := stock.Sector.Rank
	netInflow := stock.Sector.NetInflow.InexactFloat64()

	var score int
	var reason string

	switch {
	case rank <= 3 && netInflow > 2000000000:
		score = 100
		reason = "龙头板块，资金净流入>20亿"
	case rank <= 10 && netInflow > 1000000000:
		score = 80
		reason = "热门板块，资金净流入>10亿"
	case rank <= 30:
		score = 50
		reason = "板块排名中游"
	default:
		score = 20
		reason = "板块排名靠后"
	}

	return score, reason
}

// scoreInstitutional 机构行为评分
func (s *CAMPIScorer) scoreInstitutional(stock *entity.Stock) (int, string) {
	if stock.LHB == nil {
		return 50, "未上龙虎榜"
	}

	buyOrgs := stock.LHB.BuyOrgs
	sellOrgs := stock.LHB.SellOrgs
	netAmount := stock.LHB.NetAmount.InexactFloat64()

	var score int
	var reason string

	// 机构集体卖出
	if sellOrgs >= 3 && buyOrgs == 0 {
		return 0, "⚠️ 机构集体卖出"
	}

	switch {
	case buyOrgs >= 3 && sellOrgs == 0:
		score = 100
		reason = "3家以上机构买入，无机构卖出"
	case netAmount > 50000000:
		score = 80
		reason = "机构净买入>5000万"
	case netAmount > 20000000:
		score = 60
		reason = "机构净买入>2000万"
	case netAmount > 0:
		score = 40
		reason = "机构小幅净买入"
	default:
		score = 20
		reason = "机构净卖出"
	}

	return score, reason
}
