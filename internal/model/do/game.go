// =================================================================================
// Code generated and maintained by GoFrame CLI tool. DO NOT EDIT.
// =================================================================================

package do

import (
	"github.com/gogf/gf/v2/frame/g"
	"github.com/gogf/gf/v2/os/gtime"
)

// Game is the golang structure of table game for DAO operations like Where/Data.
type Game struct {
	g.Meta    `orm:"table:game, do:true"`
	Id        any         //
	Type      any         // 类型。1=体育；2=彩票；3=真人；4=电子游戏
	Name      any         // 游戏名称
	Platform  any         // 平台标识
	Code      any         // 游戏标识
	ImageCode any         // 图片标识
	Status    any         // 状态。1=启用；0=禁用
	DbName    any         // 数据表名称
	Remark    any         // 备注
	CreatedAt *gtime.Time //
	UpdatedAt *gtime.Time //
}
