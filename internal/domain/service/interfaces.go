package service

import (
	"context"

	"github.com/bingaigc/gp/internal/domain/entity"
)

// DataFetcher 数据获取接口
type DataFetcher interface {
	// FetchActiveStocks 获取活跃股票列表
	FetchActiveStocks(ctx context.Context, limit int) ([]*entity.Stock, error)

	// FetchByCode 获取单只股票
	FetchByCode(ctx context.Context, code string) (*entity.Stock, error)

	// HealthCheck 健康检查
	HealthCheck(ctx context.Context) error

	// Name 数据源名称
	Name() string
}

// Analyzer AI分析接口
type Analyzer interface {
	// Analyze 分析股票
	Analyze(ctx context.Context, stock *entity.Stock, preScore int) (*entity.Signal, error)

	// HealthCheck 健康检查
	HealthCheck(ctx context.Context) error
}

// Scorer 评分接口
type Scorer interface {
	// Score 计算评分
	Score(ctx context.Context, stock *entity.Stock) (int, []entity.FactorScore, error)
}

// RiskController 风控接口
type RiskController interface {
	// Validate 验证信号
	Validate(ctx context.Context, signal *entity.Signal) error

	// ShouldDowngrade 是否需要降级
	ShouldDowngrade(ctx context.Context) (bool, string)

	// RecordLoss 记录亏损
	RecordLoss(amount float64)

	// RecordSignal 记录信号
	RecordSignal()
}

// SignalSink 信号输出接口
type SignalSink interface {
	// Emit 输出信号
	Emit(ctx context.Context, signal *entity.Signal) error

	// Close 关闭
	Close() error

	// Name 名称
	Name() string
}
