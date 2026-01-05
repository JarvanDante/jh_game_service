// ==========================================================================
// Code generated and maintained by GoFrame CLI tool. DO NOT EDIT.
// ==========================================================================

package internal

import (
	"context"

	"github.com/gogf/gf/v2/database/gdb"
	"github.com/gogf/gf/v2/frame/g"
)

// IgPpGameDao is the data access object for the table ig_pp_game.
type IgPpGameDao struct {
	table    string             // table is the underlying table name of the DAO.
	group    string             // group is the database configuration group name of the current DAO.
	columns  IgPpGameColumns    // columns contains all the column names of Table for convenient usage.
	handlers []gdb.ModelHandler // handlers for customized model modification.
}

// IgPpGameColumns defines and stores column names for the table ig_pp_game.
type IgPpGameColumns struct {
	Id               string // 主键ID
	SiteCode         string // 站点编码
	ChannelCode      string // 渠道编码
	Username         string // 用户名
	PlatformUsername string // 平台用户名
	BetStatus        string // 投注状态：0=未结算，1=已结算
	GameCategoryCode string // 游戏类别代码
	VendorCode       string // 供应商代码
	GameNo           string // 游戏编号/注单号
	InningNo         string // 局号/回合号
	BetAmount        string // 投注金额
	ValidAmount      string // 有效投注额
	Win              string // 输赢金额（正数=赢，负数=输）
	BetTime          string // 投注时间
	WinAmount        string // 派彩金额
	GameCode         string // 厅内游戏ID
	VendorSettleTime string // 供应商结算时间（毫秒时间戳）
	Status           string // 注单状态： 1: 赢 2: 输 3: 平局
	Currency         string // 币种
	AgentId          string // 代理appId
	Platform         string // IG_PP厅的ID
	CurrencyCode     string // 币种
	CreatedAt        string // 创建时间
	UpdatedAt        string // 更新时间
}

// igPpGameColumns holds the columns for the table ig_pp_game.
var igPpGameColumns = IgPpGameColumns{
	Id:               "id",
	SiteCode:         "site_code",
	ChannelCode:      "channel_code",
	Username:         "username",
	PlatformUsername: "platform_username",
	BetStatus:        "bet_status",
	GameCategoryCode: "gameCategoryCode",
	VendorCode:       "vendorCode",
	GameNo:           "game_no",
	InningNo:         "inning_no",
	BetAmount:        "bet_amount",
	ValidAmount:      "valid_amount",
	Win:              "win",
	BetTime:          "bet_time",
	WinAmount:        "winAmount",
	GameCode:         "gameCode",
	VendorSettleTime: "vendorSettleTime",
	Status:           "status",
	Currency:         "currency",
	AgentId:          "agent_id",
	Platform:         "platform",
	CurrencyCode:     "currencyCode",
	CreatedAt:        "created_at",
	UpdatedAt:        "updated_at",
}

// NewIgPpGameDao creates and returns a new DAO object for table data access.
func NewIgPpGameDao(handlers ...gdb.ModelHandler) *IgPpGameDao {
	return &IgPpGameDao{
		group:    "default",
		table:    "ig_pp_game",
		columns:  igPpGameColumns,
		handlers: handlers,
	}
}

// DB retrieves and returns the underlying raw database management object of the current DAO.
func (dao *IgPpGameDao) DB() gdb.DB {
	return g.DB(dao.group)
}

// Table returns the table name of the current DAO.
func (dao *IgPpGameDao) Table() string {
	return dao.table
}

// Columns returns all column names of the current DAO.
func (dao *IgPpGameDao) Columns() IgPpGameColumns {
	return dao.columns
}

// Group returns the database configuration group name of the current DAO.
func (dao *IgPpGameDao) Group() string {
	return dao.group
}

// Ctx creates and returns a Model for the current DAO. It automatically sets the context for the current operation.
func (dao *IgPpGameDao) Ctx(ctx context.Context) *gdb.Model {
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
func (dao *IgPpGameDao) Transaction(ctx context.Context, f func(ctx context.Context, tx gdb.TX) error) (err error) {
	return dao.Ctx(ctx).Transaction(ctx, f)
}
