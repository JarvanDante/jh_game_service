// =================================================================================
// Code generated and maintained by GoFrame CLI tool. DO NOT EDIT.
// =================================================================================

package do

import (
	"github.com/gogf/gf/v2/frame/g"
	"github.com/gogf/gf/v2/os/gtime"
)

// UserGameFavorite is the golang structure of table user_game_favorite for DAO operations like Where/Data.
type UserGameFavorite struct {
	g.Meta    `orm:"table:user_game_favorite, do:true"`
	Id        any         // 主键ID
	SiteId    any         // 站点ID
	UserId    any         // 用户ID
	Platform  any         // 游戏平台，如 ONE
	GameCode  any         // 游戏代码，如 JL_77
	GameId    any         // 游戏ID（可选冗余）
	GameName  any         // 游戏名称（冗余）
	Vendor    any         // 厂商（冗余，如 JL）
	Style     any         // 类型（冗余，如 SLOTS）
	GameImg   any         // 游戏图片（冗余）
	CreatedAt *gtime.Time // 创建时间
	UpdatedAt *gtime.Time // 更新时间
}
