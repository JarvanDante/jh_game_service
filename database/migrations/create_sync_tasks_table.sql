-- 同步任务表
CREATE TABLE IF NOT EXISTS `sync_tasks` (
    `id` bigint(20) unsigned NOT NULL AUTO_INCREMENT COMMENT '任务ID',
    `site_id` int(11) NOT NULL COMMENT '站点ID',
    `site_code` varchar(50) NOT NULL COMMENT '站点代码',
    `platform` varchar(50) NOT NULL COMMENT '游戏平台',
    `task_type` tinyint(4) NOT NULL COMMENT '任务类型: 1=定时拉单, 2=手动补单',
    `start_time` datetime NOT NULL COMMENT '拉单开始时间',
    `end_time` datetime NOT NULL COMMENT '拉单结束时间',
    
    -- MySQL统计
    `mysql_total` int(11) NOT NULL DEFAULT 0 COMMENT 'MySQL总记录数',
    `mysql_inserted` int(11) NOT NULL DEFAULT 0 COMMENT 'MySQL新插入数',
    `mysql_duplicated` int(11) NOT NULL DEFAULT 0 COMMENT 'MySQL重复跳过数',
    
    -- MQ统计
    `mq_sent` int(11) NOT NULL DEFAULT 0 COMMENT 'MQ发送成功数',
    `mq_failed` int(11) NOT NULL DEFAULT 0 COMMENT 'MQ发送失败数',
    `mq_skipped` int(11) NOT NULL DEFAULT 0 COMMENT 'MQ跳过数(未结算或无user_id)',
    
    -- ClickHouse统计（由消费者更新）
    `ch_synced` int(11) NOT NULL DEFAULT 0 COMMENT 'ClickHouse已同步数',
    `ch_failed` int(11) NOT NULL DEFAULT 0 COMMENT 'ClickHouse同步失败数',
    
    -- 任务状态
    `status` tinyint(4) NOT NULL DEFAULT 1 COMMENT '任务状态: 1=进行中, 2=已完成, 3=失败',
    `error_msg` text COMMENT '错误信息',
    
    -- 时间戳
    `created_at` datetime NOT NULL DEFAULT CURRENT_TIMESTAMP COMMENT '创建时间',
    `updated_at` datetime NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP COMMENT '更新时间',
    `completed_at` datetime DEFAULT NULL COMMENT '完成时间',
    
    PRIMARY KEY (`id`),
    KEY `idx_site_platform` (`site_id`, `platform`),
    KEY `idx_site_code` (`site_code`),
    KEY `idx_task_type` (`task_type`),
    KEY `idx_status` (`status`),
    KEY `idx_created_at` (`created_at`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='同步任务表';
