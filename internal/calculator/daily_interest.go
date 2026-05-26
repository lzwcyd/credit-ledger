package calculator

import (
	"time"
	"github.com/yourorg/credit-ledger/pkg/decimal"
)

// DailyInterestCalculator 按日计息计算器
type DailyInterestCalculator struct {
	// 年利率
	annualRate decimal.Decimal
	// 计息基准天数（360 或 365）
	daysInYear int
}

// NewDailyInterestCalculator 创建按日计息计算器
func NewDailyInterestCalculator(annualRate decimal.Decimal, daysInYear int) *DailyInterestCalculator {
	if daysInYear != 360 && daysInYear != 365 {
		daysInYear = 360 // 默认使用360天
	}
	return &DailyInterestCalculator{
		annualRate: annualRate,
		daysInYear: daysInYear,
	}
}

// CalculateDailyInterest 计算单日利息
// principal: 本金余额
// 返回: 当日利息
func (c *DailyInterestCalculator) CalculateDailyInterest(principal decimal.Decimal) decimal.Decimal {
	// 日利率 = 年利率 / 计息基准天数
	// 使用整数计算避免精度问题
	// 日利息 = 本金 * 年利率 / 计息基准天数
	// 先乘后除，避免小数精度问题
	interest := principal.Mul(c.annualRate).Div(decimal.NewFromInt(int64(c.daysInYear))).Round(2)
	return interest
}

// CalculateInterestByDays 计算指定天数的利息
// principal: 本金余额
// days: 天数
// 返回: 利息总额
func (c *DailyInterestCalculator) CalculateInterestByDays(principal decimal.Decimal, days int) decimal.Decimal {
	// 利息 = 本金 * 年利率 * 天数 / 计息基准天数
	// 先乘后除，避免小数精度问题
	interest := principal.Mul(c.annualRate).Mul(decimal.NewFromInt(int64(days))).Div(decimal.NewFromInt(int64(c.daysInYear))).Round(2)
	return interest
}

// CalculateInterestByDateRange 计算日期范围内的利息
// principal: 本金余额
// startDate: 开始日期
// endDate: 结束日期
// 返回: 利息总额
func (c *DailyInterestCalculator) CalculateInterestByDateRange(principal decimal.Decimal, startDate, endDate time.Time) decimal.Decimal {
	days := int(endDate.Sub(startDate).Hours() / 24)
	if days < 0 {
		days = 0
	}
	return c.CalculateInterestByDays(principal, days)
}

// DailyInterestDetail 每日利息明细
type DailyInterestDetail struct {
	Date           time.Time       `json:"date"`
	Principal      decimal.Decimal `json:"principal"`
	DailyRate      decimal.Decimal `json:"daily_rate"`
	Interest       decimal.Decimal `json:"interest"`
	CumulativeInterest decimal.Decimal `json:"cumulative_interest"`
}

// CalculateDailyInterestDetails 计算每日利息明细
// principal: 初始本金
// startDate: 开始日期
// endDate: 结束日期
// repayments: 还款记录（日期 -> 还款金额）
// 返回: 每日利息明细
func (c *DailyInterestCalculator) CalculateDailyInterestDetails(
	principal decimal.Decimal,
	startDate, endDate time.Time,
	repayments map[time.Time]decimal.Decimal,
) []DailyInterestDetail {
	var details []DailyInterestDetail
	currentPrincipal := principal
	cumulativeInterest := decimal.Zero()
	
	// 日利率
	dailyRate := c.annualRate.Div(decimal.NewFromInt(int64(c.daysInYear)))
	
	// 遍历每一天
	for date := startDate; !date.After(endDate); date = date.AddDate(0, 0, 1) {
		// 检查是否有还款
		if repayment, ok := repayments[date]; ok {
			currentPrincipal = currentPrincipal.Sub(repayment)
			if currentPrincipal.IsNegative() {
				currentPrincipal = decimal.Zero()
			}
		}
		
		// 计算当日利息
		interest := currentPrincipal.Mul(dailyRate).Round(2)
		cumulativeInterest = cumulativeInterest.Add(interest)
		
		details = append(details, DailyInterestDetail{
			Date:               date,
			Principal:          currentPrincipal,
			DailyRate:          dailyRate,
			Interest:           interest,
			CumulativeInterest: cumulativeInterest,
		})
	}
	
	return details
}

// CalculateMonthlyInterest 计算月利息（按日累加）
// principal: 本金
// year: 年份
// month: 月份
// 返回: 月利息
func (c *DailyInterestCalculator) CalculateMonthlyInterest(principal decimal.Decimal, year int, month time.Month) decimal.Decimal {
	startDate := time.Date(year, month, 1, 0, 0, 0, 0, time.UTC)
	endDate := startDate.AddDate(0, 1, -1) // 当月最后一天
	
	return c.CalculateInterestByDateRange(principal, startDate, endDate)
}

// CalculateInterestSchedule 计算利息计划（按月）
// principal: 初始本金
// termMonths: 期数
// startDate: 开始日期
// monthlyRepayment: 每月还款额（用于计算剩余本金）
// 返回: 每月利息计划
type MonthlyInterestSchedule struct {
	Period      int             `json:"period"`
	StartDate   time.Time       `json:"start_date"`
	EndDate     time.Time       `json:"end_date"`
	Principal   decimal.Decimal `json:"principal"`
	Days        int             `json:"days"`
	Interest    decimal.Decimal `json:"interest"`
}

func (c *DailyInterestCalculator) CalculateInterestSchedule(
	principal decimal.Decimal,
	termMonths int,
	startDate time.Time,
	monthlyRepayment decimal.Decimal,
) []MonthlyInterestSchedule {
	var schedule []MonthlyInterestSchedule
	currentPrincipal := principal
	
	for i := 0; i < termMonths; i++ {
		period := i + 1
		periodStartDate := startDate.AddDate(0, i, 0)
		periodEndDate := startDate.AddDate(0, i+1, -1)
		
		// 计算当月天数
		days := int(periodEndDate.Sub(periodStartDate).Hours()/24) + 1
		
		// 计算当月利息
		interest := c.CalculateInterestByDays(currentPrincipal, days)
		
		schedule = append(schedule, MonthlyInterestSchedule{
			Period:    period,
			StartDate: periodStartDate,
			EndDate:   periodEndDate,
			Principal: currentPrincipal,
			Days:      days,
			Interest:  interest,
		})
		
		// 更新剩余本金（假设每月还款）
		principalPart := monthlyRepayment.Sub(interest)
		if principalPart.IsNegative() {
			principalPart = decimal.Zero()
		}
		currentPrincipal = currentPrincipal.Sub(principalPart)
		if currentPrincipal.IsNegative() {
			currentPrincipal = decimal.Zero()
		}
	}
	
	return schedule
}