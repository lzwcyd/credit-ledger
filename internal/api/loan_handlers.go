package api

import (
	"encoding/json"
	"log"
	"net/http"
	"strconv"

	"github.com/gorilla/mux"
	"github.com/yourorg/credit-ledger/internal/service"
)

// =====================================================
// 借据管理
// =====================================================

// createLoan 创建借据
func (r *Router) createLoan(w http.ResponseWriter, req *http.Request) {
	var createReq service.CreateLoanRequest
	if err := json.NewDecoder(req.Body).Decode(&createReq); err != nil {
		BadRequest(w, "请求体格式错误: "+err.Error())
		return
	}

	// 参数校验
	if err := ValidateRequired(map[string]string{
		"loan_no":            createReq.LoanNo,
		"repayment_type_code": createReq.RepaymentTypeCode,
		"value_date":         createReq.ValueDate,
		"first_due_date":     createReq.FirstDueDate,
		"maturity_date":      createReq.MaturityDate,
	}); err != nil {
		BadRequest(w, err.Error())
		return
	}

	if err := ValidateLoanNo(createReq.LoanNo); err != nil {
		BadRequest(w, err.Error())
		return
	}

	if err := ValidateDateFormat(createReq.ValueDate); err != nil {
		BadRequest(w, "起息日: "+err.Error())
		return
	}

	if err := ValidatePositiveDecimal("principal", createReq.Principal); err != nil {
		BadRequest(w, err.Error())
		return
	}

	if err := ValidateRange("term_months", createReq.TermMonths, 1, 360); err != nil {
		BadRequest(w, err.Error())
		return
	}

	loan, err := r.loanService.CreateLoan(createReq)
	if err != nil {
		requestID := GetRequestID(req)
		log.Printf("[request_id=%s] 创建借据失败: %v", requestID, err)
		InternalError(w, "服务器内部错误")
		return
	}

	Created(w, loan)
}

// getLoan 获取借据详情
func (r *Router) getLoan(w http.ResponseWriter, req *http.Request) {
	vars := mux.Vars(req)
	loanNo := vars["loan_no"]

	if err := ValidateLoanNo(loanNo); err != nil {
		BadRequest(w, err.Error())
		return
	}

	loan, err := r.loanService.GetLoanByNo(loanNo)
	if err != nil {
		NotFound(w, "借据不存在: "+loanNo)
		return
	}

	Success(w, loan)
}

// disburseLoan 放款
func (r *Router) disburseLoan(w http.ResponseWriter, req *http.Request) {
	vars := mux.Vars(req)
	loanNo := vars["loan_no"]

	var disburseReq service.DisburseRequest
	if err := json.NewDecoder(req.Body).Decode(&disburseReq); err != nil {
		BadRequest(w, "请求体格式错误: "+err.Error())
		return
	}

	disburseReq.LoanNo = loanNo

	if err := ValidateRequired(map[string]string{
		"disburse_date": disburseReq.DisburseDate,
	}); err != nil {
		BadRequest(w, err.Error())
		return
	}

	if err := ValidatePositiveDecimal("disburse_amount", disburseReq.DisburseAmount); err != nil {
		BadRequest(w, err.Error())
		return
	}

	loan, err := r.loanService.Disburse(disburseReq)
	if err != nil {
		requestID := GetRequestID(req)
		log.Printf("[request_id=%s] 放款失败: %v", requestID, err)
		InternalError(w, "服务器内部错误")
		return
	}

	Success(w, loan)
}

// =====================================================
// 还款计划
// =====================================================

// getRepaymentPlans 获取还款计划列表（支持分页）
func (r *Router) getRepaymentPlans(w http.ResponseWriter, req *http.Request) {
	vars := mux.Vars(req)
	loanNo := vars["loan_no"]

	// 分页参数
	page, _ := strconv.Atoi(req.URL.Query().Get("page"))
	size, _ := strconv.Atoi(req.URL.Query().Get("size"))
	if page < 1 {
		page = 1
	}
	if size < 1 || size > 100 {
		size = 20
	}

	plans, err := r.loanService.GetPlansByLoanNo(loanNo)
	if err != nil {
		requestID := GetRequestID(req)
		log.Printf("[request_id=%s] 查询还款计划失败: %v", requestID, err)
		InternalError(w, "服务器内部错误")
		return
	}

	// 分页处理
	total := len(plans)
	start := (page - 1) * size
	end := start + size
	if start > total {
		start = total
	}
	if end > total {
		end = total
	}

	SuccessWithPage(w, plans[start:end], total, page, size)
}

// getPlanSummary 获取还款计划汇总
func (r *Router) getPlanSummary(w http.ResponseWriter, req *http.Request) {
	vars := mux.Vars(req)
	loanNo := vars["loan_no"]

	summary, err := r.loanService.GetPlanSummary(loanNo)
	if err != nil {
		InternalError(w, "查询还款计划汇总失败: "+err.Error())
		return
	}

	Success(w, summary)
}

// =====================================================
// 还款
// =====================================================

// repaymentTrial 还款试算
func (r *Router) repaymentTrial(w http.ResponseWriter, req *http.Request) {
	var trialReq service.RepaymentTrialRequest
	if err := json.NewDecoder(req.Body).Decode(&trialReq); err != nil {
		BadRequest(w, "请求体格式错误: "+err.Error())
		return
	}

	if err := ValidateRequired(map[string]string{
		"loan_no":    trialReq.LoanNo,
		"trial_date": trialReq.TrialDate,
	}); err != nil {
		BadRequest(w, err.Error())
		return
	}

	response, err := r.loanService.RepaymentTrial(trialReq)
	if err != nil {
		InternalError(w, "还款试算失败: "+err.Error())
		return
	}

	Success(w, response)
}

// makeRepayment 还款入账
func (r *Router) makeRepayment(w http.ResponseWriter, req *http.Request) {
	var repayReq service.RepaymentRequest
	if err := json.NewDecoder(req.Body).Decode(&repayReq); err != nil {
		BadRequest(w, "请求体格式错误: "+err.Error())
		return
	}

	if err := ValidateRequired(map[string]string{
		"loan_no":      repayReq.LoanNo,
		"trial_date":   repayReq.TrialDate,
		"booking_date": repayReq.BookingDate,
	}); err != nil {
		BadRequest(w, err.Error())
		return
	}

	if err := ValidatePositiveDecimal("amount", repayReq.Amount); err != nil {
		BadRequest(w, err.Error())
		return
	}

	response, err := r.loanService.Repayment(repayReq)
	if err != nil {
		InternalError(w, "还款入账失败: "+err.Error())
		return
	}

	Created(w, response)
}

// =====================================================
// 提前结清
// =====================================================

// earlySettlementTrial 提前结清试算
func (r *Router) earlySettlementTrial(w http.ResponseWriter, req *http.Request) {
	vars := mux.Vars(req)
	loanNo := vars["loan_no"]

	var trialReq service.EarlySettlementTrialRequest
	if err := json.NewDecoder(req.Body).Decode(&trialReq); err != nil {
		BadRequest(w, "请求体格式错误: "+err.Error())
		return
	}

	trialReq.LoanNo = loanNo

	if err := ValidateRequired(map[string]string{
		"trial_date": trialReq.TrialDate,
	}); err != nil {
		BadRequest(w, err.Error())
		return
	}

	response, err := r.loanService.EarlySettlementTrial(trialReq)
	if err != nil {
		InternalError(w, "提前结清试算失败: "+err.Error())
		return
	}

	Success(w, response)
}

// earlySettlement 提前结清入账
func (r *Router) earlySettlement(w http.ResponseWriter, req *http.Request) {
	vars := mux.Vars(req)
	loanNo := vars["loan_no"]

	var settleReq service.EarlySettlementRequest
	if err := json.NewDecoder(req.Body).Decode(&settleReq); err != nil {
		BadRequest(w, "请求体格式错误: "+err.Error())
		return
	}

	settleReq.LoanNo = loanNo

	if err := ValidateRequired(map[string]string{
		"trial_date":   settleReq.TrialDate,
		"booking_date": settleReq.BookingDate,
	}); err != nil {
		BadRequest(w, err.Error())
		return
	}

	response, err := r.loanService.EarlySettlement(settleReq)
	if err != nil {
		InternalError(w, "提前结清失败: "+err.Error())
		return
	}

	Created(w, response)
}

// =====================================================
// 部分还款
// =====================================================

// partialRepaymentTrial 部分还款试算
func (r *Router) partialRepaymentTrial(w http.ResponseWriter, req *http.Request) {
	vars := mux.Vars(req)
	loanNo := vars["loan_no"]

	var trialReq service.PartialRepaymentTrialRequest
	if err := json.NewDecoder(req.Body).Decode(&trialReq); err != nil {
		BadRequest(w, "请求体格式错误: "+err.Error())
		return
	}

	trialReq.LoanNo = loanNo

	if err := ValidateRequired(map[string]string{
		"trial_date": trialReq.TrialDate,
	}); err != nil {
		BadRequest(w, err.Error())
		return
	}

	if err := ValidatePositiveDecimal("amount", trialReq.Amount); err != nil {
		BadRequest(w, err.Error())
		return
	}

	response, err := r.loanService.PartialRepaymentTrial(trialReq)
	if err != nil {
		InternalError(w, "部分还款试算失败: "+err.Error())
		return
	}

	Success(w, response)
}

// partialRepayment 部分还款入账
func (r *Router) partialRepayment(w http.ResponseWriter, req *http.Request) {
	vars := mux.Vars(req)
	loanNo := vars["loan_no"]

	var repayReq service.RepaymentRequest
	if err := json.NewDecoder(req.Body).Decode(&repayReq); err != nil {
		BadRequest(w, "请求体格式错误: "+err.Error())
		return
	}

	repayReq.LoanNo = loanNo

	if err := ValidateRequired(map[string]string{
		"trial_date":   repayReq.TrialDate,
		"booking_date": repayReq.BookingDate,
	}); err != nil {
		BadRequest(w, err.Error())
		return
	}

	if err := ValidatePositiveDecimal("amount", repayReq.Amount); err != nil {
		BadRequest(w, err.Error())
		return
	}

	response, err := r.loanService.PartialRepayment(repayReq)
	if err != nil {
		InternalError(w, "部分还款入账失败: "+err.Error())
		return
	}

	Created(w, response)
}
