# Credit Ledger - 开源信贷账务系统

[![Go](https://img.shields.io/badge/Go-1.21+-00ADD8?style=flat&logo=go)](https://go.dev/)
[![License](https://img.shields.io/badge/License-MIT-blue.svg)](LICENSE)

一个轻量级、高精度的信贷账务核心系统，支持完整的贷款生命周期管理。

## ✨ 功能特性

### 放款管理
- 借据创建与放款
- 自动生成还款计划
- 支持多种还款类型

### 还款管理
- 还款试算（不落库，实时计算）
- 还款入账（按规则自动分配）
- 支持历史入账（补账）
- 提前结清（试算 + 入账）
- 部分还款（试算 + 入账）

### 计息与费用
- 按日计息（复利/单利可配置）
- 逾期罚息自动计算
- 动态费项配置体系
- 费用触发时机可配（放款时/还款时/每日/逾期/提前还款）

### 还款类型
| 类型 | 编码 | 说明 |
|------|------|------|
| 等额本息 | EQUAL_INSTALLMENT | 每月还款金额固定 |
| 等额本金 | EQUAL_PRINCIPAL | 每月本金固定，利息递减 |
| 先息后本 | INTEREST_FIRST | 按月付息，到期还本 |
| 一次性还本付息 | BULLET | 到期一次性还本付息 |

### 还款分配规则
支持配置化的还款金额分配顺序，默认优先级：

```
罚息 > 其他费用 > 利息 > 本金
```

### 还款入账
- `trial_date`：试算基准日期（用于计算应还金额）
- `booking_date`：系统入账处理日期
- 支持入历史帐（`is_backdated` + `backdated_reason`）

### 日切跑批
- 批次表 + 任务明细表模式
- 游标分页（避免 OOM）
- `FOR UPDATE SKIP LOCKED` 并发安全

## 🏗 技术架构

```
┌─────────────────────────────────────────┐
│              API Layer (HTTP)            │
│          gorilla/mux + handlers          │
├─────────────────────────────────────────┤
│            Service Layer                 │
│    放款 / 还款 / 结清 / 跑批 / 试算      │
├─────────────────────────────────────────┤
│          Calculator Engine               │
│   等额本息 / 等额本金 / 先息后本 / 一期   │
├─────────────────────────────────────────┤
│          Repository Layer                │
│              SQL (MySQL/SQLite)          │
├─────────────────────────────────────────┤
│         pkg/decimal (高精度定点数)        │
│              Redis (缓存)                │
└─────────────────────────────────────────┘
```

### 技术栈
- **语言**：Go 1.21+
- **数据库**：MySQL / SQLite（开发环境）
- **缓存**：Redis
- **路由**：gorilla/mux
- **精度**：自研定点数库（4位小数，整数存储）

## 📁 项目结构

```
credit-ledger/
├── cmd/server/          # 服务入口
│   └── main.go
├── config/              # 配置管理
├── docs/                # API 文档
│   └── api.yaml         # OpenAPI 3.0
├── internal/
│   ├── api/             # HTTP 接口层
│   │   ├── router.go
│   │   └── loan_handlers.go
│   ├── calculator/      # 计算引擎
│   │   ├── calculator.go        # 计算器接口
│   │   ├── equal_installment.go # 等额本息
│   │   ├── equal_principal.go   # 等额本金
│   │   ├── interest_first.go    # 先息后本
│   │   ├── bullet_calculator.go # 一次性还本付息
│   │   ├── daily_interest.go    # 按日计息
│   │   └── fee_calculator.go    # 费用计算
│   ├── model/           # 数据模型
│   │   └── models.go
│   ├── repository/      # 数据访问层
│   │   └── repository.go
│   └── service/         # 业务逻辑层
│       ├── loan_service.go          # 放款/还款/试算
│       ├── settlement_service.go    # 提前结清/部分还款
│       └── batch_service.go         # 日切跑批
└── pkg/
    ├── decimal/         # 高精度定点数库
    ├── database/        # 数据库连接
    ├── cache/           # Redis 缓存
    └── logger/          # 日志
```

## 🚀 快速开始

### 环境要求
- Go 1.21+
- MySQL 8.0+ 或 SQLite（开发）

### 安装

```bash
git clone https://github.com/lzwcyd/credit-ledger.git
cd credit-ledger
go mod tidy
```

### 运行

```bash
go run cmd/server/main.go
```

服务默认启动在 `http://localhost:8080`

### 健康检查

```bash
curl http://localhost:8080/health
# {"status": "healthy"}
```

## 📖 API 文档

完整的 API 文档见 [docs/api.yaml](docs/api.yaml)（OpenAPI 3.0 格式）

### 核心接口一览

#### 借据管理
| 方法 | 路径 | 说明 |
|------|------|------|
| POST | `/api/v1/loans` | 创建借据 |
| GET | `/api/v1/loans/{loan_no}` | 获取借据详情 |
| POST | `/api/v1/loans/{loan_no}/disburse` | 放款 |

#### 还款计划
| 方法 | 路径 | 说明 |
|------|------|------|
| GET | `/api/v1/loans/{loan_no}/plans` | 获取还款计划列表 |
| GET | `/api/v1/loans/{loan_no}/plans/summary` | 获取还款计划汇总 |

#### 还款
| 方法 | 路径 | 说明 |
|------|------|------|
| POST | `/api/v1/repayments/trial` | 还款试算 |
| POST | `/api/v1/repayments` | 还款入账 |

#### 提前结清
| 方法 | 路径 | 说明 |
|------|------|------|
| POST | `/api/v1/loans/{loan_no}/early-settlement/trial` | 提前结清试算 |
| POST | `/api/v1/loans/{loan_no}/early-settlement` | 提前结清入账 |

#### 部分还款
| 方法 | 路径 | 说明 |
|------|------|------|
| POST | `/api/v1/loans/{loan_no}/partial-repayment/trial` | 部分还款试算 |
| POST | `/api/v1/loans/{loan_no}/partial-repayment` | 部分还款入账 |

## 💡 使用示例

### 1. 创建借据

```bash
curl -X POST http://localhost:8080/api/v1/loans \
  -H "Content-Type: application/json" \
  -d '{
    "loan_no": "LOAN20260001",
    "principal": 100000,
    "annual_rate": 0.06,
    "term_months": 12,
    "repayment_type_code": "EQUAL_INSTALLMENT",
    "value_date": "2026-01-01",
    "first_due_date": "2026-02-01",
    "maturity_date": "2027-01-01",
    "created_by": "system"
  }'
```

### 2. 放款

```bash
curl -X POST http://localhost:8080/api/v1/loans/LOAN20260001/disburse \
  -H "Content-Type: application/json" \
  -d '{
    "disburse_date": "2026-01-01",
    "disburse_amount": 100000,
    "created_by": "system"
  }'
```

### 3. 还款试算

```bash
curl -X POST http://localhost:8080/api/v1/repayments/trial \
  -H "Content-Type: application/json" \
  -d '{
    "loan_no": "LOAN20260001",
    "trial_date": "2026-03-01"
  }'
```

### 4. 提前结清试算

```bash
curl -X POST http://localhost:8080/api/v1/loans/LOAN20260001/early-settlement/trial \
  -H "Content-Type: application/json" \
  -d '{
    "trial_date": "2026-06-01"
  }'
```

### 5. 部分还款

```bash
curl -X POST http://localhost:8080/api/v1/loans/LOAN20260001/partial-repayment \
  -H "Content-Type: application/json" \
  -d '{
    "amount": 5000,
    "trial_date": "2026-03-01",
    "booking_date": "2026-03-01",
    "created_by": "system"
  }'
```

## 🗄 数据库表结构

核心表（13 张）：

| 表名 | 说明 |
|------|------|
| `loans` | 借据主表 |
| `loan_changes` | 借据变更记录 |
| `plans` | 还款计划 |
| `plan_changes` | 还款计划变更记录 |
| `plan_other_fees` | 其他费用明细 |
| `daily_calculations` | 每日计算明细（利息/罚息） |
| `repayments` | 还款记录 |
| `repayment_details` | 还款入账明细 |
| `repayment_types` | 还款类型配置 |
| `fee_configs` | 费项配置 |
| `allocation_rules` | 分配规则 |
| `allocation_rule_items` | 分配规则明细 |
| `batch_jobs` | 跑批批次 |

所有表通过 `loan_no` 业务字段关联，不使用数据库外键。

## ⚙️ 费项配置示例

```sql
-- 利息（每日累计）
INSERT INTO fee_configs (code, name, calc_type, calc_base, value, trigger_type, is_daily_accumulate, fee_category)
VALUES ('INTEREST', '利息', 'DAILY_RATE', 'REMAINING_PRINCIPAL', '0.0001644', 'DAILY', TRUE, 'INTEREST');

-- 逾期罚息（每日累计，1.5倍利率）
INSERT INTO fee_configs (code, name, calc_type, calc_base, value, trigger_type, is_daily_accumulate, fee_category)
VALUES ('OVERDUE_PENALTY', '逾期罚息', 'DAILY_RATE', 'OVERDUE_AMOUNT', '0.0002466', 'OVERDUE', TRUE, 'PENALTY');

-- 手续费（放款时收取）
INSERT INTO fee_configs (code, name, calc_type, calc_base, value, trigger_type, is_daily_accumulate, fee_category)
VALUES ('HANDLING_FEE', '手续费', 'PERCENTAGE', 'PRINCIPAL', '1', 'DISBURSEMENT', FALSE, 'OTHER_FEE');

-- 提前结清手续费
INSERT INTO fee_configs (code, name, calc_type, calc_base, value, trigger_type, is_daily_accumulate, fee_category)
VALUES ('EARLY_SETTLEMENT_FEE', '提前结清手续费', 'PERCENTAGE', 'REMAINING_PRINCIPAL', '1', 'EARLY_REPAYMENT', FALSE, 'OTHER_FEE');
```

## 🧪 测试

```bash
# 运行所有测试
go test ./...

# 运行指定包测试
go test ./internal/service/... -v

# 运行指定测试
go test ./pkg/decimal/... -run TestDecimal_Round -v
```

## 📋 开发计划

- [x] 阶段一：项目初始化与基础架构
- [x] 阶段二：核心计算引擎
- [x] 阶段三：业务逻辑层
- [x] 提前结清 + 部分还款
- [x] 单元测试 + API 文档
- [ ] 数据库迁移脚本（正式 MySQL）
- [ ] 日切跑批定时任务
- [ ] 分布式锁支持
- [ ] 更多费用场景
- [ ] Docker 部署支持

## 🤝 贡献

欢迎提交 Issue 和 Pull Request！

## 📄 License

MIT License
