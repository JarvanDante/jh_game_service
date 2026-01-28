package do

import (
	"github.com/gogf/gf/v2/frame/g"
	"github.com/gogf/gf/v2/os/gtime"
)

// SyncTask 同步任务DO
type SyncTask struct {
	g.Meta `orm:"table:sync_tasks, do:true"`

	Id        interface{} // 任务ID
	SiteId    interface{} // 站点ID
	SiteCode  interface{} // 站点代码
	Platform  interface{} // 游戏平台
	TaskType  interface{} // 任务类型: 1=定时拉单, 2=手动补单
	StartTime interface{} // 拉单开始时间
	EndTime   interface{} // 拉单结束时间

	// MySQL统计
	MysqlTotal      interface{} // MySQL总记录数
	MysqlInserted   interface{} // MySQL新插入数
	MysqlDuplicated interface{} // MySQL重复跳过数

	// MQ统计
	MqSent    interface{} // MQ发送成功数
	MqFailed  interface{} // MQ发送失败数
	MqSkipped interface{} // MQ跳过数

	// ClickHouse统计
	ChSynced interface{} // ClickHouse已同步数
	ChFailed interface{} // ClickHouse同步失败数

	// 任务状态
	Status   interface{} // 任务状态: 1=进行中, 2=已完成, 3=失败
	ErrorMsg interface{} // 错误信息

	// 时间戳
	CreatedAt   *gtime.Time // 创建时间
	UpdatedAt   *gtime.Time // 更新时间
	CompletedAt *gtime.Time // 完成时间
}
