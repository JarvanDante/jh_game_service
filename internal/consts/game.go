package consts

// GameType 游戏类型常量
const (
	GameTypeSport   = 1 // 体育
	GameTypeLottery = 2 // 彩票
	GameTypeLive    = 3 // 真人
	GameTypeSlot    = 4 // 电子/老虎机
	GameTypePoker   = 5 // 棋牌
)

// GameStatus 游戏状态
const (
	GameStatusDisabled = 0 // 禁用
	GameStatusEnabled  = 1 // 启用
)

// GameTypeNameMap 游戏类型名称映射
var GameTypeNameMap = map[int]string{
	GameTypeSport:   "体育",
	GameTypeLottery: "彩票",
	GameTypeLive:    "真人",
	GameTypeSlot:    "电子",
	GameTypePoker:   "棋牌",
}

// GameTypeNameEnMap 游戏类型英文名称映射
var GameTypeNameEnMap = map[int]string{
	GameTypeSport:   "Sport",
	GameTypeLottery: "Lottery",
	GameTypeLive:    "Live",
	GameTypeSlot:    "Slot",
	GameTypePoker:   "Poker",
}

// GameStatusNameMap 游戏状态名称映射
var GameStatusNameMap = map[int]string{
	GameStatusDisabled: "禁用",
	GameStatusEnabled:  "启用",
}

// GameStatusNameEnMap 游戏状态英文名称映射
var GameStatusNameEnMap = map[int]string{
	GameStatusDisabled: "Disabled",
	GameStatusEnabled:  "Enabled",
}

// GetGameTypeName 获取游戏类型名称
func GetGameTypeName(gameType int) string {
	if name, ok := GameTypeNameMap[gameType]; ok {
		return name
	}
	return "未知"
}

// GetGameTypeNameEn 获取游戏类型英文名称
func GetGameTypeNameEn(gameType int) string {
	if name, ok := GameTypeNameEnMap[gameType]; ok {
		return name
	}
	return "Unknown"
}

// GetGameStatusName 获取游戏状态名称
func GetGameStatusName(status int) string {
	if name, ok := GameStatusNameMap[status]; ok {
		return name
	}
	return "未知"
}

// GetGameStatusNameEn 获取游戏状态英文名称
func GetGameStatusNameEn(status int) string {
	if name, ok := GameStatusNameEnMap[status]; ok {
		return name
	}
	return "Unknown"
}

// IsValidGameType 验证游戏类型是否有效
func IsValidGameType(gameType int) bool {
	_, ok := GameTypeNameMap[gameType]
	return ok
}
