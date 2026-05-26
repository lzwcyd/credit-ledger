package repository

import (
	"database/sql"
	"fmt"
	"time"
	"github.com/yourorg/credit-ledger/internal/model"
	"github.com/yourorg/credit-ledger/pkg/decimal"
)

// LoanRepository 借据仓库
type LoanRepository struct {
	db *sql.DB
}

func NewLoanRepository(db *sql.DB) *LoanRepository {
	return &LoanRepository{db: db}
}

// CreateLoan 创建借据
func (r *LoanRepository) CreateLoan(loan *model.Loan) error {
	query := `
		INSERT INTO loans (
			loan_no, principal, annual_rate, term_months, repayment_type_code, allocation_rule_code,
			value_date, first_due_date, maturity_date, disbursement_date, status,
			disbursed_amount, remaining_principal,
			total_interest, total_penalty, total_other_fee,
			paid_principal, paid_interest, paid_penalty, paid_other_fee,
			overdue_days, overdue_principal,
			created_by, updated_by
		) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
	`
	
	now := time.Now()
	result, err := r.db.Exec(query,
		loan.LoanNo, loan.Principal.String(), loan.AnnualRate.String(), loan.TermMonths, 
		loan.RepaymentTypeCode, loan.AllocationRuleCode,
		loan.ValueDate, loan.FirstDueDate, loan.MaturityDate, loan.DisbursementDate, loan.Status,
		loan.DisbursedAmount.String(), loan.RemainingPrincipal.String(),
		loan.TotalInterest.String(), loan.TotalPenalty.String(), loan.TotalOtherFee.String(),
		loan.PaidPrincipal.String(), loan.PaidInterest.String(), loan.PaidPenalty.String(), loan.PaidOtherFee.String(),
		loan.OverdueDays, loan.OverduePrincipal.String(),
		loan.CreatedBy, loan.UpdatedBy,
	)
	if err != nil {
		return err
	}
	
	id, err := result.LastInsertId()
	if err != nil {
		return err
	}
	
	loan.ID = uint64(id)
	loan.CreatedAt = now
	loan.UpdatedAt = now
	
	return nil
}

// GetLoanByNo 根据借据编号获取借据
func (r *LoanRepository) GetLoanByNo(loanNo string) (*model.Loan, error) {
	query := `
		SELECT id, loan_no, principal, annual_rate, term_months, repayment_type_code, allocation_rule_code,
			value_date, first_due_date, maturity_date, settlement_date, disbursement_date, status,
			disbursed_amount, remaining_principal,
			total_interest, total_penalty, total_other_fee,
			paid_principal, paid_interest, paid_penalty, paid_other_fee,
			overdue_days, overdue_principal,
			created_by, updated_by, created_at, updated_at
		FROM loans WHERE loan_no = ?
	`
	
	loan := &model.Loan{}
	var principal, annualRate, disbursedAmount, remainingPrincipal string
	var totalInterest, totalPenalty, totalOtherFee string
	var paidPrincipal, paidInterest, paidPenalty, paidOtherFee string
	var overduePrincipal string
	var settlementDate, disbursementDate sql.NullTime
	
	err := r.db.QueryRow(query, loanNo).Scan(
		&loan.ID, &loan.LoanNo, &principal, &annualRate, &loan.TermMonths, 
		&loan.RepaymentTypeCode, &loan.AllocationRuleCode,
		&loan.ValueDate, &loan.FirstDueDate, &loan.MaturityDate, &settlementDate, &disbursementDate, &loan.Status,
		&disbursedAmount, &remainingPrincipal,
		&totalInterest, &totalPenalty, &totalOtherFee,
		&paidPrincipal, &paidInterest, &paidPenalty, &paidOtherFee,
		&loan.OverdueDays, &overduePrincipal,
		&loan.CreatedBy, &loan.UpdatedBy, &loan.CreatedAt, &loan.UpdatedAt,
	)
	if err != nil {
		return nil, err
	}
	
	// 解析 decimal
	if loan.Principal, err = decimal.NewFromString(principal); err != nil {
		return nil, fmt.Errorf("failed to parse principal: %w", err)
	}
	if loan.AnnualRate, err = decimal.NewFromString(annualRate); err != nil {
		return nil, fmt.Errorf("failed to parse annual_rate: %w", err)
	}
	if loan.DisbursedAmount, err = decimal.NewFromString(disbursedAmount); err != nil {
		return nil, fmt.Errorf("failed to parse disbursed_amount: %w", err)
	}
	if loan.RemainingPrincipal, err = decimal.NewFromString(remainingPrincipal); err != nil {
		return nil, fmt.Errorf("failed to parse remaining_principal: %w", err)
	}
	if loan.TotalInterest, err = decimal.NewFromString(totalInterest); err != nil {
		return nil, fmt.Errorf("failed to parse total_interest: %w", err)
	}
	if loan.TotalPenalty, err = decimal.NewFromString(totalPenalty); err != nil {
		return nil, fmt.Errorf("failed to parse total_penalty: %w", err)
	}
	if loan.TotalOtherFee, err = decimal.NewFromString(totalOtherFee); err != nil {
		return nil, fmt.Errorf("failed to parse total_other_fee: %w", err)
	}
	if loan.PaidPrincipal, err = decimal.NewFromString(paidPrincipal); err != nil {
		return nil, fmt.Errorf("failed to parse paid_principal: %w", err)
	}
	if loan.PaidInterest, err = decimal.NewFromString(paidInterest); err != nil {
		return nil, fmt.Errorf("failed to parse paid_interest: %w", err)
	}
	if loan.PaidPenalty, err = decimal.NewFromString(paidPenalty); err != nil {
		return nil, fmt.Errorf("failed to parse paid_penalty: %w", err)
	}
	if loan.PaidOtherFee, err = decimal.NewFromString(paidOtherFee); err != nil {
		return nil, fmt.Errorf("failed to parse paid_other_fee: %w", err)
	}
	if loan.OverduePrincipal, err = decimal.NewFromString(overduePrincipal); err != nil {
		return nil, fmt.Errorf("failed to parse overdue_principal: %w", err)
	}

	if settlementDate.Valid {
		loan.SettlementDate = &settlementDate.Time
	}
	if disbursementDate.Valid {
		loan.DisbursementDate = &disbursementDate.Time
	}
	
	return loan, nil
}

// UpdateLoan 更新借据
func (r *LoanRepository) UpdateLoan(loan *model.Loan) error {
	query := `
		UPDATE loans SET
			status = ?, settlement_date = ?, disbursement_date = ?,
			disbursed_amount = ?, remaining_principal = ?,
			total_interest = ?, total_penalty = ?, total_other_fee = ?,
			paid_principal = ?, paid_interest = ?, paid_penalty = ?, paid_other_fee = ?,
			overdue_days = ?, overdue_principal = ?,
			updated_by = ?, updated_at = ?
		WHERE id = ?
	`
	
	now := time.Now()
	_, err := r.db.Exec(query,
		loan.Status, loan.SettlementDate, loan.DisbursementDate,
		loan.DisbursedAmount.String(), loan.RemainingPrincipal.String(),
		loan.TotalInterest.String(), loan.TotalPenalty.String(), loan.TotalOtherFee.String(),
		loan.PaidPrincipal.String(), loan.PaidInterest.String(), loan.PaidPenalty.String(), loan.PaidOtherFee.String(),
		loan.OverdueDays, loan.OverduePrincipal.String(),
		loan.UpdatedBy, now, loan.ID,
	)
	
	if err != nil {
		return err
	}
	
	loan.UpdatedAt = now
	return nil
}

// CreateLoanChange 创建借据变更记录
func (r *LoanRepository) CreateLoanChange(change *model.LoanChange) error {
	query := `
		INSERT INTO loan_changes (
			loan_no, change_type, field_name, old_value, new_value, change_reason,
			related_repayment_no, batch_no, created_by, updated_by
		) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
	`
	
	now := time.Now()
	result, err := r.db.Exec(query,
		change.LoanNo, change.ChangeType, change.FieldName, change.OldValue, change.NewValue,
		change.ChangeReason, change.RelatedRepaymentNo, change.BatchNo,
		change.CreatedBy, change.UpdatedBy,
	)
	if err != nil {
		return err
	}
	
	id, err := result.LastInsertId()
	if err != nil {
		return err
	}
	
	change.ID = uint64(id)
	change.CreatedAt = now
	change.UpdatedAt = now
	
	return nil
}

// PlanRepository 还款计划仓库
type PlanRepository struct {
	db *sql.DB
}

func NewPlanRepository(db *sql.DB) *PlanRepository {
	return &PlanRepository{db: db}
}

// CreatePlan 创建还款计划
func (r *PlanRepository) CreatePlan(plan *model.Plan) error {
	query := `
		INSERT INTO plans (
			loan_no, period, due_date,
			due_principal, due_interest, due_penalty, due_other_fee, due_total,
			paid_principal, paid_interest, paid_penalty, paid_other_fee, paid_total,
			overdue_days, status,
			created_by, updated_by
		) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
	`
	
	now := time.Now()
	result, err := r.db.Exec(query,
		plan.LoanNo, plan.Period, plan.DueDate,
		plan.DuePrincipal.String(), plan.DueInterest.String(), plan.DuePenalty.String(),
		plan.DueOtherFee.String(), plan.DueTotal.String(),
		plan.PaidPrincipal.String(), plan.PaidInterest.String(), plan.PaidPenalty.String(),
		plan.PaidOtherFee.String(), plan.PaidTotal.String(),
		plan.OverdueDays, plan.Status,
		plan.CreatedBy, plan.UpdatedBy,
	)
	if err != nil {
		return err
	}
	
	id, err := result.LastInsertId()
	if err != nil {
		return err
	}
	
	plan.ID = uint64(id)
	plan.CreatedAt = now
	plan.UpdatedAt = now
	
	return nil
}

// GetPlanByLoanNoAndPeriod 根据借据编号和期数获取还款计划
func (r *PlanRepository) GetPlanByLoanNoAndPeriod(loanNo string, period int) (*model.Plan, error) {
	query := `
		SELECT id, loan_no, period, due_date,
			due_principal, due_interest, due_penalty, due_other_fee, due_total,
			paid_principal, paid_interest, paid_penalty, paid_other_fee, paid_total,
			overdue_days, status,
			created_by, updated_by, created_at, updated_at
		FROM plans WHERE loan_no = ? AND period = ?
	`
	
	plan := &model.Plan{}
	var duePrincipal, dueInterest, duePenalty, dueOtherFee, dueTotal string
	var paidPrincipal, paidInterest, paidPenalty, paidOtherFee, paidTotal string
	
	err := r.db.QueryRow(query, loanNo, period).Scan(
		&plan.ID, &plan.LoanNo, &plan.Period, &plan.DueDate,
		&duePrincipal, &dueInterest, &duePenalty, &dueOtherFee, &dueTotal,
		&paidPrincipal, &paidInterest, &paidPenalty, &paidOtherFee, &paidTotal,
		&plan.OverdueDays, &plan.Status,
		&plan.CreatedBy, &plan.UpdatedBy, &plan.CreatedAt, &plan.UpdatedAt,
	)
	if err != nil {
		return nil, err
	}
	
	if plan.DuePrincipal, err = decimal.NewFromString(duePrincipal); err != nil {
		return nil, fmt.Errorf("failed to parse due_principal: %w", err)
	}
	if plan.DueInterest, err = decimal.NewFromString(dueInterest); err != nil {
		return nil, fmt.Errorf("failed to parse due_interest: %w", err)
	}
	if plan.DuePenalty, err = decimal.NewFromString(duePenalty); err != nil {
		return nil, fmt.Errorf("failed to parse due_penalty: %w", err)
	}
	if plan.DueOtherFee, err = decimal.NewFromString(dueOtherFee); err != nil {
		return nil, fmt.Errorf("failed to parse due_other_fee: %w", err)
	}
	if plan.DueTotal, err = decimal.NewFromString(dueTotal); err != nil {
		return nil, fmt.Errorf("failed to parse due_total: %w", err)
	}
	if plan.PaidPrincipal, err = decimal.NewFromString(paidPrincipal); err != nil {
		return nil, fmt.Errorf("failed to parse paid_principal: %w", err)
	}
	if plan.PaidInterest, err = decimal.NewFromString(paidInterest); err != nil {
		return nil, fmt.Errorf("failed to parse paid_interest: %w", err)
	}
	if plan.PaidPenalty, err = decimal.NewFromString(paidPenalty); err != nil {
		return nil, fmt.Errorf("failed to parse paid_penalty: %w", err)
	}
	if plan.PaidOtherFee, err = decimal.NewFromString(paidOtherFee); err != nil {
		return nil, fmt.Errorf("failed to parse paid_other_fee: %w", err)
	}
	if plan.PaidTotal, err = decimal.NewFromString(paidTotal); err != nil {
		return nil, fmt.Errorf("failed to parse paid_total: %w", err)
	}

	return plan, nil
}

// GetPlansByLoanNo 根据借据编号获取所有还款计划
func (r *PlanRepository) GetPlansByLoanNo(loanNo string) ([]model.Plan, error) {
	query := `
		SELECT id, loan_no, period, due_date,
			due_principal, due_interest, due_penalty, due_other_fee, due_total,
			paid_principal, paid_interest, paid_penalty, paid_other_fee, paid_total,
			overdue_days, status,
			created_by, updated_by, created_at, updated_at
		FROM plans WHERE loan_no = ? ORDER BY period
	`
	
	rows, err := r.db.Query(query, loanNo)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	
	var plans []model.Plan
	for rows.Next() {
		plan := model.Plan{}
		var duePrincipal, dueInterest, duePenalty, dueOtherFee, dueTotal string
		var paidPrincipal, paidInterest, paidPenalty, paidOtherFee, paidTotal string
		
		err := rows.Scan(
			&plan.ID, &plan.LoanNo, &plan.Period, &plan.DueDate,
			&duePrincipal, &dueInterest, &duePenalty, &dueOtherFee, &dueTotal,
			&paidPrincipal, &paidInterest, &paidPenalty, &paidOtherFee, &paidTotal,
			&plan.OverdueDays, &plan.Status,
			&plan.CreatedBy, &plan.UpdatedBy, &plan.CreatedAt, &plan.UpdatedAt,
		)
		if err != nil {
			return nil, err
		}
		
		plan.DuePrincipal, _ = decimal.NewFromString(duePrincipal)
		plan.DueInterest, _ = decimal.NewFromString(dueInterest)
		plan.DuePenalty, _ = decimal.NewFromString(duePenalty)
		plan.DueOtherFee, _ = decimal.NewFromString(dueOtherFee)
		plan.DueTotal, _ = decimal.NewFromString(dueTotal)
		plan.PaidPrincipal, _ = decimal.NewFromString(paidPrincipal)
		plan.PaidInterest, _ = decimal.NewFromString(paidInterest)
		plan.PaidPenalty, _ = decimal.NewFromString(paidPenalty)
		plan.PaidOtherFee, _ = decimal.NewFromString(paidOtherFee)
		plan.PaidTotal, _ = decimal.NewFromString(paidTotal)
		
		plans = append(plans, plan)
	}
	
	return plans, nil
}

// UpdatePlan 更新还款计划
func (r *PlanRepository) UpdatePlan(plan *model.Plan) error {
	query := `
		UPDATE plans SET
			due_principal = ?, due_interest = ?, due_penalty = ?, due_other_fee = ?, due_total = ?,
			paid_principal = ?, paid_interest = ?, paid_penalty = ?, paid_other_fee = ?, paid_total = ?,
			overdue_days = ?, status = ?,
			updated_by = ?, updated_at = ?
		WHERE id = ?
	`
	
	now := time.Now()
	_, err := r.db.Exec(query,
		plan.DuePrincipal.String(), plan.DueInterest.String(), plan.DuePenalty.String(),
		plan.DueOtherFee.String(), plan.DueTotal.String(),
		plan.PaidPrincipal.String(), plan.PaidInterest.String(), plan.PaidPenalty.String(),
		plan.PaidOtherFee.String(), plan.PaidTotal.String(),
		plan.OverdueDays, plan.Status,
		plan.UpdatedBy, now, plan.ID,
	)
	
	if err != nil {
		return err
	}
	
	plan.UpdatedAt = now
	return nil
}

// RepaymentRepository 还款记录仓库
type RepaymentRepository struct {
	db *sql.DB
}

func NewRepaymentRepository(db *sql.DB) *RepaymentRepository {
	return &RepaymentRepository{db: db}
}

// CreateRepayment 创建还款记录
func (r *RepaymentRepository) CreateRepayment(repayment *model.Repayment) error {
	query := `
		INSERT INTO repayments (
			repayment_no, loan_no, plan_id, repayment_type,
			amount, principal_amount, interest_amount, penalty_amount, other_fee_amount,
			trial_date, booking_date, allocation_rule_code,
			status, description, is_backdated, backdated_reason, batch_no,
			created_by, updated_by
		) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
	`
	
	now := time.Now()
	result, err := r.db.Exec(query,
		repayment.RepaymentNo, repayment.LoanNo, repayment.PlanID, repayment.RepaymentType,
		repayment.Amount.String(), repayment.PrincipalAmount.String(), repayment.InterestAmount.String(),
		repayment.PenaltyAmount.String(), repayment.OtherFeeAmount.String(),
		repayment.TrialDate, repayment.BookingDate, repayment.AllocationRuleCode,
		repayment.Status, repayment.Description, repayment.IsBackdated, repayment.BackdatedReason, repayment.BatchNo,
		repayment.CreatedBy, repayment.UpdatedBy,
	)
	if err != nil {
		return err
	}
	
	id, err := result.LastInsertId()
	if err != nil {
		return err
	}
	
	repayment.ID = uint64(id)
	repayment.CreatedAt = now
	repayment.UpdatedAt = now
	
	return nil
}

// CreateRepaymentDetail 创建还款入账明细
func (r *RepaymentRepository) CreateRepaymentDetail(detail *model.RepaymentDetail) error {
	query := `
		INSERT INTO repayment_details (
			repayment_no, loan_no, fee_code, fee_name, fee_category, amount,
			daily_calculation_id, plan_other_fee_id,
			created_by, updated_by
		) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
	`
	
	now := time.Now()
	result, err := r.db.Exec(query,
		detail.RepaymentNo, detail.LoanNo, detail.FeeCode, detail.FeeName, detail.FeeCategory,
		detail.Amount.String(), detail.DailyCalculationID, detail.PlanOtherFeeID,
		detail.CreatedBy, detail.UpdatedBy,
	)
	if err != nil {
		return err
	}
	
	id, err := result.LastInsertId()
	if err != nil {
		return err
	}
	
	detail.ID = uint64(id)
	detail.CreatedAt = now
	detail.UpdatedAt = now
	
	return nil
}

// DailyCalculationRepository 每日计算仓库
type DailyCalculationRepository struct {
	db *sql.DB
}

func NewDailyCalculationRepository(db *sql.DB) *DailyCalculationRepository {
	return &DailyCalculationRepository{db: db}
}

// CreateDailyCalculation 创建每日计算记录
func (r *DailyCalculationRepository) CreateDailyCalculation(calc *model.DailyCalculation) error {
	query := `
		INSERT INTO daily_calculations (
			loan_no, calculation_date, fee_code, fee_category,
			base_amount, daily_rate, amount, plan_id,
			is_settled, settled_repayment_no, batch_no,
			created_by, updated_by
		) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
	`
	
	now := time.Now()
	result, err := r.db.Exec(query,
		calc.LoanNo, calc.CalculationDate, calc.FeeCode, calc.FeeCategory,
		calc.BaseAmount.String(), calc.DailyRate.String(), calc.Amount.String(), calc.PlanID,
		calc.IsSettled, calc.SettledRepaymentNo, calc.BatchNo,
		calc.CreatedBy, calc.UpdatedBy,
	)
	if err != nil {
		return err
	}
	
	id, err := result.LastInsertId()
	if err != nil {
		return err
	}
	
	calc.ID = uint64(id)
	calc.CreatedAt = now
	calc.UpdatedAt = now
	
	return nil
}

// GetUnsettledByLoanNo 获取未结清的每日计算记录
func (r *DailyCalculationRepository) GetUnsettledByLoanNo(loanNo string, feeCategory string) ([]model.DailyCalculation, error) {
	query := `
		SELECT id, loan_no, calculation_date, fee_code, fee_category,
			base_amount, daily_rate, amount, plan_id,
			is_settled, settled_repayment_no, batch_no,
			created_by, updated_by, created_at, updated_at
		FROM daily_calculations 
		WHERE loan_no = ? AND fee_category = ? AND is_settled = FALSE
		ORDER BY calculation_date
	`
	
	rows, err := r.db.Query(query, loanNo, feeCategory)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	
	var calcs []model.DailyCalculation
	for rows.Next() {
		calc := model.DailyCalculation{}
		var baseAmount, dailyRate, amount string
		
		err := rows.Scan(
			&calc.ID, &calc.LoanNo, &calc.CalculationDate, &calc.FeeCode, &calc.FeeCategory,
			&baseAmount, &dailyRate, &amount, &calc.PlanID,
			&calc.IsSettled, &calc.SettledRepaymentNo, &calc.BatchNo,
			&calc.CreatedBy, &calc.UpdatedBy, &calc.CreatedAt, &calc.UpdatedAt,
		)
		if err != nil {
			return nil, err
		}
		
		calc.BaseAmount, _ = decimal.NewFromString(baseAmount)
		calc.DailyRate, _ = decimal.NewFromString(dailyRate)
		calc.Amount, _ = decimal.NewFromString(amount)
		
		calcs = append(calcs, calc)
	}
	
	return calcs, nil
}

// SettleDailyCalculations 结清每日计算记录
func (r *DailyCalculationRepository) SettleDailyCalculations(ids []uint64, repaymentNo string) error {
	if len(ids) == 0 {
		return nil
	}
	
	query := `
		UPDATE daily_calculations 
		SET is_settled = TRUE, settled_repayment_no = ?, updated_at = ?
		WHERE id IN (?` + string(make([]byte, len(ids)-1)) + `)
	`
	
	// 构建占位符
	for i := 1; i < len(ids); i++ {
		query += ", ?"
	}
	query += ")"
	
	// 构建参数
	args := []interface{}{repaymentNo, time.Now()}
	for _, id := range ids {
		args = append(args, id)
	}
	
	_, err := r.db.Exec(query, args...)
	return err
}

// PlanOtherFeeRepository 其他费用明细仓库
type PlanOtherFeeRepository struct {
	db *sql.DB
}

func NewPlanOtherFeeRepository(db *sql.DB) *PlanOtherFeeRepository {
	return &PlanOtherFeeRepository{db: db}
}

// CreatePlanOtherFee 创建其他费用明细
func (r *PlanOtherFeeRepository) CreatePlanOtherFee(fee *model.PlanOtherFee) error {
	query := `
		INSERT INTO plan_other_fees (
			loan_no, plan_id, period, fee_code, fee_name,
			due_amount, paid_amount, status,
			created_by, updated_by
		) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
	`
	
	now := time.Now()
	result, err := r.db.Exec(query,
		fee.LoanNo, fee.PlanID, fee.Period, fee.FeeCode, fee.FeeName,
		fee.DueAmount.String(), fee.PaidAmount.String(), fee.Status,
		fee.CreatedBy, fee.UpdatedBy,
	)
	if err != nil {
		return err
	}
	
	id, err := result.LastInsertId()
	if err != nil {
		return err
	}
	
	fee.ID = uint64(id)
	fee.CreatedAt = now
	fee.UpdatedAt = now
	
	return nil
}

// GetUnpaidByPlanID 获取未还清的其他费用明细
func (r *PlanOtherFeeRepository) GetUnpaidByPlanID(planID uint64) ([]model.PlanOtherFee, error) {
	query := `
		SELECT id, loan_no, plan_id, period, fee_code, fee_name,
			due_amount, paid_amount, status,
			created_by, updated_by, created_at, updated_at
		FROM plan_other_fees 
		WHERE plan_id = ? AND status != 'PAID'
	`
	
	rows, err := r.db.Query(query, planID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	
	var fees []model.PlanOtherFee
	for rows.Next() {
		fee := model.PlanOtherFee{}
		var dueAmount, paidAmount string
		
		err := rows.Scan(
			&fee.ID, &fee.LoanNo, &fee.PlanID, &fee.Period, &fee.FeeCode, &fee.FeeName,
			&dueAmount, &paidAmount, &fee.Status,
			&fee.CreatedBy, &fee.UpdatedBy, &fee.CreatedAt, &fee.UpdatedAt,
		)
		if err != nil {
			return nil, err
		}
		
		fee.DueAmount, _ = decimal.NewFromString(dueAmount)
		fee.PaidAmount, _ = decimal.NewFromString(paidAmount)
		
		fees = append(fees, fee)
	}
	
	return fees, nil
}

// UpdatePlanOtherFee 更新其他费用明细
func (r *PlanOtherFeeRepository) UpdatePlanOtherFee(fee *model.PlanOtherFee) error {
	query := `
		UPDATE plan_other_fees SET
			paid_amount = ?, status = ?, updated_by = ?, updated_at = ?
		WHERE id = ?
	`
	
	now := time.Now()
	_, err := r.db.Exec(query,
		fee.PaidAmount.String(), fee.Status, fee.UpdatedBy, now, fee.ID,
	)
	
	if err != nil {
		return err
	}
	
	fee.UpdatedAt = now
	return nil
}

// FeeConfigRepository 费项配置仓库
type FeeConfigRepository struct {
	db *sql.DB
}

func NewFeeConfigRepository(db *sql.DB) *FeeConfigRepository {
	return &FeeConfigRepository{db: db}
}

// GetAllActive 获取所有启用的费项配置
func (r *FeeConfigRepository) GetAllActive() ([]model.FeeConfig, error) {
	query := `
		SELECT id, code, name, calc_type, calc_base, value, trigger_type,
			is_daily_accumulate, fee_category, min_amount, max_amount, is_active,
			created_by, updated_by, created_at, updated_at
		FROM fee_configs WHERE is_active = TRUE
	`
	
	rows, err := r.db.Query(query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	
	var configs []model.FeeConfig
	for rows.Next() {
		config := model.FeeConfig{}
		var value, minAmount string
		var maxAmount sql.NullString
		
		err := rows.Scan(
			&config.ID, &config.Code, &config.Name, &config.CalcType, &config.CalcBase,
			&value, &config.TriggerType, &config.IsDailyAccumulate, &config.FeeCategory,
			&minAmount, &maxAmount, &config.IsActive,
			&config.CreatedBy, &config.UpdatedBy, &config.CreatedAt, &config.UpdatedAt,
		)
		if err != nil {
			return nil, err
		}
		
		config.Value, _ = decimal.NewFromString(value)
		config.MinAmount, _ = decimal.NewFromString(minAmount)
		if maxAmount.Valid {
			maxVal, _ := decimal.NewFromString(maxAmount.String)
			config.MaxAmount = &maxVal
		}
		
		configs = append(configs, config)
	}
	
	return configs, nil
}

// GetByTriggerType 根据触发类型获取费项配置
func (r *FeeConfigRepository) GetByTriggerType(triggerType string) ([]model.FeeConfig, error) {
	query := `
		SELECT id, code, name, calc_type, calc_base, value, trigger_type,
			is_daily_accumulate, fee_category, min_amount, max_amount, is_active,
			created_by, updated_by, created_at, updated_at
		FROM fee_configs WHERE trigger_type = ? AND is_active = TRUE
	`

	rows, err := r.db.Query(query, triggerType)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var configs []model.FeeConfig
	for rows.Next() {
		config := model.FeeConfig{}
		var value, minAmount string
		var maxAmount sql.NullString

		err := rows.Scan(
			&config.ID, &config.Code, &config.Name, &config.CalcType, &config.CalcBase,
			&value, &config.TriggerType, &config.IsDailyAccumulate, &config.FeeCategory,
			&minAmount, &maxAmount, &config.IsActive,
			&config.CreatedBy, &config.UpdatedBy, &config.CreatedAt, &config.UpdatedAt,
		)
		if err != nil {
			return nil, err
		}

		config.Value, _ = decimal.NewFromString(value)
		config.MinAmount, _ = decimal.NewFromString(minAmount)
		if maxAmount.Valid {
			maxVal, _ := decimal.NewFromString(maxAmount.String)
			config.MaxAmount = &maxVal
		}

		configs = append(configs, config)
	}

	return configs, nil
}

// GetByCode 根据编码获取费项配置
func (r *FeeConfigRepository) GetByCode(code string) (*model.FeeConfig, error) {
	query := `
		SELECT id, code, name, calc_type, calc_base, value, trigger_type,
			is_daily_accumulate, fee_category, min_amount, max_amount, is_active,
			created_by, updated_by, created_at, updated_at
		FROM fee_configs WHERE code = ?
	`
	
	config := &model.FeeConfig{}
	var value, minAmount string
	var maxAmount sql.NullString
	
	err := r.db.QueryRow(query, code).Scan(
		&config.ID, &config.Code, &config.Name, &config.CalcType, &config.CalcBase,
		&value, &config.TriggerType, &config.IsDailyAccumulate, &config.FeeCategory,
		&minAmount, &maxAmount, &config.IsActive,
		&config.CreatedBy, &config.UpdatedBy, &config.CreatedAt, &config.UpdatedAt,
	)
	if err != nil {
		return nil, err
	}
	
	config.Value, _ = decimal.NewFromString(value)
	config.MinAmount, _ = decimal.NewFromString(minAmount)
	if maxAmount.Valid {
		maxVal, _ := decimal.NewFromString(maxAmount.String)
		config.MaxAmount = &maxVal
	}
	
	return config, nil
}

// AllocationRuleRepository 分配规则仓库
type AllocationRuleRepository struct {
	db *sql.DB
}

func NewAllocationRuleRepository(db *sql.DB) *AllocationRuleRepository {
	return &AllocationRuleRepository{db: db}
}

// GetDefaultRule 获取默认分配规则
func (r *AllocationRuleRepository) GetDefaultRule() (*model.AllocationRule, error) {
	query := `
		SELECT id, code, name, description, is_default, is_active,
			created_by, updated_by, created_at, updated_at
		FROM allocation_rules WHERE is_default = TRUE AND is_active = TRUE LIMIT 1
	`
	
	rule := &model.AllocationRule{}
	err := r.db.QueryRow(query).Scan(
		&rule.ID, &rule.Code, &rule.Name, &rule.Description, &rule.IsDefault, &rule.IsActive,
		&rule.CreatedBy, &rule.UpdatedBy, &rule.CreatedAt, &rule.UpdatedAt,
	)
	if err != nil {
		return nil, err
	}
	
	return rule, nil
}

// GetRuleItems 获取分配规则明细
func (r *AllocationRuleRepository) GetRuleItems(ruleCode string) ([]model.AllocationRuleItem, error) {
	query := `
		SELECT id, rule_code, priority, allocation_type, description,
			created_by, updated_by, created_at, updated_at
		FROM allocation_rule_items WHERE rule_code = ? ORDER BY priority
	`
	
	rows, err := r.db.Query(query, ruleCode)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	
	var items []model.AllocationRuleItem
	for rows.Next() {
		item := model.AllocationRuleItem{}
		
		err := rows.Scan(
			&item.ID, &item.RuleCode, &item.Priority, &item.AllocationType, &item.Description,
			&item.CreatedBy, &item.UpdatedBy, &item.CreatedAt, &item.UpdatedAt,
		)
		if err != nil {
			return nil, err
		}
		
		items = append(items, item)
	}
	
	return items, nil
}

// =====================================================
// 批次仓库
// =====================================================

// BatchJobRepository 批次仓库
type BatchJobRepository struct {
	db *sql.DB
}

func NewBatchJobRepository(db *sql.DB) *BatchJobRepository {
	return &BatchJobRepository{db: db}
}

// CreateBatchJob 创建批次
func (r *BatchJobRepository) CreateBatchJob(batch *model.BatchJob) error {
	query := `
		INSERT INTO batch_jobs (
			batch_no, batch_type, batch_date, status,
			start_time, created_by, updated_by
		) VALUES (?, ?, ?, ?, ?, ?, ?)
	`
	
	now := time.Now()
	result, err := r.db.Exec(query,
		batch.BatchNo, batch.BatchType, batch.BatchDate, batch.Status,
		batch.StartTime, batch.CreatedBy, batch.UpdatedBy,
	)
	if err != nil {
		return err
	}
	
	id, err := result.LastInsertId()
	if err != nil {
		return err
	}
	
	batch.ID = uint64(id)
	batch.CreatedAt = now
	batch.UpdatedAt = now
	
	return nil
}

// UpdateBatchJob 更新批次
func (r *BatchJobRepository) UpdateBatchJob(batch *model.BatchJob) error {
	query := `
		UPDATE batch_jobs SET
			status = ?, total_count = ?, processed_count = ?, success_count = ?, failed_count = ?,
			start_time = ?, end_time = ?, duration_ms = ?,
			error_message = ?, updated_by = ?, updated_at = ?
		WHERE id = ?
	`
	
	now := time.Now()
	_, err := r.db.Exec(query,
		batch.Status, batch.TotalCount, batch.ProcessedCount, batch.SuccessCount, batch.FailedCount,
		batch.StartTime, batch.EndTime, batch.DurationMs,
		batch.ErrorMessage, batch.UpdatedBy, now, batch.ID,
	)
	
	if err != nil {
		return err
	}
	
	batch.UpdatedAt = now
	return nil
}

// GetDisbursedLoans 获取已放款的借据
func (r *BatchJobRepository) GetDisbursedLoans() ([]model.Loan, error) {
	query := `
		SELECT id, loan_no, principal, annual_rate, term_months, repayment_type_code, allocation_rule_code,
			value_date, first_due_date, maturity_date, settlement_date, disbursement_date, status,
			disbursed_amount, remaining_principal,
			total_interest, total_penalty, total_other_fee,
			paid_principal, paid_interest, paid_penalty, paid_other_fee,
			overdue_days, overdue_principal,
			created_by, updated_by, created_at, updated_at
		FROM loans WHERE status IN ('DISBURSED', 'OVERDUE')
	`
	
	rows, err := r.db.Query(query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	
	var loans []model.Loan
	for rows.Next() {
		loan := model.Loan{}
		var principal, annualRate, disbursedAmount, remainingPrincipal string
		var totalInterest, totalPenalty, totalOtherFee string
		var paidPrincipal, paidInterest, paidPenalty, paidOtherFee string
		var overduePrincipal string
		var settlementDate, disbursementDate sql.NullTime
		
		err := rows.Scan(
			&loan.ID, &loan.LoanNo, &principal, &annualRate, &loan.TermMonths, 
			&loan.RepaymentTypeCode, &loan.AllocationRuleCode,
			&loan.ValueDate, &loan.FirstDueDate, &loan.MaturityDate, &settlementDate, &disbursementDate, &loan.Status,
			&disbursedAmount, &remainingPrincipal,
			&totalInterest, &totalPenalty, &totalOtherFee,
			&paidPrincipal, &paidInterest, &paidPenalty, &paidOtherFee,
			&loan.OverdueDays, &overduePrincipal,
			&loan.CreatedBy, &loan.UpdatedBy, &loan.CreatedAt, &loan.UpdatedAt,
		)
		if err != nil {
			return nil, err
		}
		
		loan.Principal, _ = decimal.NewFromString(principal)
		loan.AnnualRate, _ = decimal.NewFromString(annualRate)
		loan.DisbursedAmount, _ = decimal.NewFromString(disbursedAmount)
		loan.RemainingPrincipal, _ = decimal.NewFromString(remainingPrincipal)
		loan.TotalInterest, _ = decimal.NewFromString(totalInterest)
		loan.TotalPenalty, _ = decimal.NewFromString(totalPenalty)
		loan.TotalOtherFee, _ = decimal.NewFromString(totalOtherFee)
		loan.PaidPrincipal, _ = decimal.NewFromString(paidPrincipal)
		loan.PaidInterest, _ = decimal.NewFromString(paidInterest)
		loan.PaidPenalty, _ = decimal.NewFromString(paidPenalty)
		loan.PaidOtherFee, _ = decimal.NewFromString(paidOtherFee)
		loan.OverduePrincipal, _ = decimal.NewFromString(overduePrincipal)
		
		if settlementDate.Valid {
			loan.SettlementDate = &settlementDate.Time
		}
		if disbursementDate.Valid {
			loan.DisbursementDate = &disbursementDate.Time
		}
		
		loans = append(loans, loan)
	}
	
	return loans, nil
}

// GetOverduePlans 获取到期未还的还款计划
func (r *BatchJobRepository) GetOverduePlans(batchDate time.Time) ([]model.Plan, error) {
	query := `
		SELECT id, loan_no, period, due_date,
			due_principal, due_interest, due_penalty, due_other_fee, due_total,
			paid_principal, paid_interest, paid_penalty, paid_other_fee, paid_total,
			overdue_days, status,
			created_by, updated_by, created_at, updated_at
		FROM plans 
		WHERE due_date < ? AND status IN ('PENDING', 'PARTIAL')
	`
	
	rows, err := r.db.Query(query, batchDate)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	
	var plans []model.Plan
	for rows.Next() {
		plan := model.Plan{}
		var duePrincipal, dueInterest, duePenalty, dueOtherFee, dueTotal string
		var paidPrincipal, paidInterest, paidPenalty, paidOtherFee, paidTotal string
		
		err := rows.Scan(
			&plan.ID, &plan.LoanNo, &plan.Period, &plan.DueDate,
			&duePrincipal, &dueInterest, &duePenalty, &dueOtherFee, &dueTotal,
			&paidPrincipal, &paidInterest, &paidPenalty, &paidOtherFee, &paidTotal,
			&plan.OverdueDays, &plan.Status,
			&plan.CreatedBy, &plan.UpdatedBy, &plan.CreatedAt, &plan.UpdatedAt,
		)
		if err != nil {
			return nil, err
		}
		
		plan.DuePrincipal, _ = decimal.NewFromString(duePrincipal)
		plan.DueInterest, _ = decimal.NewFromString(dueInterest)
		plan.DuePenalty, _ = decimal.NewFromString(duePenalty)
		plan.DueOtherFee, _ = decimal.NewFromString(dueOtherFee)
		plan.DueTotal, _ = decimal.NewFromString(dueTotal)
		plan.PaidPrincipal, _ = decimal.NewFromString(paidPrincipal)
		plan.PaidInterest, _ = decimal.NewFromString(paidInterest)
		plan.PaidPenalty, _ = decimal.NewFromString(paidPenalty)
		plan.PaidOtherFee, _ = decimal.NewFromString(paidOtherFee)
		plan.PaidTotal, _ = decimal.NewFromString(paidTotal)
		
		plans = append(plans, plan)
	}
	
	return plans, nil
}