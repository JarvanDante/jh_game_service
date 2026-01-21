package one

import (
	"context"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	v1 "jh_game_service/api/frontend/game/v1"
	"jh_game_service/internal/model/entity"
	"jh_game_service/internal/util"
	"math/rand"
	"strings"

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
type oneUrlReq struct {
	Username  string `json:"username"`
	TraceId   string `json:"traceId"`
	GameCode  string `json:"gameCode"`
	Language  string `json:"language"`
	Platform  string `json:"platform"`
	Currency  string `json:"currency"`
	LobbyUrl  string `json:"lobbyUrl"`
	IPAddress string `json:"ipAddress"`
}

type oneToGameReq struct {
	Username       string  `json:"username"`
	TraceId        string  `json:"traceId"`
	TransferAmount float64 `json:"transferAmount"`
	Currency       string  `json:"currency"`
	referenceId    string  `json:"referenceId"`
}

//GetGameUrl
/**
 * @desc：获取ONE游戏跳转链接
 * @param req
 * @param siteGame
 * @param siteCode
 * @param username
 * @param currency
 * @param domain
 * @param clientIp
 * @return map[string]interface{}
 * @return error
 * @author : Carson
 */
func (s *sOne) GetGameUrl(req *v1.RedirectToGameReq, siteId int, userId uint, siteGame entity.SiteGame, siteCode string, username string, currency string, domain string, clientIp string) (map[string]interface{}, error) {

	// 构建游戏用户名
	buildUsername, err := s.buildUsername(username, siteCode)
	if err != nil {
		return nil, err
	}

	// 生成TraceID
	traceId, _ := s.generateUUIDv4()

	// 构建请求数据
	data := oneUrlReq{
		Username:  buildUsername,
		TraceId:   traceId,
		GameCode:  req.Key,
		Language:  "en",
		Platform:  s.getPlatformType(req.Mobile),
		Currency:  currency,
		LobbyUrl:  domain,
		IPAddress: clientIp,
	}

	// 生成签名
	signature, err := s.getONESignature(data, siteGame.ApiKey)
	if err != nil {
		return nil, err
	}

	// 构建请求头
	headers := map[string]string{
		"Content-Type": "application/json",
		"X-API-Key":    siteGame.ApiId,
		"X-Signature":  signature,
	}

	// 构建API URL
	url := strings.TrimSuffix(siteGame.ApiUrl, "/") + "/game/url"

	// 转入游戏
	amount := 180.00
	orderNo := fmt.Sprintf("GAME_%d_%d_%s", siteId, userId, traceId)

	// TODO: 调用balance服务的AdjustBalance方法进行扣款
	// 参数：
	// - user_id: userId
	// - amount: -amount (负数表示扣款)
	// - type: 4 (游戏转出)
	// - remark: fmt.Sprintf("转入%s游戏", siteGame.Name)
	// - order_no: orderNo
	// - game_id: siteGame.GameId

	// 调用三方游戏API进行转账
	err = s.transferToOne(siteGame, buildUsername, traceId, amount, currency, orderNo)
	if err != nil {
		// 失败：调用balance服务的CompensateBalance方法补偿余额
		// TODO: 调用 CompensateBalance(user_id, order_no, amount)
		return nil, err
	}

	// 成功：调用balance服务的UpdateTransferStatus方法更新状态
	// TODO: 调用 UpdateTransferStatus(order_no, status=2)

	// 进入游戏
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

func (s *sOne) transferToOne(siteGame entity.SiteGame, buildUsername string, traceId string, amount float64, currency string, orderNo string) error {

	// 构建请求数据
	data := oneToGameReq{
		Username:       buildUsername,
		TraceId:        traceId,
		TransferAmount: amount,
		Currency:       currency,
		referenceId:    orderNo,
	}

	// 生成签名
	signature, err := s.getONESignature(data, siteGame.ApiKey)
	if err != nil {
		return err
	}

	// 构建请求头
	headers := map[string]string{
		"Content-Type": "application/json",
		"X-API-Key":    siteGame.ApiId,
		"X-Signature":  signature,
	}

	// 构建API URL
	url := strings.TrimSuffix(siteGame.ApiUrl, "/") + "/cash/deposit"

	res, err := util.CurlPostHeader(url, headers, data)
	if err != nil {
		return fmt.Errorf("游戏API请求失败: %v", err)
	}

	if res["status"] == "SC_INVALID_OPERATOR" {
		return fmt.Errorf("未经授权的访问")
	}
	if res["status"] == "SC_INVALID_SIGNATURE" {
		return fmt.Errorf("验证失败")
	}
	if res["status"] == "SC_WRONG_CURRENCY" {
		return fmt.Errorf("不支持该货币")
	}
	if res["status"] == "SC_OK" {
		dataValue, exists := res["data"]
		if !exists {
			return fmt.Errorf("游戏API返回格式错误，缺少data字段")
		}
		_, ok := dataValue.(map[string]interface{})
		if !ok {
			return fmt.Errorf("游戏API返回格式错误，data字段类型不正确")
		}

		return nil
	}
	return fmt.Errorf("余额转入失败")
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
func (s *sOne) buildUsername(username string, siteCode string) (string, error) {
	newUsername := siteCode + "_" + username
	return newUsername, nil
}

//getPlatformType
/**
 * @desc：根据mobile参数返回平台类型
 * @param mobile
 * @return string
 * @author : Carson
 */
func (s *sOne) getPlatformType(mobile int32) string {
	if mobile > 0 {
		return "H5"
	}
	return "WEB"
}

//generateUUIDv4
/**
 * @desc：生成 UUID v4
 * @return string
 * @return error
 * @author : Carson
 */
func (s *sOne) generateUUIDv4() (string, error) {
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
func (s *sOne) getONESignature(data any, secretKey string) (string, error) {
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
