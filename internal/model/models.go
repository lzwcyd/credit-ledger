package model

import (
	"time"
)

// Loan 贷款模型
type Loan struct {
	ID                uint64    `json:"id" db:"id"`
	LoanNo            string    `json:"loan_no" db:"loan_no"`
	Principal         float64   `json:"principal" db:"principal"`
	AnnualInterestRate float64  `json:"annual_interest_rate" db:"annual_interest_rate"`
	TermMonths        int       `json:"term_months" db:"term_months"`
	RepaymentTypeID   uint64    `json:"repayment_type_id" db:"repayment_type_id"`
	DisbursementDate  *time.Time `json:"disbursement_date" db:"disbursement_date"`
	FirstRepaymentDate *time.Time `json:"first_repayment_date" db:"first_repayment_date"`
	MaturityDate      *time.Time `json:"maturity_date" db:"maturity_date"`
	Status            string    `json:"status" db:"status"`
	DisbursedAmount   float64   `json:"disbursed_amount" db:"disbursed_amount"`
	TotalInterest     float64   `json:"total_interest" db:"total_interest"`
	TotalFees         float64   `json:"total_fees" db:"total_fees"`
	RepaidAmount      float64   `json:"repaid_amount" db:"repaid_amount"`
	RemainingPrincipal float64  `json:"remaining_principal" db:"remaining_principal"`
	CreatedAt         time.Time `json:"created_at" db:"created_at"`
	UpdatedAt         time.Time `json:"updated_at" db:"updated_at"`
}

// LoanSchedule 还款计划模型
type LoanSchedule struct {
	ID            uint64    `json:"id" db:"id"`
	LoanID        uint64    `json:"loan_id" db:"loan_id"`
	Period        int       `json:"period" db:"period"`
	DueDate       time.Time `json:"due_date" db:"due_date"`
	PrincipalDue  float64   `json:"principal_due" db:"principal_due"`
	InterestDue   float64   `json:"interest_due" db:"interest_due"`
	TotalDue      float64   `json:"total_due" db:"total_due"`
	PrincipalPaid float64   `json:"principal_paid" db:"principal_paid"`
	InterestPaid  float64   `json:"interest_paid" db:"interest_paid"`
	TotalPaid     float64   `json:"total_paid" db:"total_paid"`
	Status        string    `json:"status" db:"status"`
	PaidDate      *time.Time `json:"paid_date" db:"paid_date"`
	CreatedAt     time.Time `json:"created_at" db:"created_at"`
	UpdatedAt     time.Time `json:"updated_at" db:"updated_at"`
}

// Transaction 交易流水模型
type Transaction struct {
	ID              uint64    `json:"id" db:"id"`
	TransactionNo   string    `json:"transaction_no" db:"transaction_no"`
	LoanID          uint64    `json:"loan_id" db:"loan_id"`
	ScheduleID      *uint64   `json:"schedule_id" db:"schedule_id"`
	TransactionType string    `json:"transaction_type" db:"transaction_type"`
	Amount          float64   `json:"amount" db:"amount"`
	PrincipalAmount float64   `json:"principal_amount" db:"principal_amount"`
	InterestAmount  float64   `json:"interest_amount" db:"interest_amount"`
	FeeAmount       float64   `json:"fee_amount" db:"fee_amount"`
	PenaltyAmount   float64   `json:"penalty_amount" db:"penalty_amount"`
	TransactionDate time.Time `json:"transaction_date" db:"transaction_date"`
	Description     string    `json:"description" db:"description"`
	CreatedAt       time.Time `json:"created_at" db:"created_at"`
}

// InterestCalculation 利息计算模型
type InterestCalculation struct {
	ID               uint64    `json:"id" db:"id"`
	LoanID           uint64    `json:"loan_id" db:"loan_id"`
	CalculationDate  time.Time `json:"calculation_date" db:"calculation_date"`
	PrincipalBalance float64   `json:"principal_balance" db:"principal_balance"`
	DailyInterestRate float64  `json:"daily_interest_rate" db:"daily_interest_rate"`
	InterestAmount   float64   `json:"interest_amount" db:"interest_amount"`
	CreatedAt        time.Time `json:"created_at" db:"created_at"`
}

// Fee 费用模型
type Fee struct {
	ID           uint64     `json:"id" db:"id"`
	LoanID       uint64     `json:"loan_id" db:"loan_id"`
	FeeConfigID  uint64     `json:"fee_config_id" db:"fee_config_id"`
	Amount       float64    `json:"amount" db:"amount"`
	Status       string     `json:"status" db:"status"`
	DueDate      *time.Time `json:"due_date" db:"due_date"`
	PaidDate     *time.Time `json:"paid_date" db:"paid_date"`
	CreatedAt    time.Time  `json:"created_at" db:"created_at"`
	UpdatedAt    time.Time  `json:"updated_at" db:"updated_at"`
}

// FeeConfig 费用配置模型
type FeeConfig struct {
	ID             uint64    `json:"id" db:"id"`
	Code           string    `json:"code" db:"code"`
	Name           string    `json:"name" db:"name"`
	FeeType        string    `json:"fee_type" db:"fee_type"`
	CalculationBase string   `json:"calculation_base" db:"calculation_base"`
	Value          float64   `json:"value" db:"value"`
	MinAmount      float64   `json:"min_amount" db:"min_amount"`
	MaxAmount      *float64  `json:"max_amount" db:"max_amount"`
	IsActive       bool      `json:"is_active" db:"is_active"`
	CreatedAt      time.Time `json:"created_at" db:"created_at"`
	UpdatedAt      time.Time `json:"updated_at" db:"updated_at"`
}

// RepaymentType 还款类型模型
type RepaymentType struct {
	ID          uint64    `json:"id" db:"id"`
	Code        string    `json:"code" db:"code"`
	Name        string    `json:"name" db:"name"`
	Description string    `json:"description" db:"description"`
	IsActive    bool      `json:"is_active" db:"is_active"`
	CreatedAt   time.Time `json:"created_at" db:"created_at"`
	UpdatedAt   time.Time `json:"updated_at" db:"updated_at"`
}