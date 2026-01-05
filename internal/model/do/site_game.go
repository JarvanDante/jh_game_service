// =================================================================================
// Code generated and maintained by GoFrame CLI tool. DO NOT EDIT.
// =================================================================================

package do

import (
	"github.com/gogf/gf/v2/frame/g"
	"github.com/gogf/gf/v2/os/gtime"
)

// SiteGame is the golang structure of table site_game for DAO operations like Where/Data.
type SiteGame struct {
	g.Meta      `orm:"table:site_game, do:true"`
	Id          any         //
	SiteId      any         // 站点ID
	Type        any         // 游戏类型。1=体育；2=彩票；3=真人；4=电子游戏
	GameId      any         // 游戏ID
	Status      any         // 游戏是否打开或者或者关闭。1=打开；0=关闭
	IsAvailable any         // 游戏是否可用。总开关。1=可用；0=不可用
	Sort        any         //
	CreatedAt   *gtime.Time //
	UpdatedAt   *gtime.Time //
}
