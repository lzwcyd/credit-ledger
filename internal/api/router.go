package api

import (
	"net/http"
	"time"

	"github.com/gorilla/mux"
	"github.com/yourorg/credit-ledger/pkg/database"
	"github.com/yourorg/credit-ledger/internal/service"
	"github.com/yourorg/credit-ledger/internal/repository"
)

var startTime = time.Now()

type Router struct {
	mux         *mux.Router
	db          *database.DB
	loanService *service.LoanService
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
	// 注册中间件（顺序：外层先执行）
	r.mux.Use(RequestIDMiddleware)
	r.mux.Use(CORSMiddleware)
	r.mux.Use(LoggingMiddleware)

	// 健康检查
	r.mux.HandleFunc("/health", r.healthCheck).Methods("GET")
	r.mux.HandleFunc("/ready", r.readinessCheck).Methods("GET")

	// API v1 路由组
	apiV1 := r.mux.PathPrefix("/api/v1").Subrouter()

	// 借据相关
	apiV1.HandleFunc("/loans", r.createLoan).Methods("POST")
	apiV1.HandleFunc("/loans/{loan_no}", r.getLoan).Methods("GET")
	apiV1.HandleFunc("/loans/{loan_no}/disburse", r.disburseLoan).Methods("POST")
	apiV1.HandleFunc("/loans/{loan_no}/plans", r.getRepaymentPlans).Methods("GET")
	apiV1.HandleFunc("/loans/{loan_no}/plans/summary", r.getPlanSummary).Methods("GET")

	// 还款相关
	apiV1.HandleFunc("/repayments/trial", r.repaymentTrial).Methods("POST")
	apiV1.HandleFunc("/repayments", r.makeRepayment).Methods("POST")

	// 提前结清
	apiV1.HandleFunc("/loans/{loan_no}/early-settlement/trial", r.earlySettlementTrial).Methods("POST")
	apiV1.HandleFunc("/loans/{loan_no}/early-settlement", r.earlySettlement).Methods("POST")

	// 部分还款
	apiV1.HandleFunc("/loans/{loan_no}/partial-repayment/trial", r.partialRepaymentTrial).Methods("POST")
	apiV1.HandleFunc("/loans/{loan_no}/partial-repayment", r.partialRepayment).Methods("POST")

	// 信贷PM核心功能
	apiV1.HandleFunc("/loans/{loan_no}/collection-status", r.updateCollectionStatus).Methods("PUT")
	apiV1.HandleFunc("/loans/{loan_no}/penalty-waiver", r.applyPenaltyWaiver).Methods("POST")
	apiV1.HandleFunc("/loans/{loan_no}/extension", r.applyExtension).Methods("POST")
	apiV1.HandleFunc("/loans/{loan_no}/write-off", r.applyWriteOff).Methods("POST")
	apiV1.HandleFunc("/loans/{loan_no}/statement", r.generateStatement).Methods("GET")
	apiV1.HandleFunc("/repayments/upcoming", r.getUpcomingDuePlans).Methods("GET")
}

func (r *Router) ServeHTTP(w http.ResponseWriter, req *http.Request) {
	r.mux.ServeHTTP(w, req)
}

// healthCheck 健康检查（详细版）
func (r *Router) healthCheck(w http.ResponseWriter, req *http.Request) {
	uptime := time.Since(startTime).Seconds()

	// 检查数据库连接
	dbStatus := "connected"
	if err := r.db.DB.Ping(); err != nil {
		dbStatus = "disconnected: " + err.Error()
	}

	Success(w, map[string]interface{}{
		"status":    "healthy",
		"version":   "1.0.0",
		"uptime_s":  int(uptime),
		"db_status": dbStatus,
	})
}

// readinessCheck 就绪检查（用于 K8s readiness probe）
func (r *Router) readinessCheck(w http.ResponseWriter, req *http.Request) {
	if err := r.db.DB.Ping(); err != nil {
		Error(w, http.StatusServiceUnavailable, 503, "服务未就绪: 数据库连接失败")
		return
	}
	Success(w, map[string]string{"status": "ready"})
}
