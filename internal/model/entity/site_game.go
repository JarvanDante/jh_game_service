// =================================================================================
// Code generated and maintained by GoFrame CLI tool. DO NOT EDIT.
// =================================================================================

package entity

import (
	"github.com/gogf/gf/v2/os/gtime"
)

// SiteGame is the golang structure for table site_game.
type SiteGame struct {
	Id          uint        `json:"id"          orm:"id"           description:""`
	SiteId      int         `json:"siteId"      orm:"site_id"      description:"站点ID"`
	Type        int         `json:"type"        orm:"type"         description:"游戏类型。1=体育；2=彩票；3=真人；4=电子游戏"`
	GameId      int         `json:"gameId"      orm:"game_id"      description:"游戏ID"`
	Status      int         `json:"status"      orm:"status"       description:"游戏是否打开或者或者关闭。1=打开；0=关闭"`
	IsAvailable int         `json:"isAvailable" orm:"is_available" description:"游戏是否可用。总开关。1=可用；0=不可用"`
	Sort        int         `json:"sort"        orm:"sort"         description:""`
	CreatedAt   *gtime.Time `json:"createdAt"   orm:"created_at"   description:""`
	UpdatedAt   *gtime.Time `json:"updatedAt"   orm:"updated_at"   description:""`
}
