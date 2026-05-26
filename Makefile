.PHONY: build run test clean docker-up docker-down

# 构建项目
build:
	go build -o bin/server ./cmd/server/

# 运行服务器
run:
	go run ./cmd/server/

# 运行测试
test:
	go test ./...

# 运行测试并显示覆盖率
test-coverage:
	go test -cover ./...

# 清理构建文件
clean:
	rm -rf bin/

# 启动依赖服务
docker-up:
	docker-compose up -d

# 停止依赖服务
docker-down:
	docker-compose down

# 代码格式化
fmt:
	go fmt ./...

# 静态检查
vet:
	go vet ./...

# 代码检查
lint: fmt vet

# 安装依赖
deps:
	go mod download

# 更新依赖
update-deps:
	go get -u ./...
	go mod tidy

# 生成文档
docs:
	godoc -http=:6060

# 运行开发环境
dev: docker-up
	@echo "等待数据库启动..."
	@sleep 5
	@echo "启动服务器..."
	@go run ./cmd/server/

# 数据库迁移
migrate:
	mysql -h localhost -u credit_user -p credit_ledger < migrations/001_initial_schema.sql

# 创建新的迁移文件
create-migration:
	@echo "创建新的迁移文件..."
	@read -p "输入迁移名称: " name; \
	touch migrations/$$(date +%Y%m%d%H%M%S)_$$name.sql

# 帮助
help:
	@echo "可用命令:"
	@echo "  build          - 构建项目"
	@echo "  run            - 运行服务器"
	@echo "  test           - 运行测试"
	@echo "  test-coverage  - 运行测试并显示覆盖率"
	@echo "  clean          - 清理构建文件"
	@echo "  docker-up      - 启动依赖服务"
	@echo "  docker-down    - 停止依赖服务"
	@echo "  fmt            - 代码格式化"
	@echo "  vet            - 静态检查"
	@echo "  lint           - 代码检查 (fmt + vet)"
	@echo "  deps           - 安装依赖"
	@echo "  update-deps    - 更新依赖"
	@echo "  docs           - 生成文档"
	@echo "  dev            - 运行开发环境"
	@echo "  migrate        - 运行数据库迁移"
	@echo "  create-migration - 创建新的迁移文件"
	@echo "  help           - 显示帮助"