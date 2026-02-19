package backend

import (
	"context"
	v1 "jh_game_service/api/grpc/agent_settlement/v1"
)

type IAgentSettlement interface {
	GetAgentDailyBetStats(ctx context.Context, req *v1.GetAgentDailyBetStatsReq) (*v1.GetAgentDailyBetStatsRes, error)
}

var localAgentSettlement IAgentSettlement

func AgentSettlement() IAgentSettlement {
	if localAgentSettlement == nil {
		panic("implement not found for interface IAgentSettlement, forgot register?")
	}
	return localAgentSettlement
}

func RegisterAgentSettlement(i IAgentSettlement) {
	localAgentSettlement = i
}
