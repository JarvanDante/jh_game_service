-- 注单表
CREATE TABLE IF NOT EXISTS bet_records (
    site_id Int32 COMMENT '站点ID',
    site_code String COMMENT '站点代码',
    game_id Int32 COMMENT '游戏ID',
    game_name String COMMENT '游戏名称',
    platform String COMMENT '游戏平台',
    order_no String COMMENT '订单号',
    username String COMMENT '用户名',
    bet_amount Float64 COMMENT '投注金额',
    win_amount Float64 COMMENT '赢取金额',
    valid_amount Float64 COMMENT '有效投注金额',
    bet_time DateTime COMMENT '投注时间',
    settle_time DateTime COMMENT '结算时间',
    status Int32 COMMENT '状态',
    raw_data String COMMENT '原始数据JSON',
    created_at DateTime DEFAULT now() COMMENT '创建时间'
) ENGINE = MergeTree()
PARTITION BY toYYYYMM(bet_time)
ORDER BY (site_id, game_id, bet_time, order_no)
SETTINGS index_granularity = 8192
COMMENT '游戏注单表';

-- 创建索引
ALTER TABLE bet_records ADD INDEX idx_order_no order_no TYPE bloom_filter GRANULARITY 1;
ALTER TABLE bet_records ADD INDEX idx_username username TYPE bloom_filter GRANULARITY 1;
