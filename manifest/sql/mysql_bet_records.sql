CREATE TABLE IF NOT EXISTS bet_records (
                                           id bigint unsigned NOT NULL AUTO_INCREMENT COMMENT '主键ID',
                                           site_id int NOT NULL COMMENT '站点ID',
                                           site_code varchar(50) NOT NULL COMMENT '站点代码',

    user_id bigint unsigned NOT NULL COMMENT '用户ID',
    username varchar(100) NOT NULL COMMENT '用户名',

    platform varchar(50) NOT NULL COMMENT '游戏平台',
    currency varchar(10) NOT NULL DEFAULT 'PHP' COMMENT '币种',

    game_id int NOT NULL COMMENT '游戏ID',
    game_name varchar(100) NOT NULL COMMENT '游戏名称',
    game_type varchar(30) DEFAULT NULL COMMENT '游戏类型',

    order_no varchar(100) NOT NULL COMMENT '订单号',
    round_id varchar(100) DEFAULT NULL COMMENT '局号/回合号',

    bet_amount decimal(18,6) NOT NULL DEFAULT 0 COMMENT '投注金额',
    win_amount decimal(18,6) NOT NULL DEFAULT 0 COMMENT '赢取金额',
    valid_amount decimal(18,6) NOT NULL DEFAULT 0 COMMENT '有效投注金额',
    net_amount decimal(18,6) NOT NULL DEFAULT 0 COMMENT '净输赢',

    bet_time datetime NOT NULL COMMENT '投注时间',
    settle_time datetime DEFAULT NULL COMMENT '结算时间',

    status tinyint NOT NULL DEFAULT 0 COMMENT '状态:0未结算,1已结算,2取消,3无效',

    raw_data json COMMENT '原始数据JSON',

    ch_sync_status tinyint NOT NULL DEFAULT 0 COMMENT '0未同步,1成功,2失败',
    ch_sync_at datetime DEFAULT NULL COMMENT '同步时间',
    ch_sync_error varchar(255) DEFAULT NULL COMMENT '同步错误',

    created_at datetime NOT NULL DEFAULT CURRENT_TIMESTAMP COMMENT '创建时间',
    updated_at datetime NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP COMMENT '更新时间',

    PRIMARY KEY (id),
    UNIQUE KEY uk_site_platform_order (site_id, platform, order_no),

    KEY idx_site_game_time (site_id, platform, bet_time),
    KEY idx_user_time (user_id, bet_time),
    KEY idx_username (username),
    KEY idx_sync (ch_sync_status)
    ) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='游戏注单表';
