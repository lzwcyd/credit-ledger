-- =====================================================
-- 优惠券/减免券 数据库迁移
-- =====================================================

USE credit_ledger;

-- =====================================================
-- 优惠券表
-- =====================================================
CREATE TABLE IF NOT EXISTS coupons (
    id               BIGINT UNSIGNED AUTO_INCREMENT PRIMARY KEY,
    coupon_code      VARCHAR(32)   NOT NULL UNIQUE COMMENT '券码',
    coupon_type      VARCHAR(16)   NOT NULL COMMENT '券类型：DISBURSEMENT/REPAYMENT/WAIVER/INTEREST_OFF',
    discount_type    VARCHAR(16)   NOT NULL COMMENT '折扣类型：FIXED/PERCENTAGE',
    face_value       VARCHAR(32)   NOT NULL COMMENT '面值（固定金额或百分比）',
    max_discount     VARCHAR(32)   NOT NULL DEFAULT '0' COMMENT '最高优惠金额（百分比券时有效）',
    min_usage_amount VARCHAR(32)   NOT NULL DEFAULT '0' COMMENT '最低使用金额',
    applicable_fee   VARCHAR(16)   NOT NULL DEFAULT 'ALL' COMMENT '适用费项：ALL/INTEREST/PENALTY/PRINCIPAL',
    valid_from       DATE          NOT NULL COMMENT '生效日期',
    valid_to         DATE          NOT NULL COMMENT '失效日期',
    loan_no          VARCHAR(32)   DEFAULT NULL COMMENT '绑定借据（空=通用券）',
    user_id          VARCHAR(64)   DEFAULT NULL COMMENT '绑定用户（空=不限）',
    status           VARCHAR(16)   NOT NULL DEFAULT 'ACTIVE' COMMENT '状态：ACTIVE/USED/EXPIRED/DISABLED',
    used_at          DATETIME      DEFAULT NULL COMMENT '使用时间',
    used_loan_no     VARCHAR(32)   DEFAULT NULL COMMENT '实际使用的借据',
    created_by       VARCHAR(64)   NOT NULL DEFAULT 'system',
    updated_by       VARCHAR(64)   NOT NULL DEFAULT 'system',
    created_at       DATETIME      NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at       DATETIME      NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,

    INDEX idx_coupon_code (coupon_code),
    INDEX idx_coupon_type (coupon_type),
    INDEX idx_status (status),
    INDEX idx_valid_from_to (valid_from, valid_to),
    INDEX idx_loan_no (loan_no),
    INDEX idx_user_id (user_id)
) ENGINE=InnoDB COMMENT='优惠券/减免券';

-- =====================================================
-- 优惠券使用记录表（详细流水）
-- =====================================================
CREATE TABLE IF NOT EXISTS coupon_usages (
    id              BIGINT UNSIGNED AUTO_INCREMENT PRIMARY KEY,
    coupon_code     VARCHAR(32)   NOT NULL COMMENT '券码',
    loan_no         VARCHAR(32)   NOT NULL COMMENT '使用借据',
    coupon_type     VARCHAR(16)   NOT NULL COMMENT '券类型',
    discount_type   VARCHAR(16)   NOT NULL COMMENT '折扣类型',
    face_value      VARCHAR(32)   NOT NULL COMMENT '面值',
    base_amount     VARCHAR(32)   NOT NULL COMMENT '使用基数',
    discount_amount VARCHAR(32)   NOT NULL COMMENT '实际优惠金额',
    applicable_fee  VARCHAR(16)   NOT NULL COMMENT '适用费项',
    related_repayment_no VARCHAR(32) DEFAULT NULL COMMENT '关联还款编号',
    created_by      VARCHAR(64)   NOT NULL DEFAULT 'system',
    created_at      DATETIME      NOT NULL DEFAULT CURRENT_TIMESTAMP,

    INDEX idx_coupon_code (coupon_code),
    INDEX idx_loan_no (loan_no)
) ENGINE=InnoDB COMMENT='优惠券使用记录';
