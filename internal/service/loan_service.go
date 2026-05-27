package service

import (
	"fmt"
	"time"
	"github.com/lzwcyd/credit-ledger/internal/calculator"
	"github.com/lzwcyd/credit-ledger/internal/model"
	"github.com/lzwcyd/credit-ledger/internal/repository"
	"github.com/lzwcyd/credit-ledger/pkg/decimal"
)

// =====================================================
// 请求/响应结构
// =====================================================

// CreateLoanRequest 创建借据请求
type CreateLoanRequest struct {
	LoanNo             string  `json:"loan_no"`
	Principal          string  `json:"principal"`          // 改为 string 类型，避免精度损失
	AnnualRate         string  `json:"annual_rate"`        // 改为 string 类型
	TermMonths         int     `json:"term_months"`
	RepaymentTypeCode  string  `json:"repayment_type_code"`
	ValueDate          string  `json:"value_date"`
	FirstDueDate       string  `json:"first_due_date"`
	MaturityDate       string  `json:"maturity_date"`
	CreatedBy          string  `json:"created_by"`
}

// DisburseRequest 放款请求
type DisburseRequest struct {
	LoanNo           string  `json:"loan_no"`
	DisburseDate     string  `json:"disburse_date"`
	DisburseAmount   string  `json:"disburse_amount"`    // 改为 string 类型，避免精度损失
	CreatedBy        string  `json:"created_by"`
}

// RepaymentTrialRequest 还款试算请求
type RepaymentTrialRequest struct {
	LoanNo    string `json:"loan_no"`
	TrialDate string `json:"trial_date"`
}

// RepaymentTrialResponse 还款试算响应
type RepaymentTrialResponse struct {
	LoanNo           string             `json:"loan_no"`
	TrialDate        string             `json:"trial_date"`
	RemainingPrincipal decimal.Decimal  `json:"remaining_principal"`
	InterestAmount   decimal.Decimal    `json:"interest_amount"`
	PenaltyAmount    decimal.Decimal    `json:"penalty_amount"`
	OtherFeeAmount   decimal.Decimal    `json:"other_fee_amount"`
	TotalAmount      decimal.Decimal    `json:"total_amount"`
	Details          []TrialDetailItem  `json:"details"`
}

// TrialDetailItem 试算明细项
type TrialDetailItem struct {
	FeeCode     string          `json:"fee_code"`
	FeeName     string          `json:"fee_name"`
	FeeCategory string          `json:"fee_category"`
	Amount      decimal.Decimal `json:"amount"`
}

// RepaymentRequest 还款入账请求
type RepaymentRequest struct {
	LoanNo        string  `json:"loan_no"`
	Amount        string  `json:"amount"`           // 改为 string 类型，避免精度损失
	TrialDate     string  `json:"trial_date"`
	BookingDate   string  `json:"booking_date"`
	RepaymentType string `json:"repayment_type"` // NORMAL, EARLY_SETTLEMENT, PARTIAL
	Description   string  `json:"description"`
	IsBackdated   bool    `json:"is_backdated"`
	BackdatedReason string `json:"backdated_reason"`
	CreatedBy     string  `json:"created_by"`
}

// RepaymentResponse 还款入账响应
type RepaymentResponse struct {
	RepaymentNo     string            `json:"repayment_no"`
	LoanNo          string            `json:"loan_no"`
	Amount          decimal.Decimal   `json:"amount"`
	Details         []model.RepaymentDetail `json:"details"`
	BookingDate     string            `json:"booking_date"`
}

// LoanService 借据服务
type LoanService struct {
	loanRepo          *repository.LoanRepository
	planRepo          *repository.PlanRepository
	repaymentRepo     *repository.RepaymentRepository
	dailyCalcRepo     *repository.DailyCalculationRepository
	planOtherFeeRepo  *repository.PlanOtherFeeRepository
	feeConfigRepo     *repository.FeeConfigRepository
	allocationRepo    *repository.AllocationRuleRepository
	calculatorFactory *calculator.CalculatorFactory
}

func NewLoanService(
	loanRepo *repository.LoanRepository,
	planRepo *repository.PlanRepository,
	repaymentRepo *repository.RepaymentRepository,
	dailyCalcRepo *repository.DailyCalculationRepository,
	planOtherFeeRepo *repository.PlanOtherFeeRepository,
	feeConfigRepo *repository.FeeConfigRepository,
	allocationRepo *repository.AllocationRuleRepository,
) *LoanService {
	return &LoanService{
		loanRepo:          loanRepo,
		planRepo:          planRepo,
		repaymentRepo:     repaymentRepo,
		dailyCalcRepo:     dailyCalcRepo,
		planOtherFeeRepo:  planOtherFeeRepo,
		feeConfigRepo:     feeConfigRepo,
		allocationRepo:    allocationRepo,
		calculatorFactory: calculator.NewCalculatorFactory(),
	}
}

// CreateLoan 创建借据
func (s *LoanService) CreateLoan(req CreateLoanRequest) (*model.Loan, error) {
	// 解析金额
	principal, err := decimal.NewFromString(req.Principal)
	if err != nil {
		return nil, fmt.Errorf("invalid principal: %w", err)
	}
	annualRate, err := decimal.NewFromString(req.AnnualRate)
	if err != nil {
		return nil, fmt.Errorf("invalid annual_rate: %w", err)
	}

	// 解析日期
	valueDate, err := time.Parse("2006-01-02", req.ValueDate)
	if err != nil {
		return nil, fmt.Errorf("invalid value_date: %w", err)
	}

	firstDueDate, err := time.Parse("2006-01-02", req.FirstDueDate)
	if err != nil {
		return nil, fmt.Errorf("invalid first_due_date: %w", err)
	}

	maturityDate, err := time.Parse("2006-01-02", req.MaturityDate)
	if err != nil {
		return nil, fmt.Errorf("invalid maturity_date: %w", err)
	}

	// 创建借据
	loan := &model.Loan{
		LoanNo:             req.LoanNo,
		Principal:          principal,
		AnnualRate:         annualRate,
		TermMonths:         req.TermMonths,
		RepaymentTypeCode:  req.RepaymentTypeCode,
		AllocationRuleCode: "DEFAULT",
		ValueDate:          valueDate,
		FirstDueDate:       firstDueDate,
		MaturityDate:       maturityDate,
		Status:             "PENDING",
		RemainingPrincipal: principal,
		CreatedBy:          req.CreatedBy,
		UpdatedBy:          req.CreatedBy,
	}
	
	err = s.loanRepo.CreateLoan(loan)
	if err != nil {
		return nil, fmt.Errorf("failed to create loan: %w", err)
	}
	
	// 生成还款计划
	err = s.generateRepaymentPlans(loan)
	if err != nil {
		return nil, fmt.Errorf("failed to generate repayment plans: %w", err)
	}
	
	return loan, nil
}

// generateRepaymentPlans 生成还款计划
func (s *LoanService) generateRepaymentPlans(loan *model.Loan) error {
	// 获取计算器
	calculator := s.calculatorFactory.GetCalculator(loan.RepaymentTypeCode)
	
	// 计算还款计划
	schedules := calculator.CalculateSchedule(
		loan.Principal,
		loan.AnnualRate,
		loan.TermMonths,
		loan.FirstDueDate,
	)
	
	// 保存还款计划
	for _, schedule := range schedules {
		plan := &model.Plan{
			LoanNo:       loan.LoanNo,
			Period:       schedule.Period,
			DueDate:      schedule.DueDate,
			DuePrincipal: schedule.PrincipalDue,
			DueInterest:  schedule.InterestDue,
			DuePenalty:   decimal.Zero(),
			DueOtherFee:  decimal.Zero(),
			DueTotal:     schedule.TotalDue,
			Status:       "PENDING",
			CreatedBy:    loan.CreatedBy,
			UpdatedBy:    loan.CreatedBy,
		}
		
		err := s.planRepo.CreatePlan(plan)
		if err != nil {
			return err
		}
	}
	
	return nil
}

// Disburse 放款
func (s *LoanService) Disburse(req DisburseRequest) (*model.Loan, error) {
	// 获取借据
	loan, err := s.loanRepo.GetLoanByNo(req.LoanNo)
	if err != nil {
		return nil, fmt.Errorf("loan not found: %w", err)
	}
	
	// 检查状态
	if loan.Status != "PENDING" {
		return nil, fmt.Errorf("loan status is %s, cannot disburse", loan.Status)
	}
	
	// 解析日期
	disburseDate, err := time.Parse("2006-01-02", req.DisburseDate)
	if err != nil {
		return nil, fmt.Errorf("invalid disburse_date: %w", err)
	}
	
	// 解析放款金额
	disburseAmount, err := decimal.NewFromString(req.DisburseAmount)
	if err != nil {
		return nil, fmt.Errorf("invalid disburse_amount: %w", err)
	}

	// 更新借据
	loan.Status = "DISBURSED"
	loan.DisbursementDate = &disburseDate
	loan.DisbursedAmount = disburseAmount
	loan.RemainingPrincipal = disburseAmount
	loan.UpdatedBy = req.CreatedBy
	
	err = s.loanRepo.UpdateLoan(loan)
	if err != nil {
		return nil, fmt.Errorf("failed to update loan: %w", err)
	}
	
	// 记录变更
	change := &model.LoanChange{
		LoanNo:       loan.LoanNo,
		ChangeType:   "DISBURSE",
		FieldName:    "status",
		OldValue:     "PENDING",
		NewValue:     "DISBURSED",
		ChangeReason: "放款",
		CreatedBy:    req.CreatedBy,
		UpdatedBy:    req.CreatedBy,
	}
	err = s.loanRepo.CreateLoanChange(change)
	if err != nil {
		return nil, fmt.Errorf("failed to create loan change: %w", err)
	}
	
	return loan, nil
}

// RepaymentTrial 还款试算
func (s *LoanService) RepaymentTrial(req RepaymentTrialRequest) (*RepaymentTrialResponse, error) {
	// 获取借据
	loan, err := s.loanRepo.GetLoanByNo(req.LoanNo)
	if err != nil {
		return nil, fmt.Errorf("loan not found: %w", err)
	}
	
	// 解析试算日期
	trialDate, err := time.Parse("2006-01-02", req.TrialDate)
	if err != nil {
		return nil, fmt.Errorf("invalid trial_date: %w", err)
	}
	
	// 计算应还利息
	interestAmount, err := s.calculateInterest(loan, trialDate)
	if err != nil {
		return nil, fmt.Errorf("failed to calculate interest: %w", err)
	}

	// 计算应还罚息
	penaltyAmount, err := s.calculatePenalty(loan, trialDate)
	if err != nil {
		return nil, fmt.Errorf("failed to calculate penalty: %w", err)
	}

	// 计算应还其他费用
	otherFeeAmount, err := s.calculateOtherFees(loan, trialDate)
	if err != nil {
		return nil, fmt.Errorf("failed to calculate other fees: %w", err)
	}
	
	// 构建明细
	var details []TrialDetailItem
	
	// 利息明细
	if !interestAmount.IsZero() {
		details = append(details, TrialDetailItem{
			FeeCode:     "INTEREST",
			FeeName:     "利息",
			FeeCategory: "INTEREST",
			Amount:      interestAmount,
		})
	}
	
	// 罚息明细
	if !penaltyAmount.IsZero() {
		details = append(details, TrialDetailItem{
			FeeCode:     "OVERDUE_PENALTY",
			FeeName:     "逾期罚息",
			FeeCategory: "PENALTY",
			Amount:      penaltyAmount,
		})
	}
	
	// 其他费用明细
	otherFees := s.getOtherFeeDetails(loan, trialDate)
	details = append(details, otherFees...)
	
	// 总金额
	totalAmount := interestAmount.Add(penaltyAmount).Add(otherFeeAmount)
	
	return &RepaymentTrialResponse{
		LoanNo:           loan.LoanNo,
		TrialDate:        req.TrialDate,
		RemainingPrincipal: loan.RemainingPrincipal,
		InterestAmount:   interestAmount,
		PenaltyAmount:    penaltyAmount,
		OtherFeeAmount:   otherFeeAmount,
		TotalAmount:      totalAmount,
		Details:          details,
	}, nil
}

// calculateInterest 计算利息
func (s *LoanService) calculateInterest(loan *model.Loan, trialDate time.Time) (decimal.Decimal, error) {
	// 获取未结清的利息记录
	unsettledInterest, err := s.dailyCalcRepo.GetUnsettledByLoanNo(loan.LoanNo, "INTEREST")
	if err != nil {
		return decimal.Zero(), fmt.Errorf("failed to get unsettled interest for loan %s: %w", loan.LoanNo, err)
	}

	// 计算累计利息
	totalInterest := decimal.Zero()
	for _, calc := range unsettledInterest {
		totalInterest = totalInterest.Add(calc.Amount)
	}

	return totalInterest, nil
}

// calculatePenalty 计算罚息
func (s *LoanService) calculatePenalty(loan *model.Loan, trialDate time.Time) (decimal.Decimal, error) {
	// 获取未结清的罚息记录
	unsettledPenalty, err := s.dailyCalcRepo.GetUnsettledByLoanNo(loan.LoanNo, "PENALTY")
	if err != nil {
		return decimal.Zero(), fmt.Errorf("failed to get unsettled penalty for loan %s: %w", loan.LoanNo, err)
	}

	// 计算累计罚息
	totalPenalty := decimal.Zero()
	for _, calc := range unsettledPenalty {
		totalPenalty = totalPenalty.Add(calc.Amount)
	}

	return totalPenalty, nil
}

// calculateOtherFees 计算其他费用
func (s *LoanService) calculateOtherFees(loan *model.Loan, trialDate time.Time) (decimal.Decimal, error) {
	// 获取所有未还清的其他费用
	plans, err := s.planRepo.GetPlansByLoanNo(loan.LoanNo)
	if err != nil {
		return decimal.Zero(), fmt.Errorf("failed to get plans for loan %s: %w", loan.LoanNo, err)
	}

	totalOtherFee := decimal.Zero()
	for _, plan := range plans {
		otherFees, err := s.planOtherFeeRepo.GetUnpaidByPlanID(plan.ID)
		if err != nil {
			return decimal.Zero(), fmt.Errorf("failed to get unpaid fees for plan %d: %w", plan.ID, err)
		}

		for _, fee := range otherFees {
			remaining := fee.DueAmount.Sub(fee.PaidAmount)
			totalOtherFee = totalOtherFee.Add(remaining)
		}
	}

	return totalOtherFee, nil
}

// getOtherFeeDetails 获取其他费用明细
func (s *LoanService) getOtherFeeDetails(loan *model.Loan, trialDate time.Time) []TrialDetailItem {
	var details []TrialDetailItem
	
	plans, err := s.planRepo.GetPlansByLoanNo(loan.LoanNo)
	if err != nil {
		return details
	}
	
	for _, plan := range plans {
		otherFees, err := s.planOtherFeeRepo.GetUnpaidByPlanID(plan.ID)
		if err != nil {
			continue
		}
		
		for _, fee := range otherFees {
			remaining := fee.DueAmount.Sub(fee.PaidAmount)
			if !remaining.IsZero() {
				details = append(details, TrialDetailItem{
					FeeCode:     fee.FeeCode,
					FeeName:     fee.FeeName,
					FeeCategory: "OTHER_FEE",
					Amount:      remaining,
				})
			}
		}
	}
	
	return details
}

// Repayment 还款入账
func (s *LoanService) Repayment(req RepaymentRequest) (*RepaymentResponse, error) {
	// 获取借据
	loan, err := s.loanRepo.GetLoanByNo(req.LoanNo)
	if err != nil {
		return nil, fmt.Errorf("loan not found: %w", err)
	}
	
	// 解析日期
	trialDate, err := time.Parse("2006-01-02", req.TrialDate)
	if err != nil {
		return nil, fmt.Errorf("invalid trial_date: %w", err)
	}
	
	bookingDate, err := time.Parse("2006-01-02", req.BookingDate)
	if err != nil {
		return nil, fmt.Errorf("invalid booking_date: %w", err)
	}
	
	// 生成还款编号
	repaymentNo := fmt.Sprintf("RP%s%s", time.Now().Format("20060102150405"), loan.LoanNo[len(loan.LoanNo)-4:])
	
	// 获取分配规则
	ruleItems, err := s.allocationRepo.GetRuleItems(loan.AllocationRuleCode)
	if err != nil {
		return nil, fmt.Errorf("failed to get allocation rule: %w", err)
	}
	
	// 按规则分配还款金额
	amount, err := decimal.NewFromString(req.Amount)
	if err != nil {
		return nil, fmt.Errorf("invalid amount: %w", err)
	}
	remaining := amount
	var details []model.RepaymentDetail
	
	for _, rule := range ruleItems {
		if remaining.IsZero() {
			break
		}
		
		// 获取该类型待还金额
		dueAmount, err := s.getDueAmountByType(loan, rule.AllocationType, trialDate)
		if err != nil {
			return nil, fmt.Errorf("failed to get due amount for %s: %w", rule.AllocationType, err)
		}
		if dueAmount.IsZero() {
			continue
		}
		
		// 计算本次可还金额
		payAmount := decimal.Min(remaining, dueAmount)
		if payAmount.IsZero() {
			continue
		}
		
		// 创建还款明细
		detail := &model.RepaymentDetail{
			RepaymentNo: repaymentNo,
			LoanNo:      loan.LoanNo,
			FeeCode:     rule.AllocationType,
			FeeName:     s.getFeeNameByType(rule.AllocationType),
			FeeCategory: s.getFeeCategoryByType(rule.AllocationType),
			Amount:      payAmount,
			CreatedBy:   req.CreatedBy,
			UpdatedBy:   req.CreatedBy,
		}
		
		err = s.repaymentRepo.CreateRepaymentDetail(detail)
		if err != nil {
			return nil, fmt.Errorf("failed to create repayment detail: %w", err)
		}
		
		details = append(details, *detail)
		remaining = remaining.Sub(payAmount)
	}
	
	// 创建还款记录
	repayment := &model.Repayment{
		RepaymentNo:        repaymentNo,
		LoanNo:             loan.LoanNo,
		RepaymentType:      req.RepaymentType,
		Amount:             amount,
		PrincipalAmount:    s.getAmountByCategory(details, "PRINCIPAL"),
		InterestAmount:     s.getAmountByCategory(details, "INTEREST"),
		PenaltyAmount:      s.getAmountByCategory(details, "PENALTY"),
		OtherFeeAmount:     s.getAmountByCategory(details, "OTHER_FEE"),
		TrialDate:          trialDate,
		BookingDate:        bookingDate,
		AllocationRuleCode: loan.AllocationRuleCode,
		Status:             "BOOKED",
		Description:        req.Description,
		IsBackdated:        req.IsBackdated,
		BackdatedReason:    req.BackdatedReason,
		CreatedBy:          req.CreatedBy,
		UpdatedBy:          req.CreatedBy,
	}
	
	err = s.repaymentRepo.CreateRepayment(repayment)
	if err != nil {
		return nil, fmt.Errorf("failed to create repayment: %w", err)
	}
	
	// 更新借据
	if err := s.updateLoanAfterRepayment(loan, amount, details, req.CreatedBy); err != nil {
		return nil, fmt.Errorf("failed to update loan after repayment: %w", err)
	}
	
	return &RepaymentResponse{
		RepaymentNo: repaymentNo,
		LoanNo:      loan.LoanNo,
		Amount:      amount,
		Details:     details,
		BookingDate: req.BookingDate,
	}, nil
}

// getDueAmountByType 获取指定类型的待还金额
func (s *LoanService) getDueAmountByType(loan *model.Loan, allocationType string, trialDate time.Time) (decimal.Decimal, error) {
	switch allocationType {
	case "PENALTY":
		return s.calculatePenalty(loan, trialDate)
	case "INTEREST":
		return s.calculateInterest(loan, trialDate)
	case "OTHER_FEE":
		return s.calculateOtherFees(loan, trialDate)
	case "PRINCIPAL":
		return loan.RemainingPrincipal, nil
	default:
		return decimal.Zero(), nil
	}
}

// getFeeNameByType 获取费用名称
func (s *LoanService) getFeeNameByType(allocationType string) string {
	switch allocationType {
	case "PENALTY":
		return "逾期罚息"
	case "INTEREST":
		return "利息"
	case "OTHER_FEE":
		return "其他费用"
	case "PRINCIPAL":
		return "本金"
	default:
		return allocationType
	}
}

// getFeeCategoryByType 获取费用类别
func (s *LoanService) getFeeCategoryByType(allocationType string) string {
	switch allocationType {
	case "PENALTY":
		return "PENALTY"
	case "INTEREST":
		return "INTEREST"
	case "OTHER_FEE":
		return "OTHER_FEE"
	case "PRINCIPAL":
		return "PRINCIPAL"
	default:
		return "OTHER_FEE"
	}
}

// getAmountByCategory 获取指定类别的金额
func (s *LoanService) getAmountByCategory(details []model.RepaymentDetail, category string) decimal.Decimal {
	total := decimal.Zero()
	for _, detail := range details {
		if detail.FeeCategory == category {
			total = total.Add(detail.Amount)
		}
	}
	return total
}

// updateLoanAfterRepayment 还款后更新借据
func (s *LoanService) updateLoanAfterRepayment(loan *model.Loan, amount decimal.Decimal, details []model.RepaymentDetail, updatedBy string) error {
	// 更新已还金额
	for _, detail := range details {
		switch detail.FeeCategory {
		case "INTEREST":
			loan.PaidInterest = loan.PaidInterest.Add(detail.Amount)
		case "PENALTY":
			loan.PaidPenalty = loan.PaidPenalty.Add(detail.Amount)
		case "OTHER_FEE":
			loan.PaidOtherFee = loan.PaidOtherFee.Add(detail.Amount)
		case "PRINCIPAL":
			loan.PaidPrincipal = loan.PaidPrincipal.Add(detail.Amount)
			loan.RemainingPrincipal = loan.RemainingPrincipal.Sub(detail.Amount)
		}
	}
	
	// 检查是否结清
	if loan.RemainingPrincipal.IsZero() &&
		loan.TotalInterest.Sub(loan.PaidInterest).IsZero() &&
		loan.TotalPenalty.Sub(loan.PaidPenalty).IsZero() &&
		loan.TotalOtherFee.Sub(loan.PaidOtherFee).IsZero() {
		now := time.Now()
		loan.Status = "REPAID"
		loan.SettlementDate = &now
	}
	
	loan.UpdatedBy = updatedBy
	return s.loanRepo.UpdateLoan(loan)
}

// GetLoanByNo 获取借据
func (s *LoanService) GetLoanByNo(loanNo string) (*model.Loan, error) {
	return s.loanRepo.GetLoanByNo(loanNo)
}

// GetPlansByLoanNo 获取还款计划列表
func (s *LoanService) GetPlansByLoanNo(loanNo string) ([]model.Plan, error) {
	return s.planRepo.GetPlansByLoanNo(loanNo)
}