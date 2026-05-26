package api

import (
	"encoding/json"
	"net/http"
	"strconv"

	"github.com/gorilla/mux"
	"github.com/yourorg/credit-ledger/internal/service"
)

// createLoan 创建贷款
func (r *Router) createLoan(w http.ResponseWriter, req *http.Request) {
	var createReq service.CreateLoanRequest
	if err := json.NewDecoder(req.Body).Decode(&createReq); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}
	
	response, err := r.loanService.CreateLoan(createReq)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(response)
}

// getLoan 获取贷款详情
func (r *Router) getLoan(w http.ResponseWriter, req *http.Request) {
	vars := mux.Vars(req)
	loanIDStr := vars["id"]
	
	loanID, err := strconv.ParseUint(loanIDStr, 10, 64)
	if err != nil {
		http.Error(w, "Invalid loan ID", http.StatusBadRequest)
		return
	}
	
	// 这里应该调用 service 获取贷款详情
	// 简化示例
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"loan_id": loanID,
		"status":  "success",
	})
}

// disburseLoan 放款
func (r *Router) disburseLoan(w http.ResponseWriter, req *http.Request) {
	vars := mux.Vars(req)
	loanIDStr := vars["id"]
	
	loanID, err := strconv.ParseUint(loanIDStr, 10, 64)
	if err != nil {
		http.Error(w, "Invalid loan ID", http.StatusBadRequest)
		return
	}
	
	var disburseReq service.DisburseLoanRequest
	if err := json.NewDecoder(req.Body).Decode(&disburseReq); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}
	
	disburseReq.LoanID = loanID
	response, err := r.loanService.DisburseLoan(disburseReq)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

// getRepaymentSchedules 获取还款计划
func (r *Router) getRepaymentSchedules(w http.ResponseWriter, req *http.Request) {
	vars := mux.Vars(req)
	loanIDStr := vars["id"]
	
	loanID, err := strconv.ParseUint(loanIDStr, 10, 64)
	if err != nil {
		http.Error(w, "Invalid loan ID", http.StatusBadRequest)
		return
	}
	
	// 这里应该调用 service 获取还款计划
	// 简化示例
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"loan_id":  loanID,
		"schedules": []interface{}{},
		"status":   "success",
	})
}