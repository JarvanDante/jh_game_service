// =================================================================================
// Code generated and maintained by GoFrame CLI tool. DO NOT EDIT.
// =================================================================================

package do

import (
	"github.com/gogf/gf/v2/frame/g"
	"github.com/gogf/gf/v2/os/gtime"
)

// OneApiGame is the golang structure of table one_api_game for DAO operations like Where/Data.
type OneApiGame struct {
	g.Meta                `orm:"table:one_api_game, do:true"`
	Id                    any         // ID
	SiteCode              any         // 商户code
	ChannelCode           any         // 渠道code
	GameNo                any         // 用于标识此下注请求交易的唯一ID。
	InningNo              any         // 用于将所有下注和赢取分组在单个回合中的游戏回合ID
	ExternalTransactionId any         // 由游戏供应商提供的外部交易ID
	Username              any         // 用户名
	CurrencyCode          any         // 币种
	GameCode              any         // OneAPI集成系统中选定游戏的游戏代码
	VendorCode            any         // 提供商code
	GameCategoryCode      any         // 游戏分类code
	BetAmount             any         // 投注金额
	WinAmount             any         // 赢取的金额，当winAmount金额>0时需要钱包贷记操作。
	Win                   any         // 绝对赢取或亏损的金额
	ValidAmount           any         // 有效投注额的金额
	JackpotAmount         any         // 奖池金额，当jackpotAmount金额>0时需要钱包贷记操作
	BetStatus             any         //
	BetTime               *gtime.Time //
	VendorSettleTime      any         //
	IsFreeSpin            any         // 用于指示下注为免费旋转下注的状态
	VendorBetId           any         //
	CreatedAt             *gtime.Time //
	UpdatedAt             *gtime.Time //
}
