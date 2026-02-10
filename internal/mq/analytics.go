package mq

// SyncUserToClickHouseMessage 同步用户到 ClickHouse
type SyncUserToClickHouseMessage struct {
	EventID      uint64 `json:"event_id"`
	SiteID       uint32 `json:"site_id"`
	UserID       uint64 `json:"user_id"`
	InviterID    uint64 `json:"inviter_id"`
	Username     string `json:"username"`
	Status       uint8  `json:"status"`
	RegisterIP   string `json:"register_ip"`
	RegisterTime string `json:"register_time"`
	CreatedAt    string `json:"created_at"`
	UpdatedAt    string `json:"updated_at"`
	PayTimes     uint32 `json:"pay_times"`
}

// SyncBalanceLogToClickHouseMessage 同步 balance_log 到 ClickHouse
type SyncBalanceLogToClickHouseMessage struct {
	EventID     uint64  `json:"event_id"`
	SiteID      uint32  `json:"site_id"`
	ID          uint64  `json:"id"`
	UserID      uint64  `json:"user_id"`
	TradeType   uint8   `json:"trade_type"`
	Gateway     uint8   `json:"gateway"`
	BankType    uint8   `json:"bank_type"`
	GameID      uint8   `json:"game_id"`
	TradeNo     string  `json:"trade_no"`
	Money       float64 `json:"money"`
	ActualMoney float64 `json:"actual_money"`
	Fee         float64 `json:"fee"`
	Status      uint8   `json:"status"`
	AdminID     uint32  `json:"admin_id"`
	Remark      string  `json:"remark"`
	CreatedAt   string  `json:"created_at"`
	UpdatedAt   string  `json:"updated_at"`
}

// SyncBalanceLogPaymentToClickHouseMessage 同步 balance_log_payment 到 ClickHouse
type SyncBalanceLogPaymentToClickHouseMessage struct {
	EventID            uint64  `json:"event_id"`
	SiteID             uint32  `json:"site_id"`
	ID                 uint64  `json:"id"`
	UserID             uint64  `json:"user_id"`
	Username           string  `json:"username"`
	ActivityRechargeID uint32  `json:"activity_recharge_id"`
	Gateway            uint8   `json:"gateway"`
	PaymentID          uint32  `json:"payment_id"`
	PaymentAccountID   uint32  `json:"payment_account_id"`
	BankValue          string  `json:"bank_value"`
	TradeNo            string  `json:"trade_no"`
	Money              float64 `json:"money"`
	ActualMoney        float64 `json:"actual_money"`
	Fee                float64 `json:"fee"`
	Status             uint8   `json:"status"`
	AdminID            uint32  `json:"admin_id"`
	Remark             string  `json:"remark"`
	CreatedAt          string  `json:"created_at"`
	UpdatedAt          string  `json:"updated_at"`
}
