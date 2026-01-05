// =================================================================================
// Code generated and maintained by GoFrame CLI tool. DO NOT EDIT.
// =================================================================================

package entity

import (
	"github.com/gogf/gf/v2/os/gtime"
)

// PtGame is the golang structure for table pt_game.
type PtGame struct {
	Id          int         `json:"id"          orm:"id"           description:"自增长id"`
	Username    string      `json:"username"    orm:"username"     description:"下注会员"`
	GameNo      string      `json:"gameNo"      orm:"game_no"      description:"唯一单号"`
	GameName    string      `json:"gameName"    orm:"game_name"    description:"游戏名"`
	TableNo     string      `json:"tableNo"     orm:"table_no"     description:"游戏桌号"`
	InningNo    string      `json:"inningNo"    orm:"inning_no"    description:"游戏局号"`
	BetAmount   float64     `json:"betAmount"   orm:"bet_amount"   description:"下注金额"`
	ValidAmount float64     `json:"validAmount" orm:"valid_amount" description:"有效金额"`
	PlayType    string      `json:"playType"    orm:"play_type"    description:"游戏玩法"`
	BetContent  string      `json:"betContent"  orm:"bet_content"  description:"游戏内容"`
	GameResult  string      `json:"gameResult"  orm:"game_result"  description:"游戏结果"`
	BetTime     *gtime.Time `json:"betTime"     orm:"bet_time"     description:"下注时间"`
	Win         float64     `json:"win"         orm:"win"          description:"输赢金额、派彩金额、退还金额"`
	OrderStatus int         `json:"orderStatus" orm:"order_status" description:"注单结算状态"`
	GameType    string      `json:"gameType"    orm:"game_type"    description:"游戏类型"`
	GameHall    int         `json:"gameHall"    orm:"game_hall"    description:"所属厅"`
	Devicetype  int         `json:"devicetype"  orm:"devicetype"   description:"设备类型(pc=0 mobile=1)"`
	BetStatus   int         `json:"betStatus"   orm:"bet_status"   description:"注单状态；0=未结算；1=已结算；默认=0"`
}
