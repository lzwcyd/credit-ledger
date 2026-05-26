-- =====================================================
-- 跑批设计（支持大数据量）
-- =====================================================

-- 跑批批次表
CREATE TABLE batch_jobs (
    id BIGINT UNSIGNED AUTO_INCREMENT PRIMARY KEY,
    batch_no VARCHAR(50) NOT NULL UNIQUE,
    batch_type ENUM('DAILY_INTEREST', 'OVERDUE_CHECK', 'PENALTY_CALC', 'STATUS_UPDATE', 'STATEMENT') NOT NULL,
    batch_date DATE NOT NULL COMMENT '业务日期',
    status ENUM('INIT', 'RUNNING', 'SUCCESS', 'FAILED', 'PARTIAL') DEFAULT 'INIT',
    
    -- 分页控制
    last_processed_id BIGINT UNSIGNED DEFAULT 0 COMMENT '上次处理到的ID（游标）',
    page_size INT DEFAULT 1000 COMMENT '每批处理数量',
    
    -- 统计
    total_count INT DEFAULT 0 COMMENT '总待处理数',
    processed_count INT DEFAULT 0 COMMENT '已处理数',
    success_count INT DEFAULT 0 COMMENT '成功数',
    failed_count INT DEFAULT 0 COMMENT '失败数',
    
    -- 时间
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
    INDEX idx_batch_type (batch_type),
    INDEX idx_status (status)
);

-- 跑批任务明细表（每批处理的记录）
CREATE TABLE batch_job_items (
    id BIGINT UNSIGNED AUTO_INCREMENT PRIMARY KEY,
    batch_no VARCHAR(50) NOT NULL,
    batch_type ENUM('DAILY_INTEREST', 'OVERDUE_CHECK', 'PENALTY_CALC', 'STATUS_UPDATE') NOT NULL,
    batch_date DATE NOT NULL,
    
    -- 业务关联
    loan_no VARCHAR(50) NOT NULL,
    plan_id BIGINT UNSIGNED COMMENT '关联计划ID',
    
    -- 处理状态
    status ENUM('PENDING', 'PROCESSING', 'SUCCESS', 'FAILED', 'SKIPPED') DEFAULT 'PENDING',
    error_message TEXT,
    
    -- 处理结果
    result_data JSON COMMENT '处理结果（JSON格式）',
    
    -- 时间
    start_time TIMESTAMP NULL,
    end_time TIMESTAMP NULL,
    
    created_by VARCHAR(64) NOT NULL DEFAULT 'system',
    updated_by VARCHAR(64) NOT NULL DEFAULT 'system',
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    
    INDEX idx_batch_no (batch_no),
    INDEX idx_batch_date (batch_date),
    INDEX idx_loan_no (loan_no),
    INDEX idx_status (status),
    INDEX idx_batch_type_status (batch_type, status)
);

-- =====================================================
-- 跑批流程说明
-- =====================================================

/*
【日切跑批流程 - 避免 OOM 和慢查询】

1. 【初始化批次】
   - 创建 batch_jobs 记录
   - batch_type = 'OVERDUE_CHECK'
   - batch_date = 业务日期
   - status = 'INIT'

2. 【统计待处理数量】
   SELECT COUNT(*) FROM plans 
   WHERE due_date < #{batch_date} 
   AND status IN ('PENDING', 'PARTIAL');
   
   - 更新 batch_jobs.total_count

3. 【分批拉取任务】
   -- 每次拉取 pageSize 条，使用游标分页
   SELECT * FROM plans 
   WHERE due_date < #{batch_date} 
   AND status IN ('PENDING', 'PARTIAL')
   AND id > #{lastProcessedId}
   ORDER BY id ASC
   LIMIT #{pageSize};
   
   - 更新 batch_jobs.last_processed_id

4. 【批量插入任务明细】
   INSERT INTO batch_job_items (batch_no, batch_type, batch_date, loan_no, plan_id, status)
   VALUES (...);
   
   -- 一次插入 100-500 条，避免单次插入过多

5. 【并行处理任务】
   -- 从 batch_job_items 拉取 PENDING 状态的任务
   SELECT * FROM batch_job_items 
   WHERE batch_no = #{batchNo} 
   AND status = 'PENDING'
   LIMIT #{concurrentSize}
   FOR UPDATE SKIP LOCKED;  -- 支持并发，避免重复处理
   
   - 更新 status = 'PROCESSING'
   - 执行业务逻辑
   - 更新 status = 'SUCCESS' 或 'FAILED'

6. 【汇总更新】
   - 统计 batch_job_items 的状态
   - 更新 batch_jobs 的 success_count, failed_count
   - 更新 batch_jobs.status = 'SUCCESS' 或 'PARTIAL'
*/

-- =====================================================
-- 分区建议（大数据量场景）
-- =====================================================

-- daily_calculations 按月分区（可选）
-- ALTER TABLE daily_calculations PARTITION BY RANGE (TO_DAYS(calculation_date)) (
--     PARTITION p202401 VALUES LESS THAN (TO_DAYS('2024-02-01')),
--     PARTITION p202402 VALUES LESS THAN (TO_DAYS('2024-03-01')),
--     ...
-- );

-- transactions 按月分区（可选）
-- ALTER TABLE transactions PARTITION BY RANGE (TO_DAYS(transaction_date)) (...);

-- =====================================================
-- 索引优化建议
-- =====================================================

-- plans 表：逾期查询优化
-- CREATE INDEX idx_overdue_query ON plans (due_date, status, id);

-- daily_calculations 表：未结清查询优化
-- CREATE INDEX idx_unsettled ON daily_calculations (loan_no, calculation_type, is_settled, calculation_date);