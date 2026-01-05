// =================================================================================
// Code generated and maintained by GoFrame CLI tool. DO NOT EDIT.
// =================================================================================

package entity

import (
	"github.com/gogf/gf/v2/os/gtime"
)

// OneApiGame is the golang structure for table one_api_game.
type OneApiGame struct {
	Id                    int64       `json:"id"                    orm:"id"                    description:"ID"`
	SiteCode              string      `json:"siteCode"              orm:"site_code"             description:"商户code"`
	ChannelCode           string      `json:"channelCode"           orm:"channel_code"          description:"渠道code"`
	GameNo                string      `json:"gameNo"                orm:"game_no"               description:"用于标识此下注请求交易的唯一ID。"`
	InningNo              string      `json:"inningNo"              orm:"inning_no"             description:"用于将所有下注和赢取分组在单个回合中的游戏回合ID"`
	ExternalTransactionId string      `json:"externalTransactionId" orm:"externalTransactionId" description:"由游戏供应商提供的外部交易ID"`
	Username              string      `json:"username"              orm:"username"              description:"用户名"`
	CurrencyCode          string      `json:"currencyCode"          orm:"currencyCode"          description:"币种"`
	GameCode              string      `json:"gameCode"              orm:"gameCode"              description:"OneAPI集成系统中选定游戏的游戏代码"`
	VendorCode            string      `json:"vendorCode"            orm:"vendorCode"            description:"提供商code"`
	GameCategoryCode      string      `json:"gameCategoryCode"      orm:"gameCategoryCode"      description:"游戏分类code"`
	BetAmount             float64     `json:"betAmount"             orm:"bet_amount"            description:"投注金额"`
	WinAmount             float64     `json:"winAmount"             orm:"winAmount"             description:"赢取的金额，当winAmount金额>0时需要钱包贷记操作。"`
	Win                   float64     `json:"win"                   orm:"win"                   description:"绝对赢取或亏损的金额"`
	ValidAmount           float64     `json:"validAmount"           orm:"valid_amount"          description:"有效投注额的金额"`
	JackpotAmount         int         `json:"jackpotAmount"         orm:"jackpotAmount"         description:"奖池金额，当jackpotAmount金额>0时需要钱包贷记操作"`
	BetStatus             int         `json:"betStatus"             orm:"bet_status"            description:""`
	BetTime               *gtime.Time `json:"betTime"               orm:"bet_time"              description:""`
	VendorSettleTime      int64       `json:"vendorSettleTime"      orm:"vendorSettleTime"      description:""`
	IsFreeSpin            int         `json:"isFreeSpin"            orm:"isFreeSpin"            description:"用于指示下注为免费旋转下注的状态"`
	VendorBetId           string      `json:"vendorBetId"           orm:"vendorBetId"           description:""`
	CreatedAt             *gtime.Time `json:"createdAt"             orm:"created_at"            description:""`
	UpdatedAt             *gtime.Time `json:"updatedAt"             orm:"updated_at"            description:""`
}
