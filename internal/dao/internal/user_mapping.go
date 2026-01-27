// ==========================================================================
// Code generated and maintained by GoFrame CLI tool. DO NOT EDIT.
// ==========================================================================

package internal

import (
	"context"

	"github.com/gogf/gf/v2/database/gdb"
	"github.com/gogf/gf/v2/frame/g"
)

// UserMappingDao is the data access object for the table user_mapping.
type UserMappingDao struct {
	table    string             // table is the underlying table name of the DAO.
	group    string             // group is the database configuration group name of the current DAO.
	columns  UserMappingColumns // columns contains all the column names of Table for convenient usage.
	handlers []gdb.ModelHandler // handlers for customized model modification.
}

// UserMappingColumns defines and stores column names for the table user_mapping.
type UserMappingColumns struct {
	Id        string //
	SiteId    string // 站点ID
	SiteCode  string // 站点code
	Username  string // 游戏用户名（含site前缀）
	UserId    string // 主系统用户ID
	CreatedAt string //
	UpdatedAt string //
}

// userMappingColumns holds the columns for the table user_mapping.
var userMappingColumns = UserMappingColumns{
	Id:        "id",
	SiteId:    "site_id",
	SiteCode:  "site_code",
	Username:  "username",
	UserId:    "user_id",
	CreatedAt: "created_at",
	UpdatedAt: "updated_at",
}

// NewUserMappingDao creates and returns a new DAO object for table data access.
func NewUserMappingDao(handlers ...gdb.ModelHandler) *UserMappingDao {
	return &UserMappingDao{
		group:    "default",
		table:    "user_mapping",
		columns:  userMappingColumns,
		handlers: handlers,
	}
}

// DB retrieves and returns the underlying raw database management object of the current DAO.
func (dao *UserMappingDao) DB() gdb.DB {
	return g.DB(dao.group)
}

// Table returns the table name of the current DAO.
func (dao *UserMappingDao) Table() string {
	return dao.table
}

// Columns returns all column names of the current DAO.
func (dao *UserMappingDao) Columns() UserMappingColumns {
	return dao.columns
}

// Group returns the database configuration group name of the current DAO.
func (dao *UserMappingDao) Group() string {
	return dao.group
}

// Ctx creates and returns a Model for the current DAO. It automatically sets the context for the current operation.
func (dao *UserMappingDao) Ctx(ctx context.Context) *gdb.Model {
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
func (dao *UserMappingDao) Transaction(ctx context.Context, f func(ctx context.Context, tx gdb.TX) error) (err error) {
	return dao.Ctx(ctx).Transaction(ctx, f)
}
