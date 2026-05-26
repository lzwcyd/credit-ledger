package api

import (
	"encoding/json"
	"net/http"

	"github.com/yourorg/credit-ledger/internal/service"
)

// repaymentTrial 还款试算
func (r *Router) repaymentTrial(w http.ResponseWriter, req *http.Request) {
	var trialReq service.RepaymentTrialRequest
	if err := json.NewDecoder(req.Body).Decode(&trialReq); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}
	
	response, err := r.loanService.RepaymentTrial(trialReq)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

// makeRepayment 还款入账
func (r *Router) makeRepayment(w http.ResponseWriter, req *http.Request) {
	var repaymentReq service.RepaymentTrialRequest
	if err := json.NewDecoder(req.Body).Decode(&repaymentReq); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}
	
	// 这里应该调用 service 处理还款入账
	// 简化示例
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"message": "还款成功",
		"status":  "success",
	})
}