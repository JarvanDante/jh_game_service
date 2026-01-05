// ==========================================================================
// Code generated and maintained by GoFrame CLI tool. DO NOT EDIT.
// ==========================================================================

package internal

import (
	"context"

	"github.com/gogf/gf/v2/database/gdb"
	"github.com/gogf/gf/v2/frame/g"
)

// EbetLiveDao is the data access object for the table ebet_live.
type EbetLiveDao struct {
	table    string             // table is the underlying table name of the DAO.
	group    string             // group is the database configuration group name of the current DAO.
	columns  EbetLiveColumns    // columns contains all the column names of Table for convenient usage.
	handlers []gdb.ModelHandler // handlers for customized model modification.
}

// EbetLiveColumns defines and stores column names for the table ebet_live.
type EbetLiveColumns struct {
	Id          string // 自增长id
	Username    string // 下注会员
	GameNo      string // 唯一单号
	GameName    string // 游戏名
	TableNo     string // 游戏桌号
	InningNo    string // 游戏局号
	BetAmount   string // 下注金额
	ValidAmount string // 有效金额
	PlayType    string // 游戏玩法
	BetContent  string // 游戏内容
	GameResult  string // 游戏结果
	BetTime     string // 下注时间
	Win         string // 输赢金额、派彩金额、退还金额
	OrderStatus string // 注单结算状态
	GameType    string // 游戏类型
	GameHall    string // 所属厅
	Devicetype  string // 设备类型(pc=0 mobile=1)
	BetStatus   string // 注单状态；0=未结算；1=已结算；默认=0
}

// ebetLiveColumns holds the columns for the table ebet_live.
var ebetLiveColumns = EbetLiveColumns{
	Id:          "id",
	Username:    "username",
	GameNo:      "game_no",
	GameName:    "game_name",
	TableNo:     "table_no",
	InningNo:    "inning_no",
	BetAmount:   "bet_amount",
	ValidAmount: "valid_amount",
	PlayType:    "play_type",
	BetContent:  "bet_content",
	GameResult:  "game_result",
	BetTime:     "bet_time",
	Win:         "win",
	OrderStatus: "order_status",
	GameType:    "game_type",
	GameHall:    "game_hall",
	Devicetype:  "devicetype",
	BetStatus:   "bet_status",
}

// NewEbetLiveDao creates and returns a new DAO object for table data access.
func NewEbetLiveDao(handlers ...gdb.ModelHandler) *EbetLiveDao {
	return &EbetLiveDao{
		group:    "default",
		table:    "ebet_live",
		columns:  ebetLiveColumns,
		handlers: handlers,
	}
}

// DB retrieves and returns the underlying raw database management object of the current DAO.
func (dao *EbetLiveDao) DB() gdb.DB {
	return g.DB(dao.group)
}

// Table returns the table name of the current DAO.
func (dao *EbetLiveDao) Table() string {
	return dao.table
}

// Columns returns all column names of the current DAO.
func (dao *EbetLiveDao) Columns() EbetLiveColumns {
	return dao.columns
}

// Group returns the database configuration group name of the current DAO.
func (dao *EbetLiveDao) Group() string {
	return dao.group
}

// Ctx creates and returns a Model for the current DAO. It automatically sets the context for the current operation.
func (dao *EbetLiveDao) Ctx(ctx context.Context) *gdb.Model {
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
func (dao *EbetLiveDao) Transaction(ctx context.Context, f func(ctx context.Context, tx gdb.TX) error) (err error) {
	return dao.Ctx(ctx).Transaction(ctx, f)
}
