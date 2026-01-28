package dao

import (
	"github.com/gogf/gf/v2/database/gdb"
	"github.com/gogf/gf/v2/frame/g"
)

// SyncTaskDao 同步任务DAO
type SyncTaskDao struct {
	table   string
	group   string
	columns SyncTaskColumns
}

// SyncTaskColumns 同步任务表字段
type SyncTaskColumns struct {
	Id              string
	SiteId          string
	SiteCode        string
	Platform        string
	TaskType        string
	StartTime       string
	EndTime         string
	MysqlTotal      string
	MysqlInserted   string
	MysqlDuplicated string
	MqSent          string
	MqFailed        string
	MqSkipped       string
	ChSynced        string
	ChFailed        string
	Status          string
	ErrorMsg        string
	CreatedAt       string
	UpdatedAt       string
	CompletedAt     string
}

var (
	// SyncTask 同步任务DAO实例
	SyncTask = SyncTaskDao{
		table: "sync_tasks",
		group: "default",
		columns: SyncTaskColumns{
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
		},
	}
)

// Table 返回表名
func (dao *SyncTaskDao) Table() string {
	return dao.table
}

// Group 返回数据库配置组
func (dao *SyncTaskDao) Group() string {
	return dao.group
}

// Columns 返回字段列表
func (dao *SyncTaskDao) Columns() SyncTaskColumns {
	return dao.columns
}

// Ctx 创建带上下文的DAO
func (dao *SyncTaskDao) Ctx(ctx g.Ctx, cacheOption ...*gdb.CacheOption) *gdb.Model {
	return g.DB(dao.group).Model(dao.table).Safe().Ctx(ctx).Cache(gdb.CacheOption{
		Duration: -1,
		Force:    false,
	})
}

// Transaction 开启事务
func (dao *SyncTaskDao) Transaction(ctx g.Ctx, f func(ctx g.Ctx, tx gdb.TX) error) error {
	return g.DB(dao.group).Transaction(ctx, f)
}

// Add your custom methods below
