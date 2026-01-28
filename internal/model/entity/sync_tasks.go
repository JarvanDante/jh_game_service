// =================================================================================
// Code generated and maintained by GoFrame CLI tool. DO NOT EDIT.
// =================================================================================

package entity

import (
	"github.com/gogf/gf/v2/os/gtime"
)

// SyncTasks is the golang structure for table sync_tasks.
type SyncTasks struct {
	Id              uint64      `json:"id"              orm:"id"               description:"任务ID"`
	SiteId          int         `json:"siteId"          orm:"site_id"          description:"站点ID"`
	SiteCode        string      `json:"siteCode"        orm:"site_code"        description:"站点代码"`
	Platform        string      `json:"platform"        orm:"platform"         description:"游戏平台"`
	TaskType        int         `json:"taskType"        orm:"task_type"        description:"任务类型: 1=定时拉单, 2=手动补单"`
	StartTime       *gtime.Time `json:"startTime"       orm:"start_time"       description:"拉单开始时间"`
	EndTime         *gtime.Time `json:"endTime"         orm:"end_time"         description:"拉单结束时间"`
	MysqlTotal      int         `json:"mysqlTotal"      orm:"mysql_total"      description:"MySQL总记录数"`
	MysqlInserted   int         `json:"mysqlInserted"   orm:"mysql_inserted"   description:"MySQL新插入数"`
	MysqlDuplicated int         `json:"mysqlDuplicated" orm:"mysql_duplicated" description:"MySQL重复跳过数"`
	MqSent          int         `json:"mqSent"          orm:"mq_sent"          description:"MQ发送成功数"`
	MqFailed        int         `json:"mqFailed"        orm:"mq_failed"        description:"MQ发送失败数"`
	MqSkipped       int         `json:"mqSkipped"       orm:"mq_skipped"       description:"MQ跳过数(未结算或无user_id)"`
	ChSynced        int         `json:"chSynced"        orm:"ch_synced"        description:"ClickHouse已同步数"`
	ChFailed        int         `json:"chFailed"        orm:"ch_failed"        description:"ClickHouse同步失败数"`
	Status          int         `json:"status"          orm:"status"           description:"任务状态: 1=进行中, 2=已完成, 3=失败"`
	ErrorMsg        string      `json:"errorMsg"        orm:"error_msg"        description:"错误信息"`
	CreatedAt       *gtime.Time `json:"createdAt"       orm:"created_at"       description:"创建时间"`
	UpdatedAt       *gtime.Time `json:"updatedAt"       orm:"updated_at"       description:"更新时间"`
	CompletedAt     *gtime.Time `json:"completedAt"     orm:"completed_at"     description:"完成时间"`
}
