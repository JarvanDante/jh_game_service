// =================================================================================
// Code generated and maintained by GoFrame CLI tool. DO NOT EDIT.
// =================================================================================

package do

import (
	"github.com/gogf/gf/v2/frame/g"
	"github.com/gogf/gf/v2/os/gtime"
)

// SyncTasks is the golang structure of table sync_tasks for DAO operations like Where/Data.
type SyncTasks struct {
	g.Meta          `orm:"table:sync_tasks, do:true"`
	Id              any         // 任务ID
	SiteId          any         // 站点ID
	SiteCode        any         // 站点代码
	Platform        any         // 游戏平台
	TaskType        any         // 任务类型: 1=定时拉单, 2=手动补单
	StartTime       *gtime.Time // 拉单开始时间
	EndTime         *gtime.Time // 拉单结束时间
	MysqlTotal      any         // MySQL总记录数
	MysqlInserted   any         // MySQL新插入数
	MysqlDuplicated any         // MySQL重复跳过数
	MqSent          any         // MQ发送成功数
	MqFailed        any         // MQ发送失败数
	MqSkipped       any         // MQ跳过数(未结算或无user_id)
	ChSynced        any         // ClickHouse已同步数
	ChFailed        any         // ClickHouse同步失败数
	Status          any         // 任务状态: 1=进行中, 2=已完成, 3=失败
	ErrorMsg        any         // 错误信息
	CreatedAt       *gtime.Time // 创建时间
	UpdatedAt       *gtime.Time // 更新时间
	CompletedAt     *gtime.Time // 完成时间
}
