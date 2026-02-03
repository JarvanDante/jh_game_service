package domain

import (
	"fmt"

	v1 "jh_game_service/api/backend/game/v1"
)

// BuildUserBetSummary 组装用户注单汇总响应
func BuildUserBetSummary(startTime, endTime string, resultBetCount int64, resultBetAmount, resultWinAmount, resultValidAmount, resultNetAmount float64) *v1.UserBetSummary {
	return &v1.UserBetSummary{
		StartTime:        startTime,
		EndTime:          endTime,
		BetCount:         fmt.Sprintf("%d", resultBetCount),
		BetAmount:        fmt.Sprintf("%.2f", resultBetAmount),
		WinAmount:        fmt.Sprintf("%.2f", resultWinAmount),
		ValidAmount:      fmt.Sprintf("%.2f", resultValidAmount),
		NetAmount:        fmt.Sprintf("%.2f", resultNetAmount),
		TurnoverMultiple: "0.00",
	}
}
