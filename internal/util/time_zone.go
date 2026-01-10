package util

import (
	"time"
)

// convertESTToTargetTimezone 将美东时间转换为目标时区
// estTime: 美东时间
// targetTimezone: 目标时区字符串，如 "Asia/Shanghai"
func ConvertESTToTargetTimezone(estTime time.Time, targetTimezone string) time.Time {
	// 加载美东时区
	est, err := time.LoadLocation("America/New_York")
	if err != nil {
		// 如果加载失败，使用固定偏移（美东8区 = UTC-5）
		est = time.FixedZone("EST", -5*3600)
	}

	// 加载目标时区
	target, err := time.LoadLocation(targetTimezone)
	if err != nil {
		// 如果加载失败，返回原时间
		return estTime
	}

	// 将时间从美东时区转换到目标时区
	return estTime.In(est).In(target)
}

// 如果目标时区固定，可以简化版本
func ConvertESTToShanghai(estTime time.Time) time.Time {
	est, _ := time.LoadLocation("America/New_York")
	shanghai, _ := time.LoadLocation("Asia/Shanghai")
	return estTime.In(est).In(shanghai)
}

// 或者用固定偏移量（更稳定，不依赖系统时区数据库）
// 美东8区 = UTC-5，上海 = UTC+8，差值 = 13小时
func ConvertESTToShanghaiFixed(estTime time.Time) time.Time {
	return estTime.Add(13 * time.Hour)
}

//使用方式
// 入库时调用
//func SaveOrder(thirdPartyOrder) {
//	// 方式1：使用时区名称
//	targetTime := ConvertESTToTargetTimezone(thirdPartyOrder.CreatedAt, "Asia/Shanghai")
//
//	// 方式2：使用简化版本
//	// targetTime := ConvertESTToShanghai(thirdPartyOrder.CreatedAt)
//
//	// 方式3：使用固定偏移（推荐，最稳定）
//	// targetTime := ConvertESTToShanghaiFixed(thirdPartyOrder.CreatedAt)
//
//	order.CreatedAt = targetTime
//	db.Save(order)
//}
