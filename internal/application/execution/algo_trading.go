package execution

import (
	"context"
	"fmt"
	"math"
	"time"

	"github.com/bingaigc/gp/pkg/errors"
)

// AlgoType represents the type of algorithm
type AlgoType string

const (
	AlgoTypeTWAP  AlgoType = "TWAP"  // Time-Weighted Average Price
	AlgoTypeVWAP  AlgoType = "VWAP"  // Volume-Weighted Average Price
	AlgoTypePOV   AlgoType = "POV"   // Percentage of Volume
	AlgoTypeIS    AlgoType = "IS"    // Implementation Shortfall
	AlgoTypeSmart AlgoType = "SMART" // Smart Order Routing
)

// Order represents a trading order
type Order struct {
	ID            string
	StockCode     string
	StockName     string
	Side          string  // BUY or SELL
	TotalQuantity int     // Total shares to trade
	LimitPrice    float64 // Max price for buy, min price for sell
	Algorithm     AlgoType
	StartTime     time.Time
	EndTime       time.Time
	Status        string // PENDING, EXECUTING, COMPLETED, FAILED
}

// ChildOrder represents a slice of parent order
type ChildOrder struct {
	ID            string
	ParentOrderID string
	StockCode     string
	Side          string
	Quantity      int
	Price         float64
	SubmitTime    time.Time
	FilledQty     int
	AvgPrice      float64
	Status        string
}

// TWAPExecutor implements Time-Weighted Average Price algorithm
type TWAPExecutor struct {
	order         *Order
	sliceCount    int           // Number of slices
	sliceInterval time.Duration // Time between slices
	childOrders   []*ChildOrder
}

// NewTWAPExecutor creates a new TWAP executor
func NewTWAPExecutor(order *Order, duration time.Duration, sliceCount int) *TWAPExecutor {
	sliceInterval := duration / time.Duration(sliceCount)

	return &TWAPExecutor{
		order:         order,
		sliceCount:    sliceCount,
		sliceInterval: sliceInterval,
		childOrders:   make([]*ChildOrder, 0, sliceCount),
	}
}

// Execute runs the TWAP algorithm
func (t *TWAPExecutor) Execute(ctx context.Context) error {
	// Calculate quantity per slice
	baseQtyPerSlice := t.order.TotalQuantity / t.sliceCount
	remainder := t.order.TotalQuantity % t.sliceCount

	// Create child orders
	for i := 0; i < t.sliceCount; i++ {
		qty := baseQtyPerSlice
		if i < remainder {
			qty++ // Distribute remainder to first few slices
		}

		childOrder := &ChildOrder{
			ID:            generateOrderID(),
			ParentOrderID: t.order.ID,
			StockCode:     t.order.StockCode,
			Side:          t.order.Side,
			Quantity:      qty,
			Price:         t.order.LimitPrice,
			SubmitTime:    t.order.StartTime.Add(time.Duration(i) * t.sliceInterval),
			Status:        "PENDING",
		}

		t.childOrders = append(t.childOrders, childOrder)
	}

	// Execute child orders at scheduled times
	for _, child := range t.childOrders {
		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-time.After(time.Until(child.SubmitTime)):
			if err := t.submitChildOrder(ctx, child); err != nil {
				return err
			}
		}
	}

	return nil
}

// submitChildOrder submits a single child order
func (t *TWAPExecutor) submitChildOrder(ctx context.Context, child *ChildOrder) error {
	// In production, this would submit to broker API
	// For now, simulate order submission
	child.Status = "SUBMITTED"

	// Simulate fill (in production, would poll for fills)
	time.Sleep(100 * time.Millisecond)
	child.FilledQty = child.Quantity
	child.AvgPrice = child.Price * (1.0 + randRange(-0.001, 0.001))
	child.Status = "FILLED"

	return nil
}

// GetExecutionStats returns execution statistics
func (t *TWAPExecutor) GetExecutionStats() *ExecutionStats {
	totalFilled := 0
	totalCost := 0.0

	for _, child := range t.childOrders {
		totalFilled += child.FilledQty
		totalCost += float64(child.FilledQty) * child.AvgPrice
	}

	avgPrice := 0.0
	if totalFilled > 0 {
		avgPrice = totalCost / float64(totalFilled)
	}

	slippage := (avgPrice - t.order.LimitPrice) / t.order.LimitPrice

	return &ExecutionStats{
		TotalQuantity:   t.order.TotalQuantity,
		FilledQuantity:  totalFilled,
		AveragePrice:    avgPrice,
		TargetPrice:     t.order.LimitPrice,
		Slippage:        slippage,
		ChildOrderCount: len(t.childOrders),
	}
}

// VWAPExecutor implements Volume-Weighted Average Price algorithm
type VWAPExecutor struct {
	order             *Order
	historicalVolume  []int   // Historical volume pattern
	totalVolume       int     // Expected total volume for the day
	participationRate float64 // Target participation rate (e.g., 0.1 for 10%)
	childOrders       []*ChildOrder
}

// NewVWAPExecutor creates a new VWAP executor
func NewVWAPExecutor(order *Order, historicalVolume []int, participationRate float64) *VWAPExecutor {
	totalVol := 0
	for _, v := range historicalVolume {
		totalVol += v
	}

	return &VWAPExecutor{
		order:             order,
		historicalVolume:  historicalVolume,
		totalVolume:       totalVol,
		participationRate: participationRate,
		childOrders:       make([]*ChildOrder, 0),
	}
}

// Execute runs the VWAP algorithm
func (v *VWAPExecutor) Execute(ctx context.Context) error {
	// Calculate target quantity for each time slice based on volume profile
	remainingQty := v.order.TotalQuantity

	for i, volSlice := range v.historicalVolume {
		if remainingQty <= 0 {
			break
		}

		// Calculate target quantity based on volume profile
		volPct := float64(volSlice) / float64(v.totalVolume)
		targetQty := int(math.Round(float64(v.order.TotalQuantity) * volPct))

		// Apply participation rate
		execQty := int(math.Round(float64(targetQty) * v.participationRate))

		if execQty > remainingQty {
			execQty = remainingQty
		}

		if execQty > 0 {
			childOrder := &ChildOrder{
				ID:            generateOrderID(),
				ParentOrderID: v.order.ID,
				StockCode:     v.order.StockCode,
				Side:          v.order.Side,
				Quantity:      execQty,
				Price:         v.order.LimitPrice,
				SubmitTime:    v.order.StartTime.Add(time.Duration(i) * 5 * time.Minute),
				Status:        "PENDING",
			}

			v.childOrders = append(v.childOrders, childOrder)
			remainingQty -= execQty
		}
	}

	// Execute child orders
	for _, child := range v.childOrders {
		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-time.After(time.Until(child.SubmitTime)):
			if err := v.submitChildOrder(ctx, child); err != nil {
				return err
			}
		}
	}

	return nil
}

func (v *VWAPExecutor) submitChildOrder(ctx context.Context, child *ChildOrder) error {
	// Similar to TWAP submission
	child.Status = "SUBMITTED"
	time.Sleep(100 * time.Millisecond)
	child.FilledQty = child.Quantity
	child.AvgPrice = child.Price * (1.0 + randRange(-0.001, 0.001))
	child.Status = "FILLED"
	return nil
}

// ExecutionStats contains execution performance metrics
type ExecutionStats struct {
	TotalQuantity   int
	FilledQuantity  int
	AveragePrice    float64
	TargetPrice     float64
	Slippage        float64
	ChildOrderCount int
	StartTime       time.Time
	EndTime         time.Time
	Duration        time.Duration
}

// SmartOrderRouter routes orders to best execution venue
type SmartOrderRouter struct {
	venues []string
}

// NewSmartOrderRouter creates a new smart order router
func NewSmartOrderRouter(venues []string) *SmartOrderRouter {
	return &SmartOrderRouter{
		venues: venues,
	}
}

// RouteOrder selects best venue for order
func (s *SmartOrderRouter) RouteOrder(ctx context.Context, order *Order) (string, error) {
	// In production, would check:
	// - Best bid/offer at each venue
	// - Liquidity available
	// - Historical fill rates
	// - Transaction costs

	// For now, simple round-robin
	if len(s.venues) == 0 {
		return "", errors.New(errors.ErrValidation, "No venues available")
	}

	// Return first venue (in production, use sophisticated routing logic)
	return s.venues[0], nil
}

// Helper functions
func generateOrderID() string {
	return fmt.Sprintf("ORD%d", time.Now().UnixNano())
}

func randRange(min, max float64) float64 {
	// Simplified - in production use proper random
	return min + (max-min)*0.5
}
