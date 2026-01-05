CREATE TABLE `egame_bet_log`
(
    `id`            BIGINT UNSIGNED NOT NULL AUTO_INCREMENT COMMENT '自增主键',

    -- 站点 & 用户
    `site_id`       INT            NOT NULL COMMENT '站点ID',
    `user_id`       BIGINT         NOT NULL COMMENT '用户ID',
    `username`      VARCHAR(64)    NOT NULL COMMENT '用户名快照',

    -- 游戏信息
    `platform_code` VARCHAR(32)    NOT NULL COMMENT '游戏平台代码 egame/jili/pg',
    `game_code`     VARCHAR(64)    NOT NULL COMMENT '游戏代码',
    `game_name`     VARCHAR(128)   NOT NULL COMMENT '游戏名称',
    `game_type`     VARCHAR(32)    NOT NULL COMMENT '游戏类型',

    -- 注单标识（幂等核心）
    `order_id`      VARCHAR(64)    NOT NULL COMMENT '厂商注单号',
    `round_id`      VARCHAR(64)             DEFAULT NULL COMMENT '局号',

    -- 金额
    `bet_amount`    DECIMAL(18, 6) NOT NULL DEFAULT 0 COMMENT '投注金额',
    `win_amount`    DECIMAL(18, 6) NOT NULL DEFAULT 0 COMMENT '中奖金额',
    `currency`      VARCHAR(16)    NOT NULL COMMENT '币种',

    -- 时间
    `bet_time`      DATETIME       NOT NULL COMMENT '投注时间',
    `settle_time`   DATETIME                DEFAULT NULL COMMENT '结算时间',

    -- 状态
    `status`        TINYINT        NOT NULL COMMENT '状态 0=未结算 1=已结算 2=取消 3=回滚',

    -- 扩展
    `raw_data`      JSON                    DEFAULT NULL COMMENT '厂商原始注单数据',
    `remark`        VARCHAR(255)            DEFAULT NULL COMMENT '备注',

    -- 维护字段
    `created_at`    DATETIME       NOT NULL DEFAULT CURRENT_TIMESTAMP COMMENT '创建时间',
    `updated_at`    DATETIME       NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP COMMENT '更新时间',

    PRIMARY KEY (`id`),

    -- 幂等索引（非常重要）
    UNIQUE KEY `uk_platform_order` (`platform_code`, `order_id`),

    -- 常用查询索引
    KEY             `idx_site_time` (`site_id`, `bet_time`),
    KEY             `idx_user_time` (`user_id`, `bet_time`),
    KEY             `idx_platform_time` (`platform_code`, `bet_time`),
    KEY             `idx_status_time` (`status`, `bet_time`)
) ENGINE=InnoDB
  DEFAULT CHARSET=utf8mb4
  COMMENT='游戏平台注单表（egame等）';
