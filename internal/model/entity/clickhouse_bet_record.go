package entity

import (
	"fmt"
	"time"
)

// ClickHouseBetRecord ClickHouse 注单记录（与 SQL 表结构完全对应）
type ClickHouseBetRecord struct {
	// =====================
	// 站点信息
	// =====================
	SiteID   uint32 `ch:"site_id" json:"site_id"`
	SiteCode string `ch:"site_code" json:"site_code"`

	// =====================
	// 用户信息
	// =====================
	UserID   uint64 `ch:"user_id" json:"user_id"`
	Username string `ch:"username" json:"username"`

	// =====================
	// 游戏信息
	// =====================
	Platform string `ch:"platform" json:"platform"`
	GameID   uint32 `ch:"game_id" json:"game_id"`
	GameName string `ch:"game_name" json:"game_name"`
	GameType string `ch:"game_type" json:"game_type"`

	// =====================
	// 订单信息
	// =====================
	OrderNo string `ch:"order_no" json:"order_no"`
	RoundID string `ch:"round_id" json:"round_id"`

	// =====================
	// 金额字段（核心统计字段）
	// =====================
	BetAmount   float64 `ch:"bet_amount" json:"bet_amount"`     // 投注金额
	WinAmount   float64 `ch:"win_amount" json:"win_amount"`     // 赢取金额
	ValidAmount float64 `ch:"valid_amount" json:"valid_amount"` // 有效投注金额 ⭐
	NetAmount   float64 `ch:"net_amount" json:"net_amount"`     // 净输赢 ⭐

	// =====================
	// 时间字段
	// =====================
	BetTime    time.Time  `ch:"bet_time" json:"bet_time"`
	SettleTime *time.Time `ch:"settle_time" json:"settle_time,omitempty"`

	// =====================
	// 状态
	// =====================
	Status uint8 `ch:"status" json:"status"` // 0=未结算, 1=已结算, 2=取消, 3=无效

	// =====================
	// 派生时间字段（自动生成，不需要手动设置）
	// =====================
	// BetDate time.Time `ch:"bet_date"` // MATERIALIZED 字段，ClickHouse 自动生成
	// BetHour uint8     `ch:"bet_hour"` // MATERIALIZED 字段，ClickHouse 自动生成
}

// BetDailyStats 每日注单统计（物化视图）
type BetDailyStats struct {
	SiteID   uint32    `ch:"site_id" json:"site_id"`
	SiteCode string    `ch:"site_code" json:"site_code"`
	StatDate time.Time `ch:"stat_date" json:"stat_date"`
	Platform string    `ch:"platform" json:"platform"`
	GameType string    `ch:"game_type" json:"game_type"`

	// 注单统计
	BetCount         uint64  `ch:"bet_count" json:"bet_count"`
	TotalBetAmount   float64 `ch:"total_bet_amount" json:"total_bet_amount"`
	TotalWinAmount   float64 `ch:"total_win_amount" json:"total_win_amount"`
	TotalValidAmount float64 `ch:"total_valid_amount" json:"total_valid_amount"` // 有效投注 ⭐
	TotalNetAmount   float64 `ch:"total_net_amount" json:"total_net_amount"`     // 净输赢 ⭐

	// 结算统计
	SettledCount       uint64  `ch:"settled_count" json:"settled_count"`
	SettledValidAmount float64 `ch:"settled_valid_amount" json:"settled_valid_amount"`
	SettledNetAmount   float64 `ch:"settled_net_amount" json:"settled_net_amount"`

	// 用户数
	UniqueUsers uint64 `ch:"unique_users" json:"unique_users"`

	UpdatedAt time.Time `ch:"updated_at" json:"updated_at"`
}

// BetHourlyStats 每小时注单统计（物化视图）
type BetHourlyStats struct {
	SiteID   uint32    `ch:"site_id" json:"site_id"`
	SiteCode string    `ch:"site_code" json:"site_code"`
	StatDate time.Time `ch:"stat_date" json:"stat_date"`
	StatHour uint8     `ch:"stat_hour" json:"stat_hour"`
	Platform string    `ch:"platform" json:"platform"`

	BetCount         uint64  `ch:"bet_count" json:"bet_count"`
	TotalBetAmount   float64 `ch:"total_bet_amount" json:"total_bet_amount"`
	TotalWinAmount   float64 `ch:"total_win_amount" json:"total_win_amount"`
	TotalValidAmount float64 `ch:"total_valid_amount" json:"total_valid_amount"`
	TotalNetAmount   float64 `ch:"total_net_amount" json:"total_net_amount"`

	UniqueUsers uint64 `ch:"unique_users" json:"unique_users"`

	UpdatedAt time.Time `ch:"updated_at" json:"updated_at"`
}

// UserBetStats 用户维度注单统计（物化视图）
type UserBetStats struct {
	SiteID   uint32    `ch:"site_id" json:"site_id"`
	UserID   uint64    `ch:"user_id" json:"user_id"`
	Username string    `ch:"username" json:"username"`
	StatDate time.Time `ch:"stat_date" json:"stat_date"`

	BetCount         uint64  `ch:"bet_count" json:"bet_count"`
	TotalBetAmount   float64 `ch:"total_bet_amount" json:"total_bet_amount"`
	TotalValidAmount float64 `ch:"total_valid_amount" json:"total_valid_amount"`
	TotalNetAmount   float64 `ch:"total_net_amount" json:"total_net_amount"`

	FirstBetTime time.Time `ch:"first_bet_time" json:"first_bet_time"`
	LastBetTime  time.Time `ch:"last_bet_time" json:"last_bet_time"`

	UpdatedAt time.Time `ch:"updated_at" json:"updated_at"`
}

// BetStatsSummary 综合统计汇总（对应前端展示）
// 注意：用户和财务数据需要从 MySQL 查询后合并
type BetStatsSummary struct {
	SiteCode string `json:"site_code"`

	// =====================
	// MySQL 查询部分
	// =====================
	// 用户统计（从 user 表）
	NewUsers          int64 `json:"new_users"`           // 新注册会员
	ActiveUsers       int64 `json:"active_users"`        // 活跃会员
	FirstDepositUsers int64 `json:"first_deposit_users"` // 首存用户数

	// 财务统计（从 balance_log 表）
	DepositAmount  float64 `json:"deposit_amount"`  // 充值总额
	WithdrawAmount float64 `json:"withdraw_amount"` // 提现总额

	// =====================
	// ClickHouse 查询部分
	// =====================
	// 注单统计
	BetCount    int64   `json:"bet_count"`    // 有效投注（注单数）
	ValidAmount float64 `json:"valid_amount"` // 有效投注金额 ⭐
	NetAmount   float64 `json:"net_amount"`   // 净输赢 ⭐
	WinAmount   float64 `json:"win_amount"`   // 派彩金额
	BetAmount   float64 `json:"bet_amount"`   // 投注金额

	// =====================
	// 计算字段
	// =====================
	ReturnRate float64 `json:"return_rate"` // 返水率 = (净输赢 / 有效投注) * 100
	WinRate    float64 `json:"win_rate"`    // 盈利率（根据业务定义）
}

// TableName 返回表名
func (ClickHouseBetRecord) TableName() string {
	return "game_analytics.bet_records"
}

// Validate 验证数据
func (r *ClickHouseBetRecord) Validate() error {
	if r.SiteID == 0 {
		return fmt.Errorf("site_id 不能为空")
	}
	if r.UserID == 0 {
		return fmt.Errorf("user_id 不能为空")
	}
	if r.OrderNo == "" {
		return fmt.Errorf("order_no 不能为空")
	}
	return nil
}

// CalculateNetAmount 计算净输赢（如果未设置）
func (r *ClickHouseBetRecord) CalculateNetAmount() {
	if r.NetAmount == 0 {
		r.NetAmount = r.WinAmount - r.BetAmount
	}
}

// IsSettled 是否已结算
func (r *ClickHouseBetRecord) IsSettled() bool {
	return r.Status == 1
}
