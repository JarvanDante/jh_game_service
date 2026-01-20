package one

import (
	"context"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"jh_game_service/internal/model/entity"
	"jh_game_service/internal/util"
	"math/rand"

	"github.com/gogf/gf/v2/frame/g"
)

type (
	sOne struct{}
)

// New 创建ONE游戏服务实例
func New() *sOne {
	return &sOne{}
}

// 请求data
type OneGameReq struct {
	Username  string `json:"username"`
	TraceId   string `json:"traceId"`
	GameCode  string `json:"gameCode"`
	Language  string `json:"language"`
	Platform  string `json:"platform"`
	Currency  string `json:"currency"`
	LobbyUrl  string `json:"lobbyUrl"`
	IPAddress string `json:"ipAddress"`
}

// GetGameUrl 获取ONE游戏跳转链接
func (s *sOne) GetGameUrl(ctx context.Context, url string, headers map[string]string, data any) (map[string]interface{}, error) {
	res, err := util.CurlPostHeader(url, headers, data)
	if err != nil {
		return nil, fmt.Errorf("游戏API请求失败: %v", err)
	}

	// ONE游戏API固定返回格式: {data: {gameUrl: "xxx", token: "xxx"}}
	dataValue, exists := res["data"]
	if !exists {
		return nil, fmt.Errorf("游戏API返回格式错误，缺少data字段")
	}

	dataMap, ok := dataValue.(map[string]interface{})
	if !ok {
		return nil, fmt.Errorf("游戏API返回格式错误，data字段类型不正确")
	}

	gameUrl, ok := dataMap["gameUrl"].(string)
	if !ok || gameUrl == "" {
		return nil, fmt.Errorf("游戏API返回格式错误，无法获取游戏链接")
	}

	return map[string]interface{}{
		"type": "url",
		"url":  gameUrl,
	}, nil
}

// AutoTransferBalance 自动转账到游戏
// TODO: 需要调用balance服务进行转账
// 建议使用消息队列异步处理，避免阻塞用户进入游戏
func (s *sOne) AutoTransferBalance(ctx context.Context, username string, siteGame *entity.SiteGame) error {
	g.Log().Infof(ctx, "ONE游戏 - 自动转账: username=%s, gameId=%d", username, siteGame.GameId)

	// TODO: 实现自动转账逻辑
	// 1. 检查用户是否开启自动转账
	// 2. 获取用户主钱包余额
	// 3. 调用balance服务进行转账
	// 4. 建议使用消息队列异步处理：
	//    - 发送转账请求到队列
	//    - balance服务消费队列处理转账
	//    - 转账结果通过回调或轮询获取

	g.Log().Warning(ctx, "自动转账功能待实现 - 需要集成balance服务和消息队列")
	return nil
}

//BuildUsername
/**
 * @desc：游戏用户名转带应用code
 * @param username
 * @param siteCode
 * @return string
 * @return error
 * @author : Carson
 */
func (s *sOne) BuildUsername(username string, siteCode string) (string, error) {
	newUsername := siteCode + "_" + username
	return newUsername, nil
}

//GenerateUUIDv4
/**
 * @desc：生成 UUID v4
 * @return string
 * @return error
 * @author : Carson
 */
func (s *sOne) GenerateUUIDv4() (string, error) {
	b := make([]byte, 16)
	_, err := rand.Read(b)
	if err != nil {
		return "", err
	}

	// 设置版本号 (4)
	b[6] = (b[6] & 0x0f) | 0x40
	// 设置 variant (RFC4122)
	b[8] = (b[8] & 0x3f) | 0x80

	uuid := fmt.Sprintf("%08x-%04x-%04x-%04x-%012x",
		b[0:4],
		b[4:6],
		b[6:8],
		b[8:10],
		b[10:16],
	)

	return uuid, nil
}

//GetONESignature
/**
 * @desc：ONE签名
 * @param data
 * @param secretKey
 * @return string
 * @return error
 * @author : Carson
 */
func (s *sOne) GetONESignature(data any, secretKey string) (string, error) {
	// 等价 json_encode($data)
	b, err := json.Marshal(data)
	if err != nil {
		return "", err
	}

	h := hmac.New(sha256.New, []byte(secretKey))
	h.Write(b)

	// PHP hash_hmac 默认输出 hex string
	return hex.EncodeToString(h.Sum(nil)), nil
}
