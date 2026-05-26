package api

import (
	"encoding/json"
	"net/http"
	"time"
)

// APIResponse 统一 API 响应格式
type APIResponse struct {
	Code      int         `json:"code"`
	Message   string      `json:"message"`
	Data      interface{} `json:"data,omitempty"`
	RequestID string      `json:"request_id,omitempty"`
	Timestamp int64       `json:"timestamp"`
}

// PaginatedData 分页数据包装
type PaginatedData struct {
	Items interface{} `json:"items"`
	Total int         `json:"total"`
	Page  int         `json:"page"`
	Size  int         `json:"size"`
}

// Success 成功响应
func Success(w http.ResponseWriter, data interface{}) {
	writeJSON(w, http.StatusOK, APIResponse{
		Code:      0,
		Message:   "success",
		Data:      data,
		Timestamp: time.Now().Unix(),
	})
}

// SuccessWithPage 分页成功响应
func SuccessWithPage(w http.ResponseWriter, items interface{}, total, page, size int) {
	writeJSON(w, http.StatusOK, APIResponse{
		Code:    0,
		Message: "success",
		Data: PaginatedData{
			Items: items,
			Total: total,
			Page:  page,
			Size:  size,
		},
		Timestamp: time.Now().Unix(),
	})
}

// Created 创建成功响应
func Created(w http.ResponseWriter, data interface{}) {
	writeJSON(w, http.StatusCreated, APIResponse{
		Code:      0,
		Message:   "created",
		Data:      data,
		Timestamp: time.Now().Unix(),
	})
}

// Error 错误响应
func Error(w http.ResponseWriter, httpStatus int, code int, message string) {
	writeJSON(w, httpStatus, APIResponse{
		Code:      code,
		Message:   message,
		Timestamp: time.Now().Unix(),
	})
}

// BadRequest 400
func BadRequest(w http.ResponseWriter, message string) {
	Error(w, http.StatusBadRequest, 400, message)
}

// NotFound 404
func NotFound(w http.ResponseWriter, message string) {
	Error(w, http.StatusNotFound, 404, message)
}

// InternalError 500
func InternalError(w http.ResponseWriter, message string) {
	Error(w, http.StatusInternalServerError, 500, message)
}

// Conflict 409
func Conflict(w http.ResponseWriter, message string) {
	Error(w, http.StatusConflict, 409, message)
}

func writeJSON(w http.ResponseWriter, status int, data interface{}) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(data)
}
