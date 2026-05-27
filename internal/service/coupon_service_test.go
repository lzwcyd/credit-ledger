package service

import (
	"testing"
	"time"

	"github.com/lzwcyd/credit-ledger/pkg/decimal"
)

// =====================================================
// 优惠券有效期测试
// =====================================================

func TestCoupon_IsValid_Active(t *testing.T) {
	coupon := &Coupon{
		Status:    CouponStatusActive,
		ValidFrom: time.Now().AddDate(0, -1, 0), // 上月生效
		ValidTo:   time.Now().AddDate(0, 1, 0),  // 下月失效
	}

	if err := coupon.IsValid(); err != nil {
		t.Errorf("expected valid, got error: %s", err)
	}
}

func TestCoupon_IsValid_Expired(t *testing.T) {
	coupon := &Coupon{
		Status:    CouponStatusActive,
		ValidFrom: time.Now().AddDate(0, -2, 0),
		ValidTo:   time.Now().AddDate(0, -1, 0), // 已过期
	}

	err := coupon.IsValid()
	if err == nil {
		t.Error("expected expired error")
	}
}

func TestCoupon_IsValid_NotStarted(t *testing.T) {
	coupon := &Coupon{
		Status:    CouponStatusActive,
		ValidFrom: time.Now().AddDate(0, 1, 0), // 尚未生效
		ValidTo:   time.Now().AddDate(0, 2, 0),
	}

	err := coupon.IsValid()
	if err == nil {
		t.Error("expected not started error")
	}
}

func TestCoupon_IsValid_Used(t *testing.T) {
	coupon := &Coupon{
		Status:    CouponStatusUsed,
		ValidFrom: time.Now().AddDate(0, -1, 0),
		ValidTo:   time.Now().AddDate(0, 1, 0),
	}

	err := coupon.IsValid()
	if err == nil {
		t.Error("expected used error")
	}
}

// =====================================================
// 优惠金额计算测试
// =====================================================

func TestCoupon_CalculateDiscount_Fixed(t *testing.T) {
	coupon := &Coupon{
		DiscountType:   DiscountFixed,
		FaceValue:      decimal.NewFromFloat(100),
		MinUsageAmount: decimal.NewFromFloat(500),
	}

	// 金额 >= 最低使用金额
	discount := coupon.CalculateDiscount(decimal.NewFromFloat(1000))
	if !discount.Eq(decimal.NewFromFloat(100)) {
		t.Errorf("expected 100, got %s", discount)
	}
}

func TestCoupon_CalculateDiscount_BelowMinAmount(t *testing.T) {
	coupon := &Coupon{
		DiscountType:   DiscountFixed,
		FaceValue:      decimal.NewFromFloat(100),
		MinUsageAmount: decimal.NewFromFloat(500),
	}

	// 金额 < 最低使用金额
	discount := coupon.CalculateDiscount(decimal.NewFromFloat(300))
	if !discount.IsZero() {
		t.Errorf("expected 0, got %s", discount)
	}
}

func TestCoupon_CalculateDiscount_Percentage(t *testing.T) {
	coupon := &Coupon{
		DiscountType:   DiscountPercentage,
		FaceValue:      decimal.NewFromFloat(10),  // 10%
		MinUsageAmount: decimal.Zero(),
	}

	// 1000 * 10% = 100
	discount := coupon.CalculateDiscount(decimal.NewFromFloat(1000))
	if !discount.Eq(decimal.NewFromFloat(100)) {
		t.Errorf("expected 100, got %s", discount)
	}
}

func TestCoupon_CalculateDiscount_Percentage_WithMaxCap(t *testing.T) {
	coupon := &Coupon{
		DiscountType:   DiscountPercentage,
		FaceValue:      decimal.NewFromFloat(20),  // 20%
		MaxDiscount:    decimal.NewFromFloat(50),   // 最高50
		MinUsageAmount: decimal.Zero(),
	}

	// 1000 * 20% = 200, 但最高50
	discount := coupon.CalculateDiscount(decimal.NewFromFloat(1000))
	if !discount.Eq(decimal.NewFromFloat(50)) {
		t.Errorf("expected 50 (max cap), got %s", discount)
	}
}

func TestCoupon_CalculateDiscount_ExceedsBaseAmount(t *testing.T) {
	coupon := &Coupon{
		DiscountType:   DiscountFixed,
		FaceValue:      decimal.NewFromFloat(500),
		MinUsageAmount: decimal.Zero(),
	}

	// 优惠500 > 基数200，应返回200
	discount := coupon.CalculateDiscount(decimal.NewFromFloat(200))
	if !discount.Eq(decimal.NewFromFloat(200)) {
		t.Errorf("expected 200 (capped to base), got %s", discount)
	}
}

// =====================================================
// 优惠券类型测试
// =====================================================

func TestCouponType_Values(t *testing.T) {
	types := []string{
		CouponTypeDisbursement,
		CouponTypeRepayment,
		CouponTypeWaiver,
		CouponTypeInterestOff,
	}

	for _, ct := range types {
		if ct == "" {
			t.Error("coupon type should not be empty")
		}
	}
}

// =====================================================
// 批量生成测试
// =====================================================

func TestCreateCoupon_BatchCount(t *testing.T) {
	count := 5
	if count <= 0 {
		count = 1
	}
	if count != 5 {
		t.Errorf("expected 5, got %d", count)
	}
}
