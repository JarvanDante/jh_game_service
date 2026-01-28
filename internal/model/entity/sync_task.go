package entity

import "time"

// SyncTask 同步任务实体
type SyncTask struct {
	Id        int64      `json:"id"         description:"任务ID"`
	SiteId    int        `json:"site_id"    description:"站点ID"`
	SiteCode  string     `json:"site_code"  description:"站点代码"`
	Platform  string     `json:"platform"   description:"游戏平台"`
	TaskType  int        `json:"task_type"  description:"任务类型: 1=定时拉单, 2=手动补单"`
	StartTime *time.Time `json:"start_time" description:"拉单开始时间"`
	EndTime   *time.Time `json:"end_time"   description:"拉单结束时间"`

	// MySQL统计
	MysqlTotal      int `json:"mysql_total"       description:"MySQL总记录数"`
	MysqlInserted   int `json:"mysql_inserted"    description:"MySQL新插入数"`
	MysqlDuplicated int `json:"mysql_duplicated"  description:"MySQL重复跳过数"`

	// MQ统计
	MqSent    int `json:"mq_sent"    description:"MQ发送成功数"`
	MqFailed  int `json:"mq_failed"  description:"MQ发送失败数"`
	MqSkipped int `json:"mq_skipped" description:"MQ跳过数"`

	// ClickHouse统计
	ChSynced int `json:"ch_synced" description:"ClickHouse已同步数"`
	ChFailed int `json:"ch_failed" description:"ClickHouse同步失败数"`

	// 任务状态
	Status   int    `json:"status"    description:"任务状态: 1=进行中, 2=已完成, 3=失败"`
	ErrorMsg string `json:"error_msg" description:"错误信息"`

	// 时间戳
	CreatedAt   *time.Time `json:"created_at"   description:"创建时间"`
	UpdatedAt   *time.Time `json:"updated_at"   description:"更新时间"`
	CompletedAt *time.Time `json:"completed_at" description:"完成时间"`
}
