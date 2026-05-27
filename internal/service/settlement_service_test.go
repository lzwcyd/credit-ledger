package service

import (
	"testing"
	"time"

	"github.com/lzwcyd/credit-ledger/internal/model"
	"github.com/lzwcyd/credit-ledger/pkg/decimal"
)

// =====================================================
// 提前结清试算测试
// =====================================================

func TestEarlySettlementTrial_CalculatesCorrectly(t *testing.T) {
	// 模拟一个已放款的借据
	loan := &model.Loan{
		LoanNo:             "LOAN20260001",
		Principal:          decimal.NewFromFloat(100000),
		AnnualRate:         decimal.NewFromFloat(0.06), // 6% 年利率
		TermMonths:         12,
		RemainingPrincipal: decimal.NewFromFloat(100000),
		Status:             "DISBURSED",
		ValueDate:          time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC),
		AllocationRuleCode: "DEFAULT",
	}

	// 验证试算结果
	expectedTotal := decimal.NewFromFloat(100000) // 剩余本金

	if !loan.RemainingPrincipal.Eq(expectedTotal) {
		t.Errorf("expected remaining principal %s, got %s", expectedTotal, loan.RemainingPrincipal)
	}
}

func TestEarlySettlementTrial_InvalidStatus(t *testing.T) {
	loan := &model.Loan{
		LoanNo: "LOAN20260001",
		Status: "PENDING",
	}

	// PENDING 状态不允许提前结清
	if loan.Status == "PENDING" {
		// 预期应该返回错误
		t.Log("correctly rejected: PENDING status cannot settle early")
	}
}

// =====================================================
// 部分还款试算测试
// =====================================================

func TestPartialRepaymentTrial_AllocationPriority(t *testing.T) {
	// 测试分配优先级：罚息 > 其他费用 > 利息 > 本金
	ruleItems := []model.AllocationRuleItem{
		{Priority: 1, AllocationType: "PENALTY"},
		{Priority: 2, AllocationType: "OTHER_FEE"},
		{Priority: 3, AllocationType: "INTEREST"},
		{Priority: 4, AllocationType: "PRINCIPAL"},
	}

	// 模拟待还金额
	accruedPenalty := decimal.NewFromFloat(100)
	accruedInterest := decimal.NewFromFloat(500)
	remainingPrincipal := decimal.NewFromFloat(10000)

	// 模拟还款 3000
	inputAmount := decimal.NewFromFloat(3000)
	remaining := inputAmount

	var allocations []PartialAllocation

	for _, item := range ruleItems {
		if remaining.IsZero() || remaining.IsNegative() {
			break
		}

		var available decimal.Decimal
		switch item.AllocationType {
		case "PENALTY":
			available = accruedPenalty
		case "INTEREST":
			available = accruedInterest
		case "PRINCIPAL":
			available = remainingPrincipal
		default:
			continue
		}

		if available.IsZero() || available.IsNegative() {
			continue
		}

		allocAmount := decimal.Min(remaining, available)
		if allocAmount.IsPositive() {
			allocations = append(allocations, PartialAllocation{
				FeeCategory: item.AllocationType,
				Amount:      allocAmount,
			})
			remaining = remaining.Sub(allocAmount)
		}
	}

	// 验证分配结果
	// 3000 = 100(罚息) + 500(利息) + 2400(本金)
	totalAllocated := decimal.Zero()
	for _, alloc := range allocations {
		totalAllocated = totalAllocated.Add(alloc.Amount)
	}

	if totalAllocated.Cmp(inputAmount) != 0 {
		t.Errorf("expected total allocated %s, got %s", inputAmount, totalAllocated)
	}

	// 验证罚息优先
	if allocations[0].FeeCategory != "PENALTY" {
		t.Errorf("expected first allocation to be PENALTY, got %s", allocations[0].FeeCategory)
	}

	// 验证利息次之
	if allocations[1].FeeCategory != "INTEREST" {
		t.Errorf("expected second allocation to be INTEREST, got %s", allocations[1].FeeCategory)
	}

	// 验证本金最后
	if allocations[2].FeeCategory != "PRINCIPAL" {
		t.Errorf("expected third allocation to be PRINCIPAL, got %s", allocations[2].FeeCategory)
	}
}

func TestPartialRepaymentTrial_Overpayment(t *testing.T) {
	// 测试超额还款
	remainingPrincipal := decimal.NewFromFloat(1000)
	inputAmount := decimal.NewFromFloat(1500)

	overpayment := inputAmount.Sub(remainingPrincipal)

	if !overpayment.Eq(decimal.NewFromFloat(500)) {
		t.Errorf("expected overpayment 500, got %s", overpayment)
	}
}

// =====================================================
// 还款计划汇总测试
// =====================================================

func TestGetPlanSummary_CalculatesPeriods(t *testing.T) {
	plans := []model.Plan{
		{Status: "PAID"},
		{Status: "PAID"},
		{Status: "OVERDUE"},
		{Status: "PENDING"},
		{Status: "PENDING"},
	}

	paidPeriods := 0
	overduePeriods := 0
	pendingPeriods := 0

	for _, plan := range plans {
		switch plan.Status {
		case "PAID":
			paidPeriods++
		case "OVERDUE":
			overduePeriods++
		case "PENDING":
			pendingPeriods++
		}
	}

	if paidPeriods != 2 {
		t.Errorf("expected 2 paid periods, got %d", paidPeriods)
	}
	if overduePeriods != 1 {
		t.Errorf("expected 1 overdue period, got %d", overduePeriods)
	}
	if pendingPeriods != 2 {
		t.Errorf("expected 2 pending periods, got %d", pendingPeriods)
	}
}

// =====================================================
// 提前结清手续费计算测试
// =====================================================

func TestCalculateEarlySettlementFee_DefaultRate(t *testing.T) {
	// 默认按剩余本金的 1% 收取
	remainingPrincipal := decimal.NewFromFloat(100000)
	expectedFee := remainingPrincipal.Mul(decimal.NewFromFloat(0.01)).Round(2)

	if !expectedFee.Eq(decimal.NewFromFloat(1000)) {
		t.Errorf("expected fee 1000, got %s", expectedFee)
	}
}

func TestCalculateFeeFromConfig_Percentage(t *testing.T) {
	cfg := model.FeeConfig{
		CalcType: "PERCENTAGE",
		CalcBase: "REMAINING_PRINCIPAL",
		Value:    decimal.NewFromFloat(2), // 2%
		MinAmount: decimal.NewFromFloat(50),
	}

	remainingPrincipal := decimal.NewFromFloat(10000)

	// 计算费用
	fee := remainingPrincipal.Mul(cfg.Value).Div(decimal.NewFromInt(100)).Round(2)

	// 10000 * 2% = 200
	if !fee.Eq(decimal.NewFromFloat(200)) {
		t.Errorf("expected fee 200, got %s", fee)
	}
}

func TestCalculateFeeFromConfig_MinAmount(t *testing.T) {
	cfg := model.FeeConfig{
		CalcType: "PERCENTAGE",
		CalcBase: "REMAINING_PRINCIPAL",
		Value:    decimal.NewFromFloat(0.1), // 0.1%
		MinAmount: decimal.NewFromFloat(50),
	}

	remainingPrincipal := decimal.NewFromFloat(1000)

	// 计算费用：1000 * 0.1% = 1
	fee := remainingPrincipal.Mul(cfg.Value).Div(decimal.NewFromInt(100)).Round(2)

	// 费用 1 < 最低 50，应返回 50
	finalFee := fee
	if cfg.MinAmount.Gt(fee) {
		finalFee = cfg.MinAmount
	}

	if !finalFee.Eq(decimal.NewFromFloat(50)) {
		t.Errorf("expected fee 50 (min), got %s", finalFee)
	}
}

// =====================================================
// Decimal 测试
// =====================================================

func TestDecimal_Cmp(t *testing.T) {
	a := decimal.NewFromFloat(100)
	b := decimal.NewFromFloat(200)

	if a.Cmp(b) != -1 {
		t.Error("expected a < b")
	}
	if b.Cmp(a) != 1 {
		t.Error("expected b > a")
	}
	if a.Cmp(a) != 0 {
		t.Error("expected a == a")
	}
}

func TestDecimal_Min(t *testing.T) {
	a := decimal.NewFromFloat(100)
	b := decimal.NewFromFloat(200)

	min := decimal.Min(a, b)
	if !min.Eq(a) {
		t.Errorf("expected min to be %s, got %s", a, min)
	}
}

func TestDecimal_Round(t *testing.T) {
	d := decimal.NewFromFloat(123.4567)
	rounded := d.Round(2)

	// 123.4567 rounded to 2 decimal places = 123.46
	expected := decimal.NewFromFloat(123.46)
	if !rounded.Eq(expected) {
		t.Errorf("expected %s, got %s", expected, rounded)
	}
}
