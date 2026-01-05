// ==========================================================================
// Code generated and maintained by GoFrame CLI tool. DO NOT EDIT.
// ==========================================================================

package internal

import (
	"context"

	"github.com/gogf/gf/v2/database/gdb"
	"github.com/gogf/gf/v2/frame/g"
)

// EgameBetLogDao is the data access object for the table egame_bet_log.
type EgameBetLogDao struct {
	table    string             // table is the underlying table name of the DAO.
	group    string             // group is the database configuration group name of the current DAO.
	columns  EgameBetLogColumns // columns contains all the column names of Table for convenient usage.
	handlers []gdb.ModelHandler // handlers for customized model modification.
}

// EgameBetLogColumns defines and stores column names for the table egame_bet_log.
type EgameBetLogColumns struct {
	Id           string // 自增主键
	SiteId       string // 站点ID
	UserId       string // 用户ID
	Username     string // 用户名快照
	PlatformCode string // 游戏平台代码 egame/jili/pg
	GameCode     string // 游戏代码
	GameName     string // 游戏名称
	GameType     string // 游戏类型
	OrderId      string // 厂商注单号
	RoundId      string // 局号
	BetAmount    string // 投注金额
	WinAmount    string // 中奖金额
	Currency     string // 币种
	BetTime      string // 投注时间
	SettleTime   string // 结算时间
	Status       string // 状态 0=未结算 1=已结算 2=取消 3=回滚
	RawData      string // 厂商原始注单数据
	Remark       string // 备注
	CreatedAt    string // 创建时间
	UpdatedAt    string // 更新时间
}

// egameBetLogColumns holds the columns for the table egame_bet_log.
var egameBetLogColumns = EgameBetLogColumns{
	Id:           "id",
	SiteId:       "site_id",
	UserId:       "user_id",
	Username:     "username",
	PlatformCode: "platform_code",
	GameCode:     "game_code",
	GameName:     "game_name",
	GameType:     "game_type",
	OrderId:      "order_id",
	RoundId:      "round_id",
	BetAmount:    "bet_amount",
	WinAmount:    "win_amount",
	Currency:     "currency",
	BetTime:      "bet_time",
	SettleTime:   "settle_time",
	Status:       "status",
	RawData:      "raw_data",
	Remark:       "remark",
	CreatedAt:    "created_at",
	UpdatedAt:    "updated_at",
}

// NewEgameBetLogDao creates and returns a new DAO object for table data access.
func NewEgameBetLogDao(handlers ...gdb.ModelHandler) *EgameBetLogDao {
	return &EgameBetLogDao{
		group:    "default",
		table:    "egame_bet_log",
		columns:  egameBetLogColumns,
		handlers: handlers,
	}
}

// DB retrieves and returns the underlying raw database management object of the current DAO.
func (dao *EgameBetLogDao) DB() gdb.DB {
	return g.DB(dao.group)
}

// Table returns the table name of the current DAO.
func (dao *EgameBetLogDao) Table() string {
	return dao.table
}

// Columns returns all column names of the current DAO.
func (dao *EgameBetLogDao) Columns() EgameBetLogColumns {
	return dao.columns
}

// Group returns the database configuration group name of the current DAO.
func (dao *EgameBetLogDao) Group() string {
	return dao.group
}

// Ctx creates and returns a Model for the current DAO. It automatically sets the context for the current operation.
func (dao *EgameBetLogDao) Ctx(ctx context.Context) *gdb.Model {
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
func (dao *EgameBetLogDao) Transaction(ctx context.Context, f func(ctx context.Context, tx gdb.TX) error) (err error) {
	return dao.Ctx(ctx).Transaction(ctx, f)
}
