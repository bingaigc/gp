package risk

import (
	"context"
	"math"
	"sort"
	"time"

	"github.com/bingaigc/gp/internal/domain/entity"
)

// RealtimeVaREngine calculates VaR and CVaR in real-time
type RealtimeVaREngine struct {
	confidenceLevel float64 // e.g., 0.95 for 95% confidence
	windowSize      int     // rolling window size in days
	returns         []float64
}

// NewRealtimeVaREngine creates a new real-time VaR engine
func NewRealtimeVaREngine(confidenceLevel float64, windowSize int) *RealtimeVaREngine {
	return &RealtimeVaREngine{
		confidenceLevel: confidenceLevel,
		windowSize:      windowSize,
		returns:         make([]float64, 0, windowSize),
	}
}

// VaRResult contains VaR and CVaR calculations
type VaRResult struct {
	VaR              float64   // Value at Risk
	CVaR             float64   // Conditional Value at Risk (Expected Shortfall)
	ConfidenceLevel  float64   // Confidence level (e.g., 0.95)
	CalculatedAt     time.Time // Calculation timestamp
	SampleSize       int       // Number of returns used
	RiskLevel        string    // HIGH, MEDIUM, LOW
	ExceedanceEvents int       // Number of times loss exceeded VaR
}

// UpdateReturns adds new return to the rolling window
func (r *RealtimeVaREngine) UpdateReturns(portfolioReturn float64) {
	r.returns = append(r.returns, portfolioReturn)
	
	// Maintain rolling window
	if len(r.returns) > r.windowSize {
		r.returns = r.returns[1:]
	}
}

// CalculateVaR computes VaR and CVaR
func (r *RealtimeVaREngine) CalculateVaR(ctx context.Context) (*VaRResult, error) {
	if len(r.returns) < 10 {
		// Need minimum sample size
		return &VaRResult{
			VaR:             0,
			CVaR:            0,
			ConfidenceLevel: r.confidenceLevel,
			CalculatedAt:    time.Now(),
			SampleSize:      len(r.returns),
			RiskLevel:       "UNKNOWN",
		}, nil
	}

	// Sort returns (ascending, so losses are at the beginning)
	sortedReturns := make([]float64, len(r.returns))
	copy(sortedReturns, r.returns)
	sort.Float64s(sortedReturns)

	// Calculate VaR (negative return at confidence level)
	varIndex := int(float64(len(sortedReturns)) * (1.0 - r.confidenceLevel))
	if varIndex >= len(sortedReturns) {
		varIndex = len(sortedReturns) - 1
	}
	var_ := -sortedReturns[varIndex] // Negate because we want positive loss value

	// Calculate CVaR (average of losses beyond VaR)
	var cvar float64
	count := 0
	for i := 0; i <= varIndex; i++ {
		cvar += sortedReturns[i]
		count++
	}
	if count > 0 {
		cvar = -cvar / float64(count) // Negate and average
	}

	// Count exceedance events
	exceedances := 0
	varThreshold := -var_ // Negative because we're comparing with actual returns
	for _, ret := range r.returns {
		if ret < varThreshold {
			exceedances++
		}
	}

	// Determine risk level
	riskLevel := "LOW"
	if var_ > 0.05 { // 5% single-day loss
		riskLevel = "HIGH"
	} else if var_ > 0.03 { // 3% single-day loss
		riskLevel = "MEDIUM"
	}

	return &VaRResult{
		VaR:              var_,
		CVaR:             cvar,
		ConfidenceLevel:  r.confidenceLevel,
		CalculatedAt:     time.Now(),
		SampleSize:       len(r.returns),
		RiskLevel:        riskLevel,
		ExceedanceEvents: exceedances,
	}, nil
}

// DynamicRiskController manages real-time risk with VaR/CVaR
type DynamicRiskController struct {
	varEngine        *RealtimeVaREngine
	varLimit         float64 // Maximum acceptable VaR (e.g., 0.03 for 3%)
	cvarLimit        float64 // Maximum acceptable CVaR
	currentPositions map[string]*entity.Signal
	portfolioValue   float64
}

// NewDynamicRiskController creates a new dynamic risk controller
func NewDynamicRiskController(varLimit, cvarLimit float64) *DynamicRiskController {
	return &DynamicRiskController{
		varEngine:        NewRealtimeVaREngine(0.95, 252), // 95% confidence, 1 year window
		varLimit:         varLimit,
		cvarLimit:        cvarLimit,
		currentPositions: make(map[string]*entity.Signal),
		portfolioValue:   1000000.0, // Default 1M
	}
}

// UpdatePortfolio updates portfolio state and recalculates risk
func (d *DynamicRiskController) UpdatePortfolio(ctx context.Context, currentValue float64, previousValue float64) (*VaRResult, error) {
	// Calculate return
	portfolioReturn := 0.0
	if previousValue > 0 {
		portfolioReturn = (currentValue - previousValue) / previousValue
	}

	// Update VaR engine
	d.varEngine.UpdateReturns(portfolioReturn)
	d.portfolioValue = currentValue

	// Calculate VaR/CVaR
	return d.varEngine.CalculateVaR(ctx)
}

// CheckRiskLimits validates if current risk exceeds limits
func (d *DynamicRiskController) CheckRiskLimits(ctx context.Context) (bool, string, error) {
	varResult, err := d.varEngine.CalculateVaR(ctx)
	if err != nil {
		return false, "", err
	}

	// Check VaR limit
	if varResult.VaR > d.varLimit {
		return false, "VaR exceeds limit: " + varResult.RiskLevel, nil
	}

	// Check CVaR limit
	if varResult.CVaR > d.cvarLimit {
		return false, "CVaR exceeds limit: " + varResult.RiskLevel, nil
	}

	return true, "", nil
}

// ShouldReducePositions determines if positions should be reduced based on risk
func (d *DynamicRiskController) ShouldReducePositions(ctx context.Context) (bool, float64, error) {
	varResult, err := d.varEngine.CalculateVaR(ctx)
	if err != nil {
		return false, 0, err
	}

	// Calculate risk ratio
	riskRatio := varResult.VaR / d.varLimit

	if riskRatio > 1.2 {
		// Risk 20% above limit, reduce by 50%
		return true, 0.5, nil
	} else if riskRatio > 1.0 {
		// Risk above limit, reduce by 30%
		return true, 0.3, nil
	}

	return false, 0, nil
}

// CalculateMaxPositionSize calculates maximum position size based on VaR
func (d *DynamicRiskController) CalculateMaxPositionSize(ctx context.Context, stockVolatility float64) (float64, error) {
	varResult, err := d.varEngine.CalculateVaR(ctx)
	if err != nil {
		return 0, err
	}

	// Use VaR to limit position size
	// maxLoss = positionSize * stockVolatility
	// We want maxLoss <= VaRLimit * portfolioValue
	maxLoss := d.varLimit * d.portfolioValue
	
	if stockVolatility > 0 {
		maxPositionValue := maxLoss / stockVolatility
		maxPositionPct := maxPositionValue / d.portfolioValue
		
		// Cap at 15% regardless
		if maxPositionPct > 0.15 {
			maxPositionPct = 0.15
		}
		
		return maxPositionPct, nil
	}

	// Default to 10% if volatility unknown
	return 0.10, nil
}

// PortfolioStressTest runs stress scenarios
func (d *DynamicRiskController) PortfolioStressTest(ctx context.Context, scenarios []float64) map[string]float64 {
	results := make(map[string]float64)
	
	for i, shock := range scenarios {
		scenarioName := ""
		if shock < 0 {
			scenarioName = fmt.Sprintf("Market_Down_%.0f%%", -shock*100)
		} else {
			scenarioName = fmt.Sprintf("Market_Up_%.0f%%", shock*100)
		}
		
		stressValue := d.portfolioValue * (1.0 + shock)
		stressLoss := d.portfolioValue - stressValue
		results[scenarioName] = stressLoss
	}
	
	// Add VaR-based scenarios
	varResult, _ := d.varEngine.CalculateVaR(ctx)
	results["VaR_95"] = d.portfolioValue * varResult.VaR
	results["CVaR_95"] = d.portfolioValue * varResult.CVaR
	
	return results
}


