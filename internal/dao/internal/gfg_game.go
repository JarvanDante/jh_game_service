// ==========================================================================
// Code generated and maintained by GoFrame CLI tool. DO NOT EDIT.
// ==========================================================================

package internal

import (
	"context"

	"github.com/gogf/gf/v2/database/gdb"
	"github.com/gogf/gf/v2/frame/g"
)

// GfgGameDao is the data access object for the table gfg_game.
type GfgGameDao struct {
	table    string             // table is the underlying table name of the DAO.
	group    string             // group is the database configuration group name of the current DAO.
	columns  GfgGameColumns     // columns contains all the column names of Table for convenient usage.
	handlers []gdb.ModelHandler // handlers for customized model modification.
}

// GfgGameColumns defines and stores column names for the table gfg_game.
type GfgGameColumns struct {
	Id                    string // 主键ID
	GameCategoryCode      string //
	Username              string //
	CurrencyCode          string // 币种
	GameCode              string // 游戏代码
	VendorCode            string // 提供商code
	AccountId             string // 用户ID
	Platform              string // 平台ID
	InningNo              string //
	MainRoundId           string // 主局号
	GameNo                string //
	ExternalTransactionId string // 由游戏供应商提供的外部交易ID
	GameResult            string // 游戏结果
	FieldId               string // 场ID
	FieldName             string // 场名称
	TableId               string // 桌ID
	Chair                 string // 椅子ID
	BetAmount             string //
	WinAmount             string // 赢取的金额
	ValidAmount           string //
	BetStatus             string // 投注状态
	VendorSettleTime      string // 供应商结算时间
	IsFreeSpin            string // 是否免费旋转
	VendorBetId           string // 供应商投注ID
	Win                   string // 赢
	Fee                   string // 手续费
	ReservedAmount        string // 预留金额
	Finished              string // 是否完成
	EnterMoney            string // 进入金额
	RoundBeginTime        string // 局开始时间
	BetTime               string //
	RoundEndTime          string // 局结束时间
	JackpotAmount         string // 奖池金额
	Ip                    string // IP地址
	Currency              string // 币种
	CreatedAt             string // 创建时间
	UpdatedAt             string // 更新时间
	AgentId               string // 代理ID，厅给站开的代理
	SiteCode              string //
	ChannelCode           string //
}

// gfgGameColumns holds the columns for the table gfg_game.
var gfgGameColumns = GfgGameColumns{
	Id:                    "id",
	GameCategoryCode:      "gameCategoryCode",
	Username:              "username",
	CurrencyCode:          "currencyCode",
	GameCode:              "gameCode",
	VendorCode:            "vendorCode",
	AccountId:             "account_id",
	Platform:              "platform",
	InningNo:              "inning_no",
	MainRoundId:           "main_round_id",
	GameNo:                "game_no",
	ExternalTransactionId: "externalTransactionId",
	GameResult:            "game_result",
	FieldId:               "field_id",
	FieldName:             "field_name",
	TableId:               "table_id",
	Chair:                 "chair",
	BetAmount:             "bet_amount",
	WinAmount:             "winAmount",
	ValidAmount:           "valid_amount",
	BetStatus:             "bet_status",
	VendorSettleTime:      "vendorSettleTime",
	IsFreeSpin:            "isFreeSpin",
	VendorBetId:           "vendorBetId",
	Win:                   "win",
	Fee:                   "fee",
	ReservedAmount:        "reserved_amount",
	Finished:              "finished",
	EnterMoney:            "enter_money",
	RoundBeginTime:        "round_begin_time",
	BetTime:               "bet_time",
	RoundEndTime:          "round_end_time",
	JackpotAmount:         "jackpot_amount",
	Ip:                    "ip",
	Currency:              "currency",
	CreatedAt:             "created_at",
	UpdatedAt:             "updated_at",
	AgentId:               "agent_id",
	SiteCode:              "site_code",
	ChannelCode:           "channel_code",
}

// NewGfgGameDao creates and returns a new DAO object for table data access.
func NewGfgGameDao(handlers ...gdb.ModelHandler) *GfgGameDao {
	return &GfgGameDao{
		group:    "default",
		table:    "gfg_game",
		columns:  gfgGameColumns,
		handlers: handlers,
	}
}

// DB retrieves and returns the underlying raw database management object of the current DAO.
func (dao *GfgGameDao) DB() gdb.DB {
	return g.DB(dao.group)
}

// Table returns the table name of the current DAO.
func (dao *GfgGameDao) Table() string {
	return dao.table
}

// Columns returns all column names of the current DAO.
func (dao *GfgGameDao) Columns() GfgGameColumns {
	return dao.columns
}

// Group returns the database configuration group name of the current DAO.
func (dao *GfgGameDao) Group() string {
	return dao.group
}

// Ctx creates and returns a Model for the current DAO. It automatically sets the context for the current operation.
func (dao *GfgGameDao) Ctx(ctx context.Context) *gdb.Model {
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
func (dao *GfgGameDao) Transaction(ctx context.Context, f func(ctx context.Context, tx gdb.TX) error) (err error) {
	return dao.Ctx(ctx).Transaction(ctx, f)
}
