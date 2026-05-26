package repository

import (
	"database/sql"
	"time"
	"github.com/yourorg/credit-ledger/internal/model"
)

// LoanRepository 贷款仓库
type LoanRepository struct {
	db *sql.DB
}

// NewLoanRepository 创建贷款仓库
func NewLoanRepository(db *sql.DB) *LoanRepository {
	return &LoanRepository{db: db}
}

// CreateLoan 创建贷款
func (r *LoanRepository) CreateLoan(loan *model.Loan) error {
	query := `
		INSERT INTO loans (
			loan_no, principal, annual_interest_rate, term_months, 
			repayment_type_id, status, created_at, updated_at
		) VALUES (?, ?, ?, ?, ?, ?, ?, ?)
	`
	
	now := time.Now()
	result, err := r.db.Exec(query,
		loan.LoanNo, loan.Principal, loan.AnnualInterestRate, loan.TermMonths,
		loan.RepaymentTypeID, loan.Status, now, now,
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

// GetLoanByID 根据ID获取贷款
func (r *LoanRepository) GetLoanByID(id uint64) (*model.Loan, error) {
	query := `
		SELECT 
			id, loan_no, principal, annual_interest_rate, term_months,
			repayment_type_id, disbursement_date, first_repayment_date,
			maturity_date, status, disbursed_amount, total_interest,
			total_fees, repaid_amount, remaining_principal,
			created_at, updated_at
		FROM loans 
		WHERE id = ?
	`
	
	loan := &model.Loan{}
	err := r.db.QueryRow(query, id).Scan(
		&loan.ID, &loan.LoanNo, &loan.Principal, &loan.AnnualInterestRate, &loan.TermMonths,
		&loan.RepaymentTypeID, &loan.DisbursementDate, &loan.FirstRepaymentDate,
		&loan.MaturityDate, &loan.Status, &loan.DisbursedAmount, &loan.TotalInterest,
		&loan.TotalFees, &loan.RepaidAmount, &loan.RemainingPrincipal,
		&loan.CreatedAt, &loan.UpdatedAt,
	)
	
	if err != nil {
		return nil, err
	}
	
	return loan, nil
}

// GetLoanByNo 根据贷款编号获取贷款
func (r *LoanRepository) GetLoanByNo(loanNo string) (*model.Loan, error) {
	query := `
		SELECT 
			id, loan_no, principal, annual_interest_rate, term_months,
			repayment_type_id, disbursement_date, first_repayment_date,
			maturity_date, status, disbursed_amount, total_interest,
			total_fees, repaid_amount, remaining_principal,
			created_at, updated_at
		FROM loans 
		WHERE loan_no = ?
	`
	
	loan := &model.Loan{}
	err := r.db.QueryRow(query, loanNo).Scan(
		&loan.ID, &loan.LoanNo, &loan.Principal, &loan.AnnualInterestRate, &loan.TermMonths,
		&loan.RepaymentTypeID, &loan.DisbursementDate, &loan.FirstRepaymentDate,
		&loan.MaturityDate, &loan.Status, &loan.DisbursedAmount, &loan.TotalInterest,
		&loan.TotalFees, &loan.RepaidAmount, &loan.RemainingPrincipal,
		&loan.CreatedAt, &loan.UpdatedAt,
	)
	
	if err != nil {
		return nil, err
	}
	
	return loan, nil
}

// UpdateLoan 更新贷款
func (r *LoanRepository) UpdateLoan(loan *model.Loan) error {
	query := `
		UPDATE loans SET
			principal = ?, annual_interest_rate = ?, term_months = ?,
			repayment_type_id = ?, disbursement_date = ?, first_repayment_date = ?,
			maturity_date = ?, status = ?, disbursed_amount = ?,
			total_interest = ?, total_fees = ?, repaid_amount = ?,
			remaining_principal = ?, updated_at = ?
		WHERE id = ?
	`
	
	now := time.Now()
	_, err := r.db.Exec(query,
		loan.Principal, loan.AnnualInterestRate, loan.TermMonths,
		loan.RepaymentTypeID, loan.DisbursementDate, loan.FirstRepaymentDate,
		loan.MaturityDate, loan.Status, loan.DisbursedAmount,
		loan.TotalInterest, loan.TotalFees, loan.RepaidAmount,
		loan.RemainingPrincipal, now, loan.ID,
	)
	
	if err != nil {
		return err
	}
	
	loan.UpdatedAt = now
	return nil
}

// CreateLoanSchedule 创建还款计划
func (r *LoanRepository) CreateLoanSchedule(schedule *model.LoanSchedule) error {
	query := `
		INSERT INTO loan_schedules (
			loan_id, period, due_date, principal_due, interest_due,
			total_due, principal_paid, interest_paid, total_paid,
			status, created_at, updated_at
		) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
	`
	
	now := time.Now()
	result, err := r.db.Exec(query,
		schedule.LoanID, schedule.Period, schedule.DueDate, schedule.PrincipalDue, schedule.InterestDue,
		schedule.TotalDue, schedule.PrincipalPaid, schedule.InterestPaid, schedule.TotalPaid,
		schedule.Status, now, now,
	)
	if err != nil {
		return err
	}
	
	id, err := result.LastInsertId()
	if err != nil {
		return err
	}
	
	schedule.ID = uint64(id)
	schedule.CreatedAt = now
	schedule.UpdatedAt = now
	
	return nil
}

// GetLoanSchedulesByLoanID 根据贷款ID获取还款计划
func (r *LoanRepository) GetLoanSchedulesByLoanID(loanID uint64) ([]model.LoanSchedule, error) {
	query := `
		SELECT 
			id, loan_id, period, due_date, principal_due, interest_due,
			total_due, principal_paid, interest_paid, total_paid,
			status, paid_date, created_at, updated_at
		FROM loan_schedules 
		WHERE loan_id = ?
		ORDER BY period
	`
	
	rows, err := r.db.Query(query, loanID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	
	var schedules []model.LoanSchedule
	for rows.Next() {
		var schedule model.LoanSchedule
		err := rows.Scan(
			&schedule.ID, &schedule.LoanID, &schedule.Period, &schedule.DueDate, &schedule.PrincipalDue, &schedule.InterestDue,
			&schedule.TotalDue, &schedule.PrincipalPaid, &schedule.InterestPaid, &schedule.TotalPaid,
			&schedule.Status, &schedule.PaidDate, &schedule.CreatedAt, &schedule.UpdatedAt,
		)
		if err != nil {
			return nil, err
		}
		schedules = append(schedules, schedule)
	}
	
	return schedules, nil
}

// CreateTransaction 创建交易记录
func (r *LoanRepository) CreateTransaction(transaction *model.Transaction) error {
	query := `
		INSERT INTO transactions (
			transaction_no, loan_id, schedule_id, transaction_type,
			amount, principal_amount, interest_amount, fee_amount,
			penalty_amount, transaction_date, description, created_at
		) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
	`
	
	now := time.Now()
	result, err := r.db.Exec(query,
		transaction.TransactionNo, transaction.LoanID, transaction.ScheduleID, transaction.TransactionType,
		transaction.Amount, transaction.PrincipalAmount, transaction.InterestAmount, transaction.FeeAmount,
		transaction.PenaltyAmount, transaction.TransactionDate, transaction.Description, now,
	)
	if err != nil {
		return err
	}
	
	id, err := result.LastInsertId()
	if err != nil {
		return err
	}
	
	transaction.ID = uint64(id)
	transaction.CreatedAt = now
	
	return nil
}

// GetTransactionsByLoanID 根据贷款ID获取交易记录
func (r *LoanRepository) GetTransactionsByLoanID(loanID uint64) ([]model.Transaction, error) {
	query := `
		SELECT 
			id, transaction_no, loan_id, schedule_id, transaction_type,
			amount, principal_amount, interest_amount, fee_amount,
			penalty_amount, transaction_date, description, created_at
		FROM transactions 
		WHERE loan_id = ?
		ORDER BY transaction_date DESC
	`
	
	rows, err := r.db.Query(query, loanID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	
	var transactions []model.Transaction
	for rows.Next() {
		var transaction model.Transaction
		err := rows.Scan(
			&transaction.ID, &transaction.TransactionNo, &transaction.LoanID, &transaction.ScheduleID, &transaction.TransactionType,
			&transaction.Amount, &transaction.PrincipalAmount, &transaction.InterestAmount, &transaction.FeeAmount,
			&transaction.PenaltyAmount, &transaction.TransactionDate, &transaction.Description, &transaction.CreatedAt,
		)
		if err != nil {
			return nil, err
		}
		transactions = append(transactions, transaction)
	}
	
	return transactions, nil
}