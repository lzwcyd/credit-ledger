-- =====================================================
-- 信贷账务系统 - 数据库初始化脚本
-- 适用于 MySQL 8.0+
-- =====================================================

CREATE DATABASE IF NOT EXISTS credit_ledger
  DEFAULT CHARACTER SET utf8mb4
  DEFAULT COLLATE utf8mb4_unicode_ci;

USE credit_ledger;

-- =====================================================
-- 1. 还款类型配置
-- =====================================================
CREATE TABLE IF NOT EXISTS repayment_types (
    id          BIGINT UNSIGNED AUTO_INCREMENT PRIMARY KEY,
    code        VARCHAR(32)  NOT NULL UNIQUE COMMENT '类型编码：EQUAL_INSTALLMENT/EQUAL_PRINCIPAL/INTEREST_FIRST/BULLET',
    name        VARCHAR(64)  NOT NULL COMMENT '类型名称',
    description VARCHAR(255) DEFAULT '' COMMENT '描述',
    is_active   TINYINT(1)   NOT NULL DEFAULT 1 COMMENT '是否启用',
    created_by  VARCHAR(64)  NOT NULL DEFAULT 'system',
    updated_by  VARCHAR(64)  NOT NULL DEFAULT 'system',
    created_at  DATETIME     NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at  DATETIME     NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    INDEX idx_code (code)
) ENGINE=InnoDB COMMENT='还款类型配置';

-- 预置还款类型
INSERT INTO repayment_types (code, name, description) VALUES
('EQUAL_INSTALLMENT', '等额本息', '每月还款金额固定，包含本金和利息'),
('EQUAL_PRINCIPAL',   '等额本金', '每月偿还固定本金，利息随剩余本金递减'),
('INTEREST_FIRST',    '先息后本', '按月付息，到期一次性偿还本金'),
('BULLET',            '一次性还本付息', '到期一次性偿还全部本金和利息')
ON DUPLICATE KEY UPDATE name=VALUES(name);

-- =====================================================
-- 2. 费项配置
-- =====================================================
CREATE TABLE IF NOT EXISTS fee_configs (
    id                  BIGINT UNSIGNED AUTO_INCREMENT PRIMARY KEY,
    code                VARCHAR(32)   NOT NULL UNIQUE COMMENT '费项编码',
    name                VARCHAR(64)   NOT NULL COMMENT '费项名称',
    calc_type           VARCHAR(16)   NOT NULL COMMENT '计算类型：FIXED/PERCENTAGE/DAILY_RATE',
    calc_base           VARCHAR(32)   NOT NULL COMMENT '计算基数：PRINCIPAL/INTEREST/TOTAL/REMAINING_PRINCIPAL/OVERDUE_AMOUNT',
    value               VARCHAR(32)   NOT NULL COMMENT '费率值（定点数字符串）',
    trigger_type        VARCHAR(16)   NOT NULL COMMENT '触发时机：DISBURSEMENT/REPAYMENT/DAILY/OVERDUE/EARLY_REPAYMENT',
    is_daily_accumulate TINYINT(1)    NOT NULL DEFAULT 0 COMMENT '是否按日累计',
    fee_category        VARCHAR(16)   NOT NULL COMMENT '费项归类：INTEREST/PENALTY/OTHER_FEE',
    min_amount          VARCHAR(32)   NOT NULL DEFAULT '0' COMMENT '最低金额',
    max_amount          VARCHAR(32)   DEFAULT NULL COMMENT '最高金额',
    is_active           TINYINT(1)    NOT NULL DEFAULT 1,
    created_by          VARCHAR(64)   NOT NULL DEFAULT 'system',
    updated_by          VARCHAR(64)   NOT NULL DEFAULT 'system',
    created_at          DATETIME      NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at          DATETIME      NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    INDEX idx_code (code),
    INDEX idx_trigger (trigger_type)
) ENGINE=InnoDB COMMENT='费项配置';

-- 预置费项
INSERT INTO fee_configs (code, name, calc_type, calc_base, value, trigger_type, is_daily_accumulate, fee_category) VALUES
('INTEREST',          '利息',   'DAILY_RATE', 'REMAINING_PRINCIPAL', '0.0001644', 'DAILY',        1, 'INTEREST'),
('OVERDUE_PENALTY',   '逾期罚息', 'DAILY_RATE', 'OVERDUE_AMOUNT',     '0.0002466', 'OVERDUE',      1, 'PENALTY'),
('HANDLING_FEE',      '手续费',  'PERCENTAGE', 'PRINCIPAL',          '1',         'DISBURSEMENT', 0, 'OTHER_FEE'),
('EARLY_SETTLEMENT_FEE', '提前结清手续费', 'PERCENTAGE', 'REMAINING_PRINCIPAL', '1', 'EARLY_REPAYMENT', 0, 'OTHER_FEE')
ON DUPLICATE KEY UPDATE name=VALUES(name);

-- =====================================================
-- 3. 分配规则
-- =====================================================
CREATE TABLE IF NOT EXISTS allocation_rules (
    id          BIGINT UNSIGNED AUTO_INCREMENT PRIMARY KEY,
    code        VARCHAR(32)  NOT NULL UNIQUE COMMENT '规则编码',
    name        VARCHAR(64)  NOT NULL COMMENT '规则名称',
    description VARCHAR(255) DEFAULT '',
    is_default  TINYINT(1)   NOT NULL DEFAULT 0 COMMENT '是否默认规则',
    is_active   TINYINT(1)   NOT NULL DEFAULT 1,
    created_by  VARCHAR(64)  NOT NULL DEFAULT 'system',
    updated_by  VARCHAR(64)  NOT NULL DEFAULT 'system',
    created_at  DATETIME     NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at  DATETIME     NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    INDEX idx_code (code)
) ENGINE=InnoDB COMMENT='分配规则';

INSERT INTO allocation_rules (code, name, description, is_default) VALUES
('DEFAULT', '默认规则', '罚息 > 其他费用 > 利息 > 本金', 1)
ON DUPLICATE KEY UPDATE name=VALUES(name);

CREATE TABLE IF NOT EXISTS allocation_rule_items (
    id              BIGINT UNSIGNED AUTO_INCREMENT PRIMARY KEY,
    rule_code       VARCHAR(32)  NOT NULL COMMENT '规则编码',
    priority        INT          NOT NULL COMMENT '优先级（小的先分配）',
    allocation_type VARCHAR(16)  NOT NULL COMMENT '分配类型：PENALTY/OTHER_FEE/INTEREST/PRINCIPAL',
    description     VARCHAR(255) DEFAULT '',
    created_by      VARCHAR(64)  NOT NULL DEFAULT 'system',
    updated_by      VARCHAR(64)  NOT NULL DEFAULT 'system',
    created_at      DATETIME     NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at      DATETIME     NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    INDEX idx_rule (rule_code, priority)
) ENGINE=InnoDB COMMENT='分配规则明细';

INSERT INTO allocation_rule_items (rule_code, priority, allocation_type, description) VALUES
('DEFAULT', 1, 'PENALTY',    '优先偿还罚息'),
('DEFAULT', 2, 'OTHER_FEE',  '其次偿还其他费用'),
('DEFAULT', 3, 'INTEREST',   '再次偿还利息'),
('DEFAULT', 4, 'PRINCIPAL',  '最后偿还本金')
ON DUPLICATE KEY UPDATE description=VALUES(description);

-- =====================================================
-- 4. 借据主表
-- =====================================================
CREATE TABLE IF NOT EXISTS loans (
    id                   BIGINT UNSIGNED AUTO_INCREMENT PRIMARY KEY,
    loan_no              VARCHAR(32)   NOT NULL UNIQUE COMMENT '借据编号',
    principal            VARCHAR(32)   NOT NULL COMMENT '贷款本金',
    annual_rate          VARCHAR(32)   NOT NULL COMMENT '年利率',
    term_months          INT           NOT NULL COMMENT '贷款期限（月）',
    repayment_type_code  VARCHAR(32)   NOT NULL COMMENT '还款类型编码',
    allocation_rule_code VARCHAR(32)   NOT NULL DEFAULT 'DEFAULT' COMMENT '分配规则编码',

    value_date           DATE          NOT NULL COMMENT '起息日',
    first_due_date       DATE          NOT NULL COMMENT '首期还款日',
    maturity_date        DATE          NOT NULL COMMENT '到期日',
    settlement_date      DATETIME      DEFAULT NULL COMMENT '结清日期',
    disbursement_date    DATETIME      DEFAULT NULL COMMENT '放款日期',

    status               VARCHAR(16)   NOT NULL DEFAULT 'PENDING' COMMENT '状态：PENDING/DISBURSED/OVERDUE/REPAID',

    disbursed_amount     VARCHAR(32)   NOT NULL DEFAULT '0' COMMENT '已放款金额',
    remaining_principal  VARCHAR(32)   NOT NULL DEFAULT '0' COMMENT '剩余本金',

    total_interest       VARCHAR(32)   NOT NULL DEFAULT '0' COMMENT '累计利息',
    total_penalty        VARCHAR(32)   NOT NULL DEFAULT '0' COMMENT '累计罚息',
    total_other_fee      VARCHAR(32)   NOT NULL DEFAULT '0' COMMENT '累计其他费用',

    paid_principal       VARCHAR(32)   NOT NULL DEFAULT '0' COMMENT '已还本金',
    paid_interest        VARCHAR(32)   NOT NULL DEFAULT '0' COMMENT '已还利息',
    paid_penalty         VARCHAR(32)   NOT NULL DEFAULT '0' COMMENT '已还罚息',
    paid_other_fee       VARCHAR(32)   NOT NULL DEFAULT '0' COMMENT '已还其他费用',

    overdue_days         INT           NOT NULL DEFAULT 0 COMMENT '逾期天数',
    overdue_principal    VARCHAR(32)   NOT NULL DEFAULT '0' COMMENT '逾期本金',

    last_interest_calc_date DATE       DEFAULT NULL COMMENT '上次计息日期',

    created_by           VARCHAR(64)   NOT NULL DEFAULT 'system',
    updated_by           VARCHAR(64)   NOT NULL DEFAULT 'system',
    created_at           DATETIME      NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at           DATETIME      NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,

    INDEX idx_loan_no (loan_no),
    INDEX idx_status (status),
    INDEX idx_value_date (value_date)
) ENGINE=InnoDB COMMENT='借据主表';

-- =====================================================
-- 5. 借据变更记录
-- =====================================================
CREATE TABLE IF NOT EXISTS loan_changes (
    id                   BIGINT UNSIGNED AUTO_INCREMENT PRIMARY KEY,
    loan_no              VARCHAR(32)  NOT NULL COMMENT '借据编号',
    change_type          VARCHAR(32)  NOT NULL COMMENT '变更类型',
    field_name           VARCHAR(64)  NOT NULL COMMENT '变更字段',
    old_value            TEXT         COMMENT '旧值',
    new_value            TEXT         COMMENT '新值',
    change_reason        VARCHAR(255) DEFAULT '' COMMENT '变更原因',
    related_repayment_no VARCHAR(32)  DEFAULT NULL COMMENT '关联还款编号',
    batch_no             VARCHAR(32)  DEFAULT NULL COMMENT '批次号',
    created_by           VARCHAR(64)  NOT NULL DEFAULT 'system',
    updated_by           VARCHAR(64)  NOT NULL DEFAULT 'system',
    created_at           DATETIME     NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at           DATETIME     NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,

    INDEX idx_loan_no (loan_no),
    INDEX idx_change_type (change_type)
) ENGINE=InnoDB COMMENT='借据变更记录';

-- =====================================================
-- 6. 还款计划
-- =====================================================
CREATE TABLE IF NOT EXISTS plans (
    id             BIGINT UNSIGNED AUTO_INCREMENT PRIMARY KEY,
    loan_no        VARCHAR(32)   NOT NULL COMMENT '借据编号',
    period         INT           NOT NULL COMMENT '期数',
    due_date       DATE          NOT NULL COMMENT '到期日',

    due_principal  VARCHAR(32)   NOT NULL DEFAULT '0' COMMENT '应还本金',
    due_interest   VARCHAR(32)   NOT NULL DEFAULT '0' COMMENT '应还利息',
    due_penalty    VARCHAR(32)   NOT NULL DEFAULT '0' COMMENT '应还罚息',
    due_other_fee  VARCHAR(32)   NOT NULL DEFAULT '0' COMMENT '应还其他费用',
    due_total      VARCHAR(32)   NOT NULL DEFAULT '0' COMMENT '应还总额',

    paid_principal VARCHAR(32)   NOT NULL DEFAULT '0' COMMENT '已还本金',
    paid_interest  VARCHAR(32)   NOT NULL DEFAULT '0' COMMENT '已还利息',
    paid_penalty   VARCHAR(32)   NOT NULL DEFAULT '0' COMMENT '已还罚息',
    paid_other_fee VARCHAR(32)   NOT NULL DEFAULT '0' COMMENT '已还其他费用',
    paid_total     VARCHAR(32)   NOT NULL DEFAULT '0' COMMENT '已还总额',

    overdue_days   INT           NOT NULL DEFAULT 0 COMMENT '逾期天数',

    status         VARCHAR(16)   NOT NULL DEFAULT 'PENDING' COMMENT '状态：PENDING/PARTIAL/PAID/OVERDUE',

    created_by     VARCHAR(64)   NOT NULL DEFAULT 'system',
    updated_by     VARCHAR(64)   NOT NULL DEFAULT 'system',
    created_at     DATETIME      NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at     DATETIME      NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,

    UNIQUE INDEX idx_loan_period (loan_no, period),
    INDEX idx_loan_no (loan_no),
    INDEX idx_due_date (due_date),
    INDEX idx_status (status)
) ENGINE=InnoDB COMMENT='还款计划';

-- =====================================================
-- 7. 还款计划变更记录
-- =====================================================
CREATE TABLE IF NOT EXISTS plan_changes (
    id                   BIGINT UNSIGNED AUTO_INCREMENT PRIMARY KEY,
    loan_no              VARCHAR(32)  NOT NULL,
    plan_id              BIGINT UNSIGNED NOT NULL COMMENT '计划 ID',
    period               INT          NOT NULL COMMENT '期数',
    change_type          VARCHAR(32)  NOT NULL COMMENT '变更类型',
    field_name           VARCHAR(64)  NOT NULL,
    old_value            TEXT,
    new_value            TEXT,
    change_reason        VARCHAR(255) DEFAULT '',
    related_repayment_no VARCHAR(32)  DEFAULT NULL,
    batch_no             VARCHAR(32)  DEFAULT NULL,
    created_by           VARCHAR(64)  NOT NULL DEFAULT 'system',
    updated_by           VARCHAR(64)  NOT NULL DEFAULT 'system',
    created_at           DATETIME     NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at           DATETIME     NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,

    INDEX idx_loan_no (loan_no),
    INDEX idx_plan_id (plan_id)
) ENGINE=InnoDB COMMENT='还款计划变更记录';

-- =====================================================
-- 8. 其他费用明细
-- =====================================================
CREATE TABLE IF NOT EXISTS plan_other_fees (
    id          BIGINT UNSIGNED AUTO_INCREMENT PRIMARY KEY,
    loan_no     VARCHAR(32)   NOT NULL,
    plan_id     BIGINT UNSIGNED NOT NULL COMMENT '关联计划 ID',
    period      INT           NOT NULL COMMENT '期数',
    fee_code    VARCHAR(32)   NOT NULL COMMENT '费项编码',
    fee_name    VARCHAR(64)   NOT NULL COMMENT '费项名称',
    due_amount  VARCHAR(32)   NOT NULL DEFAULT '0' COMMENT '应收金额',
    paid_amount VARCHAR(32)   NOT NULL DEFAULT '0' COMMENT '已收金额',
    status      VARCHAR(16)   NOT NULL DEFAULT 'UNPAID' COMMENT '状态：UNPAID/PARTIAL/PAID',
    created_by  VARCHAR(64)   NOT NULL DEFAULT 'system',
    updated_by  VARCHAR(64)   NOT NULL DEFAULT 'system',
    created_at  DATETIME      NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at  DATETIME      NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,

    INDEX idx_loan_no (loan_no),
    INDEX idx_plan_id (plan_id),
    INDEX idx_status (status)
) ENGINE=InnoDB COMMENT='其他费用明细';

-- =====================================================
-- 9. 每日计算明细
-- =====================================================
CREATE TABLE IF NOT EXISTS daily_calculations (
    id                  BIGINT UNSIGNED AUTO_INCREMENT PRIMARY KEY,
    loan_no             VARCHAR(32)   NOT NULL,
    calculation_date    DATE          NOT NULL COMMENT '计算日期',
    fee_code            VARCHAR(32)   NOT NULL COMMENT '费项编码',
    fee_category        VARCHAR(16)   NOT NULL COMMENT '费项归类：INTEREST/PENALTY/OTHER_FEE',
    base_amount         VARCHAR(32)   NOT NULL COMMENT '计算基数',
    daily_rate          VARCHAR(32)   NOT NULL COMMENT '日费率',
    amount              VARCHAR(32)   NOT NULL COMMENT '计算金额',
    plan_id             BIGINT UNSIGNED DEFAULT NULL COMMENT '关联计划 ID',
    is_settled          TINYINT(1)    NOT NULL DEFAULT 0 COMMENT '是否已结清',
    settled_repayment_no VARCHAR(32)  DEFAULT NULL COMMENT '结清关联的还款编号',
    batch_no            VARCHAR(32)   DEFAULT NULL COMMENT '批次号',
    created_by          VARCHAR(64)   NOT NULL DEFAULT 'system',
    updated_by          VARCHAR(64)   NOT NULL DEFAULT 'system',
    created_at          DATETIME      NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at          DATETIME      NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,

    INDEX idx_loan_no (loan_no),
    INDEX idx_calc_date (calculation_date),
    INDEX idx_settled (is_settled)
) ENGINE=InnoDB COMMENT='每日计算明细';

-- =====================================================
-- 10. 还款记录
-- =====================================================
CREATE TABLE IF NOT EXISTS repayments (
    id                   BIGINT UNSIGNED AUTO_INCREMENT PRIMARY KEY,
    repayment_no         VARCHAR(32)   NOT NULL UNIQUE COMMENT '还款编号',
    loan_no              VARCHAR(32)   NOT NULL COMMENT '借据编号',
    plan_id              BIGINT UNSIGNED DEFAULT NULL COMMENT '关联计划 ID',
    repayment_type       VARCHAR(16)   NOT NULL DEFAULT 'NORMAL' COMMENT '还款类型：NORMAL/PARTIAL/EARLY_SETTLEMENT',
    amount               VARCHAR(32)   NOT NULL COMMENT '还款金额',
    principal_amount     VARCHAR(32)   NOT NULL DEFAULT '0' COMMENT '本金分配',
    interest_amount      VARCHAR(32)   NOT NULL DEFAULT '0' COMMENT '利息分配',
    penalty_amount       VARCHAR(32)   NOT NULL DEFAULT '0' COMMENT '罚息分配',
    other_fee_amount     VARCHAR(32)   NOT NULL DEFAULT '0' COMMENT '其他费用分配',
    trial_date           DATE          NOT NULL COMMENT '试算基准日期',
    booking_date         DATE          NOT NULL COMMENT '入账日期',
    allocation_rule_code VARCHAR(32)   NOT NULL DEFAULT 'DEFAULT' COMMENT '分配规则',
    status               VARCHAR(16)   NOT NULL DEFAULT 'BOOKED' COMMENT '状态',
    description          VARCHAR(255)  DEFAULT '' COMMENT '备注',
    is_backdated         TINYINT(1)    NOT NULL DEFAULT 0 COMMENT '是否补历史账',
    backdated_reason     VARCHAR(255)  DEFAULT '' COMMENT '补账原因',
    batch_no             VARCHAR(32)   DEFAULT NULL COMMENT '批次号',
    created_by           VARCHAR(64)   NOT NULL DEFAULT 'system',
    updated_by           VARCHAR(64)   NOT NULL DEFAULT 'system',
    created_at           DATETIME      NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at           DATETIME      NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,

    INDEX idx_repayment_no (repayment_no),
    INDEX idx_loan_no (loan_no),
    INDEX idx_booking_date (booking_date)
) ENGINE=InnoDB COMMENT='还款记录';

-- =====================================================
-- 11. 还款入账明细
-- =====================================================
CREATE TABLE IF NOT EXISTS repayment_details (
    id                  BIGINT UNSIGNED AUTO_INCREMENT PRIMARY KEY,
    repayment_no        VARCHAR(32)   NOT NULL COMMENT '还款编号',
    loan_no             VARCHAR(32)   NOT NULL,
    fee_code            VARCHAR(32)   NOT NULL COMMENT '费项编码',
    fee_name            VARCHAR(64)   NOT NULL COMMENT '费项名称',
    fee_category        VARCHAR(16)   NOT NULL COMMENT '费项归类',
    amount              VARCHAR(32)   NOT NULL COMMENT '分配金额',
    daily_calculation_id BIGINT UNSIGNED DEFAULT NULL COMMENT '关联每日计算 ID',
    plan_other_fee_id   BIGINT UNSIGNED DEFAULT NULL COMMENT '关联其他费用 ID',
    created_by          VARCHAR(64)   NOT NULL DEFAULT 'system',
    updated_by          VARCHAR(64)   NOT NULL DEFAULT 'system',
    created_at          DATETIME      NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at          DATETIME      NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,

    INDEX idx_repayment_no (repayment_no),
    INDEX idx_loan_no (loan_no)
) ENGINE=InnoDB COMMENT='还款入账明细';

-- =====================================================
-- 12. 跑批批次
-- =====================================================
CREATE TABLE IF NOT EXISTS batch_jobs (
    id               BIGINT UNSIGNED AUTO_INCREMENT PRIMARY KEY,
    batch_no         VARCHAR(32)   NOT NULL UNIQUE COMMENT '批次号',
    batch_type       VARCHAR(32)   NOT NULL COMMENT '批次类型：DAILY_CALC/OVERDUE_CHECK',
    batch_date       DATE          NOT NULL COMMENT '批次日期',
    status           VARCHAR(16)   NOT NULL DEFAULT 'PENDING' COMMENT '状态：PENDING/RUNNING/SUCCESS/PARTIAL/FAILED',
    last_processed_id BIGINT UNSIGNED NOT NULL DEFAULT 0 COMMENT '游标：最后处理的 ID',
    page_size        INT           NOT NULL DEFAULT 1000 COMMENT '每页大小',
    total_count      INT           NOT NULL DEFAULT 0 COMMENT '总数',
    processed_count  INT           NOT NULL DEFAULT 0 COMMENT '已处理数',
    success_count    INT           NOT NULL DEFAULT 0 COMMENT '成功数',
    failed_count     INT           NOT NULL DEFAULT 0 COMMENT '失败数',
    start_time       DATETIME      DEFAULT NULL,
    end_time         DATETIME      DEFAULT NULL,
    duration_ms      BIGINT        NOT NULL DEFAULT 0 COMMENT '耗时毫秒',
    error_message    TEXT          COMMENT '错误信息',
    remark           VARCHAR(255)  DEFAULT '',
    created_by       VARCHAR(64)   NOT NULL DEFAULT 'system',
    updated_by       VARCHAR(64)   NOT NULL DEFAULT 'system',
    created_at       DATETIME      NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at       DATETIME      NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,

    INDEX idx_batch_no (batch_no),
    INDEX idx_status (status)
) ENGINE=InnoDB COMMENT='跑批批次';
