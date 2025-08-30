-- 交易机器人初始数据库结构
-- 此文件包含所有必要的表结构定义

-- 创建数据库（如果不存在）
CREATE DATABASE IF NOT EXISTS trading_bot CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci;
USE trading_bot;

-- 交易配置表
CREATE TABLE IF NOT EXISTS trading_configs (
    id BIGINT UNSIGNED AUTO_INCREMENT PRIMARY KEY,
    name VARCHAR(255) UNIQUE NOT NULL,
    symbol VARCHAR(50) NOT NULL,
    is_active BOOLEAN DEFAULT TRUE,
    max_position DECIMAL(20,8) NOT NULL,
    stop_loss DECIMAL(10,6) NOT NULL,
    take_profit DECIMAL(10,6) NOT NULL,
    leverage INT DEFAULT 1,
    risk_percent DECIMAL(5,2) NOT NULL,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    INDEX idx_symbol (symbol),
    INDEX idx_active (is_active)
);

-- 订单表
CREATE TABLE IF NOT EXISTS orders (
    id BIGINT UNSIGNED AUTO_INCREMENT PRIMARY KEY,
    exchange_order_id VARCHAR(255) UNIQUE NOT NULL,
    symbol VARCHAR(50) NOT NULL,
    side ENUM('BUY', 'SELL') NOT NULL,
    type ENUM('MARKET', 'LIMIT', 'STOP_MARKET', 'STOP_LIMIT', 'TAKE_PROFIT_MARKET') NOT NULL,
    status ENUM('NEW', 'PARTIALLY_FILLED', 'FILLED', 'CANCELED', 'REJECTED', 'EXPIRED') NOT NULL,
    quantity DECIMAL(20,8) NOT NULL,
    price DECIMAL(20,8),
    stop_price DECIMAL(20,8),
    executed_qty DECIMAL(20,8) DEFAULT 0,
    cumulative_quote DECIMAL(20,8) DEFAULT 0,
    commission DECIMAL(20,8) DEFAULT 0,
    commission_asset VARCHAR(20),
    time_in_force ENUM('GTC', 'IOC', 'FOK') DEFAULT 'GTC',
    reduce_only BOOLEAN DEFAULT FALSE,
    close_position BOOLEAN DEFAULT FALSE,
    position_side ENUM('BOTH', 'LONG', 'SHORT') DEFAULT 'BOTH',
    strategy VARCHAR(100),
    notes TEXT,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    INDEX idx_exchange_order_id (exchange_order_id),
    INDEX idx_symbol (symbol),
    INDEX idx_status (status),
    INDEX idx_created_at (created_at)
);

-- 仓位表
CREATE TABLE IF NOT EXISTS positions (
    id BIGINT UNSIGNED AUTO_INCREMENT PRIMARY KEY,
    symbol VARCHAR(50) NOT NULL,
    position_side ENUM('LONG', 'SHORT') NOT NULL,
    size DECIMAL(20,8) NOT NULL,
    entry_price DECIMAL(20,8) NOT NULL,
    mark_price DECIMAL(20,8),
    unrealized_pnl DECIMAL(20,8) DEFAULT 0,
    percentage DECIMAL(10,6) DEFAULT 0,
    leverage INT DEFAULT 1,
    margin DECIMAL(20,8) DEFAULT 0,
    maintenance_margin DECIMAL(20,8) DEFAULT 0,
    status ENUM('OPEN', 'CLOSED') DEFAULT 'OPEN',
    open_time TIMESTAMP NOT NULL,
    close_time TIMESTAMP NULL,
    closed_pnl DECIMAL(20,8) DEFAULT 0,
    strategy VARCHAR(100),
    notes TEXT,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    INDEX idx_symbol (symbol),
    INDEX idx_status (status),
    INDEX idx_open_time (open_time),
    UNIQUE KEY unique_open_position (symbol, position_side, status)
);

-- 交易记录表
CREATE TABLE IF NOT EXISTS trades (
    id BIGINT UNSIGNED AUTO_INCREMENT PRIMARY KEY,
    exchange_trade_id VARCHAR(255) UNIQUE NOT NULL,
    order_id BIGINT UNSIGNED NOT NULL,
    symbol VARCHAR(50) NOT NULL,
    side ENUM('BUY', 'SELL') NOT NULL,
    quantity DECIMAL(20,8) NOT NULL,
    price DECIMAL(20,8) NOT NULL,
    quote_qty DECIMAL(20,8) NOT NULL,
    commission DECIMAL(20,8) DEFAULT 0,
    commission_asset VARCHAR(20),
    realized_pnl DECIMAL(20,8) DEFAULT 0,
    is_maker BOOLEAN DEFAULT FALSE,
    position_side ENUM('BOTH', 'LONG', 'SHORT') DEFAULT 'BOTH',
    strategy VARCHAR(100),
    trade_time TIMESTAMP NOT NULL,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    INDEX idx_exchange_trade_id (exchange_trade_id),
    INDEX idx_order_id (order_id),
    INDEX idx_symbol (symbol),
    INDEX idx_trade_time (trade_time),
    FOREIGN KEY (order_id) REFERENCES orders(id) ON DELETE CASCADE
);

-- 账户信息表
CREATE TABLE IF NOT EXISTS accounts (
    id BIGINT UNSIGNED AUTO_INCREMENT PRIMARY KEY,
    total_wallet_balance DECIMAL(20,8) DEFAULT 0,
    total_unrealized_pnl DECIMAL(20,8) DEFAULT 0,
    total_margin_balance DECIMAL(20,8) DEFAULT 0,
    total_position_im DECIMAL(20,8) DEFAULT 0,
    total_open_order_im DECIMAL(20,8) DEFAULT 0,
    total_cross_wallet_balance DECIMAL(20,8) DEFAULT 0,
    available_balance DECIMAL(20,8) DEFAULT 0,
    max_withdraw_amount DECIMAL(20,8) DEFAULT 0,
    can_trade BOOLEAN DEFAULT TRUE,
    can_withdraw BOOLEAN DEFAULT TRUE,
    can_deposit BOOLEAN DEFAULT TRUE,
    update_time BIGINT,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    INDEX idx_updated_at (updated_at)
);

-- 余额表
CREATE TABLE IF NOT EXISTS balances (
    id BIGINT UNSIGNED AUTO_INCREMENT PRIMARY KEY,
    account_id BIGINT UNSIGNED NOT NULL,
    asset VARCHAR(20) NOT NULL,
    wallet_balance DECIMAL(20,8) DEFAULT 0,
    unrealized_pnl DECIMAL(20,8) DEFAULT 0,
    margin_balance DECIMAL(20,8) DEFAULT 0,
    maint_margin DECIMAL(20,8) DEFAULT 0,
    initial_margin DECIMAL(20,8) DEFAULT 0,
    position_im DECIMAL(20,8) DEFAULT 0,
    open_order_im DECIMAL(20,8) DEFAULT 0,
    cross_wallet_balance DECIMAL(20,8) DEFAULT 0,
    cross_un_pnl DECIMAL(20,8) DEFAULT 0,
    available_balance DECIMAL(20,8) DEFAULT 0,
    max_withdraw_amount DECIMAL(20,8) DEFAULT 0,
    margin_available BOOLEAN DEFAULT TRUE,
    update_time BIGINT,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    INDEX idx_account_id (account_id),
    INDEX idx_asset (asset),
    FOREIGN KEY (account_id) REFERENCES accounts(id) ON DELETE CASCADE
);

-- 交易对信息表
CREATE TABLE IF NOT EXISTS symbols (
    id BIGINT UNSIGNED AUTO_INCREMENT PRIMARY KEY,
    symbol VARCHAR(50) UNIQUE NOT NULL,
    pair VARCHAR(50) NOT NULL,
    contract_type VARCHAR(50),
    delivery_date BIGINT,
    onboard_date BIGINT,
    status VARCHAR(20) NOT NULL,
    maint_margin_percent DECIMAL(10,6),
    required_margin_percent DECIMAL(10,6),
    base_asset VARCHAR(20) NOT NULL,
    quote_asset VARCHAR(20) NOT NULL,
    margin_asset VARCHAR(20),
    price_precision INT,
    quantity_precision INT,
    base_asset_precision INT,
    quote_precision INT,
    underlying_type VARCHAR(50),
    trigger_protect DECIMAL(10,6),
    liquidation_fee DECIMAL(10,6),
    market_take_bound DECIMAL(10,6),
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    INDEX idx_symbol (symbol),
    INDEX idx_status (status)
);

-- 市场数据表
CREATE TABLE IF NOT EXISTS market_data (
    id BIGINT UNSIGNED AUTO_INCREMENT PRIMARY KEY,
    symbol VARCHAR(50) NOT NULL,
    price DECIMAL(20,8) NOT NULL,
    volume DECIMAL(20,8) NOT NULL,
    high DECIMAL(20,8),
    low DECIMAL(20,8),
    open DECIMAL(20,8),
    close DECIMAL(20,8),
    change DECIMAL(10,6),
    change_percent DECIMAL(10,6),
    timestamp BIGINT NOT NULL,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    INDEX idx_symbol (symbol),
    INDEX idx_timestamp (timestamp),
    INDEX idx_symbol_timestamp (symbol, timestamp)
);

-- 策略表
CREATE TABLE IF NOT EXISTS strategies (
    id BIGINT UNSIGNED AUTO_INCREMENT PRIMARY KEY,
    name VARCHAR(255) UNIQUE NOT NULL,
    type VARCHAR(100) NOT NULL,
    description TEXT,
    parameters JSON,
    is_active BOOLEAN DEFAULT TRUE,
    performance JSON,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    INDEX idx_name (name),
    INDEX idx_type (type),
    INDEX idx_active (is_active)
);

-- 风险指标表
CREATE TABLE IF NOT EXISTS risk_metrics (
    id BIGINT UNSIGNED AUTO_INCREMENT PRIMARY KEY,
    date DATE NOT NULL,
    total_pnl DECIMAL(20,8) DEFAULT 0,
    daily_pnl DECIMAL(20,8) DEFAULT 0,
    max_drawdown DECIMAL(10,6) DEFAULT 0,
    total_trades INT DEFAULT 0,
    winning_trades INT DEFAULT 0,
    losing_trades INT DEFAULT 0,
    win_rate DECIMAL(5,2) DEFAULT 0,
    avg_win DECIMAL(20,8) DEFAULT 0,
    avg_loss DECIMAL(20,8) DEFAULT 0,
    profit_factor DECIMAL(10,4) DEFAULT 0,
    sharpe_ratio DECIMAL(10,4) DEFAULT 0,
    var_95 DECIMAL(10,6) DEFAULT 0,
    max_leverage DECIMAL(5,2) DEFAULT 0,
    total_exposure DECIMAL(20,8) DEFAULT 0,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    UNIQUE KEY unique_date (date),
    INDEX idx_date (date)
);

-- 插入默认策略配置
INSERT IGNORE INTO strategies (name, type, description, parameters, is_active) VALUES
('Simple Moving Average', 'simple_moving_average', 'Basic SMA crossover strategy', 
 JSON_OBJECT('short_period', 10, 'long_period', 20, 'min_confidence', 0.7), TRUE),
('RSI Strategy', 'rsi', 'RSI overbought/oversold strategy', 
 JSON_OBJECT('period', 14, 'oversold', 30, 'overbought', 70, 'min_confidence', 0.6), TRUE),
('Grid Strategy', 'grid', 'Grid trading strategy', 
 JSON_OBJECT('grid_size', 0.01, 'num_grids', 10, 'min_confidence', 0.8), FALSE);

-- 插入默认交易配置
INSERT IGNORE INTO trading_configs (name, symbol, max_position, stop_loss, take_profit, leverage, risk_percent) VALUES
('BTCUSDT_Default', 'BTCUSDT', 1000.0, 2.0, 5.0, 5, 1.0),
('ETHUSDT_Default', 'ETHUSDT', 800.0, 2.0, 5.0, 5, 1.0),
('ADAUSDT_Default', 'ADAUSDT', 500.0, 2.5, 6.0, 3, 1.5);

-- 创建视图以便于查询
CREATE OR REPLACE VIEW active_positions AS
SELECT 
    p.*,
    (p.mark_price - p.entry_price) * p.size AS current_pnl,
    ((p.mark_price - p.entry_price) / p.entry_price) * 100 AS pnl_percentage
FROM positions p 
WHERE p.status = 'OPEN';

CREATE OR REPLACE VIEW daily_trading_summary AS
SELECT 
    DATE(created_at) as trade_date,
    symbol,
    COUNT(*) as total_trades,
    SUM(CASE WHEN side = 'BUY' THEN quantity ELSE 0 END) as total_buy_qty,
    SUM(CASE WHEN side = 'SELL' THEN quantity ELSE 0 END) as total_sell_qty,
    SUM(quote_qty) as total_volume,
    AVG(price) as avg_price
FROM trades 
GROUP BY DATE(created_at), symbol 
ORDER BY trade_date DESC, symbol;

COMMIT;
