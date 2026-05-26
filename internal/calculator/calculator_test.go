package calculator

import (
	"testing"
	"time"
)

func TestEqualInstallmentCalculator(t *testing.T) {
	calculator := NewEqualInstallmentCalculator()
	
	principal := 100000.0
	annualRate := 0.05 // 5%
	termMonths := 12
	startDate := time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC)
	
	schedules := calculator.CalculateSchedule(principal, annualRate, termMonths, startDate)
	
	// 验证还款计划数量
	if len(schedules) != termMonths {
		t.Errorf("Expected %d schedules, got %d", termMonths, len(schedules))
	}
	
	// 验证第一期
	first := schedules[0]
	if first.Period != 1 {
		t.Errorf("Expected period 1, got %d", first.Period)
	}
	
	// 验证总还款额（应该接近本金加利息）
	totalPayment := 0.0
	for _, schedule := range schedules {
		totalPayment += schedule.TotalDue
	}
	
	// 计算总利息
	totalInterest := totalPayment - principal
	t.Logf("Total payment: %.2f, Total interest: %.2f", totalPayment, totalInterest)
	
	// 验证最后一期剩余本金为0
	last := schedules[termMonths-1]
	if last.RemainingPrincipal != 0 {
		t.Errorf("Expected remaining principal 0, got %.2f", last.RemainingPrincipal)
	}
}

func TestEqualPrincipalCalculator(t *testing.T) {
	calculator := NewEqualPrincipalCalculator()
	
	principal := 100000.0
	annualRate := 0.05 // 5%
	termMonths := 12
	startDate := time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC)
	
	schedules := calculator.CalculateSchedule(principal, annualRate, termMonths, startDate)
	
	// 验证还款计划数量
	if len(schedules) != termMonths {
		t.Errorf("Expected %d schedules, got %d", termMonths, len(schedules))
	}
	
	// 验证每月本金相同（除了最后一期调整）
	for i := 0; i < termMonths-1; i++ {
		if schedules[i].PrincipalDue != schedules[0].PrincipalDue {
			t.Errorf("Period %d principal %.2f != first period principal %.2f", 
				i+1, schedules[i].PrincipalDue, schedules[0].PrincipalDue)
		}
	}
	
	// 验证利息递减
	for i := 1; i < termMonths; i++ {
		if schedules[i].InterestDue >= schedules[i-1].InterestDue {
			t.Errorf("Interest should decrease: period %d interest %.2f >= period %d interest %.2f",
				i+1, schedules[i].InterestDue, i, schedules[i-1].InterestDue)
		}
	}
}

func TestInterestOnlyCalculator(t *testing.T) {
	calculator := NewInterestOnlyCalculator()
	
	principal := 100000.0
	annualRate := 0.05 // 5%
	termMonths := 12
	startDate := time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC)
	
	schedules := calculator.CalculateSchedule(principal, annualRate, termMonths, startDate)
	
	// 验证还款计划数量
	if len(schedules) != termMonths {
		t.Errorf("Expected %d schedules, got %d", termMonths, len(schedules))
	}
	
	// 验证前11期只还利息
	for i := 0; i < termMonths-1; i++ {
		if schedules[i].PrincipalDue != 0 {
			t.Errorf("Period %d should have 0 principal, got %.2f", i+1, schedules[i].PrincipalDue)
		}
	}
	
	// 验证最后一期还本金
	last := schedules[termMonths-1]
	if last.PrincipalDue != principal {
		t.Errorf("Last period should have principal %.2f, got %.2f", principal, last.PrincipalDue)
	}
}

func TestCalculatorFactory(t *testing.T) {
	factory := NewCalculatorFactory()
	
	// 测试等额本息
	calc1 := factory.GetCalculator("EQUAL_INSTALLMENT")
	if _, ok := calc1.(*EqualInstallmentCalculator); !ok {
		t.Error("Expected EqualInstallmentCalculator")
	}
	
	// 测试等额本金
	calc2 := factory.GetCalculator("EQUAL_PRINCIPAL")
	if _, ok := calc2.(*EqualPrincipalCalculator); !ok {
		t.Error("Expected EqualPrincipalCalculator")
	}
	
	// 测试按月付息
	calc3 := factory.GetCalculator("INTEREST_ONLY")
	if _, ok := calc3.(*InterestOnlyCalculator); !ok {
		t.Error("Expected InterestOnlyCalculator")
	}
	
	// 测试默认情况
	calc4 := factory.GetCalculator("UNKNOWN")
	if _, ok := calc4.(*EqualInstallmentCalculator); !ok {
		t.Error("Expected default EqualInstallmentCalculator")
	}
}