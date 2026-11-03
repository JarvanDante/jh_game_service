-- game_list 增加自有封面字段 cover_url（后台上传的封面，三方同步不覆盖；为空时使用三方 img/img1）
-- 可重复执行：列已存在时跳过
SET @col_exists := (
  SELECT COUNT(*) FROM information_schema.COLUMNS
  WHERE TABLE_SCHEMA = DATABASE() AND TABLE_NAME = 'game_list' AND COLUMN_NAME = 'cover_url'
);
SET @ddl := IF(@col_exists = 0,
  'ALTER TABLE `game_list` ADD COLUMN `cover_url` varchar(500) CHARACTER SET utf8mb4 COLLATE utf8mb4_general_ci NOT NULL DEFAULT '''' COMMENT ''自有封面地址（后台上传，三方同步不覆盖）'' AFTER `game_id`',
  'SELECT ''cover_url already exists''');
PREPARE stmt FROM @ddl;
EXECUTE stmt;
DEALLOCATE PREPARE stmt;
