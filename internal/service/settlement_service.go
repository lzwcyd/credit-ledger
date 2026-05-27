package service

import (
	"fmt"
	"time"

	"github.com/yourorg/credit-ledger/internal/model"
	"github.com/yourorg/credit-ledger/pkg/decimal"
)

// =====================================================
// 提前结清
// =====================================================

// EarlySettlementTrialRequest 提前结清试算请求
type EarlySettlementTrialRequest struct {
	LoanNo    string `json:"loan_no"`
	TrialDate string `json:"trial_date"` // 试算基准日期
}

// EarlySettlementTrialResponse 提前结清试算响应
type EarlySettlementTrialResponse struct {
	LoanNo             string          `json:"loan_no"`
	TrialDate          string          `json:"trial_date"`
	RemainingPrincipal decimal.Decimal `json:"remaining_principal"`  // 剩余本金
	AccruedInterest    decimal.Decimal `json:"accrued_interest"`     // 应计利息
	AccruedPenalty     decimal.Decimal `json:"accrued_penalty"`      // 应计罚息
	UnsettledOtherFee  decimal.Decimal `json:"unsettled_other_fee"`  // 未结清其他费用
	EarlySettlementFee decimal.Decimal `json:"early_settlement_fee"` // 提前结清手续费
	TotalAmount        decimal.Decimal `json:"total_amount"`         // 总应还金额
	Details            []TrialDetailItem `json:"details"`             // 费用明细
}

// EarlySettlementRequest 提前结清入账请求
type EarlySettlementRequest struct {
	LoanNo    string  `json:"loan_no"`
	TrialDate string  `json:"trial_date"`
	BookingDate string `json:"booking_date"`
	Operator  string  `json:"operator"`
	Description string `json:"description"`
}

// EarlySettlementTrial 提前结清试算
func (s *LoanService) EarlySettlementTrial(req EarlySettlementTrialRequest) (*EarlySettlementTrialResponse, error) {
	loan, err := s.loanRepo.GetLoanByNo(req.LoanNo)
	if err != nil {
		return nil, fmt.Errorf("loan not found: %w", err)
	}

	if loan.Status != "DISBURSED" && loan.Status != "OVERDUE" {
		return nil, fmt.Errorf("loan status is %s, cannot settle early", loan.Status)
	}

	trialDate, err := time.Parse("2006-01-02", req.TrialDate)
	if err != nil {
		return nil, fmt.Errorf("invalid trial_date: %w", err)
	}

	// 计算应还利息
	accruedInterest, err := s.calculateInterest(loan, trialDate)
	if err != nil {
		return nil, fmt.Errorf("failed to calculate interest: %w", err)
	}

	// 计算应还罚息
	accruedPenalty, err := s.calculatePenalty(loan, trialDate)
	if err != nil {
		return nil, fmt.Errorf("failed to calculate penalty: %w", err)
	}

	// 计算应还其他费用
	unsettledOtherFee, err := s.calculateOtherFees(loan, trialDate)
	if err != nil {
		return nil, fmt.Errorf("failed to calculate other fees: %w", err)
	}

	// 计算提前结清手续费（基于剩余本金百分比）
	earlySettlementFee := s.calculateEarlySettlementFee(loan)

	// 总金额 = 剩余本金 + 应计利息 + 应计罚息 + 未结清其他费用 + 提前结清手续费
	totalAmount := loan.RemainingPrincipal.
		Add(accruedInterest).
		Add(accruedPenalty).
		Add(unsettledOtherFee).
		Add(earlySettlementFee)

	// 构建明细
	var details []TrialDetailItem

	if !loan.RemainingPrincipal.IsZero() {
		details = append(details, TrialDetailItem{
			FeeCode:     "PRINCIPAL",
			FeeName:     "剩余本金",
			FeeCategory: "PRINCIPAL",
			Amount:      loan.RemainingPrincipal,
		})
	}
	if !accruedInterest.IsZero() {
		details = append(details, TrialDetailItem{
			FeeCode:     "INTEREST",
			FeeName:     "应计利息",
			FeeCategory: "INTEREST",
			Amount:      accruedInterest,
		})
	}
	if !accruedPenalty.IsZero() {
		details = append(details, TrialDetailItem{
			FeeCode:     "OVERDUE_PENALTY",
			FeeName:     "应计罚息",
			FeeCategory: "PENALTY",
			Amount:      accruedPenalty,
		})
	}
	if !unsettledOtherFee.IsZero() {
		details = append(details, TrialDetailItem{
			FeeCode:     "OTHER_FEE",
			FeeName:     "未结清其他费用",
			FeeCategory: "OTHER_FEE",
			Amount:      unsettledOtherFee,
		})
	}
	if !earlySettlementFee.IsZero() {
		details = append(details, TrialDetailItem{
			FeeCode:     "EARLY_SETTLEMENT_FEE",
			FeeName:     "提前结清手续费",
			FeeCategory: "OTHER_FEE",
			Amount:      earlySettlementFee,
		})
	}

	return &EarlySettlementTrialResponse{
		LoanNo:             req.LoanNo,
		TrialDate:          req.TrialDate,
		RemainingPrincipal: loan.RemainingPrincipal,
		AccruedInterest:    accruedInterest,
		AccruedPenalty:     accruedPenalty,
		UnsettledOtherFee:  unsettledOtherFee,
		EarlySettlementFee: earlySettlementFee,
		TotalAmount:        totalAmount,
		Details:            details,
	}, nil
}

// EarlySettlement 提前结清入账
func (s *LoanService) EarlySettlement(req EarlySettlementRequest) (*RepaymentResponse, error) {
	// 试算应还金额
	trial, err := s.EarlySettlementTrial(EarlySettlementTrialRequest{
		LoanNo:    req.LoanNo,
		TrialDate: req.TrialDate,
	})
	if err != nil {
		return nil, err
	}

	// 转为还款入账请求
	repayReq := RepaymentRequest{
		LoanNo:        req.LoanNo,
		Amount:        trial.TotalAmount.String(), // 使用 String() 方法转换为字符串
		TrialDate:     req.TrialDate,
		BookingDate:   req.BookingDate,
		RepaymentType: "EARLY_SETTLEMENT",
		Description:   req.Description,
		CreatedBy:     req.Operator,
	}

	resp, err := s.Repayment(repayReq)
	if err != nil {
		return nil, err
	}

	// 更新借据状态为结清
	loan, err := s.loanRepo.GetLoanByNo(req.LoanNo)
	if err != nil {
		return nil, fmt.Errorf("loan not found after repayment: %w", err)
	}

	oldStatus := loan.Status
	loan.Status = "REPAID"
	now := time.Now()
	loan.SettlementDate = &now
	loan.UpdatedBy = req.Operator
	if err := s.loanRepo.UpdateLoan(loan); err != nil {
		return nil, fmt.Errorf("failed to update loan status: %w", err)
	}

	// 记录变更
	change := &model.LoanChange{
		LoanNo:       loan.LoanNo,
		ChangeType:   "EARLY_SETTLEMENT",
		FieldName:    "status",
		OldValue:     oldStatus,
		NewValue:     "REPAID",
		ChangeReason: "提前结清",
		CreatedBy:    req.Operator,
		UpdatedBy:    req.Operator,
	}
	if err := s.loanRepo.CreateLoanChange(change); err != nil {
		return nil, fmt.Errorf("failed to create loan change: %w", err)
	}

	// 将所有未结清的计划标记为结清
	plans, err := s.planRepo.GetPlansByLoanNo(req.LoanNo)
	if err == nil {
		for i := range plans {
			if plans[i].Status != "PAID" {
				plans[i].Status = "PAID"
				plans[i].UpdatedBy = req.Operator
				if err := s.planRepo.UpdatePlan(&plans[i]); err != nil {
					return nil, fmt.Errorf("failed to update plan %d: %w", plans[i].ID, err)
				}
			}
		}
	}

	return resp, nil
}

// calculateEarlySettlementFee 计算提前结清手续费
// 默认按剩余本金的 1% 收取，实际应从 fee_configs 读取
func (s *LoanService) calculateEarlySettlementFee(loan *model.Loan) decimal.Decimal {
	// 尝试从 fee_configs 获取 EARLY_REPAYMENT 类型的费项
	configs, err := s.feeConfigRepo.GetByTriggerType("EARLY_REPAYMENT")
	if err == nil && len(configs) > 0 {
		for _, cfg := range configs {
			if cfg.IsActive {
				return s.calculateFeeFromConfig(cfg, loan.RemainingPrincipal, loan.Principal)
			}
		}
	}

	// 默认：剩余本金的 1%
	return loan.RemainingPrincipal.Mul(decimal.NewFromFloat(0.01)).Round(2)
}

// calculateFeeFromConfig 根据费项配置计算费用
func (s *LoanService) calculateFeeFromConfig(cfg model.FeeConfig, remainingPrincipal, principal decimal.Decimal) decimal.Decimal {
	switch cfg.CalcType {
	case "FIXED":
		return cfg.Value
	case "PERCENTAGE":
		var base decimal.Decimal
		switch cfg.CalcBase {
		case "REMAINING_PRINCIPAL":
			base = remainingPrincipal
		case "PRINCIPAL":
			base = principal
		default:
			base = remainingPrincipal
		}
		fee := base.Mul(cfg.Value).Div(decimal.NewFromInt(100)).Round(2)
		if cfg.MinAmount.Gt(fee) {
			return cfg.MinAmount
		}
		if cfg.MaxAmount != nil && cfg.MaxAmount.Lt(fee) {
			return *cfg.MaxAmount
		}
		return fee
	default:
		return decimal.Zero()
	}
}

// =====================================================
// 部分还款
// =====================================================

// PartialRepaymentTrialRequest 部分还款试算请求
type PartialRepaymentTrialRequest struct {
	LoanNo    string  `json:"loan_no"`
	TrialDate string  `json:"trial_date"`
	Amount    string  `json:"amount"` // 改为 string 类型，避免精度损失
	RuleCode  string  `json:"rule_code,omitempty"`
}

// PartialRepaymentTrialResponse 部分还款试算响应
type PartialRepaymentTrialResponse struct {
	LoanNo                  string                          `json:"loan_no"`
	TrialDate               string                          `json:"trial_date"`
	InputAmount             decimal.Decimal                 `json:"input_amount"`
	Allocations             []PartialAllocation             `json:"allocations"`
	RemainingPrincipalAfter decimal.Decimal                 `json:"remaining_principal_after"`
	RemainingBalance        map[string]decimal.Decimal      `json:"remaining_balance"`
	Overpayment             decimal.Decimal                 `json:"overpayment"` // 超额部分
}

// PartialAllocation 部分还款分配明细
type PartialAllocation struct {
	FeeCategory string          `json:"fee_category"`
	FeeName     string          `json:"fee_name"`
	Amount      decimal.Decimal `json:"amount"`
}

// PartialRepaymentTrial 部分还款试算
func (s *LoanService) PartialRepaymentTrial(req PartialRepaymentTrialRequest) (*PartialRepaymentTrialResponse, error) {
	loan, err := s.loanRepo.GetLoanByNo(req.LoanNo)
	if err != nil {
		return nil, fmt.Errorf("loan not found: %w", err)
	}

	if loan.Status != "DISBURSED" && loan.Status != "OVERDUE" {
		return nil, fmt.Errorf("loan status is %s, cannot make partial repayment", loan.Status)
	}

	trialDate, err := time.Parse("2006-01-02", req.TrialDate)
	if err != nil {
		return nil, fmt.Errorf("invalid trial_date: %w", err)
	}

	inputAmount, err := decimal.NewFromString(req.Amount)
	if err != nil {
		return nil, fmt.Errorf("invalid amount: %w", err)
	}

	// 获取分配规则
	ruleCode := req.RuleCode
	if ruleCode == "" {
		ruleCode = loan.AllocationRuleCode
	}
	if ruleCode == "" {
		ruleCode = "DEFAULT"
	}

	ruleItems, err := s.allocationRepo.GetRuleItems(ruleCode)
	if err != nil {
		return nil, fmt.Errorf("allocation rule not found: %w", err)
	}

	// 计算各类待还金额
	accruedInterest, err := s.calculateInterest(loan, trialDate)
	if err != nil {
		return nil, fmt.Errorf("failed to calculate interest: %w", err)
	}
	accruedPenalty, err := s.calculatePenalty(loan, trialDate)
	if err != nil {
		return nil, fmt.Errorf("failed to calculate penalty: %w", err)
	}
	unsettledOtherFee, err := s.calculateOtherFees(loan, trialDate)
	if err != nil {
		return nil, fmt.Errorf("failed to calculate other fees: %w", err)
	}

	// 按优先级分配
	remaining := inputAmount
	var allocations []PartialAllocation

	for _, item := range ruleItems {
		if remaining.IsZero() || remaining.IsNegative() {
			break
		}

		var available decimal.Decimal
		switch item.AllocationType {
		case "PENALTY":
			available = accruedPenalty
		case "OTHER_FEE":
			available = unsettledOtherFee
		case "INTEREST":
			available = accruedInterest
		case "PRINCIPAL":
			available = loan.RemainingPrincipal
		default:
			continue
		}

		if available.IsZero() || available.IsNegative() {
			continue
		}

		allocAmount := decimal.Min(remaining, available)
		if allocAmount.IsPositive() {
			allocations = append(allocations, PartialAllocation{
				FeeCategory: item.AllocationType,
				FeeName:     s.getFeeNameByType(item.AllocationType),
				Amount:      allocAmount,
			})
			remaining = remaining.Sub(allocAmount)
		}
	}

	// 计算还款后的剩余本金
	remainingPrincipalAfter := loan.RemainingPrincipal
	for _, alloc := range allocations {
		if alloc.FeeCategory == "PRINCIPAL" {
			remainingPrincipalAfter = remainingPrincipalAfter.Sub(alloc.Amount)
		}
	}

	// 计算各类剩余
	remainingInterest := accruedInterest
	remainingPenalty := accruedPenalty
	remainingOtherFee := unsettledOtherFee
	remainingPrincipal := loan.RemainingPrincipal

	for _, alloc := range allocations {
		switch alloc.FeeCategory {
		case "INTEREST":
			remainingInterest = remainingInterest.Sub(alloc.Amount)
		case "PENALTY":
			remainingPenalty = remainingPenalty.Sub(alloc.Amount)
		case "OTHER_FEE":
			remainingOtherFee = remainingOtherFee.Sub(alloc.Amount)
		case "PRINCIPAL":
			remainingPrincipal = remainingPrincipal.Sub(alloc.Amount)
		}
	}

	return &PartialRepaymentTrialResponse{
		LoanNo:                  req.LoanNo,
		TrialDate:               req.TrialDate,
		InputAmount:             inputAmount,
		Allocations:             allocations,
		RemainingPrincipalAfter: remainingPrincipalAfter,
		RemainingBalance: map[string]decimal.Decimal{
			"PRINCIPAL":  remainingPrincipal,
			"INTEREST":   remainingInterest,
			"PENALTY":    remainingPenalty,
			"OTHER_FEE":  remainingOtherFee,
		},
		Overpayment: remaining,
	}, nil
}

// PartialRepayment 部分还款入账
func (s *LoanService) PartialRepayment(req RepaymentRequest) (*RepaymentResponse, error) {
	req.RepaymentType = "PARTIAL"

	loan, err := s.loanRepo.GetLoanByNo(req.LoanNo)
	if err != nil {
		return nil, fmt.Errorf("loan not found: %w", err)
	}

	if loan.Status != "DISBURSED" && loan.Status != "OVERDUE" {
		return nil, fmt.Errorf("loan status is %s, cannot make partial repayment", loan.Status)
	}

	// 复用 Repayment 逻辑
	resp, err := s.Repayment(req)
	if err != nil {
		return nil, err
	}

	return resp, nil
}

// =====================================================
// 还款计划查询（增强）
// =====================================================

// PlanSummaryResponse 还款计划汇总响应
type PlanSummaryResponse struct {
	LoanNo           string          `json:"loan_no"`
	TotalPeriods     int             `json:"total_periods"`
	PaidPeriods      int             `json:"paid_periods"`
	OverduePeriods   int             `json:"overdue_periods"`
	PendingPeriods   int             `json:"pending_periods"`
	TotalPrincipal   decimal.Decimal `json:"total_principal"`
	TotalInterest    decimal.Decimal `json:"total_interest"`
	PaidPrincipal    decimal.Decimal `json:"paid_principal"`
	PaidInterest     decimal.Decimal `json:"paid_interest"`
	RemainingPrincipal decimal.Decimal `json:"remaining_principal"`
	NextDueDate      string          `json:"next_due_date,omitempty"`
	NextDueAmount    decimal.Decimal `json:"next_due_amount,omitempty"`
	Plans            []model.Plan    `json:"plans"`
}

// GetPlanSummary 获取还款计划汇总
func (s *LoanService) GetPlanSummary(loanNo string) (*PlanSummaryResponse, error) {
	loan, err := s.loanRepo.GetLoanByNo(loanNo)
	if err != nil {
		return nil, fmt.Errorf("loan not found: %w", err)
	}

	plans, err := s.planRepo.GetPlansByLoanNo(loanNo)
	if err != nil {
		return nil, fmt.Errorf("failed to get plans: %w", err)
	}

	summary := &PlanSummaryResponse{
		LoanNo:             loanNo,
		TotalPeriods:       len(plans),
		RemainingPrincipal: loan.RemainingPrincipal,
		Plans:              plans,
	}

	var nextDueDate *time.Time

	for _, plan := range plans {
		summary.TotalPrincipal = summary.TotalPrincipal.Add(plan.DuePrincipal)
		summary.TotalInterest = summary.TotalInterest.Add(plan.DueInterest)
		summary.PaidPrincipal = summary.PaidPrincipal.Add(plan.PaidPrincipal)
		summary.PaidInterest = summary.PaidInterest.Add(plan.PaidInterest)

		switch plan.Status {
		case "PAID":
			summary.PaidPeriods++
		case "OVERDUE":
			summary.OverduePeriods++
		case "PENDING":
			summary.PendingPeriods++
			if nextDueDate == nil || plan.DueDate.Before(*nextDueDate) {
				dueDate := plan.DueDate
				nextDueDate = &dueDate
			}
		}
	}

	if nextDueDate != nil {
		summary.NextDueDate = nextDueDate.Format("2006-01-02")
		// 计算下期应还金额
		for _, plan := range plans {
			if plan.DueDate.Equal(*nextDueDate) {
				summary.NextDueAmount = plan.DueTotal.Sub(plan.PaidTotal)
				break
			}
		}
	}

	return summary, nil
}
