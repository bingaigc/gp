package pipeline

import (
	"context"
	"fmt"
	"log/slog"

	"github.com/bingaigc/gp/internal/domain/entity"
	"github.com/bingaigc/gp/internal/domain/service"
	"golang.org/x/sync/errgroup"
)

// Orchestrator 管道编排器
type Orchestrator struct {
	fetcher  service.DataFetcher
	scorer   service.Scorer
	analyzer service.Analyzer
	riskCtrl service.RiskController
	sink     service.SignalSink

	workers int
	logger  *slog.Logger
}

// ScoredStock 评分后的股票
type ScoredStock struct {
	Stock   *entity.Stock
	Score   int
	Factors []entity.FactorScore
}

// NewOrchestrator 创建编排器
func NewOrchestrator(
	fetcher service.DataFetcher,
	scorer service.Scorer,
	analyzer service.Analyzer,
	riskCtrl service.RiskController,
	sink service.SignalSink,
	workers int,
	logger *slog.Logger,
) *Orchestrator {
	return &Orchestrator{
		fetcher:  fetcher,
		scorer:   scorer,
		analyzer: analyzer,
		riskCtrl: riskCtrl,
		sink:     sink,
		workers:  workers,
		logger:   logger,
	}
}

// Run 运行管道
func (o *Orchestrator) Run(ctx context.Context, limit int) error {
	o.logger.Info("pipeline started", "workers", o.workers, "limit", limit)

	g, ctx := errgroup.WithContext(ctx)

	// Stage 1: Fetch
	stocksCh := make(chan *entity.Stock, o.workers*2)
	g.Go(func() error {
		defer close(stocksCh)
		return o.fetchStage(ctx, stocksCh, limit)
	})

	// Stage 2: Score
	scoredCh := make(chan *ScoredStock, o.workers*2)
	g.Go(func() error {
		defer close(scoredCh)
		return o.scoreStage(ctx, stocksCh, scoredCh)
	})

	// Stage 3: Analyze (Fan-out)
	signalsCh := make(chan *entity.Signal, o.workers*2)
	for i := 0; i < o.workers; i++ {
		workerID := i
		g.Go(func() error {
			return o.analyzeWorker(ctx, scoredCh, signalsCh, workerID)
		})
	}

	// Close signalsCh when all workers done
	go func() {
		g.Wait()
		close(signalsCh)
	}()

	// Stage 4: Sink
	g.Go(func() error {
		return o.sinkStage(ctx, signalsCh)
	})

	if err := g.Wait(); err != nil && err != context.Canceled {
		return fmt.Errorf("pipeline error: %w", err)
	}

	o.logger.Info("pipeline completed")
	return nil
}

// fetchStage 获取数据阶段
func (o *Orchestrator) fetchStage(ctx context.Context, out chan<- *entity.Stock, limit int) error {
	o.logger.Info("fetch stage started", "limit", limit)

	stocks, err := o.fetcher.FetchActiveStocks(ctx, limit)
	if err != nil {
		return fmt.Errorf("fetch stocks: %w", err)
	}

	o.logger.Info("fetched stocks", "count", len(stocks))

	for _, stock := range stocks {
		select {
		case out <- stock:
		case <-ctx.Done():
			return ctx.Err()
		}
	}

	return nil
}

// scoreStage 评分阶段
func (o *Orchestrator) scoreStage(ctx context.Context, in <-chan *entity.Stock, out chan<- *ScoredStock) error {
	o.logger.Info("score stage started")

	for stock := range in {
		score, factors, err := o.scorer.Score(ctx, stock)
		if err != nil {
			o.logger.Warn("score error", "stock", stock.Code, "error", err)
			continue
		}

		// 预过滤：评分低于50的不进入AI分析
		if score < 50 {
			o.logger.Debug("score too low, skipped", "stock", stock.Code, "score", score)
			continue
		}

		select {
		case out <- &ScoredStock{Stock: stock, Score: score, Factors: factors}:
		case <-ctx.Done():
			return ctx.Err()
		}
	}

	return nil
}

// analyzeWorker AI分析worker
func (o *Orchestrator) analyzeWorker(ctx context.Context, in <-chan *ScoredStock, out chan<- *entity.Signal, workerID int) error {
	o.logger.Debug("analyze worker started", "worker_id", workerID)

	for scored := range in {
		// 检查是否需要降级
		if shouldDowngrade, reason := o.riskCtrl.ShouldDowngrade(ctx); shouldDowngrade {
			o.logger.Warn("risk downgrade", "reason", reason)
			continue
		}

		// AI分析
		signal, err := o.analyzer.Analyze(ctx, scored.Stock, scored.Score)
		if err != nil {
			o.logger.Warn("analyze error", "stock", scored.Stock.Code, "error", err)
			continue
		}

		// 风控验证
		if err := o.riskCtrl.Validate(ctx, signal); err != nil {
			o.logger.Warn("risk validation failed", "stock", signal.StockCode, "reason", err)
			continue
		}

		select {
		case out <- signal:
			o.logger.Info("signal generated",
				"stock", signal.StockCode,
				"signal", signal.SignalType,
				"score", signal.TotalScore,
				"worker", workerID,
			)
		case <-ctx.Done():
			return ctx.Err()
		}
	}

	return nil
}

// sinkStage 输出阶段
func (o *Orchestrator) sinkStage(ctx context.Context, in <-chan *entity.Signal) error {
	o.logger.Info("sink stage started")

	count := 0
	for signal := range in {
		if err := o.sink.Emit(ctx, signal); err != nil {
			o.logger.Error("sink error", "signal", signal.ID, "error", err)
		}
		count++
	}

	o.logger.Info("sink stage completed", "count", count)
	return nil
}
