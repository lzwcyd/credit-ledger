package calculator

import (
	"testing"
	"time"
	"github.com/yourorg/credit-ledger/pkg/decimal"
)

func TestFeeCalculator(t *testing.T) {
	calculator := NewFeeCalculator()
	
	// 添加固定费用配置
	calculator.AddConfig(&FeeConfig{
		Code:            "HANDLING_FEE",
		Name:            "手续费",
		FeeType:         FeeTypeFixed,
		CalculationBase: BasePrincipal,
		Value:           decimal.NewFromInt(100),
		MinAmount:       decimal.NewFromInt(50),
		MaxAmount:       nil,
		IsActive:        true,
	})
	
	// 添加百分比费用配置
	maxAmount, _ := decimal.NewFromString("5000.0000")
	calculator.AddConfig(&FeeConfig{
		Code:            "MANAGEMENT_FEE",
		Name:            "管理费",
		FeeType:         FeeTypePercentage,
		CalculationBase: BasePrincipal,
		Value:           decimal.NewFromFloat(0.005), // 0.5% = 0.005
		MinAmount:       decimal.NewFromInt(0),
		MaxAmount:       &maxAmount,
		IsActive:        true,
	})
	
	// 测试固定费用
	principal, _ := decimal.NewFromString("100000.0000")
	handlingFee := calculator.CalculateFee("HANDLING_FEE", principal)
	expected, _ := decimal.NewFromString("100.0000")
	if !handlingFee.Eq(expected) {
		t.Errorf("Handling fee: expected %s, got %s", expected, handlingFee)
	}
	
	// 测试百分比费用
	managementFee := calculator.CalculateFee("MANAGEMENT_FEE", principal)
	expected, _ = decimal.NewFromString("500.0000") // 100000 * 0.5% = 500
	if !managementFee.Eq(expected) {
		t.Errorf("Management fee: expected %s, got %s", expected, managementFee)
	}
	
	// 测试最大金额限制
	largePrincipal, _ := decimal.NewFromString("2000000.0000")
	managementFee2 := calculator.CalculateFee("MANAGEMENT_FEE", largePrincipal)
	expected, _ = decimal.NewFromString("5000.0000") // 应该被限制在5000
	if !managementFee2.Eq(expected) {
		t.Errorf("Management fee with max limit: expected %s, got %s", expected, managementFee2)
	}
}

func TestOverduePenaltyCalculator(t *testing.T) {
	// 罚息利率：日利率0.05%
	penaltyRate, _ := decimal.NewFromString("0.0005")
	graceDays := 3
	calculator := NewOverduePenaltyCalculator(penaltyRate, graceDays)
	
	amount, _ := decimal.NewFromString("10000.0000")
	dueDate := time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC)
	
	// 测试宽限期内
	currentDate := time.Date(2024, 1, 3, 0, 0, 0, 0, time.UTC)
	penalty := calculator.CalculateOverduePenalty(amount, dueDate, currentDate)
	if !penalty.IsZero() {
		t.Errorf("Penalty within grace period should be 0, got %s", penalty)
	}
	
	// 测试逾期5天（扣除3天宽限期，实际逾期2天）
	currentDate = time.Date(2024, 1, 6, 0, 0, 0, 0, time.UTC)
	penalty = calculator.CalculateOverduePenalty(amount, dueDate, currentDate)
	// 10000 * 0.0005 * 2 = 10
	expected, _ := decimal.NewFromString("10.0000")
	if !penalty.Eq(expected) {
		t.Errorf("Penalty for 5 days (2 effective): expected %s, got %s", expected, penalty)
	}
}

func TestEarlyRepaymentFeeCalculator(t *testing.T) {
	// 提前还款费率：1%
	feeRate, _ := decimal.NewFromString("0.0100")
	minFee, _ := decimal.NewFromString("100.0000")
	maxFee, _ := decimal.NewFromString("5000.0000")
	calculator := NewEarlyRepaymentFeeCalculator(feeRate, minFee, &maxFee)
	
	// 测试正常计算
	remainingPrincipal, _ := decimal.NewFromString("100000.0000")
	fee := calculator.CalculateEarlyRepaymentFee(remainingPrincipal)
	// 100000 * 1% = 1000
	expected, _ := decimal.NewFromString("1000.0000")
	if !fee.Eq(expected) {
		t.Errorf("Early repayment fee: expected %s, got %s", expected, fee)
	}
	
	// 测试最低收费限制
	smallPrincipal, _ := decimal.NewFromString("5000.0000")
	fee = calculator.CalculateEarlyRepaymentFee(smallPrincipal)
	// 5000 * 1% = 50，但最低100
	expected, _ = decimal.NewFromString("100.0000")
	if !fee.Eq(expected) {
		t.Errorf("Early repayment fee with min limit: expected %s, got %s", expected, fee)
	}
	
	// 测试最高收费限制
	largePrincipal, _ := decimal.NewFromString("1000000.0000")
	fee = calculator.CalculateEarlyRepaymentFee(largePrincipal)
	// 1000000 * 1% = 10000，但最高5000
	expected, _ = decimal.NewFromString("5000.0000")
	if !fee.Eq(expected) {
		t.Errorf("Early repayment fee with max limit: expected %s, got %s", expected, fee)
	}
}

func TestFeeWaiverCalculator(t *testing.T) {
	calculator := NewFeeWaiverCalculator()
	
	// 添加固定金额减免规则
	calculator.AddRule(FeeWaiverRule{
		Code:  "VIP_DISCOUNT",
		Type:  WaiverTypeFixed,
		Value: decimal.NewFromInt(100),
		Condition: func(amount decimal.Decimal, context map[string]interface{}) bool {
			// VIP客户减免100
			isVIP, ok := context["is_vip"].(bool)
			return ok && isVIP
		},
	})
	
	// 添加百分比减免规则
	calculator.AddRule(FeeWaiverRule{
		Code:  "EARLY_BIRD",
		Type:  WaiverTypePercentage,
		Value: decimal.NewFromFloat(0.1), // 10% = 0.1
		Condition: func(amount decimal.Decimal, context map[string]interface{}) bool {
			// 提前还款减免10%
			isEarly, ok := context["is_early"].(bool)
			return ok && isEarly
		},
	})
	
	feeAmount, _ := decimal.NewFromString("1000.0000")
	
	// 测试VIP客户减免
	context := map[string]interface{}{
		"is_vip": true,
	}
	waiver := calculator.CalculateWaiver(feeAmount, context)
	expected, _ := decimal.NewFromString("100.0000")
	if !waiver.Eq(expected) {
		t.Errorf("VIP waiver: expected %s, got %s", expected, waiver)
	}
	
	// 测试提前还款减免
	context = map[string]interface{}{
		"is_early": true,
	}
	waiver = calculator.CalculateWaiver(feeAmount, context)
	expected, _ = decimal.NewFromString("100.0000") // 1000 * 10% = 100
	if !waiver.Eq(expected) {
		t.Errorf("Early bird waiver: expected %s, got %s", expected, waiver)
	}
	
	// 测试多重减免
	context = map[string]interface{}{
		"is_vip":   true,
		"is_early": true,
	}
	waiver = calculator.CalculateWaiver(feeAmount, context)
	expected, _ = decimal.NewFromString("200.0000") // 100 + 100
	if !waiver.Eq(expected) {
		t.Errorf("Multiple waivers: expected %s, got %s", expected, waiver)
	}
	
	// 测试减免金额不超过费用金额
	context = map[string]interface{}{
		"is_vip":   true,
		"is_early": true,
	}
	smallFee, _ := decimal.NewFromString("150.0000")
	waiver = calculator.CalculateWaiver(smallFee, context)
	// VIP: 100, 提前还款: 150 * 10% = 15, 总计: 115
	expected, _ = decimal.NewFromString("115.0000")
	if !waiver.Eq(expected) {
		t.Errorf("Waiver limit: expected %s, got %s", expected, waiver)
	}
}

func TestFeeSummary(t *testing.T) {
	handlingFee, _ := decimal.NewFromString("100.0000")
	managementFee, _ := decimal.NewFromString("500.0000")
	overduePenalty, _ := decimal.NewFromString("50.0000")
	earlyRepaymentFee, _ := decimal.NewFromString("1000.0000")
	waiverAmount, _ := decimal.NewFromString("200.0000")
	
	summary := CalculateFeeSummary(handlingFee, managementFee, overduePenalty, earlyRepaymentFee, waiverAmount)
	
	// 验证总费用
	expectedTotal, _ := decimal.NewFromString("1650.0000")
	if !summary.TotalFees.Eq(expectedTotal) {
		t.Errorf("Total fees: expected %s, got %s", expectedTotal, summary.TotalFees)
	}
	
	// 验证净费用
	expectedNet, _ := decimal.NewFromString("1450.0000")
	if !summary.NetFees.Eq(expectedNet) {
		t.Errorf("Net fees: expected %s, got %s", expectedNet, summary.NetFees)
	}
}