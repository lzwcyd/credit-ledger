package calculator

import (
	"time"
	"github.com/lzwcyd/credit-ledger/pkg/decimal"
)

// FeeType 费用类型
type FeeType string

const (
	FeeTypeFixed      FeeType = "FIXED"      // 固定金额
	FeeTypePercentage FeeType = "PERCENTAGE" // 百分比
	FeeTypeRate       FeeType = "RATE"       // 费率（如日费率）
)

// FeeCalculationBase 费用计算基数
type FeeCalculationBase string

const (
	BasePrincipal FeeCalculationBase = "PRINCIPAL" // 基于本金
	BaseInterest  FeeCalculationBase = "INTEREST"  // 基于利息
	BaseTotal     FeeCalculationBase = "TOTAL"     // 基于总额
)

// FeeConfig 费用配置
type FeeConfig struct {
	ID              uint64             `json:"id"`
	Code            string             `json:"code"`
	Name            string             `json:"name"`
	FeeType         FeeType            `json:"fee_type"`
	CalculationBase FeeCalculationBase `json:"calculation_base"`
	Value           decimal.Decimal    `json:"value"`
	MinAmount       decimal.Decimal    `json:"min_amount"`
	MaxAmount       *decimal.Decimal   `json:"max_amount"`
	IsActive        bool               `json:"is_active"`
}

// FeeCalculator 费用计算器
type FeeCalculator struct {
	configs map[string]*FeeConfig
}

// NewFeeCalculator 创建费用计算器
func NewFeeCalculator() *FeeCalculator {
	return &FeeCalculator{
		configs: make(map[string]*FeeConfig),
	}
}

// AddConfig 添加费用配置
func (f *FeeCalculator) AddConfig(config *FeeConfig) {
	f.configs[config.Code] = config
}

// GetConfig 获取费用配置
func (f *FeeCalculator) GetConfig(code string) *FeeConfig {
	return f.configs[code]
}

// CalculateFee 计算费用
// code: 费用代码
// baseAmount: 计算基数金额
// 返回: 费用金额
func (f *FeeCalculator) CalculateFee(code string, baseAmount decimal.Decimal) decimal.Decimal {
	config := f.GetConfig(code)
	if config == nil || !config.IsActive {
		return decimal.Zero()
	}

	var fee decimal.Decimal

	switch config.FeeType {
	case FeeTypeFixed:
		fee = config.Value
	case FeeTypePercentage:
		// Value 直接表示百分比（如 0.005 表示 0.5%）
		fee = baseAmount.Mul(config.Value).Round(2)
	case FeeTypeRate:
		fee = baseAmount.Mul(config.Value).Round(2)
	}

	// 应用最小金额限制
	if fee.Lt(config.MinAmount) {
		fee = config.MinAmount
	}

	// 应用最大金额限制
	if config.MaxAmount != nil && fee.Gt(*config.MaxAmount) {
		fee = *config.MaxAmount
	}

	return fee
}

// 手续费计算
type HandlingFeeCalculator struct {
	feeCalculator *FeeCalculator
}

func NewHandlingFeeCalculator(feeCalculator *FeeCalculator) *HandlingFeeCalculator {
	return &HandlingFeeCalculator{feeCalculator: feeCalculator}
}

// CalculateHandlingFee 计算手续费
// principal: 贷款本金
// 返回: 手续费金额
func (h *HandlingFeeCalculator) CalculateHandlingFee(principal decimal.Decimal) decimal.Decimal {
	return h.feeCalculator.CalculateFee("HANDLING_FEE", principal)
}

// 管理费计算
type ManagementFeeCalculator struct {
	feeCalculator *FeeCalculator
}

func NewManagementFeeCalculator(feeCalculator *FeeCalculator) *ManagementFeeCalculator {
	return &ManagementFeeCalculator{feeCalculator: feeCalculator}
}

// CalculateManagementFee 计算管理费
// principal: 贷款本金
// termMonths: 贷款期限（月）
// 返回: 管理费金额
func (m *ManagementFeeCalculator) CalculateManagementFee(principal decimal.Decimal, termMonths int) decimal.Decimal {
	// 计算基数：本金 * 期限
	baseAmount := principal.Mul(decimal.NewFromInt(int64(termMonths)))
	return m.feeCalculator.CalculateFee("MANAGEMENT_FEE", baseAmount)
}

// 逾期罚息计算
type OverduePenaltyCalculator struct {
	// 罚息利率（日利率）
	penaltyRate decimal.Decimal
	// 宽限期（天）
	graceDays int
}

func NewOverduePenaltyCalculator(penaltyRate decimal.Decimal, graceDays int) *OverduePenaltyCalculator {
	return &OverduePenaltyCalculator{
		penaltyRate: penaltyRate,
		graceDays:   graceDays,
	}
}

// CalculateOverduePenalty 计算逾期罚息
// amount: 逾期金额
// dueDate: 应还日期
// currentDate: 当前日期
// 返回: 罚息金额
func (o *OverduePenaltyCalculator) CalculateOverduePenalty(amount decimal.Decimal, dueDate, currentDate time.Time) decimal.Decimal {
	// 计算逾期天数
	overdueDays := int(currentDate.Sub(dueDate).Hours() / 24)
	
	// 减去宽限期
	overdueDays -= o.graceDays
	if overdueDays <= 0 {
		return decimal.Zero()
	}
	
	// 计算罚息：逾期金额 * 罚息利率 * 逾期天数
	penalty := amount.Mul(o.penaltyRate).Mul(decimal.NewFromInt(int64(overdueDays)))
	return penalty.Round(2)
}

// CalculateDailyPenalty 计算每日罚息
// amount: 逾期金额
// 返回: 每日罚息金额
func (o *OverduePenaltyCalculator) CalculateDailyPenalty(amount decimal.Decimal) decimal.Decimal {
	return amount.Mul(o.penaltyRate).Round(2)
}

// 提前还款手续费计算
type EarlyRepaymentFeeCalculator struct {
	// 提前还款费率
	feeRate decimal.Decimal
	// 最低收费
	minFee decimal.Decimal
	// 最高收费
	maxFee *decimal.Decimal
}

func NewEarlyRepaymentFeeCalculator(feeRate decimal.Decimal, minFee decimal.Decimal, maxFee *decimal.Decimal) *EarlyRepaymentFeeCalculator {
	return &EarlyRepaymentFeeCalculator{
		feeRate: feeRate,
		minFee:  minFee,
		maxFee:  maxFee,
	}
}

// CalculateEarlyRepaymentFee 计算提前还款手续费
// remainingPrincipal: 剩余本金
// 返回: 提前还款手续费
func (e *EarlyRepaymentFeeCalculator) CalculateEarlyRepaymentFee(remainingPrincipal decimal.Decimal) decimal.Decimal {
	fee := remainingPrincipal.Mul(e.feeRate).Round(2)
	
	// 应用最低收费
	if fee.Lt(e.minFee) {
		fee = e.minFee
	}
	
	// 应用最高收费
	if e.maxFee != nil && fee.Gt(*e.maxFee) {
		fee = *e.maxFee
	}
	
	return fee
}

// 费用减免计算器
type FeeWaiverCalculator struct {
	// 减免规则
	rules []FeeWaiverRule
}

// FeeWaiverRule 费用减免规则
type FeeWaiverRule struct {
	// 规则代码
	Code string
	// 减免类型
	Type WaiverType
	// 减免值
	Value decimal.Decimal
	// 条件
	Condition func(amount decimal.Decimal, context map[string]interface{}) bool
}

// WaiverType 减免类型
type WaiverType string

const (
	WaiverTypeFixed      WaiverType = "FIXED"      // 固定金额减免
	WaiverTypePercentage WaiverType = "PERCENTAGE" // 百分比减免
	WaiverTypeFull       WaiverType = "FULL"       // 全额减免
)

func NewFeeWaiverCalculator() *FeeWaiverCalculator {
	return &FeeWaiverCalculator{
		rules: make([]FeeWaiverRule, 0),
	}
}

// AddRule 添加减免规则
func (f *FeeWaiverCalculator) AddRule(rule FeeWaiverRule) {
	f.rules = append(f.rules, rule)
}

// CalculateWaiver 计算减免金额
// feeAmount: 原始费用金额
// context: 上下文信息（用于条件判断）
// 返回: 减免金额
func (f *FeeWaiverCalculator) CalculateWaiver(feeAmount decimal.Decimal, context map[string]interface{}) decimal.Decimal {
	totalWaiver := decimal.Zero()
	
	for _, rule := range f.rules {
		// 检查条件
		if rule.Condition != nil && !rule.Condition(feeAmount, context) {
			continue
		}
		
		var waiver decimal.Decimal
		
		switch rule.Type {
		case WaiverTypeFixed:
			waiver = rule.Value
		case WaiverTypePercentage:
			// Value 直接表示百分比（如 0.1 表示 10%）
			waiver = feeAmount.Mul(rule.Value).Round(2)
		case WaiverTypeFull:
			waiver = feeAmount
		}
		
		totalWaiver = totalWaiver.Add(waiver)
	}
	
	// 减免金额不能超过费用金额
	if totalWaiver.Gt(feeAmount) {
		totalWaiver = feeAmount
	}
	
	return totalWaiver
}

// FeeSummary 费用汇总
type FeeSummary struct {
	HandlingFee       decimal.Decimal `json:"handling_fee"`        // 手续费
	ManagementFee     decimal.Decimal `json:"management_fee"`      // 管理费
	OverduePenalty    decimal.Decimal `json:"overdue_penalty"`     // 逾期罚息
	EarlyRepaymentFee decimal.Decimal `json:"early_repayment_fee"` // 提前还款手续费
	TotalFees         decimal.Decimal `json:"total_fees"`          // 总费用
	WaiverAmount      decimal.Decimal `json:"waiver_amount"`       // 减免金额
	NetFees           decimal.Decimal `json:"net_fees"`            // 净费用
}

// CalculateFeeSummary 计算费用汇总
func CalculateFeeSummary(
	handlingFee, managementFee, overduePenalty, earlyRepaymentFee decimal.Decimal,
	waiverAmount decimal.Decimal,
) FeeSummary {
	totalFees := handlingFee.Add(managementFee).Add(overduePenalty).Add(earlyRepaymentFee)
	netFees := totalFees.Sub(waiverAmount)
	
	return FeeSummary{
		HandlingFee:       handlingFee,
		ManagementFee:     managementFee,
		OverduePenalty:    overduePenalty,
		EarlyRepaymentFee: earlyRepaymentFee,
		TotalFees:         totalFees,
		WaiverAmount:      waiverAmount,
		NetFees:           netFees,
	}
}