package service

import (
	"fmt"
	"time"
	"github.com/lzwcyd/credit-ledger/internal/model"
	"github.com/lzwcyd/credit-ledger/internal/repository"
	"github.com/lzwcyd/credit-ledger/pkg/decimal"
)

// BatchService 跑批服务
type BatchService struct {
	loanRepo         *repository.LoanRepository
	planRepo         *repository.PlanRepository
	dailyCalcRepo    *repository.DailyCalculationRepository
	batchRepo        *repository.BatchJobRepository
}

func NewBatchService(
	loanRepo *repository.LoanRepository,
	planRepo *repository.PlanRepository,
	dailyCalcRepo *repository.DailyCalculationRepository,
	batchRepo *repository.BatchJobRepository,
) *BatchService {
	return &BatchService{
		loanRepo:      loanRepo,
		planRepo:      planRepo,
		dailyCalcRepo: dailyCalcRepo,
		batchRepo:     batchRepo,
	}
}

// DailyInterestCalcRequest 每日计息请求
type DailyInterestCalcRequest struct {
	BatchDate time.Time
	BatchNo   string
	Operator  string
}

// DailyInterestCalcResult 每日计息结果
type DailyInterestCalcResult struct {
	BatchNo        string
	BatchDate      time.Time
	TotalCount     int
	SuccessCount   int
	FailedCount    int
	Duration       time.Duration
}

// CalculateDailyInterest 每日计息
func (s *BatchService) CalculateDailyInterest(req DailyInterestCalcRequest) (*DailyInterestCalcResult, error) {
	startTime := time.Now()
	
	// 创建批次
	batch := &model.BatchJob{
		BatchNo:   req.BatchNo,
		BatchType: "DAILY_CALC",
		BatchDate: req.BatchDate,
		Status:    "RUNNING",
		StartTime: &startTime,
		CreatedBy: req.Operator,
		UpdatedBy: req.Operator,
	}
	
	err := s.batchRepo.CreateBatchJob(batch)
	if err != nil {
		return nil, fmt.Errorf("failed to create batch job: %w", err)
	}
	
	// 获取所有已放款的借据
	loans, err := s.batchRepo.GetDisbursedLoans()
	if err != nil {
		return nil, fmt.Errorf("failed to get loans: %w", err)
	}
	
	result := &DailyInterestCalcResult{
		BatchNo:    req.BatchNo,
		BatchDate:  req.BatchDate,
		TotalCount: len(loans),
	}
	
	// 逐个计算利息
	for _, loan := range loans {
		err := s.calculateLoanDailyInterest(&loan, req.BatchDate, req.BatchNo, req.Operator)
		if err != nil {
			result.FailedCount++
			continue
		}
		result.SuccessCount++
	}
	
	// 更新批次状态
	endTime := time.Now()
	duration := endTime.Sub(startTime)
	
	batch.Status = "SUCCESS"
	if result.FailedCount > 0 {
		batch.Status = "PARTIAL"
	}
	batch.TotalCount = result.TotalCount
	batch.SuccessCount = result.SuccessCount
	batch.FailedCount = result.FailedCount
	batch.ProcessedCount = result.SuccessCount + result.FailedCount
	batch.EndTime = &endTime
	batch.DurationMs = duration.Milliseconds()
	batch.UpdatedBy = req.Operator
	
	err = s.batchRepo.UpdateBatchJob(batch)
	if err != nil {
		return nil, fmt.Errorf("failed to update batch job: %w", err)
	}
	
	result.Duration = duration
	return result, nil
}

// calculateLoanDailyInterest 计算单个借据的每日利息
func (s *BatchService) calculateLoanDailyInterest(loan *model.Loan, batchDate time.Time, batchNo string, operator string) error {
	// 计算日利率
	daysInYear := 360
	dailyRate := loan.AnnualRate.Div(decimal.NewFromInt(int64(daysInYear)))
	
	// 计算当日利息
	interest := loan.RemainingPrincipal.Mul(dailyRate).Round(2)
	
	// 创建每日计算记录
	dailyCalc := &model.DailyCalculation{
		LoanNo:          loan.LoanNo,
		CalculationDate: batchDate,
		FeeCode:         "INTEREST",
		FeeCategory:     "INTEREST",
		BaseAmount:      loan.RemainingPrincipal,
		DailyRate:       dailyRate,
		Amount:          interest,
		IsSettled:       false,
		BatchNo:         batchNo,
		CreatedBy:       operator,
		UpdatedBy:       operator,
	}
	
	err := s.dailyCalcRepo.CreateDailyCalculation(dailyCalc)
	if err != nil {
		return fmt.Errorf("failed to create daily calculation: %w", err)
	}
	
	// 更新借据累计利息
	loan.TotalInterest = loan.TotalInterest.Add(interest)
	loan.UpdatedBy = operator
	
	err = s.loanRepo.UpdateLoan(loan)
	if err != nil {
		return fmt.Errorf("failed to update loan: %w", err)
	}
	
	return nil
}

// OverdueCheckRequest 逾期检查请求
type OverdueCheckRequest struct {
	BatchDate time.Time
	BatchNo   string
	Operator  string
}

// OverdueCheckResult 逾期检查结果
type OverdueCheckResult struct {
	BatchNo        string
	BatchDate      time.Time
	TotalCount     int
	SuccessCount   int
	FailedCount    int
	OverdueCount   int
	Duration       time.Duration
}

// CheckOverdue 逾期检查
func (s *BatchService) CheckOverdue(req OverdueCheckRequest) (*OverdueCheckResult, error) {
	startTime := time.Now()
	
	// 创建批次
	batch := &model.BatchJob{
		BatchNo:   req.BatchNo,
		BatchType: "OVERDUE_CHECK",
		BatchDate: req.BatchDate,
		Status:    "RUNNING",
		StartTime: &startTime,
		CreatedBy: req.Operator,
		UpdatedBy: req.Operator,
	}
	
	err := s.batchRepo.CreateBatchJob(batch)
	if err != nil {
		return nil, fmt.Errorf("failed to create batch job: %w", err)
	}
	
	// 获取所有到期未还的还款计划
	plans, err := s.batchRepo.GetOverduePlans(req.BatchDate)
	if err != nil {
		return nil, fmt.Errorf("failed to get overdue plans: %w", err)
	}
	
	result := &OverdueCheckResult{
		BatchNo:    req.BatchNo,
		BatchDate:  req.BatchDate,
		TotalCount: len(plans),
	}
	
	// 逐个处理逾期计划
	for _, plan := range plans {
		err := s.processOverduePlan(&plan, req.BatchDate, req.BatchNo, req.Operator)
		if err != nil {
			result.FailedCount++
			continue
		}
		result.SuccessCount++
		result.OverdueCount++
	}
	
	// 更新批次状态
	endTime := time.Now()
	duration := endTime.Sub(startTime)
	
	batch.Status = "SUCCESS"
	if result.FailedCount > 0 {
		batch.Status = "PARTIAL"
	}
	batch.TotalCount = result.TotalCount
	batch.SuccessCount = result.SuccessCount
	batch.FailedCount = result.FailedCount
	batch.ProcessedCount = result.SuccessCount + result.FailedCount
	batch.EndTime = &endTime
	batch.DurationMs = duration.Milliseconds()
	batch.UpdatedBy = req.Operator
	
	err = s.batchRepo.UpdateBatchJob(batch)
	if err != nil {
		return nil, fmt.Errorf("failed to update batch job: %w", err)
	}
	
	result.Duration = duration
	return result, nil
}

// processOverduePlan 处理逾期计划
func (s *BatchService) processOverduePlan(plan *model.Plan, batchDate time.Time, batchNo string, operator string) error {
	// 计算逾期天数
	overdueDays := int(batchDate.Sub(plan.DueDate).Hours() / 24)
	if overdueDays <= 0 {
		return nil
	}
	
	// 更新计划状态
	plan.Status = "OVERDUE"
	plan.OverdueDays = overdueDays
	plan.UpdatedBy = operator
	
	err := s.planRepo.UpdatePlan(plan)
	if err != nil {
		return fmt.Errorf("failed to update plan: %w", err)
	}
	
	// 计算罚息
	// 获取借据
	loan, err := s.loanRepo.GetLoanByNo(plan.LoanNo)
	if err != nil {
		return fmt.Errorf("failed to get loan: %w", err)
	}
	
	// 计算逾期本金（应还 - 已还）
	overduePrincipal := plan.DuePrincipal.Sub(plan.PaidPrincipal)
	
	// 罚息利率（日利率0.05%）
	penaltyRate := decimal.NewFromFloat(0.0005)
	
	// 计算当日罚息
	penalty := overduePrincipal.Mul(penaltyRate).Round(2)
	
	// 创建每日计算记录
	dailyCalc := &model.DailyCalculation{
		LoanNo:          plan.LoanNo,
		CalculationDate: batchDate,
		FeeCode:         "OVERDUE_PENALTY",
		FeeCategory:     "PENALTY",
		BaseAmount:      overduePrincipal,
		DailyRate:       penaltyRate,
		Amount:          penalty,
		PlanID:          &plan.ID,
		IsSettled:       false,
		BatchNo:         batchNo,
		CreatedBy:       operator,
		UpdatedBy:       operator,
	}
	
	err = s.dailyCalcRepo.CreateDailyCalculation(dailyCalc)
	if err != nil {
		return fmt.Errorf("failed to create daily calculation: %w", err)
	}
	
	// 更新计划累计罚息
	plan.DuePenalty = plan.DuePenalty.Add(penalty)
	plan.DueTotal = plan.DueTotal.Add(penalty)
	plan.UpdatedBy = operator
	
	err = s.planRepo.UpdatePlan(plan)
	if err != nil {
		return fmt.Errorf("failed to update plan: %w", err)
	}
	
	// 更新借据逾期信息
	loan.OverdueDays = overdueDays
	loan.OverduePrincipal = overduePrincipal
	loan.TotalPenalty = loan.TotalPenalty.Add(penalty)
	loan.Status = "OVERDUE"
	loan.UpdatedBy = operator
	
	err = s.loanRepo.UpdateLoan(loan)
	if err != nil {
		return fmt.Errorf("failed to update loan: %w", err)
	}
	
	return nil
}

// BatchJobRepository 批次仓库（需要在 repository 包中实现）
type BatchJobRepository struct {
	db interface{} // 简化处理，实际应该是 *sql.DB
}

// GetDisbursedLoans 获取已放款的借据
func (s *BatchService) getDisbursedLoans() ([]model.Loan, error) {
	// 这里需要调用 repository 方法
	return nil, nil
}