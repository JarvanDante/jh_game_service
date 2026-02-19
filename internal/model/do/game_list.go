// =================================================================================
// Code generated and maintained by GoFrame CLI tool. DO NOT EDIT.
// =================================================================================

package do

import (
	"github.com/gogf/gf/v2/frame/g"
	"github.com/gogf/gf/v2/os/gtime"
)

// GameList is the golang structure of table game_list for DAO operations like Where/Data.
type GameList struct {
	g.Meta    `orm:"table:game_list, do:true"`
	Id        any         // ID
	SiteId    any         // 商户id
	Type      any         // 游戏种类
	Platform  any         // 平台
	Name      any         // 游戏名称
	AliasName any         // 别名
	Style     any         // 游戏类型
	Img       any         // 游戏图片
	Img1      any         // 游戏图片1
	Device    any         // 设备
	Vendor    any         // 提供方
	Code      any         //
	Language  any         // 语言
	Hot       any         // 热门
	Recommend any         // 推荐
	LikeNum   any         // 喜欢数量
	IsNew     any         // 新的
	Sort      any         // 排序大的靠前
	Status    any         // 状态0关闭1开启
	Available any         // 可用状态1可用2维护
	CreatedAt *gtime.Time // 新增时间
	UpdatedAt *gtime.Time // 更新时间
	GameId    any         //
}
