// =================================================================================
// Code generated and maintained by GoFrame CLI tool. DO NOT EDIT.
// =================================================================================

package entity

import (
	"github.com/gogf/gf/v2/os/gtime"
)

// UserMapping is the golang structure for table user_mapping.
type UserMapping struct {
	Id        uint64      `json:"id"        orm:"id"         description:""`
	SiteId    int         `json:"siteId"    orm:"site_id"    description:"站点ID"`
	SiteCode  string      `json:"siteCode"  orm:"site_code"  description:"站点code"`
	Username  string      `json:"username"  orm:"username"   description:"游戏用户名（含site前缀）"`
	AgentId   int         `json:"agentId"   orm:"agent_id"   description:"代理id"`
	UserId    uint64      `json:"userId"    orm:"user_id"    description:"主系统用户ID"`
	CreatedAt *gtime.Time `json:"createdAt" orm:"created_at" description:""`
	UpdatedAt *gtime.Time `json:"updatedAt" orm:"updated_at" description:""`
}
