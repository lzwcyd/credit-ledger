-- 创建数据库
CREATE DATABASE IF NOT EXISTS credit_ledger CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci;

USE credit_ledger;

-- 还款类型表
CREATE TABLE repayment_types (
    id BIGINT UNSIGNED AUTO_INCREMENT PRIMARY KEY,
    code VARCHAR(50) NOT NULL UNIQUE,
    name VARCHAR(100) NOT NULL,
    description TEXT,
    is_active BOOLEAN DEFAULT TRUE,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP
);

-- 插入默认还款类型
INSERT INTO repayment_types (code, name, description) VALUES
('EQUAL_INSTALLMENT', '等额本息', '每月还款金额固定，包含本金和利息'),
('EQUAL_PRINCIPAL', '等额本金', '每月偿还固定本金，利息随本金减少而递减'),
('INTEREST_ONLY', '按月付息到期还本', '每月只付利息，到期一次性偿还本金'),
('BULLET', '一次性还本付息', '到期一次性偿还本金和利息');

-- 费用配置表
CREATE TABLE fee_configs (
    id BIGINT UNSIGNED AUTO_INCREMENT PRIMARY KEY,
    code VARCHAR(50) NOT NULL UNIQUE,
    name VARCHAR(100) NOT NULL,
    fee_type ENUM('FIXED', 'PERCENTAGE', 'RATE') NOT NULL,
    calculation_base ENUM('PRINCIPAL', 'INTEREST', 'TOTAL') NOT NULL,
    value DECIMAL(10,6) NOT NULL,
    min_amount DECIMAL(15,2) DEFAULT 0,
    max_amount DECIMAL(15,2) DEFAULT NULL,
    is_active BOOLEAN DEFAULT TRUE,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP
);

-- 贷款主表
CREATE TABLE loans (
    id BIGINT UNSIGNED AUTO_INCREMENT PRIMARY KEY,
    loan_no VARCHAR(50) NOT NULL UNIQUE,
    principal DECIMAL(15,2) NOT NULL,
    annual_interest_rate DECIMAL(8,6) NOT NULL,
    term_months INT NOT NULL,
    repayment_type_id BIGINT UNSIGNED NOT NULL,
    disbursement_date DATE,
    first_repayment_date DATE,
    maturity_date DATE,
    status ENUM('PENDING', 'DISBURSED', 'REPAID', 'OVERDUE', 'DEFAULTED') DEFAULT 'PENDING',
    disbursed_amount DECIMAL(15,2) DEFAULT 0,
    total_interest DECIMAL(15,2) DEFAULT 0,
    total_fees DECIMAL(15,2) DEFAULT 0,
    repaid_amount DECIMAL(15,2) DEFAULT 0,
    remaining_principal DECIMAL(15,2) DEFAULT 0,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    FOREIGN KEY (repayment_type_id) REFERENCES repayment_types(id)
);

-- 还款计划表
CREATE TABLE loan_schedules (
    id BIGINT UNSIGNED AUTO_INCREMENT PRIMARY KEY,
    loan_id BIGINT UNSIGNED NOT NULL,
    period INT NOT NULL,
    due_date DATE NOT NULL,
    principal_due DECIMAL(15,2) NOT NULL,
    interest_due DECIMAL(15,2) NOT NULL,
    total_due DECIMAL(15,2) NOT NULL,
    principal_paid DECIMAL(15,2) DEFAULT 0,
    interest_paid DECIMAL(15,2) DEFAULT 0,
    total_paid DECIMAL(15,2) DEFAULT 0,
    status ENUM('PENDING', 'PAID', 'OVERDUE', 'PARTIAL') DEFAULT 'PENDING',
    paid_date DATE,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    FOREIGN KEY (loan_id) REFERENCES loans(id),
    UNIQUE KEY uk_loan_period (loan_id, period)
);

-- 交易流水表
CREATE TABLE transactions (
    id BIGINT UNSIGNED AUTO_INCREMENT PRIMARY KEY,
    transaction_no VARCHAR(50) NOT NULL UNIQUE,
    loan_id BIGINT UNSIGNED NOT NULL,
    schedule_id BIGINT UNSIGNED,
    transaction_type ENUM('DISBURSEMENT', 'REPAYMENT', 'FEE', 'INTEREST', 'PENALTY', 'ADJUSTMENT') NOT NULL,
    amount DECIMAL(15,2) NOT NULL,
    principal_amount DECIMAL(15,2) DEFAULT 0,
    interest_amount DECIMAL(15,2) DEFAULT 0,
    fee_amount DECIMAL(15,2) DEFAULT 0,
    penalty_amount DECIMAL(15,2) DEFAULT 0,
    transaction_date DATE NOT NULL,
    description TEXT,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    FOREIGN KEY (loan_id) REFERENCES loans(id),
    FOREIGN KEY (schedule_id) REFERENCES loan_schedules(id)
);

-- 利息计算表
CREATE TABLE interest_calculations (
    id BIGINT UNSIGNED AUTO_INCREMENT PRIMARY KEY,
    loan_id BIGINT UNSIGNED NOT NULL,
    calculation_date DATE NOT NULL,
    principal_balance DECIMAL(15,2) NOT NULL,
    daily_interest_rate DECIMAL(12,10) NOT NULL,
    interest_amount DECIMAL(15,2) NOT NULL,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    FOREIGN KEY (loan_id) REFERENCES loans(id),
    INDEX idx_loan_date (loan_id, calculation_date)
);

-- 费用表
CREATE TABLE fees (
    id BIGINT UNSIGNED AUTO_INCREMENT PRIMARY KEY,
    loan_id BIGINT UNSIGNED NOT NULL,
    fee_config_id BIGINT UNSIGNED NOT NULL,
    amount DECIMAL(15,2) NOT NULL,
    status ENUM('PENDING', 'PAID', 'WAIVED') DEFAULT 'PENDING',
    due_date DATE,
    paid_date DATE,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    FOREIGN KEY (loan_id) REFERENCES loans(id),
    FOREIGN KEY (fee_config_id) REFERENCES fee_configs(id)
);

-- 创建索引
CREATE INDEX idx_loans_status ON loans(status);
CREATE INDEX idx_loans_loan_no ON loans(loan_no);
CREATE INDEX idx_loan_schedules_loan_id ON loan_schedules(loan_id);
CREATE INDEX idx_loan_schedules_due_date ON loan_schedules(due_date);
CREATE INDEX idx_loan_schedules_status ON loan_schedules(status);
CREATE INDEX idx_transactions_loan_id ON transactions(loan_id);
CREATE INDEX idx_transactions_date ON transactions(transaction_date);
CREATE INDEX idx_fees_loan_id ON fees(loan_id);
CREATE INDEX idx_fees_status ON fees(status);