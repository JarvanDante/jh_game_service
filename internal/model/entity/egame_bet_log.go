// =================================================================================
// Code generated and maintained by GoFrame CLI tool. DO NOT EDIT.
// =================================================================================

package entity

import (
	"github.com/gogf/gf/v2/os/gtime"
)

// EgameBetLog is the golang structure for table egame_bet_log.
type EgameBetLog struct {
	Id           uint64      `json:"id"           orm:"id"            description:"自增主键"`
	SiteId       int         `json:"siteId"       orm:"site_id"       description:"站点ID"`
	UserId       int64       `json:"userId"       orm:"user_id"       description:"用户ID"`
	Username     string      `json:"username"     orm:"username"      description:"用户名快照"`
	PlatformCode string      `json:"platformCode" orm:"platform_code" description:"游戏平台代码 egame/jili/pg"`
	GameCode     string      `json:"gameCode"     orm:"game_code"     description:"游戏代码"`
	GameName     string      `json:"gameName"     orm:"game_name"     description:"游戏名称"`
	GameType     string      `json:"gameType"     orm:"game_type"     description:"游戏类型"`
	OrderId      string      `json:"orderId"      orm:"order_id"      description:"厂商注单号"`
	RoundId      string      `json:"roundId"      orm:"round_id"      description:"局号"`
	BetAmount    float64     `json:"betAmount"    orm:"bet_amount"    description:"投注金额"`
	WinAmount    float64     `json:"winAmount"    orm:"win_amount"    description:"中奖金额"`
	Currency     string      `json:"currency"     orm:"currency"      description:"币种"`
	BetTime      *gtime.Time `json:"betTime"      orm:"bet_time"      description:"投注时间"`
	SettleTime   *gtime.Time `json:"settleTime"   orm:"settle_time"   description:"结算时间"`
	Status       int         `json:"status"       orm:"status"        description:"状态 0=未结算 1=已结算 2=取消 3=回滚"`
	RawData      string      `json:"rawData"      orm:"raw_data"      description:"厂商原始注单数据"`
	Remark       string      `json:"remark"       orm:"remark"        description:"备注"`
	CreatedAt    *gtime.Time `json:"createdAt"    orm:"created_at"    description:"创建时间"`
	UpdatedAt    *gtime.Time `json:"updatedAt"    orm:"updated_at"    description:"更新时间"`
}
