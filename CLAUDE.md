# Credit Ledger - 信贷账务系统

## 项目概述
开源信贷账务核心系统，Go + MySQL + Redis，支持完整的贷款生命周期管理。

## 技术栈
- Go 1.21+，gorilla/mux 路由
- MySQL 8.0 / SQLite（开发），Redis 缓存
- 自研定点数库 pkg/decimal（4位小数，整数存储，避免浮点精度问题）

## 项目结构
```
cmd/server/          - 服务入口
internal/api/        - HTTP handlers + middleware + 统一响应
internal/service/    - 业务逻辑（loan/settlement/pm/coupon/batch）
internal/repository/ - 数据访问层（SQL）
internal/model/      - 数据模型
internal/calculator/ - 还款计算器（等额本息/等额本金/先息后本/一次性还本付息）
pkg/decimal/         - 高精度定点数
migrations/          - 数据库迁移 SQL
```

## 核心设计原则
- 所有表通过 loan_no 关联，不使用外键
- 还款分配顺序可配置：罚息 > 其他费用 > 利息 > 本金
- trial_date（试算日期）vs booking_date（入账日期）分离
- 所有核心表有审计字段（created_by/updated_by/created_at/updated_at）

## 代码规范
- 状态值用字符串常量（PENDING/DISBURSED/OVERDUE/REPAID）
- decimal 运算用 pkg/decimal 包，不要用 float64 直接算钱
- API 响应用统一格式：{code, message, data, timestamp}
- 错误不向客户端泄露内部细节

## 常用命令
```bash
go build ./...
go test ./...
go test ./internal/service/... -v
```

## 关键 API
```
POST /api/v1/loans                         - 创建借据
POST /api/v1/loans/{loan_no}/disburse      - 放款
POST /api/v1/repayments/trial              - 还款试算
POST /api/v1/repayments                    - 还款入账
POST /api/v1/loans/{loan_no}/early-settlement  - 提前结清
POST /api/v1/loans/{loan_no}/partial-repayment - 部分还款
POST /api/v1/loans/{loan_no}/penalty-waiver    - 罚息减免
POST /api/v1/loans/{loan_no}/extension         - 展期
POST /api/v1/loans/{loan_no}/write-off         - 坏账核销
POST /api/v1/coupons                           - 创建优惠券
POST /api/v1/coupons/trial                     - 优惠券试算
```
