CREATE DATABASE IF NOT EXISTS game_analytics;

CREATE TABLE IF NOT EXISTS game_analytics.bet_records
(
    -- =====================
    -- 站点信息
    -- =====================
    site_id     UInt32 COMMENT '站点ID',
    site_code   String COMMENT '站点代码',

    -- =====================
    -- 用户信息
    -- =====================
    user_id     UInt64 COMMENT '用户ID（来自 user_mapping）',
    username    String COMMENT '用户名（冗余，便于排查）',

    -- =====================
    -- 游戏信息
    -- =====================
    platform    LowCardinality(String) COMMENT '游戏平台',
    game_id     UInt32 COMMENT '游戏ID',
    game_name   String COMMENT '游戏名称',
    game_type   LowCardinality(String) COMMENT '游戏类型',

    -- =====================
    -- 订单信息
    -- =====================
    order_no    String COMMENT '订单号',
    round_id    String COMMENT '局号/回合号',

    -- =====================
    -- 金额字段（核心统计字段）
    -- =====================
    bet_amount    Decimal(18,6) COMMENT '投注金额',
    win_amount    Decimal(18,6) COMMENT '赢取金额',
    valid_amount  Decimal(18,6) COMMENT '有效投注金额',
    net_amount    Decimal(18,6) COMMENT '净输赢win减bet',

    -- =====================
    -- 时间字段
    -- =====================
    bet_time    DateTime COMMENT '投注时间',
    settle_time Nullable(DateTime) COMMENT '结算时间',

    -- =====================
    -- 状态
    -- =====================
    status      UInt8 COMMENT '状态:0未结算,1已结算,2取消,3无效',

    -- =====================
    -- 派生时间字段（自动生成，禁止业务写）
    -- =====================
    bet_date Date MATERIALIZED toDate(bet_time),
    bet_hour UInt8 MATERIALIZED toHour(bet_time)
    )
    ENGINE = MergeTree
    PARTITION BY toYYYYMM(bet_time)
    ORDER BY (site_id, bet_time, user_id)
    PRIMARY KEY (site_id, bet_time, user_id)
    TTL bet_time + INTERVAL 1 YEAR
    SETTINGS index_granularity = 8192
    COMMENT '游戏注单分析表';

-- 日有效投注
-- SELECT
--   bet_date,
--   sum(valid_amount)
-- FROM bet_records
-- WHERE site_id = 1
--   AND bet_date BETWEEN '2026-01-01' AND '2026-01-31'
-- GROUP BY bet_date;

-- 平台输赢
-- SELECT
--   platform,
--   sum(net_amount)
-- FROM bet_records
-- WHERE site_id = 1
--   AND bet_time >= now() - INTERVAL 7 DAY
-- GROUP BY platform;

-- 活跃用户数
-- SELECT
--   countDistinct(user_id)
-- FROM bet_records
-- WHERE site_id = 1
--   AND bet_time >= today();