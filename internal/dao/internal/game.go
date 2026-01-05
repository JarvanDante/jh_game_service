// ==========================================================================
// Code generated and maintained by GoFrame CLI tool. DO NOT EDIT.
// ==========================================================================

package internal

import (
	"context"

	"github.com/gogf/gf/v2/database/gdb"
	"github.com/gogf/gf/v2/frame/g"
)

// GameDao is the data access object for the table game.
type GameDao struct {
	table    string             // table is the underlying table name of the DAO.
	group    string             // group is the database configuration group name of the current DAO.
	columns  GameColumns        // columns contains all the column names of Table for convenient usage.
	handlers []gdb.ModelHandler // handlers for customized model modification.
}

// GameColumns defines and stores column names for the table game.
type GameColumns struct {
	Id        string //
	Type      string // 类型。1=体育；2=彩票；3=真人；4=电子游戏
	Name      string // 游戏名称
	Platform  string // 平台标识
	Code      string // 游戏标识
	ImageCode string // 图片标识
	Status    string // 状态。1=启用；0=禁用
	DbName    string // 数据表名称
	Remark    string // 备注
	CreatedAt string //
	UpdatedAt string //
}

// gameColumns holds the columns for the table game.
var gameColumns = GameColumns{
	Id:        "id",
	Type:      "type",
	Name:      "name",
	Platform:  "platform",
	Code:      "code",
	ImageCode: "image_code",
	Status:    "status",
	DbName:    "db_name",
	Remark:    "remark",
	CreatedAt: "created_at",
	UpdatedAt: "updated_at",
}

// NewGameDao creates and returns a new DAO object for table data access.
func NewGameDao(handlers ...gdb.ModelHandler) *GameDao {
	return &GameDao{
		group:    "default",
		table:    "game",
		columns:  gameColumns,
		handlers: handlers,
	}
}

// DB retrieves and returns the underlying raw database management object of the current DAO.
func (dao *GameDao) DB() gdb.DB {
	return g.DB(dao.group)
}

// Table returns the table name of the current DAO.
func (dao *GameDao) Table() string {
	return dao.table
}

// Columns returns all column names of the current DAO.
func (dao *GameDao) Columns() GameColumns {
	return dao.columns
}

// Group returns the database configuration group name of the current DAO.
func (dao *GameDao) Group() string {
	return dao.group
}

// Ctx creates and returns a Model for the current DAO. It automatically sets the context for the current operation.
func (dao *GameDao) Ctx(ctx context.Context) *gdb.Model {
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
func (dao *GameDao) Transaction(ctx context.Context, f func(ctx context.Context, tx gdb.TX) error) (err error) {
	return dao.Ctx(ctx).Transaction(ctx, f)
}
