// ==========================================================================
// Code generated and maintained by GoFrame CLI tool. DO NOT EDIT.
// ==========================================================================

package internal

import (
	"context"

	"github.com/gogf/gf/v2/database/gdb"
	"github.com/gogf/gf/v2/frame/g"
)

// BetRecordsDao is the data access object for the table bet_records.
type BetRecordsDao struct {
	table    string             // table is the underlying table name of the DAO.
	group    string             // group is the database configuration group name of the current DAO.
	columns  BetRecordsColumns  // columns contains all the column names of Table for convenient usage.
	handlers []gdb.ModelHandler // handlers for customized model modification.
}

// BetRecordsColumns defines and stores column names for the table bet_records.
type BetRecordsColumns struct {
	Id           string // 主键ID
	SiteId       string // 站点ID
	SiteCode     string // 站点代码
	UserId       string // 用户ID
	Username     string // 用户名
	Platform     string // 游戏平台
	Currency     string // 币种
	GameId       string // 游戏ID
	GameName     string // 游戏名称
	GameType     string // 游戏类型
	OrderNo      string // 订单号
	RoundId      string // 局号/回合号
	BetAmount    string // 投注金额
	WinAmount    string // 赢取金额
	ValidAmount  string // 有效投注金额
	NetAmount    string // 净输赢
	BetTime      string // 投注时间
	SettleTime   string // 结算时间
	Status       string // 状态:0未结算,1已结算,2取消,3无效
	RawData      string // 原始数据JSON
	ChSyncStatus string // 0未同步,1成功,2失败
	TaskId       string // 同步任务ID
	ChSyncAt     string // 同步时间
	ChSyncError  string // 同步错误
	CreatedAt    string // 创建时间
	UpdatedAt    string // 更新时间
}

// betRecordsColumns holds the columns for the table bet_records.
var betRecordsColumns = BetRecordsColumns{
	Id:           "id",
	SiteId:       "site_id",
	SiteCode:     "site_code",
	UserId:       "user_id",
	Username:     "username",
	Platform:     "platform",
	Currency:     "currency",
	GameId:       "game_id",
	GameName:     "game_name",
	GameType:     "game_type",
	OrderNo:      "order_no",
	RoundId:      "round_id",
	BetAmount:    "bet_amount",
	WinAmount:    "win_amount",
	ValidAmount:  "valid_amount",
	NetAmount:    "net_amount",
	BetTime:      "bet_time",
	SettleTime:   "settle_time",
	Status:       "status",
	RawData:      "raw_data",
	ChSyncStatus: "ch_sync_status",
	TaskId:       "task_id",
	ChSyncAt:     "ch_sync_at",
	ChSyncError:  "ch_sync_error",
	CreatedAt:    "created_at",
	UpdatedAt:    "updated_at",
}

// NewBetRecordsDao creates and returns a new DAO object for table data access.
func NewBetRecordsDao(handlers ...gdb.ModelHandler) *BetRecordsDao {
	return &BetRecordsDao{
		group:    "default",
		table:    "bet_records",
		columns:  betRecordsColumns,
		handlers: handlers,
	}
}

// DB retrieves and returns the underlying raw database management object of the current DAO.
func (dao *BetRecordsDao) DB() gdb.DB {
	return g.DB(dao.group)
}

// Table returns the table name of the current DAO.
func (dao *BetRecordsDao) Table() string {
	return dao.table
}

// Columns returns all column names of the current DAO.
func (dao *BetRecordsDao) Columns() BetRecordsColumns {
	return dao.columns
}

// Group returns the database configuration group name of the current DAO.
func (dao *BetRecordsDao) Group() string {
	return dao.group
}

// Ctx creates and returns a Model for the current DAO. It automatically sets the context for the current operation.
func (dao *BetRecordsDao) Ctx(ctx context.Context) *gdb.Model {
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
func (dao *BetRecordsDao) Transaction(ctx context.Context, f func(ctx context.Context, tx gdb.TX) error) (err error) {
	return dao.Ctx(ctx).Transaction(ctx, f)
}
