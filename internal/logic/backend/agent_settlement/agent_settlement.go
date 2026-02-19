package agent_settlement

import (
	"context"
	"fmt"
	v1 "jh_game_service/api/backend/game/v1"
	"jh_game_service/internal/dao"
	"jh_game_service/internal/middleware"
	"jh_game_service/internal/registry"
	"jh_game_service/internal/service/backend"

	"github.com/gogf/gf/v2/frame/g"
)

type sAgentSettlement struct{}

func init() { backend.RegisterAgentSettlement(&sAgentSettlement{}) }

func (s *sAgentSettlement) GetAgentDailyBetStats(ctx context.Context, req *v1.GetAgentDailyBetStatsReq) (*v1.GetAgentDailyBetStatsRes, error) {
	siteId, ok := middleware.GetSiteIdFromContext(ctx)
	if !ok || siteId == 0 {
		return nil, fmt.Errorf("未获取站点ID")
	}
	if req.StartDate == "" || req.EndDate == "" {
		return nil, fmt.Errorf("开始日期和结束日期不能为空")
	}
	if req.StartDate > req.EndDate {
		return nil, fmt.Errorf("开始日期不能大于结束日期")
	}

	// 优先走 ClickHouse 聚合，性能更好；失败时回退 MySQL，保证可用性
	if res, err := s.queryFromClickHouse(ctx, uint32(siteId), req); err == nil {
		g.Log().Infof(ctx, "代理结算统计数据源: ClickHouse, site_id=%d, start=%s, end=%s, agent_id=%d", siteId, req.StartDate, req.EndDate, req.AgentId)
		return res, nil
	} else {
		g.Log().Warningf(ctx, "ClickHouse代理结算统计失败，回退MySQL: %v", err)
	}

	g.Log().Infof(ctx, "代理结算统计数据源: MySQL(回退), site_id=%d, start=%s, end=%s, agent_id=%d", siteId, req.StartDate, req.EndDate, req.AgentId)
	return s.queryFromMySQL(ctx, siteId, req)
}

func (s *sAgentSettlement) queryFromClickHouse(ctx context.Context, siteId uint32, req *v1.GetAgentDailyBetStatsReq) (*v1.GetAgentDailyBetStatsRes, error) {
	chManager := registry.GetClickHouseManager()
	conn := chManager.GetConn()
	if conn == nil {
		return nil, fmt.Errorf("ClickHouse连接未初始化")
	}

	query := `
		SELECT
			toString(toDate(bet_time)) AS stat_date,
			agent_id,
			toFloat64(sum(bet_amount)) AS bet_amount,
			toFloat64(sum(valid_amount)) AS valid_bet_amount,
			toFloat64(sum(net_amount)) AS win_loss,
			toInt64(uniqExact(user_id)) AS member_count
		FROM game_analytics.bet_records
		WHERE site_id = ?
		  AND toDate(bet_time) >= toDate(?)
		  AND toDate(bet_time) <= toDate(?)
	`
	args := []interface{}{siteId, req.StartDate, req.EndDate}
	if req.AgentId > 0 {
		query += " AND agent_id = ?"
		args = append(args, req.AgentId)
	}
	query += " GROUP BY toDate(bet_time), agent_id ORDER BY toDate(bet_time) ASC, agent_id ASC"

	rows, err := conn.Query(ctx, query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	res := &v1.GetAgentDailyBetStatsRes{List: make([]*v1.AgentDailyBetStatItem, 0)}
	for rows.Next() {
		var row v1.AgentDailyBetStatItem
		if err = rows.Scan(&row.StatDate, &row.AgentId, &row.BetAmount, &row.ValidBetAmount, &row.WinLoss, &row.MemberCount); err != nil {
			return nil, err
		}
		res.List = append(res.List, &row)
	}
	return res, nil
}

func (s *sAgentSettlement) queryFromMySQL(ctx context.Context, siteId uint, req *v1.GetAgentDailyBetStatsReq) (*v1.GetAgentDailyBetStatsRes, error) {
	query := dao.BetRecords.Ctx(ctx).
		Where("site_id", siteId).
		Where("DATE(bet_time) >= ?", req.StartDate).
		Where("DATE(bet_time) <= ?", req.EndDate)
	if req.AgentId > 0 {
		query = query.Where("agent_id", req.AgentId)
	}

	var rows []struct {
		StatDate       string  `json:"stat_date"`
		AgentId        int64   `json:"agent_id"`
		BetAmount      float64 `json:"bet_amount"`
		ValidBetAmount float64 `json:"valid_bet_amount"`
		WinLoss        float64 `json:"win_loss"`
		MemberCount    int64   `json:"member_count"`
	}
	err := query.Fields(`
		DATE(bet_time) AS stat_date,
		agent_id,
		COALESCE(SUM(bet_amount),0) AS bet_amount,
		COALESCE(SUM(valid_amount),0) AS valid_bet_amount,
		COALESCE(SUM(net_amount),0) AS win_loss,
		COUNT(DISTINCT user_id) AS member_count
	`).Group("DATE(bet_time), agent_id").Order("stat_date ASC, agent_id ASC").Scan(&rows)
	if err != nil {
		return nil, err
	}

	res := &v1.GetAgentDailyBetStatsRes{List: make([]*v1.AgentDailyBetStatItem, 0, len(rows))}
	for _, row := range rows {
		res.List = append(res.List, &v1.AgentDailyBetStatItem{
			StatDate:       row.StatDate,
			AgentId:        row.AgentId,
			BetAmount:      row.BetAmount,
			ValidBetAmount: row.ValidBetAmount,
			WinLoss:        row.WinLoss,
			MemberCount:    row.MemberCount,
		})
	}
	return res, nil
}
