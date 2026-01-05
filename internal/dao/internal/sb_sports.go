// ==========================================================================
// Code generated and maintained by GoFrame CLI tool. DO NOT EDIT.
// ==========================================================================

package internal

import (
	"context"

	"github.com/gogf/gf/v2/database/gdb"
	"github.com/gogf/gf/v2/frame/g"
)

// SbSportsDao is the data access object for the table sb_sports.
type SbSportsDao struct {
	table    string             // table is the underlying table name of the DAO.
	group    string             // group is the database configuration group name of the current DAO.
	columns  SbSportsColumns    // columns contains all the column names of Table for convenient usage.
	handlers []gdb.ModelHandler // handlers for customized model modification.
}

// SbSportsColumns defines and stores column names for the table sb_sports.
type SbSportsColumns struct {
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
	IsParlay    string //
	BetStatus   string // 注单状态；0=未结算；1=已结算；默认=0
}

// sbSportsColumns holds the columns for the table sb_sports.
var sbSportsColumns = SbSportsColumns{
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
	IsParlay:    "is_parlay",
	BetStatus:   "bet_status",
}

// NewSbSportsDao creates and returns a new DAO object for table data access.
func NewSbSportsDao(handlers ...gdb.ModelHandler) *SbSportsDao {
	return &SbSportsDao{
		group:    "default",
		table:    "sb_sports",
		columns:  sbSportsColumns,
		handlers: handlers,
	}
}

// DB retrieves and returns the underlying raw database management object of the current DAO.
func (dao *SbSportsDao) DB() gdb.DB {
	return g.DB(dao.group)
}

// Table returns the table name of the current DAO.
func (dao *SbSportsDao) Table() string {
	return dao.table
}

// Columns returns all column names of the current DAO.
func (dao *SbSportsDao) Columns() SbSportsColumns {
	return dao.columns
}

// Group returns the database configuration group name of the current DAO.
func (dao *SbSportsDao) Group() string {
	return dao.group
}

// Ctx creates and returns a Model for the current DAO. It automatically sets the context for the current operation.
func (dao *SbSportsDao) Ctx(ctx context.Context) *gdb.Model {
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
func (dao *SbSportsDao) Transaction(ctx context.Context, f func(ctx context.Context, tx gdb.TX) error) (err error) {
	return dao.Ctx(ctx).Transaction(ctx, f)
}
