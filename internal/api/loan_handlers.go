package api

import (
	"encoding/json"
	"net/http"

	"github.com/gorilla/mux"
	"github.com/yourorg/credit-ledger/internal/service"
)

// createLoan 创建借据
func (r *Router) createLoan(w http.ResponseWriter, req *http.Request) {
	var createReq service.CreateLoanRequest
	if err := json.NewDecoder(req.Body).Decode(&createReq); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}
	
	loan, err := r.loanService.CreateLoan(createReq)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(loan)
}

// getLoan 获取借据详情
func (r *Router) getLoan(w http.ResponseWriter, req *http.Request) {
	vars := mux.Vars(req)
	loanNo := vars["loan_no"]
	
	loan, err := r.loanService.GetLoanByNo(loanNo)
	if err != nil {
		http.Error(w, "Loan not found", http.StatusNotFound)
		return
	}
	
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(loan)
}

// disburseLoan 放款
func (r *Router) disburseLoan(w http.ResponseWriter, req *http.Request) {
	vars := mux.Vars(req)
	loanNo := vars["loan_no"]
	
	var disburseReq service.DisburseRequest
	if err := json.NewDecoder(req.Body).Decode(&disburseReq); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}
	
	disburseReq.LoanNo = loanNo
	loan, err := r.loanService.Disburse(disburseReq)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(loan)
}

// getRepaymentPlans 获取还款计划
func (r *Router) getRepaymentPlans(w http.ResponseWriter, req *http.Request) {
	vars := mux.Vars(req)
	loanNo := vars["loan_no"]
	
	plans, err := r.loanService.GetPlansByLoanNo(loanNo)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(plans)
}

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
	var repayReq service.RepaymentRequest
	if err := json.NewDecoder(req.Body).Decode(&repayReq); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	response, err := r.loanService.Repayment(repayReq)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(response)
}

// earlySettlementTrial 提前结清试算
func (r *Router) earlySettlementTrial(w http.ResponseWriter, req *http.Request) {
	vars := mux.Vars(req)
	loanNo := vars["loan_no"]

	var trialReq service.EarlySettlementTrialRequest
	if err := json.NewDecoder(req.Body).Decode(&trialReq); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	trialReq.LoanNo = loanNo
	response, err := r.loanService.EarlySettlementTrial(trialReq)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

// earlySettlement 提前结清入账
func (r *Router) earlySettlement(w http.ResponseWriter, req *http.Request) {
	vars := mux.Vars(req)
	loanNo := vars["loan_no"]

	var settleReq service.EarlySettlementRequest
	if err := json.NewDecoder(req.Body).Decode(&settleReq); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	settleReq.LoanNo = loanNo
	response, err := r.loanService.EarlySettlement(settleReq)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(response)
}

// partialRepaymentTrial 部分还款试算
func (r *Router) partialRepaymentTrial(w http.ResponseWriter, req *http.Request) {
	vars := mux.Vars(req)
	loanNo := vars["loan_no"]

	var trialReq service.PartialRepaymentTrialRequest
	if err := json.NewDecoder(req.Body).Decode(&trialReq); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	trialReq.LoanNo = loanNo
	response, err := r.loanService.PartialRepaymentTrial(trialReq)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

// partialRepayment 部分还款入账
func (r *Router) partialRepayment(w http.ResponseWriter, req *http.Request) {
	vars := mux.Vars(req)
	loanNo := vars["loan_no"]

	var repayReq service.RepaymentRequest
	if err := json.NewDecoder(req.Body).Decode(&repayReq); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	repayReq.LoanNo = loanNo
	response, err := r.loanService.PartialRepayment(repayReq)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(response)
}

// getPlanSummary 获取还款计划汇总
func (r *Router) getPlanSummary(w http.ResponseWriter, req *http.Request) {
	vars := mux.Vars(req)
	loanNo := vars["loan_no"]

	summary, err := r.loanService.GetPlanSummary(loanNo)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(summary)
}