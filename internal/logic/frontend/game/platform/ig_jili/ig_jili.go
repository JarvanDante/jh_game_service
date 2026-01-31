package ig_jili

import (
	"context"
	"crypto/hmac"
	"crypto/md5"
	"crypto/rand"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"math"
	mrand "math/rand"
	"strconv"
	"strings"
	"time"

	"github.com/gogf/gf/v2/frame/g"
	"github.com/gogf/gf/v2/util/gconv"
	balanceV1 "jh_game_service/api/grpc/balance/v1"
	"jh_game_service/internal/grpc_client"
	"jh_game_service/internal/logic/frontend/game/platform"
	"jh_game_service/internal/util"
)

// IGJiliPlatform JILI游戏平台实现
type IGJiliPlatform struct{}

// NewIGJili 创建NewIGJili平台实例
func NewIGJili() platform.GamePlatform {
	return &IGJiliPlatform{}
}

// 注册到工厂
func init() {
	platform.GetFactory().Register("IG_JILI", func() platform.GamePlatform {
		return NewIGJili()
	})
}

//GetGameUrl
/**
 * @desc：获取游戏链接
 * @param ctx
 * @param params
 * @return map[string]interface{}
 * @return error
 * @author : Carson
 */
func (p *IGJiliPlatform) GetGameUrl(ctx context.Context, params *platform.GameParams) (map[string]interface{}, error) {
	// 1. 构建游戏用户名
	buildUsername := p.buildUsername(params.Username, params.SiteCode)

	// 2. 生成 Key
	key, _ := p.generateKey(buildUsername, params)

	// 3. 获取最后一次游戏记录和用户余额
	newCtx, balanceClient, conn, err := grpc_client.GetBalanceClientWithMetadata(ctx, params.SiteId)
	if err != nil {
		return nil, fmt.Errorf("连接balance服务失败: %v", err)
	}
	defer conn.Close()

	lastGameRes, err := balanceClient.GetLastBalanceLogGame(newCtx, &balanceV1.GetLastBalanceLogGameReq{
		SiteId: int32(params.SiteId),
		UserId: int32(params.UserId),
	})
	if err != nil {
		return nil, fmt.Errorf("获取最后游戏记录失败: %v", err)
	}

	lastGameId := int(lastGameRes.GameId)
	userBalance := float64(lastGameRes.Amount)
	currentGameId := params.SiteGame.GameId

	g.Log().Infof(ctx, "最后游戏ID: %d, 当前游戏ID: %d, 用户余额: %.2f", lastGameId, currentGameId, userBalance)

	//----------
	//4. 判断是否需要转账
	var needTransfer bool
	var amount float64
	needTransfer = false
	amount = 0
	fmt.Println(needTransfer, amount)

	if lastGameId == 0 {
		// 情况1：用户第一次进入游戏，取用户中心钱包余额全部转入当前游戏
		g.Log().Infof(ctx, "用户第一次进入游戏，余额: %.2f", userBalance)
		if userBalance > 0 {
			needTransfer = true
			amount = userBalance
		}
	} else if lastGameId == currentGameId {
		// 情况2：进入的是同一个游戏，不需要转账
		g.Log().Infof(ctx, "进入同一个游戏，无需转账")
		needTransfer = false
	} else {
		// 情况3：进入不同游戏，需要先从旧游戏转出，再转入新游戏
		g.Log().Infof(ctx, "切换游戏: %d -> %d", lastGameId, currentGameId)

		// 3.1 查询旧游戏配置
		oldGameConfig, err := platform.GetGameConfigFunc(ctx, params.SiteId, lastGameId)
		if err != nil {
			return nil, fmt.Errorf("查询旧游戏配置失败: %v", err)
		}
		g.Log().Infof(ctx, "旧游戏配置: GameId=%d, Platform=%s, Name=%s", oldGameConfig.GameId, oldGameConfig.Platform, oldGameConfig.Name)

		// 3.2 使用工厂创建旧游戏平台实例
		oldPlatform, err := platform.GetFactory().Create(oldGameConfig.Platform)
		if err != nil {
			return nil, fmt.Errorf("创建旧游戏平台实例失败: %v", err)
		}
		g.Log().Infof(ctx, "成功创建旧游戏平台实例: %s", oldGameConfig.Platform)

		// 3.3 构建旧游戏参数
		oldParams := &platform.GameParams{
			SiteId:   params.SiteId,
			UserId:   params.UserId,
			Username: params.Username,
			SiteCode: params.SiteCode,
			Currency: params.Currency,
			SiteGame: *oldGameConfig,
		}
		g.Log().Infof(ctx, "旧游戏参数: GameId=%d, ApiUrl=%s", oldParams.SiteGame.GameId, oldParams.SiteGame.ApiUrl)

		// 3.4 查询旧游戏余额
		oldGameBalance, balanceErr := oldPlatform.GetBalance(ctx, oldParams)
		if balanceErr != nil {
			g.Log().Errorf(ctx, "查询旧游戏余额失败: %v", balanceErr)
		} else {
			g.Log().Infof(ctx, "旧游戏余额: %.2f", oldGameBalance)

			// 3.5 如果旧游戏有余额，先转出到中心钱包
			if oldGameBalance > 0 {
				transferErr := oldPlatform.TransferOut(ctx, &platform.TransferParams{
					SiteId:   params.SiteId,
					UserId:   params.UserId,
					Username: params.Username,
					SiteCode: params.SiteCode,
					Amount:   oldGameBalance,
					Currency: params.Currency,
					SiteGame: *oldGameConfig, // 使用旧游戏配置
				})
				if transferErr != nil {
					return nil, fmt.Errorf("从旧游戏转出失败: %v", transferErr)
				}

				g.Log().Infof(ctx, "从旧游戏转出成功: %.2f", oldGameBalance)

				// 重新查询中心钱包余额
				lastGameRes2, err := balanceClient.GetLastBalanceLogGame(newCtx, &balanceV1.GetLastBalanceLogGameReq{
					SiteId: int32(params.SiteId),
					UserId: int32(params.UserId),
				})
				if err == nil {
					userBalance = float64(lastGameRes2.Amount)
					g.Log().Infof(ctx, "转出后中心钱包余额: %.2f", userBalance)
				} else {
					// 如果查询失败，使用计算值
					userBalance += oldGameBalance
					g.Log().Warning(ctx, "查询中心钱包余额失败，使用计算值: %.2f", userBalance)
				}
			}
		}

		// 3.6 将中心钱包余额全部转入新游戏
		if userBalance > 0 {
			needTransfer = true
			amount = userBalance
		}
	}

	//5. 执行转账
	if needTransfer && amount > 0 {
		err = p.TransferIn(ctx, &platform.TransferParams{
			SiteId:   params.SiteId,
			UserId:   params.UserId,
			Username: params.Username,
			SiteCode: params.SiteCode,
			Amount:   amount,
			Currency: params.Currency,
			SiteGame: params.SiteGame,
		})
		if err != nil {
			return nil, fmt.Errorf("转入游戏失败: %v", err)
		}
		g.Log().Infof(ctx, "转入游戏成功: %.2f", amount)
	}

	//----------

	lang := p.getLang(params.Currency)
	disableFullScreen := "0"

	// 4. 构建请求数据
	data := map[string]string{
		"Account":           buildUsername,
		"GameId":            params.Request.Key,
		"Lang":              lang,
		"HomeUrl":           params.Domain,
		"platform":          p.getPlatformType(params.Request.Mobile),
		"disableFullScreen": disableFullScreen,
		"AgentId":           params.SiteGame.ApiId,
		"Key":               key,
	}

	// 7. 调用API
	url := strings.TrimSuffix(params.SiteGame.ApiUrl, "/") + "/LoginWithoutRedirect"

	res, err := util.SendRequest(url, data)
	if err != nil {
		return nil, fmt.Errorf("游戏API请求失败: %v", err)
	}

	g.Log().Infof(ctx, "游戏API响应: %v", res)

	// 8. 解析响应
	errorCode, message, dataMap, err := p.parseAPIResponse(ctx, res)
	if err != nil {
		return nil, fmt.Errorf("解析API响应失败: %v", err)
	}

	g.Log().Infof(ctx, "API返回: ErrorCode=%d, Message=%s", errorCode, message)

	if err := p.handleAPIError(ctx, errorCode, message); err != nil {
		return nil, fmt.Errorf("获取游戏链接失败: %v", err)
	}

	gameUrl, ok := dataMap["Data"].(string)
	if !ok || gameUrl == "" {
		return nil, fmt.Errorf("无法获取游戏链接")
	}

	return map[string]interface{}{
		"type": "url",
		"url":  gameUrl,
	}, nil
}

//GetBalance
/**
 * @desc：查询游戏余额（自动处理注册）
 * @param ctx
 * @param params
 * @return float64
 * @return error
 * @author : Carson
 */
func (p *IGJiliPlatform) GetBalance(ctx context.Context, params *platform.GameParams) (float64, error) {
	buildUsername := p.buildUsername(params.Username, params.SiteCode)

	// 1. 确保用户已注册（如果未注册会自动注册）
	if err := p.ensureMemberRegistered(ctx, buildUsername, params); err != nil {
		return 0, fmt.Errorf("注册会员失败: %v", err)
	}

	// 2. 查询余额
	keyBalance, err := p.generateKeyBalance(buildUsername, params)
	if err != nil {
		return 0, err
	}

	dataBalance := map[string]string{
		"Accounts": buildUsername,
		"AgentId":  params.SiteGame.ApiId,
		"Key":      keyBalance,
	}

	// 7. 调用API
	urlBalance := strings.TrimSuffix(params.SiteGame.ApiUrl, "/") + "/GetMemberInfo"

	resBalance, err := util.SendRequest(urlBalance, dataBalance)
	if err != nil {
		return 0, fmt.Errorf("游戏API余额请求失败: %v", err)
	}

	g.Log().Infof(ctx, "游戏API响应Balance: %v", resBalance)

	// 解析响应
	errorCodeBalance, messageBalance, dataMapBalance, err := p.parseAPIResponse(ctx, resBalance)
	if err != nil {
		return 0, fmt.Errorf("解析余额响应失败: %v", err)
	}

	g.Log().Infof(ctx, "查询余额API返回: ErrorCode=%d, Message=%s", errorCodeBalance, messageBalance)

	if err := p.handleAPIError(ctx, errorCodeBalance, messageBalance); err != nil {
		return 0, fmt.Errorf("查询余额失败: %v", err)
	}

	g.Log().Infof(ctx, "查询余额成功: %s", buildUsername)

	balance, err := p.GetJsonBalance(dataMapBalance)
	if err != nil {
		return 0, err
	}
	fmt.Println("balance =", balance)

	return balance, nil
}

//TransferIn
/**
 * @desc：转入游戏
 * @param ctx
 * @param params
 * @return error
 * @author : Carson
 */
func (p *IGJiliPlatform) TransferIn(ctx context.Context, params *platform.TransferParams) error {
	// 0. 确保用户已注册
	buildUsername := p.buildUsername(params.Username, params.SiteCode)
	gameParams := &platform.GameParams{
		SiteId:   params.SiteId,
		UserId:   params.UserId,
		Username: params.Username,
		SiteCode: params.SiteCode,
		Currency: params.Currency,
		SiteGame: params.SiteGame,
	}
	if err := p.ensureMemberRegistered(ctx, buildUsername, gameParams); err != nil {
		return fmt.Errorf("注册会员失败: %v", err)
	}

	// 1. 调用balance服务扣款
	g.Log().Info(ctx, "正在连接balance服务: jh_balance_service")

	newCtx, balanceClient, conn, err := grpc_client.GetBalanceClientWithMetadata(ctx, params.SiteId)
	if err != nil {
		return fmt.Errorf("连接balance服务失败: %v", err)
	}
	defer conn.Close()

	g.Log().Infof(ctx, "balance服务连接成功，准备调用AdjustBalance: siteId=%d, userId=%d, amount=%.2f",
		params.SiteId, params.UserId, params.Amount)

	adjustRes, err := balanceClient.AdjustBalance(newCtx, &balanceV1.AdjustBalanceReq{
		SiteId:   int32(params.SiteId),
		UserId:   int32(params.UserId),
		Amount:   float32(params.Amount),
		Type:     0, // 0=转出：从中心账户转账到游戏账户
		GameName: params.SiteGame.Name,
		GameId:   strconv.Itoa(params.SiteGame.GameId),
	})

	g.Log().Infof(ctx, "balance服务调用完成: err=%v", err)

	if err != nil {
		return fmt.Errorf("调用balance服务失败: %v", err)
	}

	if !adjustRes.Success {
		return fmt.Errorf("扣款失败: %s", adjustRes.Message)
	}

	g.Log().Infof(ctx, "扣款成功，订单号: %s", adjustRes.OrderNo)

	// 2. 调用三方游戏API转账
	transactionId, _ := GenerateUUIDv4()
	params.OrderNo = transactionId
	keyTransferIn, _ := p.generateKeyTransferIn(buildUsername, params)

	data := map[string]string{
		"Account":       buildUsername,
		"TransactionId": transactionId,
		"Amount":        FormatAmount4(params.Amount),
		"TransferType":  "2", //2=平台→游戏厅
		"AgentId":       params.SiteGame.ApiId,
		"Key":           keyTransferIn,
	}

	// 注意：转出使用 /ExchangeTransferByAgentId 接口
	url := strings.TrimSuffix(params.SiteGame.ApiUrl, "/") + "/ExchangeTransferByAgentId"

	g.Log().Infof(ctx, "准备调用游戏转账API: url=%s, data=%v", url, data)

	res, err := util.SendRequest(url, data)
	if err != nil {
		g.Log().Errorf(ctx, "游戏API请求失败: %v", err)
		// 转账失败，需要补偿余额
		p.compensateBalance(newCtx, balanceClient, params.SiteId, params.UserId, adjustRes.OrderNo, params.Amount)
		return fmt.Errorf("游戏API请求失败: %v", err)
	}

	g.Log().Infof(ctx, "游戏API响应In: %v", res)

	if res["http_code"] != 200 {
		g.Log().Errorf(ctx, "网络请求失败，请稍后再试。HTTP代码:%s", res["http_code"])
		// 转账失败，需要补偿余额
		p.compensateBalance(newCtx, balanceClient, params.SiteId, params.UserId, adjustRes.OrderNo, params.Amount)
		return fmt.Errorf("网络请求失败，请稍后再试。HTTP代码:%s", res["http_code"])
	}

	// 8. 解析响应
	bodyValue, exists := res["body"]
	if !exists {
		g.Log().Errorf(ctx, "游戏API返回格式错误，缺少body字段")
		// 转账失败，需要补偿余额
		p.compensateBalance(newCtx, balanceClient, params.SiteId, params.UserId, adjustRes.OrderNo, params.Amount)
		return fmt.Errorf("游戏API返回格式错误，缺少body字段")
	}

	// body可能是字节数组或字符串，需要解析成JSON
	var bodyBytes []byte
	switch v := bodyValue.(type) {
	case []byte:
		bodyBytes = v
	case string:
		bodyBytes = []byte(v)
	case []interface{}:
		// 字节数组形式 [123, 34, ...]
		bodyBytes = make([]byte, len(v))
		for i, b := range v {
			if num, ok := b.(float64); ok {
				bodyBytes[i] = byte(num)
			}
		}
	default:
		// 转账失败，需要补偿余额
		p.compensateBalance(newCtx, balanceClient, params.SiteId, params.UserId, adjustRes.OrderNo, params.Amount)
		return fmt.Errorf("游戏API返回格式错误，body字段类型不正确: %T", bodyValue)
	}

	// 解析JSON
	var dataMap map[string]interface{}
	if err := json.Unmarshal(bodyBytes, &dataMap); err != nil {
		g.Log().Errorf(ctx, "解析游戏API响应失败: %v, body: %s", err, string(bodyBytes))
		// 转账失败，需要补偿余额
		p.compensateBalance(newCtx, balanceClient, params.SiteId, params.UserId, adjustRes.OrderNo, params.Amount)
		return fmt.Errorf("解析游戏API响应失败: %v", err)
	}

	// 解析错误码和消息
	errorCodeFloat := 0.0
	if ec, ok := dataMap["ErrorCode"].(float64); ok {
		errorCodeFloat = ec
	}
	errorCode := int(errorCodeFloat)

	message := ""
	if msg, ok := dataMap["Message"].(string); ok {
		message = msg
	}

	g.Log().Infof(ctx, "TransferIn游戏API返回: ErrorCode=%d, Message=%s", errorCode, message)

	// 处理特殊错误码
	if errorCode == 102 || errorCode == 105 {
		// 重复订单或订单已存在，视为成功
		g.Log().Infof(ctx, "订单已存在(ErrorCode=%d)，视为成功", errorCode)
		_, err = balanceClient.UpdateTransferStatus(ctx, &balanceV1.UpdateTransferStatusReq{
			OrderNo: adjustRes.OrderNo,
		})
		if err != nil {
			g.Log().Errorf(ctx, "更新转账状态失败: %v", err)
		}
		return nil
	}

	// 处理其他错误
	if errorCode != 0 {
		g.Log().Errorf(ctx, "TransferIn游戏转账失败。错误代码:%d, 消息:%s", errorCode, message)
		// 转账失败，需要补偿余额
		p.compensateBalance(newCtx, balanceClient, params.SiteId, params.UserId, adjustRes.OrderNo, params.Amount)
		return fmt.Errorf("TransferIn游戏转账失败。错误代码:%d, 消息:%s", errorCode, message)
	}
	//url=https://wb-api-2.richjourney888.com/api1/ExchangeTransferByAgentId,
	//// data=map[
	////Account:jh_0FPUozUQFpkC
	////AgentId:FJ999Trans_MMK
	////Amount:25
	////Key:0ZdGVXf1d144788b0e913ab8a7f4d8f7c03401P7Jmjd
	////TransactionId:d41c770a-cf7c-4875-9859-d3149c82b43b
	////TransferType:2

	// 成功，解析返回数据
	resData := dataMap["Data"]

	resDataMap, ok := resData.(map[string]interface{})
	if !ok {
		return fmt.Errorf("resData 类型错误，不是对象")
	}

	if gconv.Int(resDataMap["Status"]) != 1 {
		return fmt.Errorf("转账失败")
	}

	// 更新转账状态
	_, err = balanceClient.UpdateTransferStatus(newCtx, &balanceV1.UpdateTransferStatusReq{
		OrderNo: adjustRes.OrderNo,
	})
	if err != nil {
		g.Log().Errorf(ctx, "更新转账状态失败: %v", err)
	}

	g.Log().Infof(ctx, "转入成功")

	return nil
}

//TransferOut
/**
 * @desc：转出游戏（从游戏转回中心账户）
 * @param ctx
 * @param params
 * @return error
 * @author : Carson
 */
func (p *IGJiliPlatform) TransferOut(ctx context.Context, params *platform.TransferParams) error {
	// 1. 调用balance服务创建pending订单（Type=1，不扣用户余额）
	g.Log().Info(ctx, "正在连接balance服务: jh_balance_service")

	newCtx, balanceClient, conn, err := grpc_client.GetBalanceClientWithMetadata(ctx, params.SiteId)
	if err != nil {
		return fmt.Errorf("连接balance服务失败: %v", err)
	}
	defer conn.Close()

	g.Log().Infof(ctx, "balance服务连接成功，准备调用AdjustBalance: siteId=%d, userId=%d, amount=%.2f, type=1(转出)",
		params.SiteId, params.UserId, params.Amount)

	adjustRes, err := balanceClient.AdjustBalance(newCtx, &balanceV1.AdjustBalanceReq{
		SiteId:   int32(params.SiteId),
		UserId:   int32(params.UserId),
		Amount:   float32(params.Amount),
		Type:     1, // 1=转入：从游戏账户转账到中心账户
		GameName: params.SiteGame.Name,
		GameId:   strconv.Itoa(params.SiteGame.GameId),
	})

	g.Log().Infof(ctx, "balance服务调用完成: err=%v", err)

	if err != nil {
		return fmt.Errorf("调用balance服务失败: %v", err)
	}

	if !adjustRes.Success {
		return fmt.Errorf("创建转出订单失败: %s", adjustRes.Message)
	}

	g.Log().Infof(ctx, "创建转出订单成功，订单号: %s", adjustRes.OrderNo)

	// 2. 调用三方游戏API转出

	buildUsername := p.buildUsername(params.Username, params.SiteCode)
	transactionId, _ := GenerateUUIDv4()
	params.OrderNo = transactionId
	keyTransferOut, _ := p.generateKeyTransferOut(buildUsername, params)

	data := map[string]string{
		"Account":       buildUsername,
		"TransactionId": transactionId,
		"Amount":        FormatAmount4(params.Amount),
		"TransferType":  "1", //1=游戏厅→平台（全额转出）
		"AgentId":       params.SiteGame.ApiId,
		"Key":           keyTransferOut,
	}

	// 注意：转出使用 /ExchangeTransferByAgentId 接口
	url := strings.TrimSuffix(params.SiteGame.ApiUrl, "/") + "/ExchangeTransferByAgentId"

	res, err := util.SendRequest(url, data)
	if err != nil {
		// 转出失败，标记订单失败
		p.markTransferFailed(newCtx, balanceClient, adjustRes.OrderNo)
		return fmt.Errorf("游戏API请求失败: %v", err)
	}

	g.Log().Infof(ctx, "游戏API响应Out: %v", res)

	if res["http_code"] != 200 {
		g.Log().Errorf(ctx, "网络请求失败，请稍后再试。HTTP代码:%s", res["http_code"])
		// 转出失败，标记订单失败
		p.markTransferFailed(newCtx, balanceClient, adjustRes.OrderNo)
		return fmt.Errorf("网络请求失败，请稍后再试。HTTP代码:%s", res["http_code"])
	}

	// 8. 解析响应
	bodyValue, exists := res["body"]
	if !exists {
		g.Log().Errorf(ctx, "游戏API返回格式错误，缺少body字段")
		// 转出失败，标记订单失败
		p.markTransferFailed(newCtx, balanceClient, adjustRes.OrderNo)
		return fmt.Errorf("游戏API返回格式错误，缺少body字段")
	}

	// body可能是字节数组或字符串，需要解析成JSON
	var bodyBytes []byte
	switch v := bodyValue.(type) {
	case []byte:
		bodyBytes = v
	case string:
		bodyBytes = []byte(v)
	case []interface{}:
		// 字节数组形式 [123, 34, ...]
		bodyBytes = make([]byte, len(v))
		for i, b := range v {
			if num, ok := b.(float64); ok {
				bodyBytes[i] = byte(num)
			}
		}
	default:
		// 转出失败，标记订单失败
		p.markTransferFailed(newCtx, balanceClient, adjustRes.OrderNo)
		return fmt.Errorf("游戏API返回格式错误，body字段类型不正确: %T", bodyValue)
	}

	// 解析JSON
	var dataMap map[string]interface{}
	if err := json.Unmarshal(bodyBytes, &dataMap); err != nil {
		g.Log().Errorf(ctx, "解析游戏API响应失败: %v, body: %s", err, string(bodyBytes))
		// 转出失败，标记订单失败
		p.markTransferFailed(newCtx, balanceClient, adjustRes.OrderNo)
		return fmt.Errorf("解析游戏API响应失败: %v", err)
	}

	// 解析错误码和消息
	errorCodeFloat := 0.0
	if ec, ok := dataMap["ErrorCode"].(float64); ok {
		errorCodeFloat = ec
	}
	errorCode := int(errorCodeFloat)

	message := ""
	if msg, ok := dataMap["Message"].(string); ok {
		message = msg
	}

	g.Log().Infof(ctx, "TransferOut游戏API返回: ErrorCode=%d, Message=%s", errorCode, message)

	// 处理特殊错误码
	if errorCode == 102 || errorCode == 105 {
		// 重复订单或订单已存在，视为成功
		g.Log().Infof(ctx, "订单已存在(ErrorCode=%d)，视为成功", errorCode)
		_, err = balanceClient.UpdateTransferStatus(ctx, &balanceV1.UpdateTransferStatusReq{
			OrderNo: adjustRes.OrderNo,
		})
		if err != nil {
			g.Log().Errorf(ctx, "更新转账状态失败: %v", err)
		}
		return nil
	}

	// 处理其他错误
	if errorCode != 0 {
		g.Log().Errorf(ctx, "游戏转出失败。错误代码:%d, 消息:%s", errorCode, message)
		// 转出失败，标记订单失败
		p.markTransferFailed(newCtx, balanceClient, adjustRes.OrderNo)
		return fmt.Errorf("游戏转出失败。错误代码:%d, 消息:%s", errorCode, message)
	}

	// 成功，解析返回数据
	resData := dataMap["Data"]

	resDataMap, ok := resData.(map[string]interface{})
	if !ok {
		// 转出失败，标记订单失败
		p.markTransferFailed(newCtx, balanceClient, adjustRes.OrderNo)
		return fmt.Errorf("resData 类型错误，不是对象")
	}

	if gconv.Int(resDataMap["Status"]) != 1 {
		// 转出失败，标记订单失败
		p.markTransferFailed(newCtx, balanceClient, adjustRes.OrderNo)
		return fmt.Errorf("输出失败")
	}

	// 更新转账状态（增加用户余额）
	_, err = balanceClient.UpdateTransferStatus(newCtx, &balanceV1.UpdateTransferStatusReq{
		OrderNo: adjustRes.OrderNo,
	})
	if err != nil {
		g.Log().Errorf(ctx, "更新转账状态失败: %v", err)
		return fmt.Errorf("更新转账状态失败: %v", err)
	}

	g.Log().Infof(ctx, "转出成功，已增加用户余额")

	return nil
}

// PullBetRecords 拉取注单
func (p *IGJiliPlatform) PullBetRecords(ctx context.Context, params *platform.PullRecordsParams) ([]platform.BetRecord, error) {
	// TODO: 实现拉单逻辑
	g.Log().Warning(ctx, "IG_JILI平台拉单功能待实现")
	return nil, fmt.Errorf("功能待实现")
}

// GetGameSource 获取游戏列表并同步到数据库（暂不实现）
func (p *IGJiliPlatform) GetGameSource(ctx context.Context, params *platform.GetGameSourceParams) error {
	g.Log().Warning(ctx, "IG_JILI平台游戏资源同步功能暂不实现")
	return fmt.Errorf("IG_JILI平台暂不支持游戏资源同步")
}

// ========== 辅助方法 ==========

// buildUsername 构建游戏用户名
func (p *IGJiliPlatform) buildUsername(username, siteCode string) string {
	return siteCode + "_" + username
}

// getPlatformType 获取平台类型
func (p *IGJiliPlatform) getPlatformType(mobile int32) string {
	if mobile > 0 {
		return "app"
	}
	return "web"
}

//GenerateUUIDv4
/**
 * @desc：生成 UUID v4
 * @return string
 * @return error
 * @author : Carson
 */
func GenerateUUIDv4() (string, error) {
	b := make([]byte, 16)
	_, err := rand.Read(b)
	if err != nil {
		return "", err
	}

	// 版本号 v4: xxxx xxxx xxxx 4xxx
	b[6] = (b[6] & 0x0f) | 0x40

	// RFC4122 variant: 10xx xxxx
	b[8] = (b[8] & 0x3f) | 0x80

	uuid := fmt.Sprintf("%02x%02x%02x%02x-%02x%02x-%02x%02x-%02x%02x-%02x%02x%02x%02x%02x%02x",
		b[0], b[1], b[2], b[3],
		b[4], b[5],
		b[6], b[7],
		b[8], b[9],
		b[10], b[11], b[12], b[13], b[14], b[15],
	)
	return uuid, nil
}

//generateKey
/**
 * @desc：生成 Key（进入游戏）
 * @param account
 * @param params
 * @return string
 * @return error
 * @author : Carson
 */
func (p *IGJiliPlatform) generateKey(account string, params *platform.GameParams) (string, error) {
	lang := p.getLang(params.Currency)

	paramsString := "Account=" + account + "&GameId=" + params.Request.Key + "&Lang=" + lang + "&AgentId=" + params.SiteGame.ApiId

	key := CalculateKeyWithParams(paramsString, params.SiteGame.ApiId, params.SiteGame.ApiKey)

	return key, nil
}

//generateKeyRegister
/**
 * @desc：生成 Key(注册)
 * @param account
 * @param params
 * @return string
 * @return error
 * @author : Carson
 */
func (p *IGJiliPlatform) generateKeyRegister(account string, params *platform.GameParams) (string, error) {

	paramsString := "Account=" + account + "&AgentId=" + params.SiteGame.ApiId

	key := CalculateKeyWithParams(paramsString, params.SiteGame.ApiId, params.SiteGame.ApiKey)

	return key, nil
}

//generateKeyTransferOut
/**
 * @desc：生成 Key(转出)
 * @param account
 * @param params
 * @return string
 * @return error
 * @author : Carson
 */
func (p *IGJiliPlatform) generateKeyTransferOut(account string, params *platform.TransferParams) (string, error) {
	transferType := 1 //2=游戏厅→平台

	paramsString := "Account=" + account + "&TransactionId=" + params.OrderNo + "&Amount=" + FormatAmount4(params.Amount) + "&TransferType=" + gconv.String(transferType) + "&AgentId=" + params.SiteGame.ApiId

	key := CalculateKeyWithParams(paramsString, params.SiteGame.ApiId, params.SiteGame.ApiKey)

	return key, nil
}

//generateKeyTransferIn
/**
 * @desc：生成 Key(转入)
 * @param account
 * @param params
 * @return string
 * @return error
 * @author : Carson
 */
func (p *IGJiliPlatform) generateKeyTransferIn(account string, params *platform.TransferParams) (string, error) {
	transferType := 2 //2=平台→游戏厅

	paramsString := "Account=" + account + "&TransactionId=" + params.OrderNo + "&Amount=" + FormatAmount4(params.Amount) + "&TransferType=" + gconv.String(transferType) + "&AgentId=" + params.SiteGame.ApiId

	key := CalculateKeyWithParams(paramsString, params.SiteGame.ApiId, params.SiteGame.ApiKey)

	return key, nil
}

//generateKeyBalance
/**
 * @desc：生成 Key(查余额)
 * @param account
 * @param params
 * @return string
 * @return error
 * @author : Carson
 */
func (p *IGJiliPlatform) generateKeyBalance(account string, params *platform.GameParams) (string, error) {

	paramsString := "Accounts=" + account + "&AgentId=" + params.SiteGame.ApiId

	key := CalculateKeyWithParams(paramsString, params.SiteGame.ApiId, params.SiteGame.ApiKey)

	return key, nil
}

func (p *IGJiliPlatform) getLang(currency string) string {
	var lang string
	//MMK(K) -> my-MM
	if currency == "MMK(K)" {
		lang = "my-MM"
	}

	return lang
}

// ++++++++++++++++

// CalculateKeyWithParams 等价 PHP calculateKeyWithParams
func CalculateKeyWithParams(paramsString, agentId, agentKey string) string {
	// 1) 获取 UTC-4 时间的 yyMMd 格式
	yyMMd := getUtcMinus4Time()

	// 2) KeyG = md5(yyMMd + AgentId + AgentKey)
	keyGString := yyMMd + agentId + agentKey
	keyG := md5Hex(keyGString)

	// 3) middleMd5 = md5(paramsString + KeyG)
	middleMd5 := md5Hex(paramsString + keyG)

	// 4) 最终 Key = 6随机字符 + middleMd5 + 6随机字符
	return generateRandomString(6) + middleMd5 + generateRandomString(6)
}

// generateRandomString 等价 PHP generateRandomString
// 注意：PHP 用 rand()，属于伪随机，这里用 math/rand 更一致

var rng = mrand.New(mrand.NewSource(time.Now().UnixNano()))

func generateRandomString(length int) string {
	const characters = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"
	charLen := len(characters)

	b := make([]byte, length)
	for i := 0; i < length; i++ {
		b[i] = characters[rng.Intn(charLen)]
	}
	return string(b)
}

func md5Hex(s string) string {
	sum := md5.Sum([]byte(s))
	return hex.EncodeToString(sum[:]) // 小写 hex，等价 PHP md5()
}

// getUtcMinus4Time 等价 PHP getUtcMinus4Time()
// 返回 yyMMd（注意 d 不补0）
func getUtcMinus4Time() string {
	t := time.Now().UTC().Add(-4 * time.Hour)

	yy := t.Format("06")            // yy
	mm := t.Format("01")            // MM
	d := fmt.Sprintf("%d", t.Day()) // d（不补0）

	return yy + mm + d
}

//++++++++++++++++

// getSignature 生成签名
func (p *IGJiliPlatform) getSignature(data interface{}, secretKey string) (string, error) {
	b, err := json.Marshal(data)
	if err != nil {
		return "", err
	}

	h := hmac.New(sha256.New, []byte(secretKey))
	h.Write(b)

	return hex.EncodeToString(h.Sum(nil)), nil
}

// compensateBalance 补偿余额
func (p *IGJiliPlatform) compensateBalance(ctx context.Context, client balanceV1.BalanceClient, siteId int, userId uint, orderNo string, amount float64) {
	_, err := client.CompensateBalance(ctx, &balanceV1.CompensateBalanceReq{
		SiteId:  int32(siteId),
		UserId:  int32(userId),
		OrderNo: orderNo,
		Amount:  float32(amount),
	})
	if err != nil {
		g.Log().Errorf(ctx, "补偿余额失败: %v", err)
	}
}

//markTransferFailed
/**
 * @desc：标记转账失败（用于转出失败，不需要补偿余额）
 * @param ctx
 * @param client
 * @param orderNo
 * @author : Carson
 */
func (p *IGJiliPlatform) markTransferFailed(ctx context.Context, client balanceV1.BalanceClient, orderNo string) {
	/// 传入0值，balance服务会自动从订单中获取siteId、userId和amount
	// 转出失败时只标记订单失败，不会修改用户余额
	_, err := client.CompensateBalance(ctx, &balanceV1.CompensateBalanceReq{
		SiteId:  0,
		UserId:  0,
		OrderNo: orderNo,
		Amount:  0, // 0金额表示只标记失败，不补偿
	})
	if err != nil {
		g.Log().Errorf(ctx, "标记转账失败: %v", err)
	}
}

func (p *IGJiliPlatform) GetJsonBalance(responseData map[string]any) (float64, error) {
	// Data
	dataAny, ok := responseData["Data"]
	if !ok {
		return 0, fmt.Errorf("未查询到玩家信息(Data不存在)")
	}

	dataList, ok := dataAny.([]any)
	if !ok || len(dataList) == 0 {
		return 0, fmt.Errorf("未查询到玩家信息(Data为空)")
	}

	// 第一个玩家
	playerInfo, ok := dataList[0].(map[string]any)
	if !ok {
		return 0, fmt.Errorf("玩家信息格式错误(Data[0]不是对象)")
	}

	// Balance
	bAny, ok := playerInfo["Balance"]
	if !ok || bAny == nil {
		return 0, nil // 相当于 PHP 的 ?? 0
	}

	// JSON 数字默认是 float64
	switch v := bAny.(type) {
	case float64:
		return v, nil
	case int:
		return float64(v), nil
	case int64:
		return float64(v), nil
	case string:
		// 如果 Balance 是字符串 "123.45"
		val, err := strconv.ParseFloat(v, 64)
		if err != nil {
			return 0, fmt.Errorf("Balance解析失败: %v", err)
		}
		return val, nil
	default:
		return 0, fmt.Errorf("Balance类型不支持: %T", bAny)
	}
}

//FormatAmount4
/**
 * @desc：金额精度处理：最多保留4位小数，超过4位直接截断
 * @param amount
 * @return string
 * @author : Carson
 */
func FormatAmount4(amount float64) string {
	amount = math.Floor(amount*10000) / 10000
	return fmt.Sprintf("%.4f", amount)
}

// ========== 辅助方法：减少重复代码 ==========

// parseBodyBytes 解析响应body为字节数组
func (p *IGJiliPlatform) parseBodyBytes(bodyValue interface{}) ([]byte, error) {
	switch v := bodyValue.(type) {
	case []byte:
		return v, nil
	case string:
		return []byte(v), nil
	case []interface{}:
		bodyBytes := make([]byte, len(v))
		for i, b := range v {
			if num, ok := b.(float64); ok {
				bodyBytes[i] = byte(num)
			}
		}
		return bodyBytes, nil
	default:
		return nil, fmt.Errorf("body字段类型不正确: %T", bodyValue)
	}
}

// parseAPIResponse 解析API响应，返回错误码、消息和数据
func (p *IGJiliPlatform) parseAPIResponse(ctx context.Context, res map[string]interface{}) (errorCode int, message string, dataMap map[string]interface{}, err error) {
	// 检查HTTP状态码
	if httpCode, ok := res["http_code"].(int); ok && httpCode != 200 {
		return 0, "", nil, fmt.Errorf("HTTP错误: %d", httpCode)
	}

	// 解析body
	bodyValue, exists := res["body"]
	if !exists {
		return 0, "", nil, fmt.Errorf("缺少body字段")
	}

	bodyBytes, err := p.parseBodyBytes(bodyValue)
	if err != nil {
		return 0, "", nil, err
	}

	// 解析JSON
	if err := json.Unmarshal(bodyBytes, &dataMap); err != nil {
		g.Log().Errorf(ctx, "解析JSON失败: %v, body: %s", err, string(bodyBytes))
		return 0, "", nil, fmt.Errorf("解析JSON失败: %v", err)
	}

	// 解析错误码和消息
	if ec, ok := dataMap["ErrorCode"].(float64); ok {
		errorCode = int(ec)
	}
	if msg, ok := dataMap["Message"].(string); ok {
		message = msg
	}

	return errorCode, message, dataMap, nil
}

// handleAPIError 处理API错误，返回是否应该继续
func (p *IGJiliPlatform) handleAPIError(ctx context.Context, errorCode int, message string, allowedCodes ...int) error {
	if errorCode == 0 {
		return nil // 成功
	}

	// 检查是否是允许的错误码
	for _, code := range allowedCodes {
		if errorCode == code {
			g.Log().Infof(ctx, "错误码%d视为成功: %s", errorCode, message)
			return nil
		}
	}

	return fmt.Errorf("错误代码:%d, 消息:%s", errorCode, message)
}

//ensureMemberRegistered
/**
 * @desc：用API返回的错误码101（会员已存在）来避免重复注册
 * @param ctx
 * @param buildUsername
 * @param params
 * @return error
 * @author : Carson
 */
func (p *IGJiliPlatform) ensureMemberRegistered(ctx context.Context, buildUsername string, params *platform.GameParams) error {
	// 生成注册Key
	keyRegister, err := p.generateKeyRegister(buildUsername, params)
	if err != nil {
		return err
	}

	// 构建注册请求
	dataRegister := map[string]string{
		"Account": buildUsername,
		"AgentId": params.SiteGame.ApiId,
		"Key":     keyRegister,
	}

	// 调用注册API
	urlRegister := strings.TrimSuffix(params.SiteGame.ApiUrl, "/") + "/CreateMember"
	resRegister, err := util.SendRequest(urlRegister, dataRegister)
	if err != nil {
		return fmt.Errorf("注册API请求失败: %v", err)
	}

	// 解析响应
	errorCode, message, _, err := p.parseAPIResponse(ctx, resRegister)
	if err != nil {
		return fmt.Errorf("解析注册响应失败: %v", err)
	}

	g.Log().Infof(ctx, "注册API返回: ErrorCode=%d, Message=%s", errorCode, message)

	// 错误码0=注册成功，101=会员已存在（也视为成功）
	if err := p.handleAPIError(ctx, errorCode, message, 101); err != nil {
		return fmt.Errorf("注册失败: %v", err)
	}

	if errorCode == 101 {
		g.Log().Debugf(ctx, "会员已存在，跳过注册: %s", buildUsername)
	} else {
		g.Log().Infof(ctx, "会员注册成功: %s", buildUsername)
	}

	return nil
}
