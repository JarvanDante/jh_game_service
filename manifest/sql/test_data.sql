-- 插入测试游戏数据
INSERT INTO `game` (`id`, `type`, `name`, `platform`, `code`, `image_code`, `status`, `db_name`, `remark`, `created_at`, `updated_at`) VALUES
(1, 3, 'AG国际厅', 'AG', 'AG_LIVE', 'ag_live', 1, 'ag_live', 'AG真人视讯', NOW(), NOW()),
(2, 4, 'PT电子', 'PT', 'PT_GAME', 'pt_game', 1, 'pt_game', 'PT电子游戏', NOW(), NOW()),
(3, 5, 'KY棋牌', 'KY', 'KY_POKER', 'ky_poker', 1, 'ky_poker', 'KY棋牌游戏', NOW(), NOW()),
(4, 1, 'SB体育', 'SB', 'SB_SPORTS', 'sb_sports', 1, 'sb_sports', 'SB体育投注', NOW(), NOW()),
(5, 2, 'BOYA彩票', 'BOYA', 'BOYA_LOTTERY', 'boya_lottery', 1, 'boya_lottery', 'BOYA彩票', NOW(), NOW());

-- 插入测试站点游戏配置数据（假设站点ID为1）
INSERT INTO `site_game` (`id`, `site_id`, `type`, `game_id`, `status`, `is_available`, `sort`, `created_at`, `updated_at`) VALUES
(1, 1, 3, 1, 1, 1, 1, NOW(), NOW()),
(2, 1, 4, 2, 1, 1, 2, NOW(), NOW()),
(3, 1, 5, 3, 1, 1, 3, NOW(), NOW()),
(4, 1, 1, 4, 1, 1, 4, NOW(), NOW()),
(5, 1, 2, 5, 1, 1, 5, NOW(), NOW());