package config

import (
	"fmt"
	"strings"

	"github.com/spf13/viper"
)

// Config represents the application configuration
type Config struct {
	Exchange ExchangeConfig `mapstructure:"exchange"`
	Trading  TradingConfig  `mapstructure:"trading"`
	Database DatabaseConfig `mapstructure:"database"`
	Logger   LoggerConfig   `mapstructure:"logger"`
}

// ExchangeConfig holds exchange-specific configuration
type ExchangeConfig struct {
	Name      string `mapstructure:"name"`
	APIKey    string `mapstructure:"api_key"`
	SecretKey string `mapstructure:"secret_key"`
	Testnet   bool   `mapstructure:"testnet"`
	BaseURL   string `mapstructure:"base_url"`
}

// TradingConfig holds trading strategy and risk management configuration
type TradingConfig struct {
	Symbols              []string  `mapstructure:"symbols"`
	MaxPositionSize      float64   `mapstructure:"max_position_size"`
	StopLossPercent      float64   `mapstructure:"stop_loss_percent"`
	TakeProfitPercent    float64   `mapstructure:"take_profit_percent"`
	MaxDailyLoss         float64   `mapstructure:"max_daily_loss"`
	TradingInterval      int       `mapstructure:"trading_interval_seconds"`
	MinOrderValue        float64   `mapstructure:"min_order_value"`
	MaxLeverage          int       `mapstructure:"max_leverage"`
	RiskPerTrade         float64   `mapstructure:"risk_per_trade_percent"`
	EnablePaperTrading   bool      `mapstructure:"enable_paper_trading"`
	Strategy             StrategyConfig `mapstructure:"strategy"`
}

// StrategyConfig holds trading strategy parameters
type StrategyConfig struct {
	Type                string                 `mapstructure:"type"`
	Parameters          map[string]interface{} `mapstructure:"parameters"`
	EnableSignalFilters bool                   `mapstructure:"enable_signal_filters"`
}

// DatabaseConfig holds database connection configuration
type DatabaseConfig struct {
	MySQL MySQLConfig `mapstructure:"mysql"`
	Redis RedisConfig `mapstructure:"redis"`
}

// MySQLConfig holds MySQL-specific configuration
type MySQLConfig struct {
	DSN             string `mapstructure:"dsn"`
	MaxOpenConns    int    `mapstructure:"max_open_conns"`
	MaxIdleConns    int    `mapstructure:"max_idle_conns"`
	ConnMaxLifetime int    `mapstructure:"conn_max_lifetime_minutes"`
}

// RedisConfig holds Redis-specific configuration
type RedisConfig struct {
	Addr     string `mapstructure:"addr"`
	Password string `mapstructure:"password"`
	DB       int    `mapstructure:"db"`
	PoolSize int    `mapstructure:"pool_size"`
}

// LoggerConfig holds logging configuration
type LoggerConfig struct {
	Level  string `mapstructure:"level"`
	Format string `mapstructure:"format"`
	Output string `mapstructure:"output"`
}

// Load reads and parses the configuration from file and environment variables
func Load() (*Config, error) {
	viper.SetConfigName("config")
	viper.SetConfigType("yaml")
	viper.AddConfigPath("./config")
	viper.AddConfigPath(".")

	// Set environment variable prefix
	viper.SetEnvPrefix("TRADER")
	viper.SetEnvKeyReplacer(strings.NewReplacer(".", "_"))
	viper.AutomaticEnv()

	// Set default values
	setDefaults()

	// Read configuration file
	if err := viper.ReadInConfig(); err != nil {
		if _, ok := err.(viper.ConfigFileNotFoundError); ok {
			// Config file not found; use defaults and environment variables
		} else {
			return nil, fmt.Errorf("error reading config file: %w", err)
		}
	}

	var config Config
	if err := viper.Unmarshal(&config); err != nil {
		return nil, fmt.Errorf("error unmarshaling config: %w", err)
	}

	// Validate configuration
	if err := validateConfig(&config); err != nil {
		return nil, fmt.Errorf("configuration validation failed: %w", err)
	}

	return &config, nil
}

// setDefaults sets default configuration values
func setDefaults() {
	// Exchange defaults
	viper.SetDefault("exchange.name", "binance")
	viper.SetDefault("exchange.testnet", true)
	viper.SetDefault("exchange.base_url", "")

	// Trading defaults
	viper.SetDefault("trading.symbols", []string{"BTCUSDT", "ETHUSDT"})
	viper.SetDefault("trading.max_position_size", 1000.0)
	viper.SetDefault("trading.stop_loss_percent", 2.0)
	viper.SetDefault("trading.take_profit_percent", 5.0)
	viper.SetDefault("trading.max_daily_loss", 500.0)
	viper.SetDefault("trading.trading_interval_seconds", 60)
	viper.SetDefault("trading.min_order_value", 10.0)
	viper.SetDefault("trading.max_leverage", 5)
	viper.SetDefault("trading.risk_per_trade_percent", 1.0)
	viper.SetDefault("trading.enable_paper_trading", true)
	viper.SetDefault("trading.strategy.type", "simple_moving_average")
	viper.SetDefault("trading.strategy.enable_signal_filters", true)

	// Database defaults
	viper.SetDefault("database.mysql.max_open_conns", 25)
	viper.SetDefault("database.mysql.max_idle_conns", 5)
	viper.SetDefault("database.mysql.conn_max_lifetime_minutes", 30)
	viper.SetDefault("database.redis.db", 0)
	viper.SetDefault("database.redis.pool_size", 10)

	// Logger defaults
	viper.SetDefault("logger.level", "info")
	viper.SetDefault("logger.format", "json")
	viper.SetDefault("logger.output", "stdout")
}

// validateConfig validates the configuration values
func validateConfig(config *Config) error {
	// Validate exchange configuration
	if config.Exchange.APIKey == "" {
		return fmt.Errorf("exchange API key is required")
	}
	if config.Exchange.SecretKey == "" {
		return fmt.Errorf("exchange secret key is required")
	}

	// Validate trading configuration
	if len(config.Trading.Symbols) == 0 {
		return fmt.Errorf("at least one trading symbol is required")
	}
	if config.Trading.MaxPositionSize <= 0 {
		return fmt.Errorf("max position size must be positive")
	}
	if config.Trading.StopLossPercent <= 0 || config.Trading.StopLossPercent > 50 {
		return fmt.Errorf("stop loss percent must be between 0 and 50")
	}
	if config.Trading.TakeProfitPercent <= 0 || config.Trading.TakeProfitPercent > 100 {
		return fmt.Errorf("take profit percent must be between 0 and 100")
	}
	if config.Trading.MaxLeverage < 1 || config.Trading.MaxLeverage > 125 {
		return fmt.Errorf("max leverage must be between 1 and 125")
	}
	if config.Trading.RiskPerTrade < 0.1 || config.Trading.RiskPerTrade > 10 {
		return fmt.Errorf("risk per trade percent must be between 0.1 and 10")
	}

	// Validate database configuration
	if config.Database.MySQL.DSN == "" {
		return fmt.Errorf("MySQL DSN is required")
	}
	if config.Database.Redis.Addr == "" {
		return fmt.Errorf("Redis address is required")
	}

	return nil
}
