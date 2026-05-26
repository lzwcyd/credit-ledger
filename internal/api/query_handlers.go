package api

import (
	"encoding/json"
	"net/http"

	"github.com/gorilla/mux"
)

// getLoanDetails 获取贷款详情
func (r *Router) getLoanDetails(w http.ResponseWriter, req *http.Request) {
	vars := mux.Vars(req)
	loanID := vars["id"]

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"loan_id": loanID,
		"details": map[string]interface{}{},
		"status":  "success",
	})
}

// getTransactions 获取交易记录
func (r *Router) getTransactions(w http.ResponseWriter, req *http.Request) {
	vars := mux.Vars(req)
	loanID := vars["id"]

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"loan_id":     loanID,
		"transactions": []interface{}{},
		"status":      "success",
	})
}