# 加密货币合约交易机器人

这是一个使用Go语言开发的加密货币合约交易机器人，支持币安等主流交易所的期货合约交易。

## 功能特性

- 🔄 **多交易所支持**: 目前支持币安期货，架构支持扩展其他交易所
- 📊 **多种交易策略**: 简单移动平均线(SMA)、RSI、网格交易等策略
- 🛡️ **风险管理**: 完善的风险控制系统，包括止损、止盈、仓位管理
- 📁 **数据持久化**: 使用MySQL存储交易数据，Redis缓存实时数据
- ⚙️ **灵活配置**: 支持YAML配置文件和环境变量
- 📈 **实时监控**: 账户、仓位、订单实时监控
- 🧪 **纸上交易**: 支持模拟交易模式，安全测试策略

## 快速开始

### 1. 环境要求

- Go 1.21+
- MySQL 8.0+
- Redis 6.0+

### 2. 获取API密钥

1. 注册币安账户
2. 在API管理中创建API密钥
3. **强烈建议先使用测试网进行测试**

### 3. 配置数据库

#### MySQL
```sql
CREATE DATABASE trading_bot CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci;
CREATE USER 'trader'@'localhost' IDENTIFIED BY 'your_password';
GRANT ALL PRIVILEGES ON trading_bot.* TO 'trader'@'localhost';
FLUSH PRIVILEGES;
```

#### Redis
```bash
# 启动Redis服务
redis-server
```

### 4. 配置环境变量

```bash
# 复制环境变量模板
cp .env.example .env

# 编辑环境变量文件
vim .env
```

填入以下信息：
```bash
BINANCE_API_KEY=your_api_key
BINANCE_SECRET_KEY=your_secret_key
MYSQL_DSN=trader:password@tcp(localhost:3306)/trading_bot?charset=utf8mb4&parseTime=True&loc=Local
REDIS_ADDR=localhost:6379
REDIS_PASSWORD=
```

### 5. 运行机器人

```bash
# 下载依赖
go mod download

# 编译并运行
go run cmd/trader/main.go
```

## 配置说明

### 主要配置项

```yaml
# 交易所配置
exchange:
  testnet: true                    # 建议先使用测试网
  
# 交易配置  
trading:
  symbols: ["BTCUSDT", "ETHUSDT"]  # 交易标的
  max_position_size: 1000.0        # 最大仓位大小
  stop_loss_percent: 2.0           # 止损百分比
  take_profit_percent: 5.0         # 止盈百分比
  enable_paper_trading: true       # 开启纸上交易模式
```

### 策略配置

#### 简单移动平均线策略
```yaml
strategy:
  type: "simple_moving_average"
  parameters:
    short_period: 10               # 短期均线
    long_period: 20                # 长期均线
    min_confidence: 0.7            # 最小信号置信度
```

#### RSI策略
```yaml
strategy:
  type: "rsi"
  parameters:
    period: 14                     # RSI周期
    oversold: 30                   # 超卖阈值
    overbought: 70                 # 超买阈值
```

## 项目结构

```
contract_playground/
├── cmd/trader/                    # 主程序入口
├── internal/
│   ├── config/                    # 配置管理
│   ├── database/                  # 数据库操作
│   ├── exchange/                  # 交易所客户端
│   ├── models/                    # 数据模型
│   └── trading/                   # 交易引擎和策略
├── config/                        # 配置文件
├── migrations/                    # 数据库迁移
└── pkg/utils/                     # 工具函数
```

## 安全提醒

⚠️ **重要安全建议**:

1. **首先使用测试网**: 设置 `exchange.testnet: true`
2. **启用纸上交易**: 设置 `trading.enable_paper_trading: true`
3. **设置合理的风险限制**: 不要将所有资金投入交易
4. **定期备份数据**: 备份数据库和配置文件
5. **监控日志**: 定期检查交易日志和系统状态
6. **API权限**: 仅授予必要的API权限，禁用提币权限

## 风险控制

系统内置多层风险控制机制：

- **止损止盈**: 自动止损和止盈
- **仓位控制**: 限制单笔和总仓位大小
- **日亏损限制**: 达到日亏损上限自动停止交易
- **杠杆控制**: 限制最大杠杆倍数
- **实时监控**: 监控账户余额和仓位变化

## 策略开发

### 实现自定义策略

1. 实现 `Strategy` 接口：
```go
type Strategy interface {
    Name() string
    ShouldBuy(ctx context.Context, symbol string, data *MarketData) (*Signal, error)
    ShouldSell(ctx context.Context, symbol string, data *MarketData, position *models.Position) (*Signal, error)
    Initialize(config map[string]interface{}) error
}
```

2. 在交易引擎中注册策略

### 信号结构
```go
type Signal struct {
    Action       string  // BUY, SELL, HOLD
    Quantity     float64
    Price        float64
    StopLoss     float64
    TakeProfit   float64
    Confidence   float64 // 0.0 to 1.0
    Reason       string
    PositionSide string  // LONG, SHORT
}
```

## 监控和日志

系统提供详细的日志记录：

- 交易信号和执行
- 风险管理决策
- 账户变化
- 错误和异常

日志格式支持JSON和文本格式，便于集成监控系统。

## 性能优化

- 使用连接池管理数据库连接
- Redis缓存实时数据
- 并发处理多个交易对
- 异步执行非关键任务

## 贡献指南

欢迎提交问题和改进建议！

1. Fork项目
2. 创建功能分支
3. 提交更改
4. 创建Pull Request

## 免责声明

⚠️ **投资风险提示**:

- 加密货币交易存在极高风险
- 本软件仅供学习和研究使用
- 使用本软件进行实盘交易的风险由用户自行承担
- 开发者不对任何交易损失负责
- 在进行实盘交易前，请充分了解相关风险并考虑自身财务状况

## 许可证

MIT License

## 联系方式

如有问题，请通过GitHub Issues联系。
