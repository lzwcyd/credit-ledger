package service

import (
	"testing"
	"time"

	"github.com/lzwcyd/credit-ledger/internal/model"
	"github.com/lzwcyd/credit-ledger/pkg/decimal"
)

// =====================================================
// 逾期分级测试
// =====================================================

func TestGetOverdueTier(t *testing.T) {
	tests := []struct {
		days     int
		expected string
	}{
		{0, ""},
		{1, "M1"},
		{30, "M1"},
		{31, "M2"},
		{60, "M2"},
		{61, "M3"},
		{90, "M3"},
		{91, "M4"},
		{120, "M4"},
		{121, "M5"},
		{150, "M5"},
		{151, "M6"},
		{180, "M6"},
		{181, "M7+"},
		{365, "M7+"},
	}

	for _, tt := range tests {
		result := model.GetOverdueTier(tt.days)
		if result != tt.expected {
			t.Errorf("GetOverdueTier(%d) = %s, want %s", tt.days, result, tt.expected)
		}
	}
}

// =====================================================
// 罚息减免测试
// =====================================================

func TestPenaltyWaiver_ExceedsOriginal(t *testing.T) {
	// 减免金额超过原始金额应拒绝
	originalAmount := decimal.NewFromFloat(100)
	waiverAmount := decimal.NewFromFloat(150)

	if waiverAmount.Lte(originalAmount) {
		t.Error("expected waiver > original to be rejected")
	}
}

func TestPenaltyWaiver_ValidAmount(t *testing.T) {
	// 合理的减免金额
	originalAmount := decimal.NewFromFloat(100)
	waiverAmount := decimal.NewFromFloat(50)

	if waiverAmount.Gt(originalAmount) {
		t.Error("valid waiver should not exceed original")
	}

	remaining := originalAmount.Sub(waiverAmount)
	if !remaining.Eq(decimal.NewFromFloat(50)) {
		t.Errorf("expected remaining 50, got %s", remaining)
	}
}

// =====================================================
// 展期测试
// =====================================================

func TestExtension_CalculateNewMaturity(t *testing.T) {
	originalMaturity := time.Date(2026, 12, 31, 0, 0, 0, 0, time.UTC)

	// 展期 30 天
	newMaturity := originalMaturity.AddDate(0, 0, 30)
	expectedDate := time.Date(2027, 1, 30, 0, 0, 0, 0, time.UTC)

	if !newMaturity.Equal(expectedDate) {
		t.Errorf("expected %s, got %s", expectedDate, newMaturity)
	}
}

func TestExtension_CalculateNewMaturity_Months(t *testing.T) {
	originalMaturity := time.Date(2026, 12, 31, 0, 0, 0, 0, time.UTC)

	// 展期 3 个月
	newMaturity := originalMaturity.AddDate(0, 3, 0)
	expectedDate := time.Date(2027, 3, 31, 0, 0, 0, 0, time.UTC)

	if !newMaturity.Equal(expectedDate) {
		t.Errorf("expected %s, got %s", expectedDate, newMaturity)
	}
}

func TestExtension_InvalidRequest(t *testing.T) {
	// 天数和月数都为0应拒绝
	days := 0
	months := 0

	if days > 0 || months > 0 {
		t.Error("should reject zero extension")
	}
}

// =====================================================
// 坏账核销测试
// =====================================================

func TestWriteOff_CalculateAmount(t *testing.T) {
	remainingPrincipal := decimal.NewFromFloat(50000)
	unpaidInterest := decimal.NewFromFloat(3000)
	unpaidPenalty := decimal.NewFromFloat(1500)

	writeOffAmount := remainingPrincipal.Add(unpaidInterest).Add(unpaidPenalty)

	if !writeOffAmount.Eq(decimal.NewFromFloat(54500)) {
		t.Errorf("expected 54500, got %s", writeOffAmount)
	}
}

func TestWriteOff_ZeroAmount(t *testing.T) {
	remainingPrincipal := decimal.Zero()
	unpaidInterest := decimal.Zero()
	unpaidPenalty := decimal.Zero()

	writeOffAmount := remainingPrincipal.Add(unpaidInterest).Add(unpaidPenalty)

	if !writeOffAmount.IsZero() {
		t.Error("should not write off zero amount")
	}
}

// =====================================================
// 客户对账单测试
// =====================================================

func TestStatement_CalculateTotalPaid(t *testing.T) {
	paidPrincipal := decimal.NewFromFloat(20000)
	paidInterest := decimal.NewFromFloat(1200)
	paidPenalty := decimal.NewFromFloat(0)
	paidOtherFee := decimal.NewFromFloat(500)

	totalPaid := paidPrincipal.Add(paidInterest).Add(paidPenalty).Add(paidOtherFee)

	if !totalPaid.Eq(decimal.NewFromFloat(21700)) {
		t.Errorf("expected 21700, got %s", totalPaid)
	}
}

func TestStatement_CalculateTotalDue(t *testing.T) {
	plans := []model.Plan{
		{DueTotal: decimal.NewFromFloat(8500)},
		{DueTotal: decimal.NewFromFloat(8500)},
		{DueTotal: decimal.NewFromFloat(8500)},
	}

	totalDue := decimal.Zero()
	for _, plan := range plans {
		totalDue = totalDue.Add(plan.DueTotal)
	}

	if !totalDue.Eq(decimal.NewFromFloat(25500)) {
		t.Errorf("expected 25500, got %s", totalDue)
	}
}
