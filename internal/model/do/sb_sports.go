// =================================================================================
// Code generated and maintained by GoFrame CLI tool. DO NOT EDIT.
// =================================================================================

package do

import (
	"github.com/gogf/gf/v2/frame/g"
	"github.com/gogf/gf/v2/os/gtime"
)

// SbSports is the golang structure of table sb_sports for DAO operations like Where/Data.
type SbSports struct {
	g.Meta      `orm:"table:sb_sports, do:true"`
	Id          any         // 自增长id
	Username    any         // 下注会员
	GameNo      any         // 唯一单号
	GameName    any         // 游戏名
	TableNo     any         // 游戏桌号
	InningNo    any         // 游戏局号
	BetAmount   any         // 下注金额
	ValidAmount any         // 有效金额
	PlayType    any         // 游戏玩法
	BetContent  any         // 游戏内容
	GameResult  any         // 游戏结果
	BetTime     *gtime.Time // 下注时间
	Win         any         // 输赢金额、派彩金额、退还金额
	OrderStatus any         // 注单结算状态
	GameType    any         // 游戏类型
	GameHall    any         // 所属厅
	Devicetype  any         // 设备类型(pc=0 mobile=1)
	IsParlay    any         //
	BetStatus   any         // 注单状态；0=未结算；1=已结算；默认=0
}
