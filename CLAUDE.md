# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Project Overview

This is a cryptocurrency futures trading bot written in Go that supports automated trading on exchanges like Binance. The bot includes multiple trading strategies (SMA, RSI, Grid), comprehensive risk management, and supports both paper trading and live trading modes.

## Common Development Commands

### Build and Run
- `make build` - Compile the trading bot binary
- `make run` - Start the trading bot (requires configuration)
- `make dev` - Start in development mode with `go run`
- `go run cmd/trader/main.go` - Direct run without make

### Testing and Configuration
- `make test` - Run all Go tests
- `make config-test` - Test configuration loading (useful for debugging config issues)
- `go run cmd/test/main.go` - Run configuration test directly

### Database and Infrastructure
- `make docker-up` - Start MySQL and Redis containers
- `make docker-down` - Stop containers
- `make reset-db` - Reset database and run migrations
- `make db-status` - Check database and Redis connectivity
- `make mysql-cli` - Connect to MySQL CLI
- `make redis-cli` - Connect to Redis CLI

### Environment Setup
- `make setup` - Download dependencies and initialize project
- `make full-setup` - Complete setup including infrastructure
- `make generate-env` - Create example environment file
- `make check-env` - Verify system requirements

## High-Level Architecture

### Core Components
1. **Trading Engine** (`internal/trading/engine.go`) - Main orchestrator that manages:
   - Market data collection
   - Strategy signal processing
   - Order execution
   - Risk monitoring
   - Account monitoring

2. **Strategy System** (`internal/trading/strategy.go`) - Pluggable strategy interface with implementations:
   - Simple Moving Average (SMA) crossover strategy
   - RSI overbought/oversold strategy  
   - Grid trading strategy
   - All strategies implement the same `Strategy` interface

3. **Risk Manager** (`internal/trading/risk.go`) - Validates all orders against:
   - Position size limits
   - Daily loss limits
   - Leverage restrictions
   - Account balance checks

4. **Exchange Layer** (`internal/exchange/`) - Abstracted exchange interface currently supporting Binance futures
5. **Data Layer** (`internal/database/` and `internal/models/`) - MySQL + Redis persistence

### Key Data Flow
1. Engine collects market data via exchange client
2. Strategy analyzes data and generates buy/sell signals
3. Risk manager validates signals
4. Orders executed through exchange client
5. Positions and trades stored in database
6. Performance metrics tracked continuously

### Database Schema
The system uses MySQL with the following key tables:
- `orders` - All order records from exchange
- `positions` - Current and historical positions
- `trades` - Individual trade executions
- `accounts` - Account balance and status snapshots
- `market_data` - Historical price/volume data
- `risk_metrics` - Daily performance and risk statistics

### Configuration System
- Main config: `config/config.yaml` (supports environment variable interpolation)
- Environment variables required: `BINANCE_API_KEY`, `BINANCE_SECRET_KEY`, `MYSQL_DSN`, `REDIS_ADDR`
- Paper trading mode configurable via `trading.enable_paper_trading`
- Testnet mode configurable via `exchange.testnet`

## Development Guidelines

### Safety First
- Always test with `exchange.testnet: true` first
- Use `trading.enable_paper_trading: true` for strategy testing
- Never commit API keys or secrets
- Use the provided Makefile commands for consistent development

### Testing New Strategies
1. Implement the `Strategy` interface in `internal/trading/strategy.go`
2. Add strategy type to engine initialization in `NewEngine()`
3. Configure strategy parameters in `config.yaml`
4. Test with paper trading mode enabled

### Database Changes
- Create new migration files in `migrations/` directory
- Use `make reset-db` to apply migrations during development
- Database models are defined in `internal/models/models.go`

### Adding Exchange Support
- Implement the `exchange.Client` interface in `internal/exchange/`
- Update configuration to support new exchange parameters
- Ensure proper error handling and rate limiting

### Key Safety Mechanisms
- Risk manager validates all orders before execution
- Stop loss and take profit automatically applied
- Position size limits enforced
- Daily loss limits monitored
- Paper trading mode for safe testing
- Comprehensive logging and error handling