// ==========================================================================
// Code generated and maintained by GoFrame CLI tool. DO NOT EDIT.
// ==========================================================================

package internal

import (
	"context"

	"github.com/gogf/gf/v2/database/gdb"
	"github.com/gogf/gf/v2/frame/g"
)

// UserGameFavoriteDao is the data access object for the table user_game_favorite.
type UserGameFavoriteDao struct {
	table    string                  // table is the underlying table name of the DAO.
	group    string                  // group is the database configuration group name of the current DAO.
	columns  UserGameFavoriteColumns // columns contains all the column names of Table for convenient usage.
	handlers []gdb.ModelHandler      // handlers for customized model modification.
}

// UserGameFavoriteColumns defines and stores column names for the table user_game_favorite.
type UserGameFavoriteColumns struct {
	Id        string // 主键ID
	SiteId    string // 站点ID
	UserId    string // 用户ID
	Platform  string // 游戏平台，如 ONE
	GameCode  string // 游戏代码，如 JL_77
	GameId    string // 游戏ID（可选冗余）
	GameName  string // 游戏名称（冗余）
	Vendor    string // 厂商（冗余，如 JL）
	Style     string // 类型（冗余，如 SLOTS）
	GameImg   string // 游戏图片（冗余）
	CreatedAt string // 创建时间
	UpdatedAt string // 更新时间
}

// userGameFavoriteColumns holds the columns for the table user_game_favorite.
var userGameFavoriteColumns = UserGameFavoriteColumns{
	Id:        "id",
	SiteId:    "site_id",
	UserId:    "user_id",
	Platform:  "platform",
	GameCode:  "game_code",
	GameId:    "game_id",
	GameName:  "game_name",
	Vendor:    "vendor",
	Style:     "style",
	GameImg:   "game_img",
	CreatedAt: "created_at",
	UpdatedAt: "updated_at",
}

// NewUserGameFavoriteDao creates and returns a new DAO object for table data access.
func NewUserGameFavoriteDao(handlers ...gdb.ModelHandler) *UserGameFavoriteDao {
	return &UserGameFavoriteDao{
		group:    "default",
		table:    "user_game_favorite",
		columns:  userGameFavoriteColumns,
		handlers: handlers,
	}
}

// DB retrieves and returns the underlying raw database management object of the current DAO.
func (dao *UserGameFavoriteDao) DB() gdb.DB {
	return g.DB(dao.group)
}

// Table returns the table name of the current DAO.
func (dao *UserGameFavoriteDao) Table() string {
	return dao.table
}

// Columns returns all column names of the current DAO.
func (dao *UserGameFavoriteDao) Columns() UserGameFavoriteColumns {
	return dao.columns
}

// Group returns the database configuration group name of the current DAO.
func (dao *UserGameFavoriteDao) Group() string {
	return dao.group
}

// Ctx creates and returns a Model for the current DAO. It automatically sets the context for the current operation.
func (dao *UserGameFavoriteDao) Ctx(ctx context.Context) *gdb.Model {
	model := dao.DB().Model(dao.table)
	for _, handler := range dao.handlers {
		model = handler(model)
	}
	return model.Safe().Ctx(ctx)
}

// Transaction wraps the transaction logic using function f.
// It rolls back the transaction and returns the error if function f returns a non-nil error.
// It commits the transaction and returns nil if function f returns nil.
//
// Note: Do not commit or roll back the transaction in function f,
// as it is automatically handled by this function.
func (dao *UserGameFavoriteDao) Transaction(ctx context.Context, f func(ctx context.Context, tx gdb.TX) error) (err error) {
	return dao.Ctx(ctx).Transaction(ctx, f)
}
