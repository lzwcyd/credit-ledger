-- =====================================================
-- 信贷PM核心功能 - 数据库迁移脚本
-- =====================================================

USE credit_ledger;

-- =====================================================
-- 借据表新增字段（逾期分级 + 催收）
-- =====================================================
ALTER TABLE loans
  ADD COLUMN IF NOT EXISTS overdue_tier VARCHAR(8) DEFAULT '' COMMENT '逾期等级：M1-M7+' AFTER overdue_principal,
  ADD COLUMN IF NOT EXISTS collection_status VARCHAR(16) DEFAULT 'NORMAL' COMMENT '催收状态：NORMAL/IN_COLLECTION/LEGAL/WRITTEN_OFF' AFTER overdue_tier,
  ADD COLUMN IF NOT EXISTS last_collection_date DATETIME DEFAULT NULL COMMENT '最后催收日期' AFTER collection_status,
  ADD COLUMN IF NOT EXISTS collection_notes TEXT COMMENT '催收备注' AFTER last_collection_date,
  ADD COLUMN IF NOT EXISTS last_interest_calc_date DATE DEFAULT NULL COMMENT '上次计息日期' AFTER collection_notes;

ALTER TABLE loans ADD INDEX IF NOT EXISTS idx_overdue_tier (overdue_tier);
ALTER TABLE loans ADD INDEX IF NOT EXISTS idx_collection_status (collection_status);

-- =====================================================
-- 罚息减免记录表
-- =====================================================
CREATE TABLE IF NOT EXISTS penalty_waivers (
    id              BIGINT UNSIGNED AUTO_INCREMENT PRIMARY KEY,
    waiver_no       VARCHAR(32)   NOT NULL UNIQUE COMMENT '减免编号',
    loan_no         VARCHAR(32)   NOT NULL COMMENT '借据编号',
    waiver_type     VARCHAR(16)   NOT NULL COMMENT '减免类型：PENALTY/INTEREST/OTHER_FEE',
    waiver_amount   VARCHAR(32)   NOT NULL COMMENT '减免金额',
    original_amount VARCHAR(32)   NOT NULL COMMENT '原始待减免金额',
    reason          VARCHAR(512)  NOT NULL COMMENT '减免原因',
    approved_by     VARCHAR(64)   DEFAULT NULL COMMENT '审批人',
    status          VARCHAR(16)   NOT NULL DEFAULT 'APPLIED' COMMENT '状态：PENDING/APPROVED/REJECTED/APPLIED',
    created_by      VARCHAR(64)   NOT NULL DEFAULT 'system',
    updated_by      VARCHAR(64)   NOT NULL DEFAULT 'system',
    created_at      DATETIME      NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at      DATETIME      NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,

    INDEX idx_loan_no (loan_no),
    INDEX idx_status (status)
) ENGINE=InnoDB COMMENT='罚息减免记录';

-- =====================================================
-- 借据展期记录表
-- =====================================================
CREATE TABLE IF NOT EXISTS loan_extensions (
    id               BIGINT UNSIGNED AUTO_INCREMENT PRIMARY KEY,
    extension_no     VARCHAR(32)   NOT NULL UNIQUE COMMENT '展期编号',
    loan_no          VARCHAR(32)   NOT NULL COMMENT '借据编号',
    original_maturity DATE         NOT NULL COMMENT '原到期日',
    new_maturity     DATE          NOT NULL COMMENT '新到期日',
    extension_days   INT           NOT NULL DEFAULT 0 COMMENT '展期天数',
    extension_months INT           NOT NULL DEFAULT 0 COMMENT '展期月数',
    reason           VARCHAR(512)  NOT NULL COMMENT '展期原因',
    status           VARCHAR(16)   NOT NULL DEFAULT 'APPLIED' COMMENT '状态：PENDING/APPROVED/REJECTED/APPLIED',
    created_by       VARCHAR(64)   NOT NULL DEFAULT 'system',
    updated_by       VARCHAR(64)   NOT NULL DEFAULT 'system',
    created_at       DATETIME      NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at       DATETIME      NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,

    INDEX idx_loan_no (loan_no),
    INDEX idx_status (status)
) ENGINE=InnoDB COMMENT='借据展期记录';

-- =====================================================
-- 坏账核销记录表
-- =====================================================
CREATE TABLE IF NOT EXISTS write_offs (
    id               BIGINT UNSIGNED AUTO_INCREMENT PRIMARY KEY,
    write_off_no     VARCHAR(32)   NOT NULL UNIQUE COMMENT '核销编号',
    loan_no          VARCHAR(32)   NOT NULL COMMENT '借据编号',
    write_off_amount VARCHAR(32)   NOT NULL COMMENT '核销总金额',
    principal_amount VARCHAR(32)   NOT NULL COMMENT '核销本金',
    interest_amount  VARCHAR(32)   NOT NULL COMMENT '核销利息',
    penalty_amount   VARCHAR(32)   NOT NULL COMMENT '核销罚息',
    reason           VARCHAR(512)  NOT NULL COMMENT '核销原因',
    approved_by      VARCHAR(64)   DEFAULT NULL COMMENT '审批人',
    status           VARCHAR(16)   NOT NULL DEFAULT 'APPLIED' COMMENT '状态：PENDING/APPROVED/REJECTED/APPLIED',
    created_by       VARCHAR(64)   NOT NULL DEFAULT 'system',
    updated_by       VARCHAR(64)   NOT NULL DEFAULT 'system',
    created_at       DATETIME      NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at       DATETIME      NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,

    INDEX idx_loan_no (loan_no),
    INDEX idx_status (status)
) ENGINE=InnoDB COMMENT='坏账核销记录';

-- =====================================================
-- 还款提醒记录表
-- =====================================================
CREATE TABLE IF NOT EXISTS repayment_reminders (
    id            BIGINT UNSIGNED AUTO_INCREMENT PRIMARY KEY,
    loan_no       VARCHAR(32)   NOT NULL COMMENT '借据编号',
    plan_id       BIGINT UNSIGNED NOT NULL COMMENT '计划ID',
    period        INT           NOT NULL COMMENT '期数',
    due_date      DATE          NOT NULL COMMENT '到期日',
    days_before   INT           NOT NULL COMMENT '提前天数',
    reminder_type VARCHAR(16)   NOT NULL DEFAULT 'SMS' COMMENT '提醒方式：SMS/EMAIL/PUSH',
    status        VARCHAR(16)   NOT NULL DEFAULT 'PENDING' COMMENT '状态：PENDING/SENT/FAILED',
    sent_at       DATETIME      DEFAULT NULL COMMENT '发送时间',
    created_by    VARCHAR(64)   NOT NULL DEFAULT 'system',
    updated_by    VARCHAR(64)   NOT NULL DEFAULT 'system',
    created_at    DATETIME      NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at    DATETIME      NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,

    INDEX idx_loan_no (loan_no),
    INDEX idx_due_date (due_date),
    INDEX idx_status (status)
) ENGINE=InnoDB COMMENT='还款提醒记录';
