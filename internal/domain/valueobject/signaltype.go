package valueobject

// SignalType 信号类型
type SignalType string

const (
	SignalStrongBuy   SignalType = "STRONG_BUY"
	SignalBuy         SignalType = "BUY"
	SignalWatch       SignalType = "WATCH"
	SignalAvoid       SignalType = "AVOID"
	SignalStrongAvoid SignalType = "STRONG_AVOID"
)

// IsValid 验证信号类型是否有效
func (s SignalType) IsValid() bool {
	switch s {
	case SignalStrongBuy, SignalBuy, SignalWatch, SignalAvoid, SignalStrongAvoid:
		return true
	}
	return false
}

// IsBullish 是否为看涨信号
func (s SignalType) IsBullish() bool {
	return s == SignalStrongBuy || s == SignalBuy
}

// IsBearish 是否为看跌信号
func (s SignalType) IsBearish() bool {
	return s == SignalAvoid || s == SignalStrongAvoid
}

// FromScore 根据评分转换信号
func FromScore(score int) SignalType {
	switch {
	case score >= 85:
		return SignalStrongBuy
	case score >= 75:
		return SignalBuy
	case score >= 60:
		return SignalWatch
	case score >= 50:
		return SignalAvoid
	default:
		return SignalStrongAvoid
	}
}

// RiskLevel 风险等级
type RiskLevel string

const (
	RiskLow    RiskLevel = "LOW"
	RiskMedium RiskLevel = "MEDIUM"
	RiskHigh   RiskLevel = "HIGH"
)
