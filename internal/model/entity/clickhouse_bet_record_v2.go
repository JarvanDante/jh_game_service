package entity

import (
	"time"
)

// ClickHouseBetRecord ClickHouse注单记录（字段名与 MySQL 一致）
type ClickHouseBetRecord struct {
	// 主键
	ID uint64 `ch:"id"`

	// 站点信息
	SiteID   uint32 `ch:"site_id"`
	SiteCode string `ch:"site_code"`

	// 用户信息
	UserID   uint64 `ch:"user_id"`
	Username string `ch:"username"`

	// 游戏信息
	Platform string `ch:"platform"`
	Currency string `ch:"currency"`
	GameID   uint32 `ch:"game_id"`
	GameName string `ch:"game_name"`
	GameType string `ch:"game_type"`

	// 订单信息
	OrderNo string `ch:"order_no"`
	RoundID string `ch:"round_id"`

	// 注单金额（与 MySQL 字段名一致）
	BetAmount   float64 `ch:"bet_amount"`   // 投注金额
	WinAmount   float64 `ch:"win_amount"`   // 赢取金额
	ValidAmount float64 `ch:"valid_amount"` // 有效投注金额 ⭐
	NetAmount   float64 `ch:"net_amount"`   // 净盈亏 ⭐

	// 时间字段
	BetTime    time.Time  `ch:"bet_time"`
	SettleTime *time.Time `ch:"settle_time"`

	// 注单状态
	Status uint8 `ch:"status"` // 0=未结算, 1=已结算, 2=取消, 3=无效

	// 原始数据
	RawData *string `ch:"raw_data"`

	// 同步状态
	ChSyncStatus uint8      `ch:"ch_sync_status"`
	ChSyncAt     *time.Time `ch:"ch_sync_at"`
	ChSyncError  *string    `ch:"ch_sync_error"`

	// 时间分区字段
	BetDate time.Time `ch:"bet_date"`
	BetHour uint8     `ch:"bet_hour"`

	// 时间戳
	CreatedAt time.Time `ch:"created_at"`
	UpdatedAt time.Time `ch:"updated_at"`
}

// BetDailyStats 每日注单统计（从物化视图查询）
type BetDailyStats struct {
	SiteID   uint32    `ch:"site_id"`
	SiteCode string    `ch:"site_code"`
	StatDate time.Time `ch:"stat_date"`
	Platform string    `ch:"platform"`
	GameType string    `ch:"game_type"`

	// 注单统计
	BetCount         uint64  `ch:"bet_count"`
	TotalBetAmount   float64 `ch:"total_bet_amount"`
	TotalWinAmount   float64 `ch:"total_win_amount"`
	TotalValidAmount float64 `ch:"total_valid_amount"` // 有效投注 ⭐
	TotalNetAmount   float64 `ch:"total_net_amount"`   // 净盈亏 ⭐

	// 结算统计
	SettledCount       uint64  `ch:"settled_count"`
	SettledValidAmount float64 `ch:"settled_valid_amount"`
	SettledNetAmount   float64 `ch:"settled_net_amount"`

	// 用户数
	UniqueUsers uint64 `ch:"unique_users"`

	UpdatedAt time.Time `ch:"updated_at"`
}

// BetStatsSummary 综合统计汇总（对应前端展示）
// 注意：用户和财务数据需要从 MySQL 查询后合并
type BetStatsSummary struct {
	SiteCode string `json:"site_code"`

	// === MySQL 查询部分 ===
	// 用户统计（从 user 表）
	NewUsers          int64 `json:"new_users"`           // 新注册会员
	ActiveUsers       int64 `json:"active_users"`        // 活跃会员
	FirstDepositUsers int64 `json:"first_deposit_users"` // 首存用户数

	// 财务统计（从 balance_log 表）
	DepositAmount  float64 `json:"deposit_amount"`  // 充值总额
	WithdrawAmount float64 `json:"withdraw_amount"` // 提现总额

	// === ClickHouse 查询部分 ===
	// 注单统计（字段名与 MySQL 一致）
	BetCount    int64   `json:"bet_count"`    // 有效投注（注单数）
	ValidAmount float64 `json:"valid_amount"` // 有效投注金额 ⭐
	NetAmount   float64 `json:"net_amount"`   // 净盈亏 ⭐
	WinAmount   float64 `json:"win_amount"`   // 派彩金额

	// === 计算字段 ===
	ReturnRate float64 `json:"return_rate"` // 返水率 = (净盈亏 / 有效投注) * 100
	WinRate    float64 `json:"win_rate"`    // 盈利率（根据业务定义）
}

// FromMySQLBetLog 从 MySQL 注单转换为 ClickHouse 注单
func (r *ClickHouseBetRecord) FromMySQLBetLog(mysqlBet interface{}) {
	// 这里可以添加从 MySQL egame_bet_log 转换的逻辑
	// 确保字段名一致，直接映射即可
}
