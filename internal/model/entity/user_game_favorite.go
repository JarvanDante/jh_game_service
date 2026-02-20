// =================================================================================
// Code generated and maintained by GoFrame CLI tool. DO NOT EDIT.
// =================================================================================

package entity

import (
	"github.com/gogf/gf/v2/os/gtime"
)

// UserGameFavorite is the golang structure for table user_game_favorite.
type UserGameFavorite struct {
	Id        uint64      `json:"id"        orm:"id"         description:"主键ID"`
	SiteId    int         `json:"siteId"    orm:"site_id"    description:"站点ID"`
	UserId    uint64      `json:"userId"    orm:"user_id"    description:"用户ID"`
	Platform  string      `json:"platform"  orm:"platform"   description:"游戏平台，如 ONE"`
	GameCode  string      `json:"gameCode"  orm:"game_code"  description:"游戏代码，如 JL_77"`
	GameId    int         `json:"gameId"    orm:"game_id"    description:"游戏ID（可选冗余）"`
	GameName  string      `json:"gameName"  orm:"game_name"  description:"游戏名称（冗余）"`
	Vendor    string      `json:"vendor"    orm:"vendor"     description:"厂商（冗余，如 JL）"`
	Style     string      `json:"style"     orm:"style"      description:"类型（冗余，如 SLOTS）"`
	GameImg   string      `json:"gameImg"   orm:"game_img"   description:"游戏图片（冗余）"`
	CreatedAt *gtime.Time `json:"createdAt" orm:"created_at" description:"创建时间"`
	UpdatedAt *gtime.Time `json:"updatedAt" orm:"updated_at" description:"更新时间"`
}
