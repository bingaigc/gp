package sink

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/bingaigc/gp/internal/domain/entity"
)

// JSONFileSink JSON文件输出
type JSONFileSink struct {
	outputDir string
	pretty    bool
}

// NewJSONFileSink 创建JSON文件输出
func NewJSONFileSink(outputDir string, pretty bool) *JSONFileSink {
	// 确保输出目录存在
	if outputDir == "" {
		outputDir = "output/signals"
	}
	os.MkdirAll(outputDir, 0755)

	return &JSONFileSink{
		outputDir: outputDir,
		pretty:    pretty,
	}
}

// Emit 输出信号到JSON文件
func (s *JSONFileSink) Emit(ctx context.Context, signal *entity.Signal) error {
	// 生成文件名：signals_YYYYMMDD_HHMMSS_STOCKCODE.json
	timestamp := signal.GeneratedAt.Format("20060102_150405")
	filename := fmt.Sprintf("signal_%s_%s.json", timestamp, signal.StockCode)
	filepath := filepath.Join(s.outputDir, filename)

	// 序列化信号
	var data []byte
	var err error
	if s.pretty {
		data, err = json.MarshalIndent(signal, "", "  ")
	} else {
		data, err = json.Marshal(signal)
	}
	if err != nil {
		return fmt.Errorf("marshal signal: %w", err)
	}

	// 写入文件
	if err := os.WriteFile(filepath, data, 0644); err != nil {
		return fmt.Errorf("write file: %w", err)
	}

	return nil
}

// Close 关闭
func (s *JSONFileSink) Close() error {
	return nil
}

// Name 名称
func (s *JSONFileSink) Name() string {
	return "json_file"
}

// JSONLineSink JSON Lines输出（追加模式，一行一个信号）
type JSONLineSink struct {
	filepath string
	file     *os.File
}

// NewJSONLineSink 创建JSON Lines输出
func NewJSONLineSink(filepath string) (*JSONLineSink, error) {
	if filepath == "" {
		// 默认文件名：signals_YYYYMMDD.jsonl
		filepath = fmt.Sprintf("output/signals_%s.jsonl", time.Now().Format("20060102"))
	}

	// 确保目录存在
	dir := filepath[:len(filepath)-len(filepath[len(filepath)-1:])]
	if dir != "" {
		os.MkdirAll(dir, 0755)
	}

	// 打开文件（追加模式）
	file, err := os.OpenFile(filepath, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0644)
	if err != nil {
		return nil, fmt.Errorf("open file: %w", err)
	}

	return &JSONLineSink{
		filepath: filepath,
		file:     file,
	}, nil
}

// Emit 输出信号（一行一个JSON）
func (s *JSONLineSink) Emit(ctx context.Context, signal *entity.Signal) error {
	data, err := json.Marshal(signal)
	if err != nil {
		return fmt.Errorf("marshal signal: %w", err)
	}

	// 写入一行
	if _, err := s.file.Write(append(data, '\n')); err != nil {
		return fmt.Errorf("write line: %w", err)
	}

	return nil
}

// Close 关闭文件
func (s *JSONLineSink) Close() error {
	if s.file != nil {
		return s.file.Close()
	}
	return nil
}

// Name 名称
func (s *JSONLineSink) Name() string {
	return "json_line"
}
