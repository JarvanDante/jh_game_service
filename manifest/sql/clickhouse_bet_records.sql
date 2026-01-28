CREATE DATABASE IF NOT EXISTS game_analytics;

CREATE TABLE IF NOT EXISTS game_analytics.bet_records
(
    site_id     UInt32,
    site_code   String,
    user_id     UInt64,
    username    String,
    platform    LowCardinality(String),
    game_id     UInt32,
    game_name   String,
    game_type   LowCardinality(String),
    order_no    String,
    round_id    String,
    bet_amount    Decimal(18,6),
    win_amount    Decimal(18,6),
    valid_amount  Decimal(18,6),
    net_amount    Decimal(18,6),
    bet_time    DateTime,
    settle_time Nullable(DateTime),
    status      UInt8,
    updated_at  DateTime DEFAULT now(),
    bet_date Date MATERIALIZED toDate(bet_time),
    bet_hour UInt8 MATERIALIZED toHour(bet_time)
)
ENGINE = ReplacingMergeTree(updated_at)
PARTITION BY toYYYYMM(bet_time)
ORDER BY (site_id, order_no)
PRIMARY KEY (site_id, order_no)
TTL bet_time + INTERVAL 1 YEAR
SETTINGS index_granularity = 8192;

-- 字段说明：
-- site_id: 站点ID
-- site_code: 站点代码
-- user_id: 用户ID
-- username: 用户名
-- platform: 游戏平台
-- game_id: 游戏ID
-- game_name: 游戏名称
-- game_type: 游戏类型
-- order_no: 订单号（唯一键，用于去重）
-- round_id: 局号
-- bet_amount: 投注金额
-- win_amount: 赢取金额
-- valid_amount: 有效投注金额
-- net_amount: 净输赢
-- bet_time: 投注时间
-- settle_time: 结算时间
-- status: 状态 (0=未结算, 1=已结算, 2=取消, 3=无效)
-- updated_at: 更新时间（用于 ReplacingMergeTree 保留最新版本）
-- bet_date: 投注日期（自动生成）
-- bet_hour: 投注小时（自动生成）

-- 查询示例：

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
