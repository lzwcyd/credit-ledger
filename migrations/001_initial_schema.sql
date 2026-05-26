-- 创建数据库
CREATE DATABASE IF NOT EXISTS credit_ledger CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci;

USE credit_ledger;

-- =====================================================
-- 基础配置表
-- =====================================================

-- 还款类型表
CREATE TABLE repayment_types (
    id BIGINT UNSIGNED AUTO_INCREMENT PRIMARY KEY,
    code VARCHAR(50) NOT NULL UNIQUE,
    name VARCHAR(100) NOT NULL,
    description TEXT,
    is_active BOOLEAN DEFAULT TRUE,
    created_by VARCHAR(64) NOT NULL DEFAULT 'system',
    updated_by VARCHAR(64) NOT NULL DEFAULT 'system',
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP
);

INSERT INTO repayment_types (code, name, description, created_by, updated_by) VALUES
('EQUAL_INSTALLMENT', '等额本息', '每月还款金额固定，包含本金和利息', 'system', 'system'),
('EQUAL_PRINCIPAL', '等额本金', '每月偿还固定本金，利息随本金减少而递减', 'system', 'system'),
('INTEREST_ONLY', '按月付息到期还本', '每月只付利息，到期一次性偿还本金', 'system', 'system'),
('BULLET', '一次性还本付息', '到期一次性偿还本金和利息', 'system', 'system');

-- 费项配置
CREATE TABLE fee_configs (
    id BIGINT UNSIGNED AUTO_INCREMENT PRIMARY KEY,
    code VARCHAR(50) NOT NULL UNIQUE,
    name VARCHAR(100) NOT NULL,
    
    calc_type ENUM('FIXED', 'PERCENTAGE', 'DAILY_RATE') NOT NULL,
    calc_base ENUM('PRINCIPAL', 'INTEREST', 'TOTAL', 'REMAINING_PRINCIPAL', 'OVERDUE_AMOUNT') NOT NULL,
    value DECIMAL(10,6) NOT NULL,
    
    trigger_type ENUM('DISBURSEMENT', 'REPAYMENT', 'DAILY', 'OVERDUE', 'EARLY_REPAYMENT') NOT NULL,
    is_daily_accumulate BOOLEAN DEFAULT FALSE,
    
    fee_category ENUM('INTEREST', 'PENALTY', 'OTHER_FEE') NOT NULL,
    
    min_amount DECIMAL(15,2) DEFAULT 0,
    max_amount DECIMAL(15,2) DEFAULT NULL,
    
    is_active BOOLEAN DEFAULT TRUE,
    created_by VARCHAR(64) NOT NULL DEFAULT 'system',
    updated_by VARCHAR(64) NOT NULL DEFAULT 'system',
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP
);

INSERT INTO fee_configs (code, name, calc_type, calc_base, value, trigger_type, is_daily_accumulate, fee_category, created_by, updated_by) VALUES
('INTEREST',        '利息',           'DAILY_RATE', 'REMAINING_PRINCIPAL', 0.000138889, 'DAILY',           TRUE,  'INTEREST',  'system', 'system'),
('OVERDUE_PENALTY', '逾期罚息',       'DAILY_RATE', 'OVERDUE_AMOUNT',      0.0005,     'OVERDUE',         TRUE,  'PENALTY',   'system', 'system'),
('HANDLING_FEE',    '手续费',         'PERCENTAGE', 'PRINCIPAL',           0.01,       'DISBURSEMENT',    FALSE, 'OTHER_FEE', 'system', 'system'),
('MANAGEMENT_FEE',  '管理费',         'PERCENTAGE', 'PRINCIPAL',           0.005,      'DISBURSEMENT',    FALSE, 'OTHER_FEE', 'system', 'system'),
('SERVICE_FEE',     '服务费',         'PERCENTAGE', 'PRINCIPAL',           0.003,      'DISBURSEMENT',    FALSE, 'OTHER_FEE', 'system', 'system'),
('RISK_FEE',        '风险金',         'PERCENTAGE', 'PRINCIPAL',           0.002,      'DISBURSEMENT',    FALSE, 'OTHER_FEE', 'system', 'system'),
('EARLY_REPAY_FEE', '提前还款手续费', 'PERCENTAGE', 'REMAINING_PRINCIPAL', 0.01,       'EARLY_REPAYMENT', FALSE, 'OTHER_FEE', 'system', 'system');

-- 分配规则
CREATE TABLE allocation_rules (
    id BIGINT UNSIGNED AUTO_INCREMENT PRIMARY KEY,
    code VARCHAR(50) NOT NULL UNIQUE,
    name VARCHAR(100) NOT NULL,
    description TEXT,
    is_default BOOLEAN DEFAULT FALSE,
    is_active BOOLEAN DEFAULT TRUE,
    created_by VARCHAR(64) NOT NULL DEFAULT 'system',
    updated_by VARCHAR(64) NOT NULL DEFAULT 'system',
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP
);

CREATE TABLE allocation_rule_items (
    id BIGINT UNSIGNED AUTO_INCREMENT PRIMARY KEY,
    rule_code VARCHAR(50) NOT NULL,
    priority INT NOT NULL,
    allocation_type ENUM('PENALTY', 'OTHER_FEE', 'INTEREST', 'PRINCIPAL') NOT NULL,
    description VARCHAR(200),
    created_by VARCHAR(64) NOT NULL DEFAULT 'system',
    updated_by VARCHAR(64) NOT NULL DEFAULT 'system',
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    UNIQUE KEY uk_rule_priority (rule_code, priority)
);

INSERT INTO allocation_rules (code, name, description, is_default, created_by, updated_by) VALUES
('DEFAULT', '默认规则', '罚息->其他费用->利息->本金', TRUE, 'system', 'system');

INSERT INTO allocation_rule_items (rule_code, priority, allocation_type, description, created_by, updated_by) VALUES
('DEFAULT', 1, 'PENALTY',    '先还罚息',    'system', 'system'),
('DEFAULT', 2, 'OTHER_FEE',  '再还其他费用', 'system', 'system'),
('DEFAULT', 3, 'INTEREST',   '再还利息',    'system', 'system'),
('DEFAULT', 4, 'PRINCIPAL',  '最后还本金',  'system', 'system');

-- =====================================================
-- 借据主表
-- =====================================================
CREATE TABLE loans (
    id BIGINT UNSIGNED AUTO_INCREMENT PRIMARY KEY,
    loan_no VARCHAR(50) NOT NULL UNIQUE,
    
    principal DECIMAL(15,2) NOT NULL,
    annual_rate DECIMAL(8,6) NOT NULL,
    term_months INT NOT NULL,
    repayment_type_code VARCHAR(50) NOT NULL,
    allocation_rule_code VARCHAR(50) NOT NULL DEFAULT 'DEFAULT',
    
    -- 关键日期
    value_date DATE NOT NULL COMMENT '起息日',
    first_due_date DATE NOT NULL COMMENT '首个还款日',
    maturity_date DATE NOT NULL COMMENT '到期日',
    settlement_date DATE COMMENT '结清日期',
    disbursement_date DATE COMMENT '放款日期',
    
    status ENUM('PENDING', 'DISBURSED', 'REPAID', 'OVERDUE', 'DEFAULTED') DEFAULT 'PENDING',
    
    disbursed_amount DECIMAL(15,2) DEFAULT 0,
    remaining_principal DECIMAL(15,2) DEFAULT 0,
    
    total_interest DECIMAL(15,2) DEFAULT 0,
    total_penalty DECIMAL(15,2) DEFAULT 0,
    total_other_fee DECIMAL(15,2) DEFAULT 0,
    
    paid_principal DECIMAL(15,2) DEFAULT 0,
    paid_interest DECIMAL(15,2) DEFAULT 0,
    paid_penalty DECIMAL(15,2) DEFAULT 0,
    paid_other_fee DECIMAL(15,2) DEFAULT 0,
    
    overdue_days INT DEFAULT 0,
    overdue_principal DECIMAL(15,2) DEFAULT 0,
    
    created_by VARCHAR(64) NOT NULL DEFAULT 'system',
    updated_by VARCHAR(64) NOT NULL DEFAULT 'system',
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    
    INDEX idx_loan_no (loan_no),
    INDEX idx_status (status)
);

-- 借据变更表
CREATE TABLE loan_changes (
    id BIGINT UNSIGNED AUTO_INCREMENT PRIMARY KEY,
    loan_no VARCHAR(50) NOT NULL,
    change_type ENUM('DISBURSE', 'REPAYMENT', 'OVERDUE', 'EARLY_SETTLEMENT', 'STATUS_CHANGE', 'ADJUSTMENT') NOT NULL,
    field_name VARCHAR(100) NOT NULL,
    old_value TEXT,
    new_value TEXT,
    change_reason TEXT,
    related_repayment_no VARCHAR(50),
    batch_no VARCHAR(50),
    created_by VARCHAR(64) NOT NULL DEFAULT 'system',
    updated_by VARCHAR(64) NOT NULL DEFAULT 'system',
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    INDEX idx_loan_no (loan_no)
);

-- =====================================================
-- 还款计划表
-- =====================================================
CREATE TABLE plans (
    id BIGINT UNSIGNED AUTO_INCREMENT PRIMARY KEY,
    loan_no VARCHAR(50) NOT NULL,
    period INT NOT NULL,
    
    due_date DATE NOT NULL COMMENT '应还日',
    
    due_principal DECIMAL(15,2) NOT NULL DEFAULT 0,
    due_interest DECIMAL(15,2) NOT NULL DEFAULT 0,
    due_penalty DECIMAL(15,2) NOT NULL DEFAULT 0,
    due_other_fee DECIMAL(15,2) NOT NULL DEFAULT 0,
    due_total DECIMAL(15,2) NOT NULL DEFAULT 0,
    
    paid_principal DECIMAL(15,2) NOT NULL DEFAULT 0,
    paid_interest DECIMAL(15,2) NOT NULL DEFAULT 0,
    paid_penalty DECIMAL(15,2) NOT NULL DEFAULT 0,
    paid_other_fee DECIMAL(15,2) NOT NULL DEFAULT 0,
    paid_total DECIMAL(15,2) NOT NULL DEFAULT 0,
    
    overdue_days INT DEFAULT 0,
    
    status ENUM('PENDING', 'PAID', 'OVERDUE', 'PARTIAL') DEFAULT 'PENDING',
    
    created_by VARCHAR(64) NOT NULL DEFAULT 'system',
    updated_by VARCHAR(64) NOT NULL DEFAULT 'system',
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    
    UNIQUE KEY uk_loan_period (loan_no, period),
    INDEX idx_loan_no (loan_no),
    INDEX idx_due_date (due_date),
    INDEX idx_status (status)
);

-- 还款计划变更表
CREATE TABLE plan_changes (
    id BIGINT UNSIGNED AUTO_INCREMENT PRIMARY KEY,
    loan_no VARCHAR(50) NOT NULL,
    plan_id BIGINT UNSIGNED NOT NULL,
    period INT NOT NULL,
    change_type ENUM('PAYMENT', 'OVERDUE', 'PENALTY', 'ADJUSTMENT', 'WAIVER') NOT NULL,
    field_name VARCHAR(100) NOT NULL,
    old_value TEXT,
    new_value TEXT,
    change_reason TEXT,
    related_repayment_no VARCHAR(50),
    batch_no VARCHAR(50),
    created_by VARCHAR(64) NOT NULL DEFAULT 'system',
    updated_by VARCHAR(64) NOT NULL DEFAULT 'system',
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    INDEX idx_loan_no (loan_no),
    INDEX idx_plan_id (plan_id)
);

-- 还款计划-其他费用明细
CREATE TABLE plan_other_fees (
    id BIGINT UNSIGNED AUTO_INCREMENT PRIMARY KEY,
    loan_no VARCHAR(50) NOT NULL,
    plan_id BIGINT UNSIGNED NOT NULL,
    period INT NOT NULL,
    
    fee_code VARCHAR(50) NOT NULL,
    fee_name VARCHAR(100) NOT NULL,
    
    due_amount DECIMAL(15,2) NOT NULL DEFAULT 0,
    paid_amount DECIMAL(15,2) NOT NULL DEFAULT 0,
    
    status ENUM('PENDING', 'PAID', 'WAIVED') DEFAULT 'PENDING',
    
    created_by VARCHAR(64) NOT NULL DEFAULT 'system',
    updated_by VARCHAR(64) NOT NULL DEFAULT 'system',
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    
    INDEX idx_plan_id (plan_id),
    INDEX idx_loan_no (loan_no)
);

-- =====================================================
-- 每日计算明细（利息、罚息）
-- =====================================================
CREATE TABLE daily_calculations (
    id BIGINT UNSIGNED AUTO_INCREMENT PRIMARY KEY,
    loan_no VARCHAR(50) NOT NULL,
    calculation_date DATE NOT NULL COMMENT '计息日',
    
    fee_code VARCHAR(50) NOT NULL,
    fee_category ENUM('INTEREST', 'PENALTY', 'OTHER_FEE') NOT NULL,
    
    base_amount DECIMAL(15,2) NOT NULL,
    daily_rate DECIMAL(12,10) NOT NULL,
    amount DECIMAL(15,2) NOT NULL,
    
    plan_id BIGINT UNSIGNED,
    
    is_settled BOOLEAN DEFAULT FALSE,
    settled_repayment_no VARCHAR(50) COMMENT '结清的还款编号',
    
    batch_no VARCHAR(50),
    created_by VARCHAR(64) NOT NULL DEFAULT 'system',
    updated_by VARCHAR(64) NOT NULL DEFAULT 'system',
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    
    INDEX idx_loan_no (loan_no),
    INDEX idx_calculation_date (calculation_date),
    INDEX idx_is_settled (is_settled)
);

-- =====================================================
-- 还款记录
-- =====================================================
CREATE TABLE repayments (
    id BIGINT UNSIGNED AUTO_INCREMENT PRIMARY KEY,
    repayment_no VARCHAR(50) NOT NULL UNIQUE,
    loan_no VARCHAR(50) NOT NULL,
    plan_id BIGINT UNSIGNED,
    
    repayment_type ENUM('NORMAL', 'EARLY_SETTLEMENT', 'PARTIAL', 'ADVANCE') NOT NULL,
    
    amount DECIMAL(15,2) NOT NULL COMMENT '还款总额',
    principal_amount DECIMAL(15,2) DEFAULT 0,
    interest_amount DECIMAL(15,2) DEFAULT 0,
    penalty_amount DECIMAL(15,2) DEFAULT 0,
    other_fee_amount DECIMAL(15,2) DEFAULT 0,
    
    -- 日期
    trial_date DATE NOT NULL COMMENT '试算基准日期（用于计算应还金额）',
    booking_date DATE NOT NULL COMMENT '入账处理日期（系统实际入账时间）',
    
    allocation_rule_code VARCHAR(50),
    
    status ENUM('PENDING', 'BOOKED', 'REVERSED') DEFAULT 'PENDING',
    
    description TEXT,
    is_backdated BOOLEAN DEFAULT FALSE,
    backdated_reason TEXT,
    
    batch_no VARCHAR(50),
    created_by VARCHAR(64) NOT NULL DEFAULT 'system',
    updated_by VARCHAR(64) NOT NULL DEFAULT 'system',
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    
    INDEX idx_loan_no (loan_no),
    INDEX idx_trial_date (trial_date),
    INDEX idx_booking_date (booking_date),
    INDEX idx_repayment_no (repayment_no)
);

-- 还款入账明细
CREATE TABLE repayment_details (
    id BIGINT UNSIGNED AUTO_INCREMENT PRIMARY KEY,
    repayment_no VARCHAR(50) NOT NULL,
    loan_no VARCHAR(50) NOT NULL,
    
    fee_code VARCHAR(50) NOT NULL,
    fee_name VARCHAR(100) NOT NULL,
    fee_category ENUM('INTEREST', 'PENALTY', 'OTHER_FEE') NOT NULL,
    
    amount DECIMAL(15,2) NOT NULL,
    
    daily_calculation_id BIGINT UNSIGNED,
    plan_other_fee_id BIGINT UNSIGNED,
    
    created_by VARCHAR(64) NOT NULL DEFAULT 'system',
    updated_by VARCHAR(64) NOT NULL DEFAULT 'system',
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    
    INDEX idx_repayment_no (repayment_no),
    INDEX idx_loan_no (loan_no)
);

-- =====================================================
-- 跑批批次表
-- =====================================================
CREATE TABLE batch_jobs (
    id BIGINT UNSIGNED AUTO_INCREMENT PRIMARY KEY,
    batch_no VARCHAR(50) NOT NULL UNIQUE,
    batch_type ENUM('DAILY_CALC', 'OVERDUE_CHECK', 'STATUS_UPDATE') NOT NULL,
    batch_date DATE NOT NULL,
    status ENUM('INIT', 'RUNNING', 'SUCCESS', 'FAILED', 'PARTIAL') DEFAULT 'INIT',
    
    last_processed_id BIGINT UNSIGNED DEFAULT 0,
    page_size INT DEFAULT 1000,
    
    total_count INT DEFAULT 0,
    processed_count INT DEFAULT 0,
    success_count INT DEFAULT 0,
    failed_count INT DEFAULT 0,
    
    start_time TIMESTAMP NULL,
    end_time TIMESTAMP NULL,
    duration_ms BIGINT DEFAULT 0,
    
    error_message TEXT,
    remark TEXT,
    
    created_by VARCHAR(64) NOT NULL DEFAULT 'system',
    updated_by VARCHAR(64) NOT NULL DEFAULT 'system',
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    
    INDEX idx_batch_date (batch_date),
    INDEX idx_status (status)
);