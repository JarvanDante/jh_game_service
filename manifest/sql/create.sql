-- 游戏表
CREATE TABLE `game` (
    `id` int unsigned NOT NULL AUTO_INCREMENT COMMENT '游戏ID',
    `type` tinyint NOT NULL COMMENT '类型。1=体育；2=彩票；3=真人；4=电子游戏；5=棋牌',
    `name` varchar(20) NOT NULL COMMENT '游戏名称',
    `platform` varchar(20) NOT NULL COMMENT '平台标识',
    `code` varchar(20) NOT NULL COMMENT '游戏标识',
    `image_code` varchar(20) NOT NULL COMMENT '图片标识',
    `status` tinyint(1) NOT NULL DEFAULT '1' COMMENT '状态。1=启用；0=禁用',
    `db_name` varchar(20) NOT NULL COMMENT '数据表名称',
    `remark` varchar(255) DEFAULT NULL COMMENT '备注',
    `created_at` timestamp NULL DEFAULT NULL COMMENT '创建时间',
    `updated_at` timestamp NULL DEFAULT NULL COMMENT '更新时间',
    PRIMARY KEY (`id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='游戏表';

-- 站点游戏表
CREATE TABLE `site_game` (
    `id` int unsigned NOT NULL AUTO_INCREMENT COMMENT '主键ID',
    `site_id` int NOT NULL COMMENT '站点ID',
    `type` tinyint NOT NULL COMMENT '游戏类型。1=体育；2=彩票；3=真人；4=电子游戏；5=棋牌',
    `game_id` int NOT NULL COMMENT '游戏ID',
    `status` tinyint(1) NOT NULL DEFAULT '1' COMMENT '游戏是否打开或者关闭。1=打开；0=关闭',
    `is_available` tinyint(1) NOT NULL DEFAULT '1' COMMENT '游戏是否可用。总开关。1=可用；0=不可用',
    `sort` int NOT NULL DEFAULT '0' COMMENT '排序',
    `created_at` timestamp NULL DEFAULT NULL COMMENT '创建时间',
    `updated_at` timestamp NULL DEFAULT NULL COMMENT '更新时间',
    PRIMARY KEY (`id`),
    KEY `idx_site_id` (`site_id`),
    KEY `idx_game_id` (`game_id`),
    KEY `idx_type` (`type`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='站点游戏表';

-- 用户表（示例）
CREATE TABLE `user` (
    `id` int unsigned NOT NULL AUTO_INCREMENT COMMENT 'User ID',
    `passport` varchar(45) NOT NULL COMMENT 'User Passport',
    `password` varchar(45) NOT NULL COMMENT 'User Password',
    `nickname` varchar(45) NOT NULL COMMENT 'User Nickname',
    `create_at` datetime DEFAULT NULL COMMENT 'Created Time',
    `update_at` datetime DEFAULT NULL COMMENT 'Updated Time',
    `delete_at` datetime DEFAULT NULL COMMENT 'Deleted Time',
    PRIMARY KEY (`id`),
    UNIQUE KEY `uniq_passport` (`passport`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4;
