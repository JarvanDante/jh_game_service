// =================================================================================
// Code generated and maintained by GoFrame CLI tool. DO NOT EDIT.
// =================================================================================

package entity

import (
	"github.com/gogf/gf/v2/os/gtime"
)

// GfgGame is the golang structure for table gfg_game.
type GfgGame struct {
	Id                    uint64      `json:"id"                    orm:"id"                    description:"主键ID"`
	GameCategoryCode      string      `json:"gameCategoryCode"      orm:"gameCategoryCode"      description:""`
	Username              string      `json:"username"              orm:"username"              description:""`
	CurrencyCode          string      `json:"currencyCode"          orm:"currencyCode"          description:"币种"`
	GameCode              string      `json:"gameCode"              orm:"gameCode"              description:"游戏代码"`
	VendorCode            string      `json:"vendorCode"            orm:"vendorCode"            description:"提供商code"`
	AccountId             int         `json:"accountId"             orm:"account_id"            description:"用户ID"`
	Platform              int         `json:"platform"              orm:"platform"              description:"平台ID"`
	InningNo              string      `json:"inningNo"              orm:"inning_no"             description:""`
	MainRoundId           string      `json:"mainRoundId"           orm:"main_round_id"         description:"主局号"`
	GameNo                string      `json:"gameNo"                orm:"game_no"               description:""`
	ExternalTransactionId string      `json:"externalTransactionId" orm:"externalTransactionId" description:"由游戏供应商提供的外部交易ID"`
	GameResult            int         `json:"gameResult"            orm:"game_result"           description:"游戏结果"`
	FieldId               int         `json:"fieldId"               orm:"field_id"              description:"场ID"`
	FieldName             string      `json:"fieldName"             orm:"field_name"            description:"场名称"`
	TableId               int         `json:"tableId"               orm:"table_id"              description:"桌ID"`
	Chair                 int         `json:"chair"                 orm:"chair"                 description:"椅子ID"`
	BetAmount             float64     `json:"betAmount"             orm:"bet_amount"            description:""`
	WinAmount             float64     `json:"winAmount"             orm:"winAmount"             description:"赢取的金额"`
	ValidAmount           float64     `json:"validAmount"           orm:"valid_amount"          description:""`
	BetStatus             int         `json:"betStatus"             orm:"bet_status"            description:"投注状态"`
	VendorSettleTime      int64       `json:"vendorSettleTime"      orm:"vendorSettleTime"      description:"供应商结算时间"`
	IsFreeSpin            int         `json:"isFreeSpin"            orm:"isFreeSpin"            description:"是否免费旋转"`
	VendorBetId           string      `json:"vendorBetId"           orm:"vendorBetId"           description:"供应商投注ID"`
	Win                   float64     `json:"win"                   orm:"win"                   description:"赢"`
	Fee                   float64     `json:"fee"                   orm:"fee"                   description:"手续费"`
	ReservedAmount        float64     `json:"reservedAmount"        orm:"reserved_amount"       description:"预留金额"`
	Finished              int         `json:"finished"              orm:"finished"              description:"是否完成"`
	EnterMoney            float64     `json:"enterMoney"            orm:"enter_money"           description:"进入金额"`
	RoundBeginTime        *gtime.Time `json:"roundBeginTime"        orm:"round_begin_time"      description:"局开始时间"`
	BetTime               *gtime.Time `json:"betTime"               orm:"bet_time"              description:""`
	RoundEndTime          *gtime.Time `json:"roundEndTime"          orm:"round_end_time"        description:"局结束时间"`
	JackpotAmount         float64     `json:"jackpotAmount"         orm:"jackpot_amount"        description:"奖池金额"`
	Ip                    string      `json:"ip"                    orm:"ip"                    description:"IP地址"`
	Currency              string      `json:"currency"              orm:"currency"              description:"币种"`
	CreatedAt             *gtime.Time `json:"createdAt"             orm:"created_at"            description:"创建时间"`
	UpdatedAt             *gtime.Time `json:"updatedAt"             orm:"updated_at"            description:"更新时间"`
	AgentId               string      `json:"agentId"               orm:"agent_id"              description:"代理ID，厅给站开的代理"`
	SiteCode              string      `json:"siteCode"              orm:"site_code"             description:""`
	ChannelCode           string      `json:"channelCode"           orm:"channel_code"          description:""`
}
