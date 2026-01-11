package sink

import (
	"context"
	"fmt"
	"sync"

	"github.com/bingaigc/gp/internal/domain/entity"
	"github.com/bingaigc/gp/internal/domain/service"
)

// MultiSink 多重输出（同时输出到多个sink）
type MultiSink struct {
	sinks []service.SignalSink
	mu    sync.Mutex
}

// NewMultiSink 创建多重输出
func NewMultiSink(sinks ...service.SignalSink) *MultiSink {
	return &MultiSink{
		sinks: sinks,
	}
}

// AddSink 添加sink
func (m *MultiSink) AddSink(sink service.SignalSink) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.sinks = append(m.sinks, sink)
}

// Emit 输出到所有sink
func (m *MultiSink) Emit(ctx context.Context, signal *entity.Signal) error {
	m.mu.Lock()
	sinks := make([]service.SignalSink, len(m.sinks))
	copy(sinks, m.sinks)
	m.mu.Unlock()

	var errs []error
	for _, sink := range sinks {
		if err := sink.Emit(ctx, signal); err != nil {
			errs = append(errs, fmt.Errorf("%s: %w", sink.Name(), err))
		}
	}

	if len(errs) > 0 {
		return fmt.Errorf("multi-sink errors: %v", errs)
	}

	return nil
}

// Close 关闭所有sink
func (m *MultiSink) Close() error {
	m.mu.Lock()
	defer m.mu.Unlock()

	var errs []error
	for _, sink := range m.sinks {
		if err := sink.Close(); err != nil {
			errs = append(errs, fmt.Errorf("%s: %w", sink.Name(), err))
		}
	}

	if len(errs) > 0 {
		return fmt.Errorf("multi-sink close errors: %v", errs)
	}

	return nil
}

// Name 名称
func (m *MultiSink) Name() string {
	return "multi_sink"
}

// GetSinks 获取所有sink
func (m *MultiSink) GetSinks() []service.SignalSink {
	m.mu.Lock()
	defer m.mu.Unlock()

	sinks := make([]service.SignalSink, len(m.sinks))
	copy(sinks, m.sinks)
	return sinks
}
