-- 为 bet_records 表添加 task_id 字段
ALTER TABLE `bet_records` ADD COLUMN `task_id` bigint(20) DEFAULT 0 COMMENT '同步任务ID' AFTER `ch_sync_status`;

-- 添加索引
ALTER TABLE `bet_records` ADD INDEX `idx_task_id` (`task_id`);
