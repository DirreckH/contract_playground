# 交易机器人 Makefile

.PHONY: help build run test clean docker-up docker-down setup config-test

# 默认目标
help:
	@echo "加密货币交易机器人 - 可用命令:"
	@echo ""
	@echo "  build         - 编译交易机器人"
	@echo "  run           - 运行交易机器人"
	@echo "  test          - 运行测试"
	@echo "  config-test   - 测试配置加载"
	@echo "  clean         - 清理编译文件"
	@echo "  setup         - 初始化项目（下载依赖）"
	@echo "  docker-up     - 启动Docker环境（MySQL + Redis）"
	@echo "  docker-down   - 停止Docker环境"
	@echo "  docker-logs   - 查看Docker日志"
	@echo "  mysql-cli     - 连接到MySQL命令行"
	@echo "  redis-cli     - 连接到Redis命令行"
	@echo ""

# 编译项目
build:
	@echo "编译交易机器人..."
	go build -o trader cmd/trader/main.go
	@echo "编译完成！"

# 运行交易机器人
run:
	@echo "启动交易机器人..."
	@echo "⚠️  请确保已经配置好环境变量和数据库"
	./trader

# 测试配置加载
config-test:
	@echo "测试配置加载..."
	go run cmd/test/main.go

# 运行测试
test:
	@echo "运行测试..."
	go test ./... -v

# 清理编译文件
clean:
	@echo "清理编译文件..."
	rm -f trader
	go clean
	@echo "清理完成！"

# 初始化项目
setup:
	@echo "初始化项目..."
	@echo "下载Go依赖..."
	go mod download
	go mod tidy
	@echo "项目初始化完成！"

# 启动Docker环境
docker-up:
	@echo "启动Docker环境..."
	docker-compose up -d
	@echo "等待服务启动..."
	@sleep 10
	@echo "Docker环境已启动！"
	@echo ""
	@echo "服务地址:"
	@echo "  MySQL:        localhost:3306"
	@echo "  Redis:        localhost:6379"
	@echo "  phpMyAdmin:   http://localhost:8080"
	@echo "  Redis Commander: http://localhost:8081"
	@echo ""

# 停止Docker环境
docker-down:
	@echo "停止Docker环境..."
	docker-compose down
	@echo "Docker环境已停止！"

# 查看Docker日志
docker-logs:
	docker-compose logs -f

# 连接MySQL命令行
mysql-cli:
	@echo "连接到MySQL..."
	docker exec -it trading_mysql mysql -u trader -ppassword trading_bot

# 连接Redis命令行
redis-cli:
	@echo "连接到Redis..."
	docker exec -it trading_redis redis-cli

# 重置数据库
reset-db:
	@echo "重置数据库..."
	docker exec -i trading_mysql mysql -u root -prootpassword -e "DROP DATABASE IF EXISTS trading_bot; CREATE DATABASE trading_bot CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci;"
	docker exec -i trading_mysql mysql -u root -prootpassword trading_bot < migrations/001_initial_schema.sql
	@echo "数据库重置完成！"

# 查看数据库状态
db-status:
	@echo "数据库连接状态:"
	@docker exec trading_mysql mysqladmin -u trader -ppassword ping
	@echo ""
	@echo "Redis连接状态:"
	@docker exec trading_redis redis-cli ping

# 检查环境
check-env:
	@echo "检查环境配置..."
	@echo "Go版本:"
	@go version
	@echo ""
	@echo "Docker版本:"
	@docker --version
	@echo ""
	@echo "Docker Compose版本:"
	@docker-compose --version
	@echo ""

# 生成示例环境变量文件
generate-env:
	@echo "生成示例环境变量文件..."
	@echo "# 币安API配置" > .env.example
	@echo "BINANCE_API_KEY=your_binance_api_key_here" >> .env.example
	@echo "BINANCE_SECRET_KEY=your_binance_secret_key_here" >> .env.example
	@echo "" >> .env.example
	@echo "# 数据库配置" >> .env.example
	@echo "MYSQL_DSN=trader:password@tcp(localhost:3306)/trading_bot?charset=utf8mb4&parseTime=True&loc=Local" >> .env.example
	@echo "REDIS_ADDR=localhost:6379" >> .env.example
	@echo "REDIS_PASSWORD=" >> .env.example
	@echo ""
	@echo "示例环境变量文件已生成: .env.example"
	@echo "请复制并重命名为 .env，然后填入真实的API密钥"

# 完整设置（推荐新用户使用）
full-setup: check-env setup generate-env docker-up
	@echo ""
	@echo "🎉 完整设置已完成！"
	@echo ""
	@echo "下一步:"
	@echo "1. 复制 .env.example 为 .env"
	@echo "2. 在 .env 中填入您的币安API密钥"
	@echo "3. 运行 'make config-test' 测试配置"
	@echo "4. 运行 'make build' 编译程序"
	@echo "5. 运行 'make run' 启动交易机器人"
	@echo ""

# 开发模式（包含实时重载）
dev:
	@echo "启动开发模式..."
	go run cmd/trader/main.go

# 查看项目结构
tree:
	@echo "项目结构:"
	tree -I 'vendor|.git|trader'
