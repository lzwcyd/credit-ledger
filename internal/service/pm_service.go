package service

import (
	"fmt"
	"time"

	"github.com/lzwcyd/credit-ledger/internal/model"
	"github.com/lzwcyd/credit-ledger/pkg/decimal"
)

// =====================================================
// 信贷PM核心功能服务
// =====================================================

// --- 1. 催收状态管理 ---

// UpdateCollectionStatusRequest 更新催收状态请求
type UpdateCollectionStatusRequest struct {
	LoanNo     string `json:"loan_no"`
	Status     string `json:"status"`      // NORMAL/IN_COLLECTION/LEGAL/WRITTEN_OFF
	Notes      string `json:"notes"`       // 催收备注
	Operator   string `json:"operator"`
}

// UpdateCollectionStatus 更新催收状态
func (s *LoanService) UpdateCollectionStatus(req UpdateCollectionStatusRequest) (*model.Loan, error) {
	loan, err := s.loanRepo.GetLoanByNo(req.LoanNo)
	if err != nil {
		return nil, fmt.Errorf("借据不存在: %w", err)
	}

	oldStatus := loan.CollectionStatus
	if req.Status != "NORMAL" && req.Status != "IN_COLLECTION" && req.Status != "LEGAL" && req.Status != "WRITTEN_OFF" {
		return nil, fmt.Errorf("无效的催收状态: %s", req.Status)
	}
	loan.CollectionStatus = req.Status
	loan.CollectionNotes = req.Notes
	now := time.Now()
	loan.LastCollectionDate = &now
	loan.UpdatedBy = req.Operator

	if err := s.loanRepo.UpdateLoan(loan); err != nil {
		return nil, fmt.Errorf("更新催收状态失败: %w", err)
	}

	// 记录变更
	change := &model.LoanChange{
		LoanNo:       loan.LoanNo,
		ChangeType:   "COLLECTION_STATUS_CHANGE",
		FieldName:    "collection_status",
		OldValue:     oldStatus,
		NewValue:     req.Status,
		ChangeReason: "更新催收状态",
		CreatedBy:    req.Operator,
		UpdatedBy:    req.Operator,
	}
	if err := s.loanRepo.CreateLoanChange(change); err != nil {
		return nil, fmt.Errorf("记录变更失败: %w", err)
	}

	return loan, nil
}

// --- 2. 罚息减免 ---

// PenaltyWaiverRequest 罚息减免请求
type PenaltyWaiverRequest struct {
	LoanNo         string  `json:"loan_no"`
	WaiverType     string  `json:"waiver_type"`   // PENALTY/INTEREST/OTHER_FEE
	WaiverAmount   string  `json:"waiver_amount"` // 改为 string 类型，避免精度损失
	Reason         string  `json:"reason"`
	ApprovedBy     string  `json:"approved_by"`
	Operator       string  `json:"operator"`
}

// PenaltyWaiverResponse 罚息减免响应
type PenaltyWaiverResponse struct {
	WaiverNo       string          `json:"waiver_no"`
	LoanNo         string          `json:"loan_no"`
	WaiverType     string          `json:"waiver_type"`
	WaiverAmount   decimal.Decimal `json:"waiver_amount"`
	OriginalAmount decimal.Decimal `json:"original_amount"`
	Status         string          `json:"status"`
}

// ApplyPenaltyWaiver 申请罚息减免
func (s *LoanService) ApplyPenaltyWaiver(req PenaltyWaiverRequest) (*PenaltyWaiverResponse, error) {
	loan, err := s.loanRepo.GetLoanByNo(req.LoanNo)
	if err != nil {
		return nil, fmt.Errorf("借据不存在: %w", err)
	}

	if loan.Status != "DISBURSED" && loan.Status != "OVERDUE" {
		return nil, fmt.Errorf("借据状态 %s 不允许减免", loan.Status)
	}

	if req.WaiverType != "PENALTY" && req.WaiverType != "INTEREST" && req.WaiverType != "OTHER_FEE" {
		return nil, fmt.Errorf("无效的减免类型: %s", req.WaiverType)
	}

	// 获取原始金额
	var originalAmount decimal.Decimal
	switch req.WaiverType {
	case "PENALTY":
		originalAmount = loan.TotalPenalty.Sub(loan.PaidPenalty)
	case "INTEREST":
		originalAmount = loan.TotalInterest.Sub(loan.PaidInterest)
	case "OTHER_FEE":
		originalAmount = loan.TotalOtherFee.Sub(loan.PaidOtherFee)
	}

	waiverAmount, err := decimal.NewFromString(req.WaiverAmount)
	if err != nil {
		return nil, fmt.Errorf("减免金额格式错误: %w", err)
	}

	if waiverAmount.Lte(decimal.Zero()) {
		return nil, fmt.Errorf("减免金额必须大于0")
	}

	// 校验减免金额不超过原始金额
	if waiverAmount.Gt(originalAmount) {
		return nil, fmt.Errorf("减免金额 %s 超过待减免金额 %s", waiverAmount, originalAmount)
	}

	// 生成减免编号
	suffix := loan.LoanNo
	if len(suffix) > 4 {
		suffix = suffix[len(suffix)-4:]
	}
	waiverNo := fmt.Sprintf("WV%s%s", time.Now().Format("20060102150405"), suffix)

	// 直接应用减免（简化流程，实际可能需要审批）
	// 更新借据的已还金额
	switch req.WaiverType {
	case "PENALTY":
		loan.PaidPenalty = loan.PaidPenalty.Add(waiverAmount)
	case "INTEREST":
		loan.PaidInterest = loan.PaidInterest.Add(waiverAmount)
	case "OTHER_FEE":
		loan.PaidOtherFee = loan.PaidOtherFee.Add(waiverAmount)
	}
	loan.UpdatedBy = req.Operator

	if err := s.loanRepo.UpdateLoan(loan); err != nil {
		return nil, fmt.Errorf("更新借据失败: %w", err)
	}

	return &PenaltyWaiverResponse{
		WaiverNo:       waiverNo,
		LoanNo:         req.LoanNo,
		WaiverType:     req.WaiverType,
		WaiverAmount:   waiverAmount,
		OriginalAmount: originalAmount,
		Status:         "APPLIED",
	}, nil
}

// --- 3. 借据展期 ---

// ExtensionRequest 展期请求
type ExtensionRequest struct {
	LoanNo          string `json:"loan_no"`
	ExtensionDays   int    `json:"extension_days"`   // 展期天数
	ExtensionMonths int    `json:"extension_months"` // 展期月数
	Reason          string `json:"reason"`
	Operator        string `json:"operator"`
}

// ExtensionResponse 展期响应
type ExtensionResponse struct {
	ExtensionNo     string    `json:"extension_no"`
	LoanNo          string    `json:"loan_no"`
	OriginalMaturity time.Time `json:"original_maturity"`
	NewMaturity     time.Time `json:"new_maturity"`
	ExtensionDays   int       `json:"extension_days"`
	ExtensionMonths int       `json:"extension_months"`
}

// ApplyExtension 申请展期
func (s *LoanService) ApplyExtension(req ExtensionRequest) (*ExtensionResponse, error) {
	loan, err := s.loanRepo.GetLoanByNo(req.LoanNo)
	if err != nil {
		return nil, fmt.Errorf("借据不存在: %w", err)
	}

	if loan.Status != "DISBURSED" && loan.Status != "OVERDUE" {
		return nil, fmt.Errorf("借据状态 %s 不允许展期", loan.Status)
	}

	if req.ExtensionDays <= 0 && req.ExtensionMonths <= 0 {
		return nil, fmt.Errorf("展期天数或月数必须大于0")
	}

	// 计算新到期日
	originalMaturity := loan.MaturityDate
	newMaturity := originalMaturity.AddDate(0, req.ExtensionMonths, req.ExtensionDays)

	// 生成展期编号
	suffix := loan.LoanNo
	if len(suffix) > 4 {
		suffix = suffix[len(suffix)-4:]
	}
	extensionNo := fmt.Sprintf("EX%s%s", time.Now().Format("20060102150405"), suffix)

	// 更新借据到期日
	oldMaturity := loan.MaturityDate
	loan.MaturityDate = newMaturity
	loan.UpdatedBy = req.Operator

	if err := s.loanRepo.UpdateLoan(loan); err != nil {
		return nil, fmt.Errorf("更新借据到期日失败: %w", err)
	}

	// 记录变更
	change := &model.LoanChange{
		LoanNo:       loan.LoanNo,
		ChangeType:   "EXTENSION",
		FieldName:    "maturity_date",
		OldValue:     oldMaturity.Format("2006-01-02"),
		NewValue:     newMaturity.Format("2006-01-02"),
		ChangeReason: fmt.Sprintf("展期%d天%d月: %s", req.ExtensionDays, req.ExtensionMonths, req.Reason),
		CreatedBy:    req.Operator,
		UpdatedBy:    req.Operator,
	}
	if err := s.loanRepo.CreateLoanChange(change); err != nil {
		return nil, fmt.Errorf("记录变更失败: %w", err)
	}

	return &ExtensionResponse{
		ExtensionNo:      extensionNo,
		LoanNo:           req.LoanNo,
		OriginalMaturity: originalMaturity,
		NewMaturity:      newMaturity,
		ExtensionDays:    req.ExtensionDays,
		ExtensionMonths:  req.ExtensionMonths,
	}, nil
}

// --- 4. 坏账核销 ---

// WriteOffRequest 核销请求
type WriteOffRequest struct {
	LoanNo   string  `json:"loan_no"`
	Reason   string  `json:"reason"`
	Operator string  `json:"operator"`
}

// WriteOffResponse 核销响应
type WriteOffResponse struct {
	WriteOffNo      string          `json:"write_off_no"`
	LoanNo          string          `json:"loan_no"`
	WriteOffAmount  decimal.Decimal `json:"write_off_amount"`
	PrincipalAmount decimal.Decimal `json:"principal_amount"`
	InterestAmount  decimal.Decimal `json:"interest_amount"`
	PenaltyAmount   decimal.Decimal `json:"penalty_amount"`
}

// ApplyWriteOff 申请坏账核销
func (s *LoanService) ApplyWriteOff(req WriteOffRequest) (*WriteOffResponse, error) {
	loan, err := s.loanRepo.GetLoanByNo(req.LoanNo)
	if err != nil {
		return nil, fmt.Errorf("借据不存在: %w", err)
	}

	if loan.Status == "WRITTEN_OFF" {
		return nil, fmt.Errorf("借据已核销，不可重复核销")
	}
	if loan.Status == "REPAID" {
		return nil, fmt.Errorf("借据已结清，不可核销")
	}

	// 计算核销金额 = 剩余本金 + 未还利息 + 未还罚息
	principalAmount := loan.RemainingPrincipal
	interestAmount := loan.TotalInterest.Sub(loan.PaidInterest)
	penaltyAmount := loan.TotalPenalty.Sub(loan.PaidPenalty)
	writeOffAmount := principalAmount.Add(interestAmount).Add(penaltyAmount)

	if writeOffAmount.IsZero() || writeOffAmount.IsNegative() {
		return nil, fmt.Errorf("无可核销金额")
	}

	// 生成核销编号
	suffix := loan.LoanNo
	if len(suffix) > 4 {
		suffix = suffix[len(suffix)-4:]
	}
	writeOffNo := fmt.Sprintf("WO%s%s", time.Now().Format("20060102150405"), suffix)

	// 更新借据状态
	oldStatus := loan.Status
	loan.Status = "WRITTEN_OFF"
	loan.CollectionStatus = model.CollectionStatusWrittenOff
	// 将所有未还金额标记为已还（核销）
	loan.PaidPrincipal = loan.PaidPrincipal.Add(principalAmount)
	loan.PaidInterest = loan.PaidInterest.Add(interestAmount)
	loan.PaidPenalty = loan.PaidPenalty.Add(penaltyAmount)
	loan.RemainingPrincipal = decimal.Zero()
	now := time.Now()
	loan.SettlementDate = &now
	loan.UpdatedBy = req.Operator

	if err := s.loanRepo.UpdateLoan(loan); err != nil {
		return nil, fmt.Errorf("更新借据状态失败: %w", err)
	}

	// 记录变更
	change := &model.LoanChange{
		LoanNo:       loan.LoanNo,
		ChangeType:   "WRITE_OFF",
		FieldName:    "status",
		OldValue:     oldStatus,
		NewValue:     "WRITTEN_OFF",
		ChangeReason: fmt.Sprintf("坏账核销: %s，核销金额: %s", req.Reason, writeOffAmount),
		CreatedBy:    req.Operator,
		UpdatedBy:    req.Operator,
	}
	if err := s.loanRepo.CreateLoanChange(change); err != nil {
		return nil, fmt.Errorf("记录变更失败: %w", err)
	}

	// 将所有未结清计划标记为结清
	plans, err := s.planRepo.GetPlansByLoanNo(req.LoanNo)
	if err != nil {
		return nil, fmt.Errorf("查询还款计划失败: %w", err)
	}
	for i := range plans {
		if plans[i].Status != "PAID" {
			plans[i].Status = "PAID"
			plans[i].UpdatedBy = req.Operator
			if err := s.planRepo.UpdatePlan(&plans[i]); err != nil {
				return nil, fmt.Errorf("更新还款计划失败: %w", err)
			}
		}
	}

	return &WriteOffResponse{
		WriteOffNo:      writeOffNo,
		LoanNo:          req.LoanNo,
		WriteOffAmount:  writeOffAmount,
		PrincipalAmount: principalAmount,
		InterestAmount:  interestAmount,
		PenaltyAmount:   penaltyAmount,
	}, nil
}

// --- 5. 还款提醒 ---

// GetUpcomingDuePlans 获取即将到期的还款计划
func (s *LoanService) GetUpcomingDuePlans(daysBefore int) ([]model.Plan, error) {
	if daysBefore <= 0 {
		daysBefore = 3
	}
	targetDate := time.Now().AddDate(0, 0, daysBefore)
	return s.planRepo.GetPlansDueBefore(targetDate)
}

// --- 6. 客户对账单 ---

// GenerateStatement 生成客户对账单
func (s *LoanService) GenerateStatement(loanNo string) (*model.CustomerStatement, error) {
	loan, err := s.loanRepo.GetLoanByNo(loanNo)
	if err != nil {
		return nil, fmt.Errorf("借据不存在: %w", err)
	}

	plans, err := s.planRepo.GetPlansByLoanNo(loanNo)
	if err != nil {
		return nil, fmt.Errorf("查询还款计划失败: %w", err)
	}

	// 计算应还总额
	totalDue := decimal.Zero()
	for _, plan := range plans {
		totalDue = totalDue.Add(plan.DueTotal)
	}

	// 计算已还总额
	totalPaid := loan.PaidPrincipal.Add(loan.PaidInterest).Add(loan.PaidPenalty).Add(loan.PaidOtherFee)

	// 获取最近的还款记录（取最近5条）
	// 注意：这里简化处理，实际应该从 repayment 表查询

	return &model.CustomerStatement{
		LoanNo:             loanNo,
		StatementDate:      time.Now(),
		Principal:          loan.Principal,
		AnnualRate:         loan.AnnualRate,
		RemainingPrincipal: loan.RemainingPrincipal,
		TotalPaid:          totalPaid,
		TotalDue:           totalDue,
		OverdueDays:        loan.OverdueDays,
		OverdueTier:        loan.OverdueTier,
		Status:             loan.Status,
		Plans:              plans,
	}, nil
}
