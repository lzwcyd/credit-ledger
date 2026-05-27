package calculator

import (
	"time"
	"github.com/lzwcyd/credit-ledger/pkg/decimal"
)

// RepaymentCalculator 还款计算器接口
type RepaymentCalculator interface {
	CalculateSchedule(principal decimal.Decimal, annualRate decimal.Decimal, termMonths int, startDate time.Time) []RepaymentSchedule
}

// RepaymentSchedule 还款计划
type RepaymentSchedule struct {
	Period        int             `json:"period"`
	DueDate       time.Time       `json:"due_date"`
	PrincipalDue  decimal.Decimal `json:"principal_due"`
	InterestDue   decimal.Decimal `json:"interest_due"`
	TotalDue      decimal.Decimal `json:"total_due"`
	RemainingPrincipal decimal.Decimal `json:"remaining_principal"`
}

// EqualInstallmentCalculator 等额本息计算器
type EqualInstallmentCalculator struct{}

// NewEqualInstallmentCalculator 创建等额本息计算器
func NewEqualInstallmentCalculator() *EqualInstallmentCalculator {
	return &EqualInstallmentCalculator{}
}

// CalculateSchedule 计算等额本息还款计划
func (c *EqualInstallmentCalculator) CalculateSchedule(principal decimal.Decimal, annualRate decimal.Decimal, termMonths int, startDate time.Time) []RepaymentSchedule {
	// 月利率
	monthlyRate := annualRate.Div(decimal.NewFromInt(12))
	
	// 计算每月还款额
	// 公式：M = P * r * (1+r)^n / ((1+r)^n - 1)
	// 其中：M=每月还款额, P=贷款本金, r=月利率, n=还款期数
	if monthlyRate.IsZero() {
		// 零利率情况
		return c.calculateZeroRateSchedule(principal, termMonths, startDate)
	}
	
	// 计算 (1+r)^n
	onePlusRate := decimal.One().Add(monthlyRate)
	compoundFactor := decimal.One()
	for i := 0; i < termMonths; i++ {
		compoundFactor = compoundFactor.Mul(onePlusRate)
	}
	
	// 计算每月还款额
	// M = P * r * (1+r)^n / ((1+r)^n - 1)
	numerator := principal.Mul(monthlyRate).Mul(compoundFactor)
	denominator := compoundFactor.Sub(decimal.One())
	monthlyPayment := numerator.Div(denominator).Round(2)
	
	schedules := make([]RepaymentSchedule, termMonths)
	remainingPrincipal := principal
	
	for i := 0; i < termMonths; i++ {
		period := i + 1
		dueDate := startDate.AddDate(0, period, 0)
		
		// 计算当月利息
		interestDue := remainingPrincipal.Mul(monthlyRate).Round(2)
		
		// 计算当月本金
		principalDue := monthlyPayment.Sub(interestDue)
		
		// 最后一期调整
		if i == termMonths-1 {
			principalDue = remainingPrincipal
			monthlyPayment = principalDue.Add(interestDue)
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

// calculateZeroRateSchedule 计算零利率还款计划
func (c *EqualInstallmentCalculator) calculateZeroRateSchedule(principal decimal.Decimal, termMonths int, startDate time.Time) []RepaymentSchedule {
	monthlyPrincipal := principal.Div(decimal.NewFromInt(int64(termMonths))).Round(2)
	
	schedules := make([]RepaymentSchedule, termMonths)
	remainingPrincipal := principal
	
	for i := 0; i < termMonths; i++ {
		period := i + 1
		dueDate := startDate.AddDate(0, period, 0)
		
		// 最后一期调整
		principalDue := monthlyPrincipal
		if i == termMonths-1 {
			principalDue = remainingPrincipal
		}
		
		// 更新剩余本金
		remainingPrincipal = remainingPrincipal.Sub(principalDue)
		
		schedules[i] = RepaymentSchedule{
			Period:             period,
			DueDate:            dueDate,
			PrincipalDue:       principalDue,
			InterestDue:        decimal.Zero(),
			TotalDue:           principalDue,
			RemainingPrincipal: remainingPrincipal,
		}
	}
	
	return schedules
}

// EqualPrincipalCalculator 等额本金计算器
type EqualPrincipalCalculator struct{}

// NewEqualPrincipalCalculator 创建等额本金计算器
func NewEqualPrincipalCalculator() *EqualPrincipalCalculator {
	return &EqualPrincipalCalculator{}
}

// CalculateSchedule 计算等额本金还款计划
func (c *EqualPrincipalCalculator) CalculateSchedule(principal decimal.Decimal, annualRate decimal.Decimal, termMonths int, startDate time.Time) []RepaymentSchedule {
	// 月利率
	monthlyRate := annualRate.Div(decimal.NewFromInt(12))
	
	// 每月固定本金
	monthlyPrincipal := principal.Div(decimal.NewFromInt(int64(termMonths))).Round(2)
	
	schedules := make([]RepaymentSchedule, termMonths)
	remainingPrincipal := principal
	
	for i := 0; i < termMonths; i++ {
		period := i + 1
		dueDate := startDate.AddDate(0, period, 0)
		
		// 计算当月利息
		interestDue := remainingPrincipal.Mul(monthlyRate).Round(2)
		
		// 最后一期本金调整
		principalDue := monthlyPrincipal
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

// InterestOnlyCalculator 按月付息到期还本计算器
type InterestOnlyCalculator struct{}

// NewInterestOnlyCalculator 创建按月付息到期还本计算器
func NewInterestOnlyCalculator() *InterestOnlyCalculator {
	return &InterestOnlyCalculator{}
}

// CalculateSchedule 计算按月付息到期还本还款计划
func (c *InterestOnlyCalculator) CalculateSchedule(principal decimal.Decimal, annualRate decimal.Decimal, termMonths int, startDate time.Time) []RepaymentSchedule {
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

// CalculatorFactory 计算器工厂
type CalculatorFactory struct{}

// NewCalculatorFactory 创建计算器工厂
func NewCalculatorFactory() *CalculatorFactory {
	return &CalculatorFactory{}
}

// GetCalculator 根据还款类型获取计算器
func (f *CalculatorFactory) GetCalculator(repaymentType string) RepaymentCalculator {
	switch repaymentType {
	case "EQUAL_INSTALLMENT":
		return NewEqualInstallmentCalculator()
	case "EQUAL_PRINCIPAL":
		return NewEqualPrincipalCalculator()
	case "INTEREST_ONLY":
		return NewInterestOnlyCalculator()
	case "BULLET":
		return NewBulletCalculator()
	default:
		return NewEqualInstallmentCalculator()
	}
}