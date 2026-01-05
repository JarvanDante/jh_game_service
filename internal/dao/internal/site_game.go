// ==========================================================================
// Code generated and maintained by GoFrame CLI tool. DO NOT EDIT.
// ==========================================================================

package internal

import (
	"context"

	"github.com/gogf/gf/v2/database/gdb"
	"github.com/gogf/gf/v2/frame/g"
)

// SiteGameDao is the data access object for the table site_game.
type SiteGameDao struct {
	table    string             // table is the underlying table name of the DAO.
	group    string             // group is the database configuration group name of the current DAO.
	columns  SiteGameColumns    // columns contains all the column names of Table for convenient usage.
	handlers []gdb.ModelHandler // handlers for customized model modification.
}

// SiteGameColumns defines and stores column names for the table site_game.
type SiteGameColumns struct {
	Id          string //
	SiteId      string // 站点ID
	Type        string // 游戏类型。1=体育；2=彩票；3=真人；4=电子游戏
	GameId      string // 游戏ID
	Status      string // 游戏是否打开或者或者关闭。1=打开；0=关闭
	IsAvailable string // 游戏是否可用。总开关。1=可用；0=不可用
	Sort        string //
	CreatedAt   string //
	UpdatedAt   string //
}

// siteGameColumns holds the columns for the table site_game.
var siteGameColumns = SiteGameColumns{
	Id:          "id",
	SiteId:      "site_id",
	Type:        "type",
	GameId:      "game_id",
	Status:      "status",
	IsAvailable: "is_available",
	Sort:        "sort",
	CreatedAt:   "created_at",
	UpdatedAt:   "updated_at",
}

// NewSiteGameDao creates and returns a new DAO object for table data access.
func NewSiteGameDao(handlers ...gdb.ModelHandler) *SiteGameDao {
	return &SiteGameDao{
		group:    "default",
		table:    "site_game",
		columns:  siteGameColumns,
		handlers: handlers,
	}
}

// DB retrieves and returns the underlying raw database management object of the current DAO.
func (dao *SiteGameDao) DB() gdb.DB {
	return g.DB(dao.group)
}

// Table returns the table name of the current DAO.
func (dao *SiteGameDao) Table() string {
	return dao.table
}

// Columns returns all column names of the current DAO.
func (dao *SiteGameDao) Columns() SiteGameColumns {
	return dao.columns
}

// Group returns the database configuration group name of the current DAO.
func (dao *SiteGameDao) Group() string {
	return dao.group
}

// Ctx creates and returns a Model for the current DAO. It automatically sets the context for the current operation.
func (dao *SiteGameDao) Ctx(ctx context.Context) *gdb.Model {
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
func (dao *SiteGameDao) Transaction(ctx context.Context, f func(ctx context.Context, tx gdb.TX) error) (err error) {
	return dao.Ctx(ctx).Transaction(ctx, f)
}
