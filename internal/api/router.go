package api

import (
	"net/http"

	"github.com/gorilla/mux"
	"github.com/yourorg/credit-ledger/pkg/database"
	"github.com/yourorg/credit-ledger/pkg/cache"
	"github.com/yourorg/credit-ledger/internal/service"
)

type Router struct {
	mux         *mux.Router
	db          *database.DB
	redis       *cache.Redis
	loanService *service.LoanService
}

func NewRouter(db *database.DB, redis *cache.Redis) *Router {
	r := &Router{
		mux:         mux.NewRouter(),
		db:          db,
		redis:       redis,
		loanService: service.NewLoanService(),
	}

	r.setupRoutes()
	return r
}

func (r *Router) setupRoutes() {
	// 健康检查
	r.mux.HandleFunc("/health", r.healthCheck).Methods("GET")

	// API v1 路由组
	apiV1 := r.mux.PathPrefix("/api/v1").Subrouter()

	// 贷款相关路由
	loans := apiV1.PathPrefix("/loans").Subrouter()
	loans.HandleFunc("", r.createLoan).Methods("POST")
	loans.HandleFunc("/{id}", r.getLoan).Methods("GET")
	loans.HandleFunc("/{id}/disburse", r.disburseLoan).Methods("POST")
	loans.HandleFunc("/{id}/repayment-schedules", r.getRepaymentSchedules).Methods("GET")

	// 还款相关路由
	repayments := apiV1.PathPrefix("/repayments").Subrouter()
	repayments.HandleFunc("/trial", r.repaymentTrial).Methods("POST")
	repayments.HandleFunc("", r.makeRepayment).Methods("POST")

	// 查询相关路由
	queries := apiV1.PathPrefix("/queries").Subrouter()
	queries.HandleFunc("/loans/{id}/details", r.getLoanDetails).Methods("GET")
	queries.HandleFunc("/loans/{id}/transactions", r.getTransactions).Methods("GET")

	// 管理接口
	admin := apiV1.PathPrefix("/admin").Subrouter()
	admin.HandleFunc("/fee-configs", r.createFeeConfig).Methods("POST")
	admin.HandleFunc("/fee-configs", r.getFeeConfigs).Methods("GET")
}

func (r *Router) ServeHTTP(w http.ResponseWriter, req *http.Request) {
	r.mux.ServeHTTP(w, req)
}

func (r *Router) healthCheck(w http.ResponseWriter, req *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	w.Write([]byte(`{"status": "healthy"}`))
}