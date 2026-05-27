package service

import (
	"fmt"
	"time"

	"github.com/lzwcyd/credit-ledger/pkg/decimal"
)

// =====================================================
// 优惠券 / 减免体系
// =====================================================

// CouponType 优惠券类型
const (
	CouponTypeDisbursement = "DISBURSEMENT" // 放款券（放款时抵扣）
	CouponTypeRepayment    = "REPAYMENT"    // 还款券（还款时抵扣）
	CouponTypeWaiver       = "WAIVER"       // 减免券（直接减免待还金额）
	CouponTypeInterestOff  = "INTEREST_OFF" // 息费折扣券（利息打折）
)

// DiscountType 折扣类型
const (
	DiscountFixed      = "FIXED"      // 固定金额
	DiscountPercentage = "PERCENTAGE" // 百分比折扣
)

// CouponStatus 优惠券状态
const (
	CouponStatusActive   = "ACTIVE"   // 可用
	CouponStatusUsed     = "USED"     // 已使用
	CouponStatusExpired  = "EXPIRED"  // 已过期
	CouponStatusDisabled = "DISABLED" // 已停用
)

// Coupon 优惠券
type Coupon struct {
	ID              uint64          `json:"id" db:"id"`
	CouponCode      string          `json:"coupon_code" db:"coupon_code"`
	CouponType      string          `json:"coupon_type" db:"coupon_type"`     // DISBURSEMENT/REPAYMENT/WAIVER/INTEREST_OFF
	DiscountType    string          `json:"discount_type" db:"discount_type"` // FIXED/PERCENTAGE
	FaceValue       decimal.Decimal `json:"face_value" db:"face_value"`       // 面值（固定金额或百分比）
	MaxDiscount     decimal.Decimal `json:"max_discount" db:"max_discount"`   // 最高优惠金额（百分比券时有效）
	MinUsageAmount  decimal.Decimal `json:"min_usage_amount" db:"min_usage_amount"` // 最低使用金额
	ApplicableFee   string          `json:"applicable_fee" db:"applicable_fee"` // 适用费项：ALL/INTEREST/PENALTY/PRINCIPAL
	ValidFrom       time.Time       `json:"valid_from" db:"valid_from"`       // 生效日期
	ValidTo         time.Time       `json:"valid_to" db:"valid_to"`           // 失效日期
	LoanNo          string          `json:"loan_no" db:"loan_no"`             // 绑定借据（空=通用券）
	UserID          string          `json:"user_id" db:"user_id"`             // 绑定用户（空=不限）
	Status          string          `json:"status" db:"status"`
	UsedAt          *time.Time      `json:"used_at" db:"used_at"`
	UsedLoanNo      string          `json:"used_loan_no" db:"used_loan_no"` // 实际使用的借据
	CreatedBy       string          `json:"created_by" db:"created_by"`
	UpdatedBy       string          `json:"updated_by" db:"updated_by"`
	CreatedAt       time.Time       `json:"created_at" db:"created_at"`
	UpdatedAt       time.Time       `json:"updated_at" db:"updated_at"`
}

// IsValid 检查优惠券是否有效
func (c *Coupon) IsValid() error {
	now := time.Now()
	if c.Status != CouponStatusActive {
		return fmt.Errorf("优惠券状态异常: %s", c.Status)
	}
	if now.Before(c.ValidFrom) {
		return fmt.Errorf("优惠券尚未生效，生效日期: %s", c.ValidFrom.Format("2006-01-02"))
	}
	if now.After(c.ValidTo) {
		return fmt.Errorf("优惠券已过期，失效日期: %s", c.ValidTo.Format("2006-01-02"))
	}
	return nil
}

// CalculateDiscount 计算优惠金额
func (c *Coupon) CalculateDiscount(baseAmount decimal.Decimal) decimal.Decimal {
	if baseAmount.Lt(c.MinUsageAmount) {
		return decimal.Zero()
	}

	var discount decimal.Decimal
	switch c.DiscountType {
	case DiscountFixed:
		discount = c.FaceValue
	case DiscountPercentage:
		discount = baseAmount.Mul(c.FaceValue).Div(decimal.NewFromInt(100)).Round(2)
		// 限制最高优惠
		if !c.MaxDiscount.IsZero() && discount.Gt(c.MaxDiscount) {
			discount = c.MaxDiscount
		}
	}
	// 优惠不超过基数
	if discount.Gt(baseAmount) {
		discount = baseAmount
	}
	return discount
}

// =====================================================
// 请求/响应
// =====================================================

// CreateCouponRequest 创建优惠券请求
type CreateCouponRequest struct {
	CouponType     string  `json:"coupon_type"`
	DiscountType   string  `json:"discount_type"`
	FaceValue      string  `json:"face_value"`           // 改为 string 类型
	MaxDiscount    string  `json:"max_discount,omitempty"` // 改为 string 类型
	MinUsageAmount string  `json:"min_usage_amount,omitempty"` // 改为 string 类型
	ApplicableFee  string  `json:"applicable_fee,omitempty"`
	ValidFrom      string  `json:"valid_from"`
	ValidTo        string  `json:"valid_to"`
	LoanNo         string  `json:"loan_no,omitempty"`
	UserID         string  `json:"user_id,omitempty"`
	Count          int     `json:"count,omitempty"` // 批量生成数量
	Operator       string  `json:"operator"`
}

// ApplyCouponRequest 使用优惠券请求
type ApplyCouponRequest struct {
	CouponCode string  `json:"coupon_code"`
	LoanNo     string  `json:"loan_no"`
	BaseAmount string  `json:"base_amount"` // 改为 string 类型，避免精度损失
	Operator   string  `json:"operator"`
}

// ApplyCouponResponse 使用优惠券响应
type ApplyCouponResponse struct {
	CouponCode    string          `json:"coupon_code"`
	LoanNo        string          `json:"loan_no"`
	DiscountType  string          `json:"discount_type"`
	FaceValue     decimal.Decimal `json:"face_value"`
	DiscountAmount decimal.Decimal `json:"discount_amount"` // 实际优惠金额
}

// CouponTrialRequest 优惠券试算请求
type CouponTrialRequest struct {
	CouponCode string  `json:"coupon_code"`
	LoanNo     string  `json:"loan_no"`
	BaseAmount string  `json:"base_amount"` // 改为 string 类型，避免精度损失
}

// CouponTrialResponse 优惠券试算响应
type CouponTrialResponse struct {
	CouponCode     string          `json:"coupon_code"`
	IsValid        bool            `json:"is_valid"`
	InvalidReason  string          `json:"invalid_reason,omitempty"`
	DiscountType   string          `json:"discount_type"`
	FaceValue      decimal.Decimal `json:"face_value"`
	DiscountAmount decimal.Decimal `json:"discount_amount"`
	ValidFrom      time.Time       `json:"valid_from"`
	ValidTo        time.Time       `json:"valid_to"`
}

// =====================================================
// 服务方法
// =====================================================

// CreateCoupon 创建优惠券
func (s *LoanService) CreateCoupon(req CreateCouponRequest) ([]Coupon, error) {
	// 参数校验
	if req.CouponType == "" {
		return nil, fmt.Errorf("优惠券类型不能为空")
	}
	if req.DiscountType == "" {
		return nil, fmt.Errorf("折扣类型不能为空")
	}
	faceValue, err := decimal.NewFromString(req.FaceValue)
	if err != nil {
		return nil, fmt.Errorf("面值格式错误: %w", err)
	}
	if !faceValue.IsPositive() {
		return nil, fmt.Errorf("面值必须大于0")
	}

	validFrom, err := time.Parse("2006-01-02", req.ValidFrom)
	if err != nil {
		return nil, fmt.Errorf("生效日期格式错误: %w", err)
	}
	validTo, err := time.Parse("2006-01-02", req.ValidTo)
	if err != nil {
		return nil, fmt.Errorf("失效日期格式错误: %w", err)
	}
	if !validTo.After(validFrom) {
		return nil, fmt.Errorf("失效日期必须晚于生效日期")
	}

	if req.ApplicableFee == "" {
		req.ApplicableFee = "ALL"
	}

	count := req.Count
	if count <= 0 {
		count = 1
	}

	var coupons []Coupon
	now := time.Now()

	for i := 0; i < count; i++ {
		coupon := Coupon{
			CouponCode:      generateCouponCode(),
			CouponType:      req.CouponType,
			DiscountType:    req.DiscountType,
			FaceValue:       faceValue,
			MaxDiscount:     decimal.Zero(), // 默认为0，后面根据需要解析
			MinUsageAmount:  decimal.Zero(), // 默认为0，后面根据需要解析
			ApplicableFee:   req.ApplicableFee,
			ValidFrom:       validFrom,
			ValidTo:         validTo,
			LoanNo:          req.LoanNo,
			UserID:          req.UserID,
			Status:          CouponStatusActive,
			CreatedBy:       req.Operator,
			UpdatedBy:       req.Operator,
			CreatedAt:       now,
			UpdatedAt:       now,
		}
		// 解析 MaxDiscount
		if req.MaxDiscount != "" {
			maxDiscount, err := decimal.NewFromString(req.MaxDiscount)
			if err != nil {
				return nil, fmt.Errorf("最大折扣金额格式错误: %w", err)
			}
			coupon.MaxDiscount = maxDiscount
		}
		// 解析 MinUsageAmount
		if req.MinUsageAmount != "" {
			minUsageAmount, err := decimal.NewFromString(req.MinUsageAmount)
			if err != nil {
				return nil, fmt.Errorf("最低使用金额格式错误: %w", err)
			}
			coupon.MinUsageAmount = minUsageAmount
		}
		// TODO: 持久化到数据库
		coupons = append(coupons, coupon)
	}

	return coupons, nil
}

// CouponTrial 优惠券试算
func (s *LoanService) CouponTrial(req CouponTrialRequest) (*CouponTrialResponse, error) {
	coupon, err := s.getCouponByCode(req.CouponCode)
	if err != nil {
		return &CouponTrialResponse{
			CouponCode:    req.CouponCode,
			IsValid:       false,
			InvalidReason: "优惠券不存在",
		}, nil
	}

	baseAmount, err := decimal.NewFromString(req.BaseAmount)
	if err != nil {
		return nil, fmt.Errorf("基数金额格式错误: %w", err)
	}

	// 校验有效性
	if err := coupon.IsValid(); err != nil {
		return &CouponTrialResponse{
			CouponCode:    req.CouponCode,
			IsValid:       false,
			InvalidReason: err.Error(),
			ValidFrom:     coupon.ValidFrom,
			ValidTo:       coupon.ValidTo,
		}, nil
	}

	// 校验最低使用金额
	if baseAmount.Lt(coupon.MinUsageAmount) {
		return &CouponTrialResponse{
			CouponCode:    req.CouponCode,
			IsValid:       false,
			InvalidReason: fmt.Sprintf("使用金额 %s 低于最低要求 %s", baseAmount, coupon.MinUsageAmount),
			ValidFrom:     coupon.ValidFrom,
			ValidTo:       coupon.ValidTo,
		}, nil
	}

	discountAmount := coupon.CalculateDiscount(baseAmount)

	return &CouponTrialResponse{
		CouponCode:     req.CouponCode,
		IsValid:        true,
		DiscountType:   coupon.DiscountType,
		FaceValue:      coupon.FaceValue,
		DiscountAmount: discountAmount,
		ValidFrom:      coupon.ValidFrom,
		ValidTo:        coupon.ValidTo,
	}, nil
}

// ApplyCoupon 使用优惠券
func (s *LoanService) ApplyCoupon(req ApplyCouponRequest) (*ApplyCouponResponse, error) {
	coupon, err := s.getCouponByCode(req.CouponCode)
	if err != nil {
		return nil, fmt.Errorf("优惠券不存在: %w", err)
	}

	// 校验有效性
	if err := coupon.IsValid(); err != nil {
		return nil, fmt.Errorf("优惠券不可用: %w", err)
	}

	baseAmount, err := decimal.NewFromString(req.BaseAmount)
	if err != nil {
		return nil, fmt.Errorf("基数金额格式错误: %w", err)
	}

	// 校验最低使用金额
	if baseAmount.Lt(coupon.MinUsageAmount) {
		return nil, fmt.Errorf("使用金额 %s 低于最低要求 %s", baseAmount, coupon.MinUsageAmount)
	}

	discountAmount := coupon.CalculateDiscount(baseAmount)

	// 标记为已使用
	now := time.Now()
	coupon.Status = CouponStatusUsed
	coupon.UsedAt = &now
	coupon.UsedLoanNo = req.LoanNo
	coupon.UpdatedBy = req.Operator
	// TODO: 持久化到数据库

	// 根据优惠券类型执行不同逻辑
	switch coupon.CouponType {
	case CouponTypeWaiver:
		// 减免券：直接减免对应金额
		if err := s.applyWaiverCoupon(req.LoanNo, coupon, discountAmount, req.Operator); err != nil {
			return nil, fmt.Errorf("应用减免券失败: %w", err)
		}
	case CouponTypeRepayment:
		// 还款券：在下次还款时抵扣（这里只记录，实际抵扣在还款时处理）
	case CouponTypeDisbursement:
		// 放款券：在放款时抵扣（这里只记录，实际抵扣在放款时处理）
	}

	return &ApplyCouponResponse{
		CouponCode:     req.CouponCode,
		LoanNo:         req.LoanNo,
		DiscountType:   coupon.DiscountType,
		FaceValue:      coupon.FaceValue,
		DiscountAmount: discountAmount,
	}, nil
}

// applyWaiverCoupon 应用减免券
func (s *LoanService) applyWaiverCoupon(loanNo string, coupon *Coupon, amount decimal.Decimal, operator string) error {
	loan, err := s.loanRepo.GetLoanByNo(loanNo)
	if err != nil {
		return fmt.Errorf("借据不存在: %w", err)
	}

	switch coupon.ApplicableFee {
	case "INTEREST":
		loan.PaidInterest = loan.PaidInterest.Add(amount)
	case "PENALTY":
		loan.PaidPenalty = loan.PaidPenalty.Add(amount)
	case "PRINCIPAL":
		loan.PaidPrincipal = loan.PaidPrincipal.Add(amount)
		loan.RemainingPrincipal = loan.RemainingPrincipal.Sub(amount)
	default:
		// ALL: 按优先级减免（罚息 > 利息 > 本金）
		remaining := amount
		unpaidPenalty := loan.TotalPenalty.Sub(loan.PaidPenalty)
		if remaining.IsPositive() && unpaidPenalty.IsPositive() {
			d := decimal.Min(remaining, unpaidPenalty)
			loan.PaidPenalty = loan.PaidPenalty.Add(d)
			remaining = remaining.Sub(d)
		}
		unpaidInterest := loan.TotalInterest.Sub(loan.PaidInterest)
		if remaining.IsPositive() && unpaidInterest.IsPositive() {
			d := decimal.Min(remaining, unpaidInterest)
			loan.PaidInterest = loan.PaidInterest.Add(d)
			remaining = remaining.Sub(d)
		}
		if remaining.IsPositive() {
			loan.PaidPrincipal = loan.PaidPrincipal.Add(remaining)
			loan.RemainingPrincipal = loan.RemainingPrincipal.Sub(remaining)
		}
	}

	loan.UpdatedBy = operator
	return s.loanRepo.UpdateLoan(loan)
}

// getCouponByCode 根据编码获取优惠券（临时实现，后续走数据库）
func (s *LoanService) getCouponByCode(code string) (*Coupon, error) {
	// TODO: 从数据库查询
	return nil, fmt.Errorf("优惠券不存在: %s", code)
}

// generateCouponCode 生成优惠券编码
func generateCouponCode() string {
	return fmt.Sprintf("CP%s%04d", time.Now().Format("20060102"), time.Now().Nanosecond()%10000)
}
