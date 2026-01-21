package one

import (
	"context"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	v1 "jh_game_service/api/frontend/game/v1"
	balanceV1 "jh_game_service/api/grpc/balance/v1"
	"jh_game_service/internal/model/entity"
	"jh_game_service/internal/util"
	"math/rand"
	"strconv"
	"strings"
	"time"

	"github.com/gogf/gf/v2/frame/g"
	"google.golang.org/grpc"
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
	TransferAmount float32 `json:"transferAmount"`
	Currency       string  `json:"currency"`
	ReferenceId    string  `json:"referenceId"`
}

type oneBalanceReq struct {
	Username string `json:"username"`
	TraceId  string `json:"traceId"`
	Currency string `json:"currency"`
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
func (s *sOne) GetGameUrl(ctx context.Context, req *v1.RedirectToGameReq, siteId int, userId uint, siteGame entity.SiteGame, siteCode string, username string, currency string, domain string, clientIp string) (map[string]interface{}, error) {

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

	// 生成UUID作为订单号（三方API要求UUID格式）
	orderNo, _ := s.generateUUIDv4()

	balanceClient := s.GetBalanceClient()
	// 成功：更新转账状态
	lastRes, lastErr := balanceClient.GetLastBalanceLogGame(ctx, &balanceV1.GetLastBalanceLogGameReq{
		SiteId: int32(siteId),
		UserId: int32(userId),
	})
	if lastErr != nil {
		return nil, fmt.Errorf("获取最后一次进入游戏失败： %v ", lastErr)
	}

	var amount float32
	if lastRes.GameId == 0 {
		g.Log().Infof(ctx, "用户从来没有进入游戏，取用户余额")
		amount = lastRes.Amount //用户余额
	} else {
		//不等于自己
		if lastRes.GameId != req.GameId {
			g.Log().Infof(ctx, "用户之前进入游戏ID: %d", lastRes.GameId)
			amount, err = s.GetBalance(ctx, siteGame, siteCode, username, currency)
		} else {
			// 等于自己直接进入游戏
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

	}

	// 调用balance服务的AdjustBalance方法进行扣款
	g.Log().Infof(ctx, "准备调用balance服务扣款: siteId=%d, userId=%d, amount=%.2f, gameName=%s",
		siteId, userId, amount, siteGame.Name)
	// 设置超时时间为10秒
	callCtx, cancel := context.WithTimeout(ctx, 10*time.Second)
	defer cancel()

	g.Log().Infof(ctx, "开始调用balance服务AdjustBalance方法")

	adjustRes, err := balanceClient.AdjustBalance(callCtx, &balanceV1.AdjustBalanceReq{
		SiteId:   int32(siteId),
		UserId:   int32(userId),
		Type:     0, // 0=转出：从中心账户转账到游戏账户
		GameName: siteGame.Name,
		GameId:   strconv.Itoa(siteGame.GameId),
	})

	if err != nil {
		g.Log().Errorf(ctx, "调用balance服务失败: %v", err)
	} else {
		g.Log().Infof(ctx, "调用balance服务成功: success=%v, message=%s", adjustRes.Success, adjustRes.Message)
	}

	g.Log().Infof(ctx, "balance服务调用完成: err=%v", err)

	if err != nil {
		return nil, fmt.Errorf("调用balance服务失败: %v", err)
	}

	if !adjustRes.Success {
		return nil, fmt.Errorf("扣款失败: %s", adjustRes.Message)
	}

	// 调用三方游戏API进行转账
	err = s.transferToOne(ctx, siteGame, buildUsername, traceId, amount, currency, orderNo)
	if err != nil {
		// 失败：补偿余额
		compensateRes, compensateErr := balanceClient.CompensateBalance(ctx, &balanceV1.CompensateBalanceReq{
			SiteId:  int32(siteId),
			UserId:  int32(userId),
			OrderNo: adjustRes.OrderNo,
			Amount:  float32(amount), // 传入正数，由type参数判断
		})
		if compensateErr != nil {
			g.Log().Errorf(ctx, "补偿余额失败: %v", compensateErr)
		} else if !compensateRes.Success {
			g.Log().Errorf(ctx, "补偿余额失败: %s", compensateRes.Message)
		}
		return nil, err
	}

	// 成功：更新转账状态
	updateRes, updateErr := balanceClient.UpdateTransferStatus(ctx, &balanceV1.UpdateTransferStatusReq{
		OrderNo: adjustRes.OrderNo,
	})
	if updateErr != nil {
		g.Log().Errorf(ctx, "更新转账状态失败: %v", updateErr)
	} else if !updateRes.Success {
		g.Log().Errorf(ctx, "更新转账状态失败: %s", updateRes.Message)
	}

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

//transferToOne
/**
 * @desc：ONE转入三方钱包
 * @param ctx
 * @param siteGame
 * @param buildUsername
 * @param traceId
 * @param amount
 * @param currency
 * @param orderNo
 * @return error
 * @author : Carson
 */
func (s *sOne) transferToOne(ctx context.Context, siteGame entity.SiteGame, buildUsername string, traceId string, amount float32, currency string, orderNo string) error {

	// 构建请求数据
	data := oneToGameReq{
		Username:       buildUsername,
		TraceId:        traceId,
		TransferAmount: amount,
		Currency:       currency,
		ReferenceId:    orderNo,
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

	g.Log().Infof(ctx, "转入请求参数: %v,%v,%v", url, headers, data)
	res, err := util.CurlPostHeader(url, headers, data)
	if err != nil {
		return fmt.Errorf("游戏API请求失败: %v", err)
	}

	g.Log().Infof(ctx, "转入请求结果: %v", res)

	if res["status"] == "SC_INVALID_OPERATOR" {
		return fmt.Errorf("未经授权的访问")
	}
	if res["status"] == "SC_INVALID_SIGNATURE" {
		return fmt.Errorf("验证失败")
	}
	if res["status"] == "SC_WRONG_CURRENCY" {
		return fmt.Errorf("不支持该货币")
	}
	if res["status"] == "SC_INVALID_REQUEST" {
		return fmt.Errorf("请求无效")
	}
	if res["status"] == "SC_UNKNOWN_ERROR" {
		return fmt.Errorf("未知错误")
	}
	if res["status"] == "SC_DUPLICATE_REQUEST" {
		return fmt.Errorf("请求重复")
	}
	if res["status"] == "SC_REFERENCE_ID_DUPLICATED" {
		return fmt.Errorf("已发送重复的参考号")
	}
	if res["status"] == "SC_WALLET_NOT_SUPPORTED" {
		return fmt.Errorf("钱包类型不支持")
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

func (s *sOne) GetBalance(ctx context.Context, siteGame entity.SiteGame, siteCode string, username string, currency string) (float32, error) {

	// 构建游戏用户名
	buildUsername, err := s.buildUsername(username, siteCode)
	if err != nil {
		return 0, err
	}

	// 生成TraceID
	traceId, _ := s.generateUUIDv4()
	// 构建请求数据
	data := oneBalanceReq{
		Username: buildUsername,
		TraceId:  traceId,
		Currency: currency,
	}

	// 生成签名
	signature, err := s.getONESignature(data, siteGame.ApiKey)
	if err != nil {
		return 0, err
	}

	// 构建请求头
	headers := map[string]string{
		"Content-Type": "application/json",
		"X-API-Key":    siteGame.ApiId,
		"X-Signature":  signature,
	}

	// 构建API URL
	url := strings.TrimSuffix(siteGame.ApiUrl, "/") + "/cash/balance"

	g.Log().Infof(ctx, "转入请求参数: %v,%v,%v", url, headers, data)
	res, err := util.CurlPostHeader(url, headers, data)
	if err != nil {
		return 0, fmt.Errorf("游戏API请求失败: %v", err)
	}

	g.Log().Infof(ctx, "转入请求结果: %v", res)

	if res["status"] == "SC_INVALID_OPERATOR" {
		return 0, fmt.Errorf("未经授权的访问")
	}
	if res["status"] == "SC_INVALID_SIGNATURE" {
		return 0, fmt.Errorf("验证失败")
	}
	if res["status"] == "SC_WRONG_CURRENCY" {
		return 0, fmt.Errorf("不支持该货币")
	}
	if res["status"] == "SC_INVALID_REQUEST" {
		return 0, fmt.Errorf("请求无效")
	}
	if res["status"] == "SC_UNKNOWN_ERROR" {
		return 0, fmt.Errorf("未知错误")
	}
	if res["status"] == "SC_DUPLICATE_REQUEST" {
		return 0, fmt.Errorf("请求重复")
	}
	if res["status"] == "SC_WALLET_NOT_SUPPORTED" {
		return 0, fmt.Errorf("钱包类型不支持")
	}

	if res["status"] == "SC_OK" {
		dataValue, exists := res["data"]
		if !exists {
			return 0, fmt.Errorf("游戏API返回格式错误，缺少data字段")
		}
		dataMap, ok := dataValue.(map[string]interface{})
		if !ok {
			return 0, fmt.Errorf("游戏API返回格式错误，data字段类型不正确")
		}

		amountAny, ok := dataMap["amount"]
		if !ok {
			return 0, fmt.Errorf("缺少字段 data.amount")
		}
		amount, ok := amountAny.(float32)
		if !ok {
			return 0, fmt.Errorf("data.amount 不是 float64")
		}

		return amount, nil
	}
	return 0, fmt.Errorf("余额转入失败")
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

// GetBalanceClient 获取balance服务的gRPC客户端
func (s *sOne) GetBalanceClient() balanceV1.BalanceClient {
	g.Log().Info(context.Background(), "正在连接balance服务: mydocker-jh_balance_service-1:8003")

	// 直接使用容器名和端口，不阻塞连接
	conn, err := grpc.Dial("mydocker-jh_balance_service-1:8003",
		grpc.WithInsecure())
	if err != nil {
		g.Log().Errorf(context.Background(), "连接balance服务失败: %v", err)
		panic(fmt.Sprintf("无法连接到balance服务: %v", err))
	}

	g.Log().Infof(context.Background(), "balance服务连接创建成功")

	return balanceV1.NewBalanceClient(conn)
}
