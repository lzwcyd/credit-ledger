package api

import (
	"encoding/json"
	"net/http"
	"strconv"

	"github.com/gorilla/mux"
	"github.com/lzwcyd/credit-ledger/internal/service"
)

// =====================================================
// 信贷PM核心功能 API
// =====================================================

// updateCollectionStatus 更新催收状态
func (r *Router) updateCollectionStatus(w http.ResponseWriter, req *http.Request) {
	vars := mux.Vars(req)
	loanNo := vars["loan_no"]

	var updateReq service.UpdateCollectionStatusRequest
	if err := json.NewDecoder(req.Body).Decode(&updateReq); err != nil {
		BadRequest(w, "请求体格式错误: "+err.Error())
		return
	}

	updateReq.LoanNo = loanNo

	if err := ValidateRequired(map[string]string{
		"status": updateReq.Status,
	}); err != nil {
		BadRequest(w, err.Error())
		return
	}

	loan, err := r.loanService.UpdateCollectionStatus(updateReq)
	if err != nil {
		InternalError(w, "更新催收状态失败: "+err.Error())
		return
	}

	Success(w, loan)
}

// applyPenaltyWaiver 申请罚息减免
func (r *Router) applyPenaltyWaiver(w http.ResponseWriter, req *http.Request) {
	vars := mux.Vars(req)
	loanNo := vars["loan_no"]

	var waiverReq service.PenaltyWaiverRequest
	if err := json.NewDecoder(req.Body).Decode(&waiverReq); err != nil {
		BadRequest(w, "请求体格式错误: "+err.Error())
		return
	}

	waiverReq.LoanNo = loanNo

	if err := ValidateRequired(map[string]string{
		"waiver_type": waiverReq.WaiverType,
		"reason":      waiverReq.Reason,
	}); err != nil {
		BadRequest(w, err.Error())
		return
	}

	if err := ValidatePositiveDecimal("waiver_amount", waiverReq.WaiverAmount); err != nil {
		BadRequest(w, err.Error())
		return
	}

	resp, err := r.loanService.ApplyPenaltyWaiver(waiverReq)
	if err != nil {
		InternalError(w, "申请罚息减免失败: "+err.Error())
		return
	}

	Created(w, resp)
}

// applyExtension 申请展期
func (r *Router) applyExtension(w http.ResponseWriter, req *http.Request) {
	vars := mux.Vars(req)
	loanNo := vars["loan_no"]

	var extReq service.ExtensionRequest
	if err := json.NewDecoder(req.Body).Decode(&extReq); err != nil {
		BadRequest(w, "请求体格式错误: "+err.Error())
		return
	}

	extReq.LoanNo = loanNo

	if extReq.ExtensionDays <= 0 && extReq.ExtensionMonths <= 0 {
		BadRequest(w, "展期天数或月数必须大于0")
		return
	}

	resp, err := r.loanService.ApplyExtension(extReq)
	if err != nil {
		InternalError(w, "申请展期失败: "+err.Error())
		return
	}

	Created(w, resp)
}

// applyWriteOff 申请坏账核销
func (r *Router) applyWriteOff(w http.ResponseWriter, req *http.Request) {
	vars := mux.Vars(req)
	loanNo := vars["loan_no"]

	var woReq service.WriteOffRequest
	if err := json.NewDecoder(req.Body).Decode(&woReq); err != nil {
		BadRequest(w, "请求体格式错误: "+err.Error())
		return
	}

	woReq.LoanNo = loanNo

	if err := ValidateRequired(map[string]string{
		"reason": woReq.Reason,
	}); err != nil {
		BadRequest(w, err.Error())
		return
	}

	resp, err := r.loanService.ApplyWriteOff(woReq)
	if err != nil {
		InternalError(w, "申请坏账核销失败: "+err.Error())
		return
	}

	Created(w, resp)
}

// getUpcomingDuePlans 获取即将到期的还款计划
func (r *Router) getUpcomingDuePlans(w http.ResponseWriter, req *http.Request) {
	daysStr := req.URL.Query().Get("days")
	days, _ := strconv.Atoi(daysStr)
	if days <= 0 {
		days = 3
	}

	plans, err := r.loanService.GetUpcomingDuePlans(days)
	if err != nil {
		InternalError(w, "查询到期计划失败: "+err.Error())
		return
	}

	Success(w, map[string]interface{}{
		"days":  days,
		"plans": plans,
	})
}

// generateStatement 生成客户对账单
func (r *Router) generateStatement(w http.ResponseWriter, req *http.Request) {
	vars := mux.Vars(req)
	loanNo := vars["loan_no"]

	statement, err := r.loanService.GenerateStatement(loanNo)
	if err != nil {
		InternalError(w, "生成对账单失败: "+err.Error())
		return
	}

	Success(w, statement)
}
