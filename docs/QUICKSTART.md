# Quick Start Guide

## Prerequisites

- Go 1.21 or higher
- Git

## Installation

### 1. Clone the repository

```bash
git clone https://github.com/bingaigc/gp.git
cd gp
```

### 2. Download dependencies

```bash
go mod download
```

### 3. Build the application

```bash
make build
```

This will create the binary at `bin/alpha-detector`.

## Configuration

### Option 1: Environment Variable (Recommended)

```bash
export ALPHA_KIMI_API_KEY="your-kimi-api-key-here"
```

### Option 2: Configuration File

Edit `configs/config.yaml`:

```yaml
ai:
  api_key: "your-kimi-api-key-here"
```

## Running the Application

### Scan Market

```bash
# Using environment variable
ALPHA_KIMI_API_KEY=your-key ./bin/alpha-detector scan

# Or if you set it globally
./bin/alpha-detector scan

# Using custom config file
./bin/alpha-detector scan --config configs/config.yaml
```

### Using Make

```bash
make run
```

## Expected Output

The system will display:

1. **System Header** - Welcome banner and configuration info
2. **Scanning Progress** - Real-time progress updates
3. **Signal Details** - For each qualifying stock:
   - Signal type (STRONG_BUY, BUY, WATCH, AVOID, STRONG_AVOID)
   - Comprehensive score (0-100)
   - Factor breakdown with reasoning
   - AI analysis (bullish/bearish factors)
   - Price targets
   - Risk assessment
   - Position advice
4. **Statistics** - Summary and risk control stats

## Example Output

```
╔═══════════════════════════════════════════════════════════╗
║   CAMP-I Distributed Alpha Signal Detector v1.0.0        ║
║   企业级分布式Alpha信号探测系统                              ║
╚═══════════════════════════════════════════════════════════╝

📋 配置: CAMP-I Alpha Detector (development)
🔧 数据源: eastmoney
🤖 AI模型: kimi (moonshot-v1-8k)
⚙️  并发数: 3 workers

🚀 开始扫描市场...

═══════════════════════════════════════════════════════════
📊 宁德时代 (300750) - SZ
═══════════════════════════════════════════════════════════

🎯 信号类型: 🚀🚀🚀 STRONG_BUY
📈 综合评分: 88/100
🎲 置信度: 88.00%

📊 因子评分:
  • 资金面 (权重30%): 80/100 - 主力连续流入>5000万
  • 技术面 (权重25%): 100/100 - 放量上涨，换手率健康
  • 估值面 (权重20%): 80/100 - PE合理，大盘股
  • 板块轮动 (权重15%): 100/100 - 龙头板块，资金净流入>20亿
  • 机构行为 (权重10%): 80/100 - 机构净买入>5000万

💰 价格目标:
  当前价格: 185.20
  支撑位: 175.94
  压力位: 203.72
  止损价: 179.64
  止盈价: 200.02

⚠️  风险等级: 🟢 LOW (评分: 12/100)

📋 仓位建议:
  操作: BUY
  建议仓位: 10.0%
  入场价格: 183.35-187.05
  止损比例: 3.0%
  止盈比例: 8.0%
  持有周期: 5-10天
  建议理由: 强买入信号，建议10%仓位

✅ 扫描完成！生成信号数: 2
```

## Troubleshooting

### Issue: "AI API key is required"

**Solution**: Set the ALPHA_KIMI_API_KEY environment variable:

```bash
export ALPHA_KIMI_API_KEY="your-key"
```

### Issue: Build fails with "missing go.sum entry"

**Solution**: Run `go mod tidy`:

```bash
go mod tidy
```

### Issue: Command not found

**Solution**: Make sure you're in the project directory and the binary is built:

```bash
cd /path/to/gp
make build
./bin/alpha-detector --help
```

## Configuration Options

### System Settings

```yaml
system:
  name: "CAMP-I Alpha Detector"
  version: "1.0.0"
  environment: "development"  # development, production
  log_level: "info"           # debug, info, warn, error
```

### Data Source

```yaml
data_source:
  primary: "eastmoney"        # Primary data source
  timeout: 10s                # Request timeout
  max_retry: 3                # Maximum retry attempts
```

### AI Configuration

```yaml
ai:
  provider: "kimi"
  model: "moonshot-v1-8k"
  base_url: "https://api.moonshot.cn/v1"
  api_key: ""                 # Set via environment variable
  timeout: 30s
  max_retry: 2
```

### Risk Control

```yaml
risk:
  daily_max_loss: 50000.0     # Maximum daily loss in yuan
  max_signals_per_day: 20     # Maximum signals per day
  max_consecutive_losses: 3   # Trigger cooldown after N losses
  min_confidence: 0.6         # Minimum confidence threshold (60%)
```

### Pipeline

```yaml
pipeline:
  workers: 3                  # Number of concurrent workers
  batch_size: 20              # Batch size for processing
  limit: 100                  # Maximum stocks to scan
```

## Development

### Run Tests

```bash
make test
```

### Format Code

```bash
make fmt
```

### Code Check

```bash
make vet
```

### Build for Production

```bash
CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build \
    -ldflags="-w -s" \
    -o bin/alpha-detector \
    ./cmd/alpha
```

## Docker Deployment

### Build Docker Image

```bash
docker build -t alpha-detector:latest .
```

### Run Container

```bash
docker run -e ALPHA_KIMI_API_KEY=your-key alpha-detector:latest
```

## Next Steps

1. ✅ **Learn the System** - Read the [Architecture Documentation](docs/architecture.md)
2. 🔄 **Customize Configuration** - Adjust settings in `configs/config.yaml`
3. 🚀 **Run Your First Scan** - Execute `make run` to see it in action
4. 📊 **Analyze Results** - Review the signal output and scoring logic
5. 🔧 **Extend Functionality** - Add custom filters, scorers, or sinks

## Support

- 📖 Documentation: [docs/](docs/)
- 🐛 Issues: https://github.com/bingaigc/gp/issues
- 📧 Email: your-email@example.com

## License

MIT License - see LICENSE file for details
