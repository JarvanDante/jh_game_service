package agent_settlement

import (
	"context"
	"fmt"
	v1 "jh_game_service/api/grpc/agent_settlement/v1"
	"jh_game_service/internal/dao"
	"jh_game_service/internal/middleware"
	"jh_game_service/internal/service/backend"
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
