# 开源信贷账务系统 (Credit Ledger)

一个基于 Go + MySQL + Redis 的开源信贷账务系统，支持完整的贷款生命周期管理。

## 功能特性

### 核心功能
- **放款管理**：放款试算、入账处理
- **还款管理**：还款试算、入账处理、部分还款、提前结清
- **还款计划**：自动生成还款计划、计划查询
- **计息系统**：按日计息、高精度计算

### 还款类型支持
- 等额本息 (Equal Installment)
- 等额本金 (Equal Principal)
- 按月付息到期还本 (Interest-Only)
- 一次性还本付息 (Bullet)

### 费用计算体系
- 手续费
- 管理费
- 逾期罚息
- 提前还款手续费
- 费用减免机制

## 技术栈

- **后端**：Go 1.21+
- **数据库**：MySQL 8.0
- **缓存**：Redis 7
- **API**：RESTful JSON API
- **部署**：Docker, Docker Compose

## 项目结构

```
credit-ledger/
├── cmd/                    # 主程序入口
│   └── server/            # HTTP 服务器
├── internal/              # 内部包
│   ├── model/            # 数据模型
│   ├── service/          # 业务逻辑
│   ├── repository/       # 数据访问
│   ├── calculator/       # 计算引擎
│   └── api/             # API 路由
├── pkg/                  # 公共包
│   ├── database/        # 数据库连接
│   ├── cache/          # 缓存
│   └── logger/         # 日志
├── config/              # 配置
├── migrations/          # 数据库迁移
├── tests/              # 测试
├── docs/               # 文档
├── scripts/            # 脚本
├── docker-compose.yml  # Docker 编排
└── .env.example        # 环境变量示例
```

## 快速开始

### 1. 克隆项目
```bash
git clone https://github.com/yourorg/credit-ledger.git
cd credit-ledger
```

### 2. 启动依赖服务
```bash
docker-compose up -d
```

### 3. 配置环境变量
```bash
cp .env.example .env
# 编辑 .env 文件，配置数据库和 Redis 连接
```

### 4. 运行数据库迁移
```bash
# 使用 MySQL 客户端执行迁移脚本
mysql -h localhost -u credit_user -p credit_ledger < migrations/001_initial_schema.sql
```

### 5. 启动服务
```bash
go run cmd/server/main.go
```

## API 文档

### 健康检查
```
GET /health
```

### 贷款管理
```
POST /api/v1/loans                    # 创建贷款
GET  /api/v1/loans/{id}               # 获取贷款详情
POST /api/v1/loans/{id}/disburse      # 放款
GET  /api/v1/loans/{id}/repayment-schedules  # 获取还款计划
```

### 还款管理
```
POST /api/v1/repayments/trial         # 还款试算
POST /api/v1/repayments               # 还款入账
```

### 查询接口
```
GET /api/v1/queries/loans/{id}/details      # 贷款详情查询
GET /api/v1/queries/loans/{id}/transactions # 交易记录查询
```

### 管理接口
```
POST /api/v1/admin/fee-configs        # 创建费用配置
GET  /api/v1/admin/fee-configs        # 获取费用配置列表
```

## 开发指南

### 运行测试
```bash
go test ./...
```

### 代码格式化
```bash
go fmt ./...
```

### 静态检查
```bash
go vet ./...
```

## 部署

### Docker 部署
```bash
# 构建镜像
docker build -t credit-ledger .

# 运行容器
docker run -p 8080:8080 credit-ledger
```

### Kubernetes 部署
参考 `docs/kubernetes.md` 文档。

## 贡献指南

1. Fork 项目
2. 创建功能分支 (`git checkout -b feature/AmazingFeature`)
3. 提交更改 (`git commit -m 'Add some AmazingFeature'`)
4. 推送到分支 (`git push origin feature/AmazingFeature`)
5. 打开 Pull Request

## 许可证

本项目采用 MIT 许可证 - 查看 [LICENSE](LICENSE) 文件了解详情。

## 联系方式

- 项目主页: https://github.com/yourorg/credit-ledger
- 问题反馈: https://github.com/yourorg/credit-ledger/issues

## 致谢

感谢所有为这个项目做出贡献的开发者。