package api

import (
	"net/http"

	"github.com/gorilla/mux"
	"github.com/yourorg/credit-ledger/pkg/database"
	"github.com/yourorg/credit-ledger/internal/service"
	"github.com/yourorg/credit-ledger/internal/repository"
)

type Router struct {
	mux          *mux.Router
	db           *database.DB
	loanService  *service.LoanService
}

func NewRouter(db *database.DB) *Router {
	// 初始化仓库
	loanRepo := repository.NewLoanRepository(db.DB)
	planRepo := repository.NewPlanRepository(db.DB)
	repaymentRepo := repository.NewRepaymentRepository(db.DB)
	dailyCalcRepo := repository.NewDailyCalculationRepository(db.DB)
	planOtherFeeRepo := repository.NewPlanOtherFeeRepository(db.DB)
	feeConfigRepo := repository.NewFeeConfigRepository(db.DB)
	allocationRepo := repository.NewAllocationRuleRepository(db.DB)
	
	// 初始化服务
	loanService := service.NewLoanService(
		loanRepo, planRepo, repaymentRepo, dailyCalcRepo,
		planOtherFeeRepo, feeConfigRepo, allocationRepo,
	)
	
	r := &Router{
		mux:         mux.NewRouter(),
		db:          db,
		loanService: loanService,
	}

	r.setupRoutes()
	return r
}

func (r *Router) setupRoutes() {
	// 健康检查
	r.mux.HandleFunc("/health", r.healthCheck).Methods("GET")

	// API v1 路由组
	apiV1 := r.mux.PathPrefix("/api/v1").Subrouter()

	// 借据相关
	apiV1.HandleFunc("/loans", r.createLoan).Methods("POST")
	apiV1.HandleFunc("/loans/{loan_no}", r.getLoan).Methods("GET")
	apiV1.HandleFunc("/loans/{loan_no}/disburse", r.disburseLoan).Methods("POST")
	apiV1.HandleFunc("/loans/{loan_no}/plans", r.getRepaymentPlans).Methods("GET")

	// 还款相关
	apiV1.HandleFunc("/repayments/trial", r.repaymentTrial).Methods("POST")
	apiV1.HandleFunc("/repayments", r.makeRepayment).Methods("POST")

	// 提前结清相关
	apiV1.HandleFunc("/loans/{loan_no}/early-settlement/trial", r.earlySettlementTrial).Methods("POST")
	apiV1.HandleFunc("/loans/{loan_no}/early-settlement", r.earlySettlement).Methods("POST")

	// 部分还款相关
	apiV1.HandleFunc("/loans/{loan_no}/partial-repayment/trial", r.partialRepaymentTrial).Methods("POST")
	apiV1.HandleFunc("/loans/{loan_no}/partial-repayment", r.partialRepayment).Methods("POST")

	// 还款计划汇总
	apiV1.HandleFunc("/loans/{loan_no}/plans/summary", r.getPlanSummary).Methods("GET")
}

func (r *Router) ServeHTTP(w http.ResponseWriter, req *http.Request) {
	r.mux.ServeHTTP(w, req)
}

func (r *Router) healthCheck(w http.ResponseWriter, req *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	w.Write([]byte(`{"status": "healthy"}`))
}