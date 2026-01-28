package sync_task

import (
	"context"
	"fmt"
	"jh_game_service/internal/dao"
	"jh_game_service/internal/model/do"
	"jh_game_service/internal/model/entity"

	"github.com/gogf/gf/v2/frame/g"
	"github.com/gogf/gf/v2/os/gtime"
)

// CreateTask 创建同步任务
func CreateTask(ctx context.Context, siteId int, siteCode, platform string, taskType int, startTime, endTime string) (int64, error) {
	// 解析时间
	start, err := gtime.StrToTime(startTime)
	if err != nil {
		return 0, fmt.Errorf("解析开始时间失败: %v", err)
	}
	end, err := gtime.StrToTime(endTime)
	if err != nil {
		return 0, fmt.Errorf("解析结束时间失败: %v", err)
	}

	task := &do.SyncTask{
		SiteId:    siteId,
		SiteCode:  siteCode,
		Platform:  platform,
		TaskType:  taskType,
		StartTime: start,
		EndTime:   end,
		Status:    1, // 进行中
		CreatedAt: gtime.Now(),
		UpdatedAt: gtime.Now(),
	}

	result, err := dao.SyncTask.Ctx(ctx).Insert(task)
	if err != nil {
		return 0, fmt.Errorf("创建同步任务失败: %v", err)
	}

	taskId, err := result.LastInsertId()
	if err != nil {
		return 0, fmt.Errorf("获取任务ID失败: %v", err)
	}

	g.Log().Infof(ctx, "创建同步任务成功: task_id=%d, site=%s, platform=%s, type=%d",
		taskId, siteCode, platform, taskType)

	return taskId, nil
}

// UpdateMySQLStats 更新MySQL统计
func UpdateMySQLStats(ctx context.Context, taskId int64, total, inserted, duplicated int) error {
	_, err := dao.SyncTask.Ctx(ctx).
		Where("id", taskId).
		Data(do.SyncTask{
			MysqlTotal:      total,
			MysqlInserted:   inserted,
			MysqlDuplicated: duplicated,
			UpdatedAt:       gtime.Now(),
		}).
		Update()

	if err != nil {
		return fmt.Errorf("更新MySQL统计失败: %v", err)
	}

	g.Log().Infof(ctx, "更新MySQL统计: task_id=%d, total=%d, inserted=%d, duplicated=%d",
		taskId, total, inserted, duplicated)

	return nil
}

// UpdateMQStats 更新MQ统计
func UpdateMQStats(ctx context.Context, taskId int64, sent, failed, skipped int) error {
	_, err := dao.SyncTask.Ctx(ctx).
		Where("id", taskId).
		Data(do.SyncTask{
			MqSent:    sent,
			MqFailed:  failed,
			MqSkipped: skipped,
			UpdatedAt: gtime.Now(),
		}).
		Update()

	if err != nil {
		return fmt.Errorf("更新MQ统计失败: %v", err)
	}

	g.Log().Infof(ctx, "更新MQ统计: task_id=%d, sent=%d, failed=%d, skipped=%d",
		taskId, sent, failed, skipped)

	return nil
}

// UpdateClickHouseStats 更新ClickHouse统计（由消费者调用）
func UpdateClickHouseStats(ctx context.Context, taskId int64, synced, failed int) error {
	_, err := dao.SyncTask.Ctx(ctx).
		Where("id", taskId).
		Data(do.SyncTask{
			ChSynced:  synced,
			ChFailed:  failed,
			UpdatedAt: gtime.Now(),
		}).
		Update()

	if err != nil {
		return fmt.Errorf("更新ClickHouse统计失败: %v", err)
	}

	return nil
}

// IncrementClickHouseSynced 增加ClickHouse已同步数（原子操作）
func IncrementClickHouseSynced(ctx context.Context, taskId int64, count int) error {
	_, err := dao.SyncTask.Ctx(ctx).
		Where("id", taskId).
		Increment("ch_synced", count)

	if err != nil {
		return fmt.Errorf("增加ClickHouse同步数失败: %v", err)
	}

	return nil
}

// IncrementClickHouseFailed 增加ClickHouse失败数（原子操作）
func IncrementClickHouseFailed(ctx context.Context, taskId int64, count int) error {
	_, err := dao.SyncTask.Ctx(ctx).
		Where("id", taskId).
		Increment("ch_failed", count)

	if err != nil {
		return fmt.Errorf("增加ClickHouse失败数失败: %v", err)
	}

	return nil
}

// CompleteTask 完成任务
func CompleteTask(ctx context.Context, taskId int64, status int, errorMsg string) error {
	updateData := do.SyncTask{
		Status:      status,
		UpdatedAt:   gtime.Now(),
		CompletedAt: gtime.Now(),
	}

	if errorMsg != "" {
		updateData.ErrorMsg = errorMsg
	}

	_, err := dao.SyncTask.Ctx(ctx).
		Where("id", taskId).
		Data(updateData).
		Update()

	if err != nil {
		return fmt.Errorf("完成任务失败: %v", err)
	}

	g.Log().Infof(ctx, "任务完成: task_id=%d, status=%d", taskId, status)

	return nil
}

// GetTask 获取任务详情
func GetTask(ctx context.Context, taskId int64) (*entity.SyncTask, error) {
	var task entity.SyncTask
	err := dao.SyncTask.Ctx(ctx).
		Where("id", taskId).
		Scan(&task)

	if err != nil {
		return nil, fmt.Errorf("获取任务失败: %v", err)
	}

	return &task, nil
}

// GetTaskList 获取任务列表
func GetTaskList(ctx context.Context, siteId int, platform string, taskType int, page, pageSize int) ([]*entity.SyncTask, int, error) {
	query := dao.SyncTask.Ctx(ctx)

	// 过滤条件
	if siteId > 0 {
		query = query.Where("site_id", siteId)
	}
	if platform != "" {
		query = query.Where("platform", platform)
	}
	if taskType > 0 {
		query = query.Where("task_type", taskType)
	}

	// 获取总数
	total, err := query.Count()
	if err != nil {
		return nil, 0, fmt.Errorf("获取任务总数失败: %v", err)
	}

	// 分页查询
	var tasks []*entity.SyncTask
	err = query.
		Order("id DESC").
		Limit(pageSize).
		Offset((page - 1) * pageSize).
		Scan(&tasks)

	if err != nil {
		return nil, 0, fmt.Errorf("获取任务列表失败: %v", err)
	}

	return tasks, total, nil
}
