// =================================================================================
// Code generated and maintained by GoFrame CLI tool. DO NOT EDIT.
// =================================================================================

package do

import (
	"github.com/gogf/gf/v2/frame/g"
	"github.com/gogf/gf/v2/os/gtime"
)

// UserMapping is the golang structure of table user_mapping for DAO operations like Where/Data.
type UserMapping struct {
	g.Meta    `orm:"table:user_mapping, do:true"`
	Id        any         //
	SiteId    any         // 站点ID
	SiteCode  any         // 站点code
	Username  any         // 游戏用户名（含site前缀）
	UserId    any         // 主系统用户ID
	CreatedAt *gtime.Time //
	UpdatedAt *gtime.Time //
}
