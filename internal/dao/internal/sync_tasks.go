// ==========================================================================
// Code generated and maintained by GoFrame CLI tool. DO NOT EDIT.
// ==========================================================================

package internal

import (
	"context"

	"github.com/gogf/gf/v2/database/gdb"
	"github.com/gogf/gf/v2/frame/g"
)

// SyncTasksDao is the data access object for the table sync_tasks.
type SyncTasksDao struct {
	table    string             // table is the underlying table name of the DAO.
	group    string             // group is the database configuration group name of the current DAO.
	columns  SyncTasksColumns   // columns contains all the column names of Table for convenient usage.
	handlers []gdb.ModelHandler // handlers for customized model modification.
}

// SyncTasksColumns defines and stores column names for the table sync_tasks.
type SyncTasksColumns struct {
	Id              string // 任务ID
	SiteId          string // 站点ID
	SiteCode        string // 站点代码
	Platform        string // 游戏平台
	TaskType        string // 任务类型: 1=定时拉单, 2=手动补单
	StartTime       string // 拉单开始时间
	EndTime         string // 拉单结束时间
	MysqlTotal      string // MySQL总记录数
	MysqlInserted   string // MySQL新插入数
	MysqlDuplicated string // MySQL重复跳过数
	MqSent          string // MQ发送成功数
	MqFailed        string // MQ发送失败数
	MqSkipped       string // MQ跳过数(未结算或无user_id)
	ChSynced        string // ClickHouse已同步数
	ChFailed        string // ClickHouse同步失败数
	Status          string // 任务状态: 1=进行中, 2=已完成, 3=失败
	ErrorMsg        string // 错误信息
	CreatedAt       string // 创建时间
	UpdatedAt       string // 更新时间
	CompletedAt     string // 完成时间
}

// syncTasksColumns holds the columns for the table sync_tasks.
var syncTasksColumns = SyncTasksColumns{
	Id:              "id",
	SiteId:          "site_id",
	SiteCode:        "site_code",
	Platform:        "platform",
	TaskType:        "task_type",
	StartTime:       "start_time",
	EndTime:         "end_time",
	MysqlTotal:      "mysql_total",
	MysqlInserted:   "mysql_inserted",
	MysqlDuplicated: "mysql_duplicated",
	MqSent:          "mq_sent",
	MqFailed:        "mq_failed",
	MqSkipped:       "mq_skipped",
	ChSynced:        "ch_synced",
	ChFailed:        "ch_failed",
	Status:          "status",
	ErrorMsg:        "error_msg",
	CreatedAt:       "created_at",
	UpdatedAt:       "updated_at",
	CompletedAt:     "completed_at",
}

// NewSyncTasksDao creates and returns a new DAO object for table data access.
func NewSyncTasksDao(handlers ...gdb.ModelHandler) *SyncTasksDao {
	return &SyncTasksDao{
		group:    "default",
		table:    "sync_tasks",
		columns:  syncTasksColumns,
		handlers: handlers,
	}
}

// DB retrieves and returns the underlying raw database management object of the current DAO.
func (dao *SyncTasksDao) DB() gdb.DB {
	return g.DB(dao.group)
}

// Table returns the table name of the current DAO.
func (dao *SyncTasksDao) Table() string {
	return dao.table
}

// Columns returns all column names of the current DAO.
func (dao *SyncTasksDao) Columns() SyncTasksColumns {
	return dao.columns
}

// Group returns the database configuration group name of the current DAO.
func (dao *SyncTasksDao) Group() string {
	return dao.group
}

// Ctx creates and returns a Model for the current DAO. It automatically sets the context for the current operation.
func (dao *SyncTasksDao) Ctx(ctx context.Context) *gdb.Model {
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
func (dao *SyncTasksDao) Transaction(ctx context.Context, f func(ctx context.Context, tx gdb.TX) error) (err error) {
	return dao.Ctx(ctx).Transaction(ctx, f)
}
