package calculator

import (
	"testing"
	"time"
	"github.com/yourorg/credit-ledger/pkg/decimal"
)

func TestDailyInterestCalculator(t *testing.T) {
	annualRate, _ := decimal.NewFromString("0.0500") // 5%
	calculator := NewDailyInterestCalculator(annualRate, 360)
	
	principal, _ := decimal.NewFromString("100000.0000")
	
	// 测试单日利息
	dailyInterest := calculator.CalculateDailyInterest(principal)
	// 日利率 = 5% / 360 = 0.000138888...
	// 单日利息 = 100000 * 0.000138888... = 13.89
	expected, _ := decimal.NewFromString("13.8900")
	if !dailyInterest.Eq(expected) {
		t.Errorf("Daily interest: expected %s, got %s", expected, dailyInterest)
	}
	
	// 测试指定天数利息
	interest := calculator.CalculateInterestByDays(principal, 30)
	// 30天利息 = 100000 * 0.000138888... * 30 = 416.67
	expected, _ = decimal.NewFromString("416.6700")
	if !interest.Eq(expected) {
		t.Errorf("30 days interest: expected %s, got %s", expected, interest)
	}
	
	// 测试日期范围利息
	startDate := time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC)
	endDate := time.Date(2024, 1, 31, 0, 0, 0, 0, time.UTC)
	interest = calculator.CalculateInterestByDateRange(principal, startDate, endDate)
	expected, _ = decimal.NewFromString("416.6700")
	if !interest.Eq(expected) {
		t.Errorf("Date range interest: expected %s, got %s", expected, interest)
	}
}

func TestDailyInterestCalculator365(t *testing.T) {
	annualRate, _ := decimal.NewFromString("0.0500") // 5%
	calculator := NewDailyInterestCalculator(annualRate, 365)
	
	principal, _ := decimal.NewFromString("100000.0000")
	
	// 测试单日利息（365天基准）
	dailyInterest := calculator.CalculateDailyInterest(principal)
	// 日利率 = 5% / 365 = 0.000136986...
	// 单日利息 = 100000 * 0.000136986... = 13.70
	expected, _ := decimal.NewFromString("13.7000")
	if !dailyInterest.Eq(expected) {
		t.Errorf("Daily interest (365): expected %s, got %s", expected, dailyInterest)
	}
}

func TestMonthlyInterestSchedule(t *testing.T) {
	annualRate, _ := decimal.NewFromString("0.0500") // 5%
	calculator := NewDailyInterestCalculator(annualRate, 360)
	
	principal, _ := decimal.NewFromString("100000.0000")
	termMonths := 12
	startDate := time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC)
	monthlyRepayment, _ := decimal.NewFromString("8560.7500")
	
	schedule := calculator.CalculateInterestSchedule(principal, termMonths, startDate, monthlyRepayment)
	
	// 验证计划数量
	if len(schedule) != termMonths {
		t.Errorf("Expected %d schedules, got %d", termMonths, len(schedule))
	}
	
	// 验证第一期
	first := schedule[0]
	if first.Period != 1 {
		t.Errorf("Expected period 1, got %d", first.Period)
	}
	
	// 验证第一期利息
	// 1月有31天，利息 = 100000 * 5% / 360 * 31 = 430.56
	expectedInterest, _ := decimal.NewFromString("430.5600")
	if !first.Interest.Eq(expectedInterest) {
		t.Errorf("First period interest: expected %s, got %s", expectedInterest, first.Interest)
	}
	
	t.Logf("First period: principal=%s, days=%d, interest=%s", first.Principal, first.Days, first.Interest)
}

func TestDailyInterestDetails(t *testing.T) {
	annualRate, _ := decimal.NewFromString("0.0360") // 3.6% (方便计算，3.6% / 360 = 0.0001)
	calculator := NewDailyInterestCalculator(annualRate, 360)
	
	principal, _ := decimal.NewFromString("100000.0000")
	startDate := time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC)
	endDate := time.Date(2024, 1, 5, 0, 0, 0, 0, time.UTC)
	
	// 无还款
	repayments := make(map[time.Time]decimal.Decimal)
	
	details := calculator.CalculateDailyInterestDetails(principal, startDate, endDate, repayments)
	
	// 验证天数
	if len(details) != 5 {
		t.Errorf("Expected 5 days, got %d", len(details))
	}
	
	// 验证每日利息
	// 日利率 = 3.6% / 360 = 0.0001
	// 单日利息 = 100000 * 0.0001 = 10.00
	expectedDailyInterest, _ := decimal.NewFromString("10.0000")
	for _, detail := range details {
		if !detail.Interest.Eq(expectedDailyInterest) {
			t.Errorf("Daily interest: expected %s, got %s", expectedDailyInterest, detail.Interest)
			break
		}
	}
	
	// 验证累计利息
	// 5天累计 = 10.00 * 5 = 50.00
	lastDetail := details[len(details)-1]
	expectedCumulative, _ := decimal.NewFromString("50.0000")
	if !lastDetail.CumulativeInterest.Eq(expectedCumulative) {
		t.Errorf("Cumulative interest: expected %s, got %s", expectedCumulative, lastDetail.CumulativeInterest)
	}
}

func TestDailyInterestDetailsWithRepayment(t *testing.T) {
	annualRate, _ := decimal.NewFromString("0.0500") // 5%
	calculator := NewDailyInterestCalculator(annualRate, 360)
	
	principal, _ := decimal.NewFromString("10000.0000")
	startDate := time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC)
	endDate := time.Date(2024, 1, 10, 0, 0, 0, 0, time.UTC)
	
	// 第5天还款2000
	repayments := map[time.Time]decimal.Decimal{
		time.Date(2024, 1, 5, 0, 0, 0, 0, time.UTC): decimal.NewFromInt(2000),
	}
	
	details := calculator.CalculateDailyInterestDetails(principal, startDate, endDate, repayments)
	
	// 验证前4天本金为10000
	for i := 0; i < 4; i++ {
		if !details[i].Principal.Eq(principal) {
			t.Errorf("Day %d principal: expected %s, got %s", i+1, principal, details[i].Principal)
		}
	}
	
	// 验证第5天及之后本金为8000
	expectedPrincipal, _ := decimal.NewFromString("8000.0000")
	for i := 4; i < len(details); i++ {
		if !details[i].Principal.Eq(expectedPrincipal) {
			t.Errorf("Day %d principal: expected %s, got %s", i+1, expectedPrincipal, details[i].Principal)
		}
	}
}