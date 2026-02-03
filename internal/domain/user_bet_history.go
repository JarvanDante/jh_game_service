package domain

import (
	"math"

	v1 "jh_game_service/api/frontend/game/v1"
	"jh_game_service/internal/dao"
)

// BuildUserBetHistoryRes 组装用户历史投注响应
func BuildUserBetHistoryRes(startTime, endTime string, rows []dao.UserBetHistoryRow, total int) *v1.GetUserBetHistoryRes {
	items := make([]*v1.UserBetHistoryItem, 0, len(rows))
	for _, row := range rows {
		betTime := ""
		if row.BetTime != nil {
			betTime = row.BetTime.Format("2006-01-02 15:04:05")
		}

		items = append(items, &v1.UserBetHistoryItem{
			GameName:    row.GameName,
			ValidAmount: trunc2(row.ValidAmount),
			BetTime:     betTime,
			WinAmount:   trunc2(row.WinAmount),
			Status:      int32(row.Status),
			OrderNo:     row.OrderNo,
			NetAmount:   trunc2(row.NetAmount),
		})
	}

	return &v1.GetUserBetHistoryRes{
		List:  items,
		Total: int32(total),
	}
}

func trunc2(v float64) float64 {
	return math.Trunc(v*100) / 100
}
