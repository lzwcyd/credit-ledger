package api

import (
	"encoding/json"
	"net/http"

	"github.com/gorilla/mux"
	"github.com/yourorg/credit-ledger/internal/service"
)

// createCoupons 创建优惠券（支持批量）
func (r *Router) createCoupons(w http.ResponseWriter, req *http.Request) {
	var createReq service.CreateCouponRequest
	if err := json.NewDecoder(req.Body).Decode(&createReq); err != nil {
		BadRequest(w, "请求体格式错误: "+err.Error())
		return
	}

	if err := ValidateRequired(map[string]string{
		"coupon_type":   createReq.CouponType,
		"discount_type": createReq.DiscountType,
		"valid_from":    createReq.ValidFrom,
		"valid_to":      createReq.ValidTo,
	}); err != nil {
		BadRequest(w, err.Error())
		return
	}

	if err := ValidatePositiveFloat("face_value", createReq.FaceValue); err != nil {
		BadRequest(w, err.Error())
		return
	}

	coupons, err := r.loanService.CreateCoupon(createReq)
	if err != nil {
		InternalError(w, "创建优惠券失败: "+err.Error())
		return
	}

	Created(w, map[string]interface{}{
		"count":   len(coupons),
		"coupons": coupons,
	})
}

// couponTrial 优惠券试算
func (r *Router) couponTrial(w http.ResponseWriter, req *http.Request) {
	var trialReq service.CouponTrialRequest
	if err := json.NewDecoder(req.Body).Decode(&trialReq); err != nil {
		BadRequest(w, "请求体格式错误: "+err.Error())
		return
	}

	if err := ValidateRequired(map[string]string{
		"coupon_code": trialReq.CouponCode,
	}); err != nil {
		BadRequest(w, err.Error())
		return
	}

	resp, err := r.loanService.CouponTrial(trialReq)
	if err != nil {
		InternalError(w, "优惠券试算失败: "+err.Error())
		return
	}

	Success(w, resp)
}

// applyCoupon 使用优惠券
func (r *Router) applyCoupon(w http.ResponseWriter, req *http.Request) {
	vars := mux.Vars(req)
	loanNo := vars["loan_no"]

	var applyReq service.ApplyCouponRequest
	if err := json.NewDecoder(req.Body).Decode(&applyReq); err != nil {
		BadRequest(w, "请求体格式错误: "+err.Error())
		return
	}

	applyReq.LoanNo = loanNo

	if err := ValidateRequired(map[string]string{
		"coupon_code": applyReq.CouponCode,
	}); err != nil {
		BadRequest(w, err.Error())
		return
	}

	resp, err := r.loanService.ApplyCoupon(applyReq)
	if err != nil {
		InternalError(w, "使用优惠券失败: "+err.Error())
		return
	}

	Success(w, resp)
}
