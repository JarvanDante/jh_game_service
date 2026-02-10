package entity

import "time"

// ClickHouseBalanceLog ClickHouse balance_log 记录
type ClickHouseBalanceLog struct {
	EventID     uint64    `ch:"event_id" json:"event_id"`
	SiteID      uint32    `ch:"site_id" json:"site_id"`
	ID          uint64    `ch:"id" json:"id"`
	UserID      uint64    `ch:"user_id" json:"user_id"`
	TradeType   uint8     `ch:"trade_type" json:"trade_type"`
	Gateway     uint8     `ch:"gateway" json:"gateway"`
	BankType    uint8     `ch:"bank_type" json:"bank_type"`
	GameID      uint8     `ch:"game_id" json:"game_id"`
	TradeNo     string    `ch:"trade_no" json:"trade_no"`
	Money       float64   `ch:"money" json:"money"`
	ActualMoney float64   `ch:"actual_money" json:"actual_money"`
	Fee         float64   `ch:"fee" json:"fee"`
	Status      uint8     `ch:"status" json:"status"`
	AdminID     uint32    `ch:"admin_id" json:"admin_id"`
	Remark      string    `ch:"remark" json:"remark"`
	CreatedAt   time.Time `ch:"created_at" json:"created_at"`
	UpdatedAt   time.Time `ch:"updated_at" json:"updated_at"`
}
