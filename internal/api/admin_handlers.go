package api

import (
	"encoding/json"
	"net/http"
)

// createFeeConfig 创建费用配置
func (r *Router) createFeeConfig(w http.ResponseWriter, req *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(map[string]interface{}{
		"message": "费用配置创建成功",
		"status":  "success",
	})
}

// getFeeConfigs 获取费用配置列表
func (r *Router) getFeeConfigs(w http.ResponseWriter, req *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"fee_configs": []interface{}{},
		"status":      "success",
	})
}