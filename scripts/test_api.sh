#!/bin/bash

# 测试脚本
echo "测试信贷账务系统 API"

# 等待服务器启动
echo "等待服务器启动..."
sleep 3

# 测试健康检查
echo "测试健康检查..."
curl -s http://localhost:8080/health | jq .

# 测试创建贷款
echo "测试创建贷款..."
curl -s -X POST http://localhost:8080/api/v1/loans \
  -H "Content-Type: application/json" \
  -d '{
    "loan_no": "LOAN2024001",
    "principal": 100000,
    "annual_interest_rate": 0.05,
    "term_months": 12,
    "repayment_type": "EQUAL_INSTALLMENT"
  }' | jq .

# 测试获取贷款详情
echo "测试获取贷款详情..."
curl -s http://localhost:8080/api/v1/loans/1 | jq .

# 测试放款
echo "测试放款..."
curl -s -X POST http://localhost:8080/api/v1/loans/1/disburse \
  -H "Content-Type: application/json" \
  -d '{
    "disbursement_date": "2024-01-01"
  }' | jq .

# 测试获取还款计划
echo "测试获取还款计划..."
curl -s http://localhost:8080/api/v1/loans/1/repayment-schedules | jq .

# 测试还款试算
echo "测试还款试算..."
curl -s -X POST http://localhost:8080/api/v1/repayments/trial \
  -H "Content-Type: application/json" \
  -d '{
    "loan_id": 1,
    "repayment_date": "2024-02-01",
    "amount": 8560.75,
    "repayment_type": "FULL"
  }' | jq .

# 测试还款入账
echo "测试还款入账..."
curl -s -X POST http://localhost:8080/api/v1/repayments \
  -H "Content-Type: application/json" \
  -d '{
    "loan_id": 1,
    "repayment_date": "2024-02-01",
    "amount": 8560.75,
    "repayment_type": "FULL"
  }' | jq .

echo "测试完成"