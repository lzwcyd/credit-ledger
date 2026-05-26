package model

import (
	"time"
	"github.com/yourorg/credit-ledger/pkg/decimal"
)

// =====================================================
// 基础配置模型
// =====================================================

// RepaymentType 还款类型
type RepaymentType struct {
	ID          uint64    `json:"id" db:"id"`
	Code        string    `json:"code" db:"code"`
	Name        string    `json:"name" db:"name"`
	Description string    `json:"description" db:"description"`
	IsActive    bool      `json:"is_active" db:"is_active"`
	CreatedBy   string    `json:"created_by" db:"created_by"`
	UpdatedBy   string    `json:"updated_by" db:"updated_by"`
	CreatedAt   time.Time `json:"created_at" db:"created_at"`
	UpdatedAt   time.Time `json:"updated_at" db:"updated_at"`
}

// FeeConfig 费项配置
type FeeConfig struct {
	ID                uint64          `json:"id" db:"id"`
	Code              string          `json:"code" db:"code"`
	Name              string          `json:"name" db:"name"`
	CalcType          string          `json:"calc_type" db:"calc_type"`       // FIXED, PERCENTAGE, DAILY_RATE
	CalcBase          string          `json:"calc_base" db:"calc_base"`       // PRINCIPAL, INTEREST, TOTAL, REMAINING_PRINCIPAL, OVERDUE_AMOUNT
	Value             decimal.Decimal `json:"value" db:"value"`
	TriggerType       string          `json:"trigger_type" db:"trigger_type"` // DISBURSEMENT, REPAYMENT, DAILY, OVERDUE, EARLY_REPAYMENT
	IsDailyAccumulate bool            `json:"is_daily_accumulate" db:"is_daily_accumulate"`
	FeeCategory       string          `json:"fee_category" db:"fee_category"` // INTEREST, PENALTY, OTHER_FEE
	MinAmount         decimal.Decimal `json:"min_amount" db:"min_amount"`
	MaxAmount         *decimal.Decimal `json:"max_amount" db:"max_amount"`
	IsActive          bool            `json:"is_active" db:"is_active"`
	CreatedBy         string          `json:"created_by" db:"created_by"`
	UpdatedBy         string          `json:"updated_by" db:"updated_by"`
	CreatedAt         time.Time       `json:"created_at" db:"created_at"`
	UpdatedAt         time.Time       `json:"updated_at" db:"updated_at"`
}

// AllocationRule 分配规则
type AllocationRule struct {
	ID          uint64    `json:"id" db:"id"`
	Code        string    `json:"code" db:"code"`
	Name        string    `json:"name" db:"name"`
	Description string    `json:"description" db:"description"`
	IsDefault   bool      `json:"is_default" db:"is_default"`
	IsActive    bool      `json:"is_active" db:"is_active"`
	CreatedBy   string    `json:"created_by" db:"created_by"`
	UpdatedBy   string    `json:"updated_by" db:"updated_by"`
	CreatedAt   time.Time `json:"created_at" db:"created_at"`
	UpdatedAt   time.Time `json:"updated_at" db:"updated_at"`
}

// AllocationRuleItem 分配规则明细
type AllocationRuleItem struct {
	ID             uint64    `json:"id" db:"id"`
	RuleCode       string    `json:"rule_code" db:"rule_code"`
	Priority       int       `json:"priority" db:"priority"`
	AllocationType string    `json:"allocation_type" db:"allocation_type"` // PENALTY, OTHER_FEE, INTEREST, PRINCIPAL
	Description    string    `json:"description" db:"description"`
	CreatedBy      string    `json:"created_by" db:"created_by"`
	UpdatedBy      string    `json:"updated_by" db:"updated_by"`
	CreatedAt      time.Time `json:"created_at" db:"created_at"`
	UpdatedAt      time.Time `json:"updated_at" db:"updated_at"`
}

// =====================================================
// 借据模型
// =====================================================

// Loan 借据主表
type Loan struct {
	ID                   uint64          `json:"id" db:"id"`
	LoanNo               string          `json:"loan_no" db:"loan_no"`
	Principal            decimal.Decimal `json:"principal" db:"principal"`
	AnnualRate           decimal.Decimal `json:"annual_rate" db:"annual_rate"`
	TermMonths           int             `json:"term_months" db:"term_months"`
	RepaymentTypeCode    string          `json:"repayment_type_code" db:"repayment_type_code"`
	AllocationRuleCode   string          `json:"allocation_rule_code" db:"allocation_rule_code"`
	
	ValueDate            time.Time       `json:"value_date" db:"value_date"`
	FirstDueDate         time.Time       `json:"first_due_date" db:"first_due_date"`
	MaturityDate         time.Time       `json:"maturity_date" db:"maturity_date"`
	SettlementDate       *time.Time      `json:"settlement_date" db:"settlement_date"`
	DisbursementDate     *time.Time      `json:"disbursement_date" db:"disbursement_date"`
	
	Status               string          `json:"status" db:"status"`
	
	DisbursedAmount      decimal.Decimal `json:"disbursed_amount" db:"disbursed_amount"`
	RemainingPrincipal   decimal.Decimal `json:"remaining_principal" db:"remaining_principal"`
	
	TotalInterest        decimal.Decimal `json:"total_interest" db:"total_interest"`
	TotalPenalty         decimal.Decimal `json:"total_penalty" db:"total_penalty"`
	TotalOtherFee        decimal.Decimal `json:"total_other_fee" db:"total_other_fee"`
	
	PaidPrincipal        decimal.Decimal `json:"paid_principal" db:"paid_principal"`
	PaidInterest         decimal.Decimal `json:"paid_interest" db:"paid_interest"`
	PaidPenalty          decimal.Decimal `json:"paid_penalty" db:"paid_penalty"`
	PaidOtherFee         decimal.Decimal `json:"paid_other_fee" db:"paid_other_fee"`
	
	OverdueDays          int             `json:"overdue_days" db:"overdue_days"`
	OverduePrincipal     decimal.Decimal `json:"overdue_principal" db:"overdue_principal"`
	
	CreatedBy            string          `json:"created_by" db:"created_by"`
	UpdatedBy            string          `json:"updated_by" db:"updated_by"`
	CreatedAt            time.Time       `json:"created_at" db:"created_at"`
	UpdatedAt            time.Time       `json:"updated_at" db:"updated_at"`
}

// LoanChange 借据变更记录
type LoanChange struct {
	ID                   uint64    `json:"id" db:"id"`
	LoanNo               string    `json:"loan_no" db:"loan_no"`
	ChangeType           string    `json:"change_type" db:"change_type"`
	FieldName            string    `json:"field_name" db:"field_name"`
	OldValue             string    `json:"old_value" db:"old_value"`
	NewValue             string    `json:"new_value" db:"new_value"`
	ChangeReason         string    `json:"change_reason" db:"change_reason"`
	RelatedRepaymentNo   string    `json:"related_repayment_no" db:"related_repayment_no"`
	BatchNo              string    `json:"batch_no" db:"batch_no"`
	CreatedBy            string    `json:"created_by" db:"created_by"`
	UpdatedBy            string    `json:"updated_by" db:"updated_by"`
	CreatedAt            time.Time `json:"created_at" db:"created_at"`
	UpdatedAt            time.Time `json:"updated_at" db:"updated_at"`
}

// =====================================================
// 还款计划模型
// =====================================================

// Plan 还款计划
type Plan struct {
	ID             uint64          `json:"id" db:"id"`
	LoanNo         string          `json:"loan_no" db:"loan_no"`
	Period         int             `json:"period" db:"period"`
	DueDate        time.Time       `json:"due_date" db:"due_date"`
	
	DuePrincipal   decimal.Decimal `json:"due_principal" db:"due_principal"`
	DueInterest    decimal.Decimal `json:"due_interest" db:"due_interest"`
	DuePenalty     decimal.Decimal `json:"due_penalty" db:"due_penalty"`
	DueOtherFee    decimal.Decimal `json:"due_other_fee" db:"due_other_fee"`
	DueTotal       decimal.Decimal `json:"due_total" db:"due_total"`
	
	PaidPrincipal  decimal.Decimal `json:"paid_principal" db:"paid_principal"`
	PaidInterest   decimal.Decimal `json:"paid_interest" db:"paid_interest"`
	PaidPenalty    decimal.Decimal `json:"paid_penalty" db:"paid_penalty"`
	PaidOtherFee   decimal.Decimal `json:"paid_other_fee" db:"paid_other_fee"`
	PaidTotal      decimal.Decimal `json:"paid_total" db:"paid_total"`
	
	OverdueDays    int             `json:"overdue_days" db:"overdue_days"`
	
	Status         string          `json:"status" db:"status"`
	
	CreatedBy      string          `json:"created_by" db:"created_by"`
	UpdatedBy      string          `json:"updated_by" db:"updated_by"`
	CreatedAt      time.Time       `json:"created_at" db:"created_at"`
	UpdatedAt      time.Time       `json:"updated_at" db:"updated_at"`
}

// PlanChange 还款计划变更记录
type PlanChange struct {
	ID                 uint64    `json:"id" db:"id"`
	LoanNo             string    `json:"loan_no" db:"loan_no"`
	PlanID             uint64    `json:"plan_id" db:"plan_id"`
	Period             int       `json:"period" db:"period"`
	ChangeType         string    `json:"change_type" db:"change_type"`
	FieldName          string    `json:"field_name" db:"field_name"`
	OldValue           string    `json:"old_value" db:"old_value"`
	NewValue           string    `json:"new_value" db:"new_value"`
	ChangeReason       string    `json:"change_reason" db:"change_reason"`
	RelatedRepaymentNo string    `json:"related_repayment_no" db:"related_repayment_no"`
	BatchNo            string    `json:"batch_no" db:"batch_no"`
	CreatedBy          string    `json:"created_by" db:"created_by"`
	UpdatedBy          string    `json:"updated_by" db:"updated_by"`
	CreatedAt          time.Time `json:"created_at" db:"created_at"`
	UpdatedAt          time.Time `json:"updated_at" db:"updated_at"`
}

// PlanOtherFee 还款计划-其他费用明细
type PlanOtherFee struct {
	ID         uint64          `json:"id" db:"id"`
	LoanNo     string          `json:"loan_no" db:"loan_no"`
	PlanID     uint64          `json:"plan_id" db:"plan_id"`
	Period     int             `json:"period" db:"period"`
	FeeCode    string          `json:"fee_code" db:"fee_code"`
	FeeName    string          `json:"fee_name" db:"fee_name"`
	DueAmount  decimal.Decimal `json:"due_amount" db:"due_amount"`
	PaidAmount decimal.Decimal `json:"paid_amount" db:"paid_amount"`
	Status     string          `json:"status" db:"status"`
	CreatedBy  string          `json:"created_by" db:"created_by"`
	UpdatedBy  string          `json:"updated_by" db:"updated_by"`
	CreatedAt  time.Time       `json:"created_at" db:"created_at"`
	UpdatedAt  time.Time       `json:"updated_at" db:"updated_at"`
}

// =====================================================
// 每日计算模型
// =====================================================

// DailyCalculation 每日计算明细
type DailyCalculation struct {
	ID                  uint64          `json:"id" db:"id"`
	LoanNo              string          `json:"loan_no" db:"loan_no"`
	CalculationDate     time.Time       `json:"calculation_date" db:"calculation_date"`
	FeeCode             string          `json:"fee_code" db:"fee_code"`
	FeeCategory         string          `json:"fee_category" db:"fee_category"`
	BaseAmount          decimal.Decimal `json:"base_amount" db:"base_amount"`
	DailyRate           decimal.Decimal `json:"daily_rate" db:"daily_rate"`
	Amount              decimal.Decimal `json:"amount" db:"amount"`
	PlanID              *uint64         `json:"plan_id" db:"plan_id"`
	IsSettled           bool            `json:"is_settled" db:"is_settled"`
	SettledRepaymentNo  string          `json:"settled_repayment_no" db:"settled_repayment_no"`
	BatchNo             string          `json:"batch_no" db:"batch_no"`
	CreatedBy           string          `json:"created_by" db:"created_by"`
	UpdatedBy           string          `json:"updated_by" db:"updated_by"`
	CreatedAt           time.Time       `json:"created_at" db:"created_at"`
	UpdatedAt           time.Time       `json:"updated_at" db:"updated_at"`
}

// =====================================================
// 还款记录模型
// =====================================================

// Repayment 还款记录
type Repayment struct {
	ID                 uint64          `json:"id" db:"id"`
	RepaymentNo        string          `json:"repayment_no" db:"repayment_no"`
	LoanNo             string          `json:"loan_no" db:"loan_no"`
	PlanID             *uint64         `json:"plan_id" db:"plan_id"`
	RepaymentType      string          `json:"repayment_type" db:"repayment_type"`
	Amount             decimal.Decimal `json:"amount" db:"amount"`
	PrincipalAmount    decimal.Decimal `json:"principal_amount" db:"principal_amount"`
	InterestAmount     decimal.Decimal `json:"interest_amount" db:"interest_amount"`
	PenaltyAmount      decimal.Decimal `json:"penalty_amount" db:"penalty_amount"`
	OtherFeeAmount     decimal.Decimal `json:"other_fee_amount" db:"other_fee_amount"`
	TrialDate          time.Time       `json:"trial_date" db:"trial_date"`
	BookingDate        time.Time       `json:"booking_date" db:"booking_date"`
	AllocationRuleCode string          `json:"allocation_rule_code" db:"allocation_rule_code"`
	Status             string          `json:"status" db:"status"`
	Description        string          `json:"description" db:"description"`
	IsBackdated        bool            `json:"is_backdated" db:"is_backdated"`
	BackdatedReason    string          `json:"backdated_reason" db:"backdated_reason"`
	BatchNo            string          `json:"batch_no" db:"batch_no"`
	CreatedBy          string          `json:"created_by" db:"created_by"`
	UpdatedBy          string          `json:"updated_by" db:"updated_by"`
	CreatedAt          time.Time       `json:"created_at" db:"created_at"`
	UpdatedAt          time.Time       `json:"updated_at" db:"updated_at"`
}

// RepaymentDetail 还款入账明细
type RepaymentDetail struct {
	ID                  uint64          `json:"id" db:"id"`
	RepaymentNo         string          `json:"repayment_no" db:"repayment_no"`
	LoanNo              string          `json:"loan_no" db:"loan_no"`
	FeeCode             string          `json:"fee_code" db:"fee_code"`
	FeeName             string          `json:"fee_name" db:"fee_name"`
	FeeCategory         string          `json:"fee_category" db:"fee_category"`
	Amount              decimal.Decimal `json:"amount" db:"amount"`
	DailyCalculationID  *uint64         `json:"daily_calculation_id" db:"daily_calculation_id"`
	PlanOtherFeeID      *uint64         `json:"plan_other_fee_id" db:"plan_other_fee_id"`
	CreatedBy           string          `json:"created_by" db:"created_by"`
	UpdatedBy           string          `json:"updated_by" db:"updated_by"`
	CreatedAt           time.Time       `json:"created_at" db:"created_at"`
	UpdatedAt           time.Time       `json:"updated_at" db:"updated_at"`
}

// =====================================================
// 跑批模型
// =====================================================

// BatchJob 跑批批次
type BatchJob struct {
	ID              uint64    `json:"id" db:"id"`
	BatchNo         string    `json:"batch_no" db:"batch_no"`
	BatchType       string    `json:"batch_type" db:"batch_type"`
	BatchDate       time.Time `json:"batch_date" db:"batch_date"`
	Status          string    `json:"status" db:"status"`
	LastProcessedID uint64    `json:"last_processed_id" db:"last_processed_id"`
	PageSize        int       `json:"page_size" db:"page_size"`
	TotalCount      int       `json:"total_count" db:"total_count"`
	ProcessedCount  int       `json:"processed_count" db:"processed_count"`
	SuccessCount    int       `json:"success_count" db:"success_count"`
	FailedCount     int       `json:"failed_count" db:"failed_count"`
	StartTime       *time.Time `json:"start_time" db:"start_time"`
	EndTime         *time.Time `json:"end_time" db:"end_time"`
	DurationMs      int64     `json:"duration_ms" db:"duration_ms"`
	ErrorMessage    string    `json:"error_message" db:"error_message"`
	Remark          string    `json:"remark" db:"remark"`
	CreatedBy       string    `json:"created_by" db:"created_by"`
	UpdatedBy       string    `json:"updated_by" db:"updated_by"`
	CreatedAt       time.Time `json:"created_at" db:"created_at"`
	UpdatedAt       time.Time `json:"updated_at" db:"updated_at"`
}