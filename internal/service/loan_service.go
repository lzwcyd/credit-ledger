package service

import (
	"time"
	"github.com/yourorg/credit-ledger/internal/calculator"
	"github.com/yourorg/credit-ledger/pkg/decimal"
)

// LoanService 贷款服务
type LoanService struct {
	calculatorFactory *calculator.CalculatorFactory
}

// NewLoanService 创建贷款服务
func NewLoanService() *LoanService {
	return &LoanService{
		calculatorFactory: calculator.NewCalculatorFactory(),
	}
}

// CreateLoanRequest 创建贷款请求
type CreateLoanRequest struct {
	LoanNo            string  `json:"loan_no"`
	Principal         float64 `json:"principal"`
	AnnualInterestRate float64 `json:"annual_interest_rate"`
	TermMonths        int     `json:"term_months"`
	RepaymentType     string  `json:"repayment_type"`
}

// CreateLoanResponse 创建贷款响应
type CreateLoanResponse struct {
	LoanID   uint64 `json:"loan_id"`
	LoanNo   string `json:"loan_no"`
	Message  string `json:"message"`
}

// DisburseLoanRequest 放款请求
type DisburseLoanRequest struct {
	LoanID           uint64 `json:"loan_id"`
	DisbursementDate string `json:"disbursement_date"`
}

// DisburseLoanResponse 放款响应
type DisburseLoanResponse struct {
	LoanID      uint64 `json:"loan_id"`
	Message     string `json:"message"`
	ScheduleID  uint64 `json:"schedule_id,omitempty"`
}

// RepaymentTrialRequest 还款试算请求
type RepaymentTrialRequest struct {
	LoanID         uint64  `json:"loan_id"`
	RepaymentDate  string  `json:"repayment_date"`
	Amount         float64 `json:"amount"`
	RepaymentType  string  `json:"repayment_type"` // FULL, PARTIAL, EARLY_SETTLEMENT
}

// RepaymentTrialResponse 还款试算响应
type RepaymentTrialResponse struct {
	LoanID           uint64  `json:"loan_id"`
	RepaymentDate    string  `json:"repayment_date"`
	PrincipalAmount  float64 `json:"principal_amount"`
	InterestAmount   float64 `json:"interest_amount"`
	FeeAmount        float64 `json:"fee_amount"`
	TotalAmount      float64 `json:"total_amount"`
	Message          string  `json:"message"`
}

// CreateLoan 创建贷款
func (s *LoanService) CreateLoan(req CreateLoanRequest) (*CreateLoanResponse, error) {
	// 这里应该实现创建贷款的业务逻辑
	// 包括验证、数据库操作等
	
	// 计算还款计划
	startDate := time.Now()
	calculator := s.calculatorFactory.GetCalculator(req.RepaymentType)
	principal := decimal.NewFromFloat(req.Principal)
	annualRate := decimal.NewFromFloat(req.AnnualInterestRate)
	_ = calculator.CalculateSchedule(principal, annualRate, req.TermMonths, startDate)
	
	// 保存贷款和还款计划到数据库
	// 这里只是示例，实际应该调用 repository 层
	
	return &CreateLoanResponse{
		LoanID:  1, // 示例 ID
		LoanNo:  req.LoanNo,
		Message: "贷款创建成功",
	}, nil
}

// DisburseLoan 放款
func (s *LoanService) DisburseLoan(req DisburseLoanRequest) (*DisburseLoanResponse, error) {
	// 这里应该实现放款的业务逻辑
	// 包括验证贷款状态、生成放款记录等
	
	return &DisburseLoanResponse{
		LoanID:  req.LoanID,
		Message: "放款成功",
	}, nil
}

// RepaymentTrial 还款试算
func (s *LoanService) RepaymentTrial(req RepaymentTrialRequest) (*RepaymentTrialResponse, error) {
	// 这里应该实现还款试算的业务逻辑
	// 包括计算应还本金、利息、费用等
	
	// 示例计算
	principalAmount := req.Amount * 0.8  // 假设80%是本金
	interestAmount := req.Amount * 0.15  // 假设15%是利息
	feeAmount := req.Amount * 0.05       // 假设5%是费用
	
	return &RepaymentTrialResponse{
		LoanID:          req.LoanID,
		RepaymentDate:   req.RepaymentDate,
		PrincipalAmount: principalAmount,
		InterestAmount:  interestAmount,
		FeeAmount:       feeAmount,
		TotalAmount:     req.Amount,
		Message:         "还款试算成功",
	}, nil
}