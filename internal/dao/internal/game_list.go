// ==========================================================================
// Code generated and maintained by GoFrame CLI tool. DO NOT EDIT.
// ==========================================================================

package internal

import (
	"context"

	"github.com/gogf/gf/v2/database/gdb"
	"github.com/gogf/gf/v2/frame/g"
)

// GameListDao is the data access object for the table game_list.
type GameListDao struct {
	table    string             // table is the underlying table name of the DAO.
	group    string             // group is the database configuration group name of the current DAO.
	columns  GameListColumns    // columns contains all the column names of Table for convenient usage.
	handlers []gdb.ModelHandler // handlers for customized model modification.
}

// GameListColumns defines and stores column names for the table game_list.
type GameListColumns struct {
	Id        string // ID
	SiteId    string // 商户id
	Platform  string // 平台
	Name      string // 游戏名称
	AliasName string // 别名
	Style     string // 游戏类型
	Img       string // 游戏图片
	Img1      string // 游戏图片1
	Device    string // 设备
	Vendor    string // 提供方
	Code      string // 游戏代码
	Language  string // 语言
	Hot       string // 热门
	Recommend string // 推荐
	LikeNum   string // 喜欢数量
	IsNew     string // 新的
	Sort      string // 排序大的靠前
	Status    string // 状态0关闭1开启
	Available string // 可用状态1可用2维护
	CreatedAt string // 新增时间
	UpdatedAt string // 更新时间
	GameId    string //
}

// gameListColumns holds the columns for the table game_list.
var gameListColumns = GameListColumns{
	Id:        "id",
	SiteId:    "site_id",
	Platform:  "platform",
	Name:      "name",
	AliasName: "alias_name",
	Style:     "style",
	Img:       "img",
	Img1:      "img1",
	Device:    "device",
	Vendor:    "vendor",
	Code:      "code",
	Language:  "language",
	Hot:       "hot",
	Recommend: "recommend",
	LikeNum:   "like_num",
	IsNew:     "is_new",
	Sort:      "sort",
	Status:    "status",
	Available: "available",
	CreatedAt: "created_at",
	UpdatedAt: "updated_at",
	GameId:    "game_id",
}

// NewGameListDao creates and returns a new DAO object for table data access.
func NewGameListDao(handlers ...gdb.ModelHandler) *GameListDao {
	return &GameListDao{
		group:    "default",
		table:    "game_list",
		columns:  gameListColumns,
		handlers: handlers,
	}
}

// DB retrieves and returns the underlying raw database management object of the current DAO.
func (dao *GameListDao) DB() gdb.DB {
	return g.DB(dao.group)
}

// Table returns the table name of the current DAO.
func (dao *GameListDao) Table() string {
	return dao.table
}

// Columns returns all column names of the current DAO.
func (dao *GameListDao) Columns() GameListColumns {
	return dao.columns
}

// Group returns the database configuration group name of the current DAO.
func (dao *GameListDao) Group() string {
	return dao.group
}

// Ctx creates and returns a Model for the current DAO. It automatically sets the context for the current operation.
func (dao *GameListDao) Ctx(ctx context.Context) *gdb.Model {
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
func (dao *GameListDao) Transaction(ctx context.Context, f func(ctx context.Context, tx gdb.TX) error) (err error) {
	return dao.Ctx(ctx).Transaction(ctx, f)
}
