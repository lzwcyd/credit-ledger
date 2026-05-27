package database

import (
	"database/sql"
	"fmt"
	"time"

	_ "github.com/mattn/go-sqlite3"
	"github.com/lzwcyd/credit-ledger/config"
)

type DB struct {
	*sql.DB
}

func NewMySQL(cfg config.DatabaseConfig) (*DB, error) {
	// 开发环境使用 SQLite
	dbPath := "./credit_ledger.db"
	
	db, err := sql.Open("sqlite3", dbPath)
	if err != nil {
		return nil, fmt.Errorf("failed to open database: %w", err)
	}

	// 设置连接池参数
	db.SetMaxOpenConns(1) // SQLite 单连接
	db.SetMaxIdleConns(1)
	db.SetConnMaxLifetime(5 * time.Minute)
	db.SetConnMaxIdleTime(5 * time.Minute)

	// 测试连接
	if err := db.Ping(); err != nil {
		return nil, fmt.Errorf("failed to ping database: %w", err)
	}

	// 创建表
	if err := createTables(db); err != nil {
		return nil, fmt.Errorf("failed to create tables: %w", err)
	}

	return &DB{db}, nil
}

func createTables(db *sql.DB) error {
	// 还款类型表
	_, err := db.Exec(`
		CREATE TABLE IF NOT EXISTS repayment_types (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			code TEXT NOT NULL UNIQUE,
			name TEXT NOT NULL,
			description TEXT,
			is_active BOOLEAN DEFAULT 1,
			created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
			updated_at DATETIME DEFAULT CURRENT_TIMESTAMP
		)
	`)
	if err != nil {
		return err
	}

	// 插入默认还款类型
	_, err = db.Exec(`
		INSERT OR IGNORE INTO repayment_types (code, name, description) VALUES
		('EQUAL_INSTALLMENT', '等额本息', '每月还款金额固定，包含本金和利息'),
		('EQUAL_PRINCIPAL', '等额本金', '每月偿还固定本金，利息随本金减少而递减'),
		('INTEREST_ONLY', '按月付息到期还本', '每月只付利息，到期一次性偿还本金'),
		('BULLET', '一次性还本付息', '到期一次性偿还本金和利息')
	`)
	if err != nil {
		return err
	}

	// 费用配置表
	_, err = db.Exec(`
		CREATE TABLE IF NOT EXISTS fee_configs (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			code TEXT NOT NULL UNIQUE,
			name TEXT NOT NULL,
			fee_type TEXT NOT NULL CHECK(fee_type IN ('FIXED', 'PERCENTAGE', 'RATE')),
			calculation_base TEXT NOT NULL CHECK(calculation_base IN ('PRINCIPAL', 'INTEREST', 'TOTAL')),
			value REAL NOT NULL,
			min_amount REAL DEFAULT 0,
			max_amount REAL,
			is_active BOOLEAN DEFAULT 1,
			created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
			updated_at DATETIME DEFAULT CURRENT_TIMESTAMP
		)
	`)
	if err != nil {
		return err
	}

	// 贷款主表
	_, err = db.Exec(`
		CREATE TABLE IF NOT EXISTS loans (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			loan_no TEXT NOT NULL UNIQUE,
			principal REAL NOT NULL,
			annual_interest_rate REAL NOT NULL,
			term_months INTEGER NOT NULL,
			repayment_type_id INTEGER NOT NULL,
			disbursement_date DATE,
			first_repayment_date DATE,
			maturity_date DATE,
			status TEXT DEFAULT 'PENDING' CHECK(status IN ('PENDING', 'DISBURSED', 'REPAID', 'OVERDUE', 'DEFAULTED')),
			disbursed_amount REAL DEFAULT 0,
			total_interest REAL DEFAULT 0,
			total_fees REAL DEFAULT 0,
			repaid_amount REAL DEFAULT 0,
			remaining_principal REAL DEFAULT 0,
			created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
			updated_at DATETIME DEFAULT CURRENT_TIMESTAMP,
			FOREIGN KEY (repayment_type_id) REFERENCES repayment_types(id)
		)
	`)
	if err != nil {
		return err
	}

	// 还款计划表
	_, err = db.Exec(`
		CREATE TABLE IF NOT EXISTS loan_schedules (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			loan_id INTEGER NOT NULL,
			period INTEGER NOT NULL,
			due_date DATE NOT NULL,
			principal_due REAL NOT NULL,
			interest_due REAL NOT NULL,
			total_due REAL NOT NULL,
			principal_paid REAL DEFAULT 0,
			interest_paid REAL DEFAULT 0,
			total_paid REAL DEFAULT 0,
			status TEXT DEFAULT 'PENDING' CHECK(status IN ('PENDING', 'PAID', 'OVERDUE', 'PARTIAL')),
			paid_date DATE,
			created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
			updated_at DATETIME DEFAULT CURRENT_TIMESTAMP,
			FOREIGN KEY (loan_id) REFERENCES loans(id),
			UNIQUE(loan_id, period)
		)
	`)
	if err != nil {
		return err
	}

	// 交易流水表
	_, err = db.Exec(`
		CREATE TABLE IF NOT EXISTS transactions (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			transaction_no TEXT NOT NULL UNIQUE,
			loan_id INTEGER NOT NULL,
			schedule_id INTEGER,
			transaction_type TEXT NOT NULL CHECK(transaction_type IN ('DISBURSEMENT', 'REPAYMENT', 'FEE', 'INTEREST', 'PENALTY', 'ADJUSTMENT')),
			amount REAL NOT NULL,
			principal_amount REAL DEFAULT 0,
			interest_amount REAL DEFAULT 0,
			fee_amount REAL DEFAULT 0,
			penalty_amount REAL DEFAULT 0,
			transaction_date DATE NOT NULL,
			description TEXT,
			created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
			FOREIGN KEY (loan_id) REFERENCES loans(id),
			FOREIGN KEY (schedule_id) REFERENCES loan_schedules(id)
		)
	`)
	if err != nil {
		return err
	}

	// 利息计算表
	_, err = db.Exec(`
		CREATE TABLE IF NOT EXISTS interest_calculations (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			loan_id INTEGER NOT NULL,
			calculation_date DATE NOT NULL,
			principal_balance REAL NOT NULL,
			daily_interest_rate REAL NOT NULL,
			interest_amount REAL NOT NULL,
			created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
			FOREIGN KEY (loan_id) REFERENCES loans(id)
		)
	`)
	if err != nil {
		return err
	}

	// 费用表
	_, err = db.Exec(`
		CREATE TABLE IF NOT EXISTS fees (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			loan_id INTEGER NOT NULL,
			fee_config_id INTEGER NOT NULL,
			amount REAL NOT NULL,
			status TEXT DEFAULT 'PENDING' CHECK(status IN ('PENDING', 'PAID', 'WAIVED')),
			due_date DATE,
			paid_date DATE,
			created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
			updated_at DATETIME DEFAULT CURRENT_TIMESTAMP,
			FOREIGN KEY (loan_id) REFERENCES loans(id),
			FOREIGN KEY (fee_config_id) REFERENCES fee_configs(id)
		)
	`)
	if err != nil {
		return err
	}

	// 创建索引
	_, err = db.Exec(`CREATE INDEX IF NOT EXISTS idx_loans_status ON loans(status)`)
	if err != nil {
		return err
	}

	_, err = db.Exec(`CREATE INDEX IF NOT EXISTS idx_loans_loan_no ON loans(loan_no)`)
	if err != nil {
		return err
	}

	_, err = db.Exec(`CREATE INDEX IF NOT EXISTS idx_loan_schedules_loan_id ON loan_schedules(loan_id)`)
	if err != nil {
		return err
	}

	_, err = db.Exec(`CREATE INDEX IF NOT EXISTS idx_loan_schedules_due_date ON loan_schedules(due_date)`)
	if err != nil {
		return err
	}

	_, err = db.Exec(`CREATE INDEX IF NOT EXISTS idx_loan_schedules_status ON loan_schedules(status)`)
	if err != nil {
		return err
	}

	_, err = db.Exec(`CREATE INDEX IF NOT EXISTS idx_transactions_loan_id ON transactions(loan_id)`)
	if err != nil {
		return err
	}

	_, err = db.Exec(`CREATE INDEX IF NOT EXISTS idx_transactions_date ON transactions(transaction_date)`)
	if err != nil {
		return err
	}

	_, err = db.Exec(`CREATE INDEX IF NOT EXISTS idx_fees_loan_id ON fees(loan_id)`)
	if err != nil {
		return err
	}

	_, err = db.Exec(`CREATE INDEX IF NOT EXISTS idx_fees_status ON fees(status)`)
	if err != nil {
		return err
	}

	return nil
}

func (db *DB) Close() error {
	return db.DB.Close()
}