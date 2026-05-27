package calculator

import (
	"time"
	"github.com/lzwcyd/credit-ledger/pkg/decimal"
)

// BulletCalculator 一次性还本付息计算器
// 特点：按月付息，到期一次性偿还本金
type BulletCalculator struct{}

// NewBulletCalculator 创建一次性还本付息计算器
func NewBulletCalculator() *BulletCalculator {
	return &BulletCalculator{}
}

// CalculateSchedule 计算一次性还本付息还款计划
func (c *BulletCalculator) CalculateSchedule(principal decimal.Decimal, annualRate decimal.Decimal, termMonths int, startDate time.Time) []RepaymentSchedule {
	// 月利率
	monthlyRate := annualRate.Div(decimal.NewFromInt(12))
	
	schedules := make([]RepaymentSchedule, termMonths)
	remainingPrincipal := principal
	
	for i := 0; i < termMonths; i++ {
		period := i + 1
		dueDate := startDate.AddDate(0, period, 0)
		
		// 计算当月利息
		interestDue := remainingPrincipal.Mul(monthlyRate).Round(2)
		
		// 本金在最后一期偿还
		principalDue := decimal.Zero()
		if i == termMonths-1 {
			principalDue = remainingPrincipal
		}
		
		// 更新剩余本金
		remainingPrincipal = remainingPrincipal.Sub(principalDue)
		
		schedules[i] = RepaymentSchedule{
			Period:             period,
			DueDate:            dueDate,
			PrincipalDue:       principalDue,
			InterestDue:        interestDue,
			TotalDue:           principalDue.Add(interestDue),
			RemainingPrincipal: remainingPrincipal,
		}
	}
	
	return schedules
}

// CalculateEarlySettlement 提前结清计算
// settlementDate: 结清日期
// 返回: 应还本金 + 应还利息 + 可能的提前还款手续费
type EarlySettlementResult struct {
	Principal         decimal.Decimal `json:"principal"`          // 剩余本金
	Interest          decimal.Decimal `json:"interest"`           // 应还利息
	EarlyRepaymentFee decimal.Decimal `json:"early_repayment_fee"` // 提前还款手续费
	TotalAmount       decimal.Decimal `json:"total_amount"`       // 总金额
	SettlementDate    time.Time       `json:"settlement_date"`    // 结清日期
	DaysSinceLastPayment int          `json:"days_since_last_payment"` // 距上次还款天数
}

func (c *BulletCalculator) CalculateEarlySettlement(
	principal decimal.Decimal,
	annualRate decimal.Decimal,
	lastPaymentDate time.Time,
	settlementDate time.Time,
	earlyRepaymentFeeRate decimal.Decimal,
) EarlySettlementResult {
	// 计算日利率
	daysInYear := 360
	dailyRate := annualRate.Div(decimal.NewFromInt(int64(daysInYear)))
	
	// 计算天数
	days := int(settlementDate.Sub(lastPaymentDate).Hours() / 24)
	if days < 0 {
		days = 0
	}
	
	// 计算应还利息
	interest := principal.Mul(dailyRate).Mul(decimal.NewFromInt(int64(days))).Round(2)
	
	// 计算提前还款手续费
	earlyRepaymentFee := principal.Mul(earlyRepaymentFeeRate).Round(2)
	
	// 总金额
	totalAmount := principal.Add(interest).Add(earlyRepaymentFee)
	
	return EarlySettlementResult{
		Principal:            principal,
		Interest:             interest,
		EarlyRepaymentFee:    earlyRepaymentFee,
		TotalAmount:          totalAmount,
		SettlementDate:       settlementDate,
		DaysSinceLastPayment: days,
	}
}

// CalculatePartialRepayment 部分还款计算
// repaymentAmount: 还款金额
// 返回: 本金减少额、利息支付额
type PartialRepaymentResult struct {
	PrincipalReduction decimal.Decimal `json:"principal_reduction"` // 本金减少额
	InterestPayment    decimal.Decimal `json:"interest_payment"`    // 利息支付额
	RemainingPrincipal decimal.Decimal `json:"remaining_principal"` // 剩余本金
	RemainingInterest  decimal.Decimal `json:"remaining_interest"`  // 剩余利息（如果有）
}

func (c *BulletCalculator) CalculatePartialRepayment(
	currentPrincipal decimal.Decimal,
	annualRate decimal.Decimal,
	lastPaymentDate time.Time,
	repaymentDate time.Time,
	repaymentAmount decimal.Decimal,
) PartialRepaymentResult {
	// 计算日利率
	daysInYear := 360
	dailyRate := annualRate.Div(decimal.NewFromInt(int64(daysInYear)))
	
	// 计算天数
	days := int(repaymentDate.Sub(lastPaymentDate).Hours() / 24)
	if days < 0 {
		days = 0
	}
	
	// 计算应还利息
	interestDue := currentPrincipal.Mul(dailyRate).Mul(decimal.NewFromInt(int64(days))).Round(2)
	
	// 分配还款金额
	var principalReduction, interestPayment, remainingInterest decimal.Decimal
	
	if repaymentAmount.Gte(interestDue) {
		// 还款金额 >= 应还利息
		interestPayment = interestDue
		principalReduction = repaymentAmount.Sub(interestDue)
		remainingInterest = decimal.Zero()
	} else {
		// 还款金额 < 应还利息
		interestPayment = repaymentAmount
		principalReduction = decimal.Zero()
		remainingInterest = interestDue.Sub(repaymentAmount)
	}
	
	// 确保本金减少额不超过剩余本金
	if principalReduction.Gt(currentPrincipal) {
		principalReduction = currentPrincipal
	}
	
	// 计算剩余本金
	remainingPrincipal := currentPrincipal.Sub(principalReduction)
	
	return PartialRepaymentResult{
		PrincipalReduction: principalReduction,
		InterestPayment:    interestPayment,
		RemainingPrincipal: remainingPrincipal,
		RemainingInterest:  remainingInterest,
	}
}