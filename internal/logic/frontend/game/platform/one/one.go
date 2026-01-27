package one

import (
	"context"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"github.com/gogf/gf/v2/frame/g"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	balanceV1 "jh_game_service/api/grpc/balance/v1"
	"jh_game_service/internal/consts"
	"jh_game_service/internal/logic/frontend/game/platform"
	"jh_game_service/internal/mq"
	"jh_game_service/internal/registry"
	"jh_game_service/internal/util"
	"math/rand"
	"strconv"
	"strings"
	"time"
)

// OnePlatform ONE游戏平台实现
type OnePlatform struct{}

// New 创建ONE平台实例
func New() platform.GamePlatform {
	return &OnePlatform{}
}

// 注册到工厂
func init() {
	platform.GetFactory().Register("ONE", func() platform.GamePlatform {
		return New()
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
func (p *OnePlatform) GetGameUrl(ctx context.Context, params *platform.GameParams) (map[string]interface{}, error) {
	// 1. 构建游戏用户名
	buildUsername := util.BuildUsername(params.Username, params.SiteCode)

	// 2. 生成TraceID
	traceId, _ := p.generateUUID()

	// 3. 获取最后一次游戏记录和用户余额
	balanceClient, conn, err := p.getBalanceClient(ctx)
	if err != nil {
		return nil, fmt.Errorf("连接balance服务失败: %v", err)
	}
	defer conn.Close()

	lastGameRes, err := balanceClient.GetLastBalanceLogGame(ctx, &balanceV1.GetLastBalanceLogGameReq{
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

	// 4. 判断是否需要转账
	var needTransfer bool
	var amount float64

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
				lastGameRes2, err := balanceClient.GetLastBalanceLogGame(ctx, &balanceV1.GetLastBalanceLogGameReq{
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

	// 5. 执行转账
	if needTransfer && amount > 0 {
		orderNo, _ := p.generateUUID()
		err = p.TransferIn(ctx, &platform.TransferParams{
			SiteId:   params.SiteId,
			UserId:   params.UserId,
			Username: params.Username,
			SiteCode: params.SiteCode,
			Amount:   amount,
			Currency: params.Currency,
			OrderNo:  orderNo,
			SiteGame: params.SiteGame,
		})
		if err != nil {
			return nil, fmt.Errorf("转入游戏失败: %v", err)
		}
		g.Log().Infof(ctx, "转入游戏成功: %.2f", amount)
	}

	// 4. 构建请求数据
	data := map[string]interface{}{
		"username":  buildUsername,
		"traceId":   traceId,
		"gameCode":  params.Request.Key,
		"language":  "en",
		"platform":  p.getPlatformType(params.Request.Mobile),
		"currency":  params.Currency,
		"lobbyUrl":  params.Domain,
		"ipAddress": params.ClientIP,
	}

	// 5. 生成签名
	signature, err := p.getSignature(data, params.SiteGame.ApiKey)
	if err != nil {
		return nil, err
	}

	// 6. 构建请求头
	headers := map[string]string{
		"Content-Type": "application/json",
		"X-API-Key":    params.SiteGame.ApiId,
		"X-Signature":  signature,
	}

	// 7. 调用API
	url := strings.TrimSuffix(params.SiteGame.ApiUrl, "/") + "/game/url"
	res, err := util.CurlPostHeader(url, headers, data)
	if err != nil {
		return nil, fmt.Errorf("游戏API请求失败: %v", err)
	}

	g.Log().Infof(ctx, "游戏API响应: %v", res)

	// 8. 解析响应
	dataValue, exists := res["data"]
	if !exists {
		g.Log().Errorf(ctx, "游戏API返回格式错误，完整响应: %v", res)
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

//GetBalance
/**
 * @desc：查询游戏余额
 * @param ctx
 * @param params
 * @return float64
 * @return error
 * @author : Carson
 */
func (p *OnePlatform) GetBalance(ctx context.Context, params *platform.GameParams) (float64, error) {
	// 1. 构建游戏用户名
	buildUsername := util.BuildUsername(params.Username, params.SiteCode)

	// 2. 生成TraceID
	traceId, _ := p.generateUUID()

	// 3. 构建请求数据
	data := map[string]interface{}{
		"username": buildUsername,
		"traceId":  traceId,
		"currency": params.Currency,
	}

	// 4. 生成签名
	signature, err := p.getSignature(data, params.SiteGame.ApiKey)
	if err != nil {
		return 0, err
	}

	// 5. 构建请求头
	headers := map[string]string{
		"Content-Type": "application/json",
		"X-API-Key":    params.SiteGame.ApiId,
		"X-Signature":  signature,
	}

	// 6. 调用API
	url := strings.TrimSuffix(params.SiteGame.ApiUrl, "/") + "/cash/balance"

	g.Log().Infof(ctx, "查询余额请求: url=%s, username=%s", url, buildUsername)
	res, err := util.CurlPostHeader(url, headers, data)
	if err != nil {
		return 0, fmt.Errorf("游戏API请求失败: %v", err)
	}

	g.Log().Infof(ctx, "查询余额响应: %v", res)

	// 7. 检查响应状态
	status, _ := res["status"].(string)
	switch status {
	case "SC_OK":
		// 解析余额
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

		// 处理不同的数字类型
		var amount float64
		switch v := amountAny.(type) {
		case float64:
			amount = v
		case float32:
			amount = float64(v)
		case int:
			amount = float64(v)
		case int64:
			amount = float64(v)
		default:
			return 0, fmt.Errorf("data.amount 类型不正确: %T", amountAny)
		}

		return amount, nil

	case "SC_INVALID_OPERATOR":
		return 0, fmt.Errorf("未经授权的访问")
	case "SC_INVALID_SIGNATURE":
		return 0, fmt.Errorf("验证失败")
	case "SC_WRONG_CURRENCY":
		return 0, fmt.Errorf("不支持该货币")
	case "SC_INVALID_REQUEST":
		return 0, fmt.Errorf("请求无效")
	case "SC_UNKNOWN_ERROR":
		return 0, fmt.Errorf("未知错误")
	case "SC_DUPLICATE_REQUEST":
		return 0, fmt.Errorf("请求重复")
	case "SC_WALLET_NOT_SUPPORTED":
		return 0, fmt.Errorf("钱包类型不支持")
	default:
		return 0, fmt.Errorf("查询余额失败: %s", status)
	}
}

//TransferIn
/**
 * @desc：转入游戏
 * @param ctx
 * @param params
 * @return error
 * @author : Carson
 */
func (p *OnePlatform) TransferIn(ctx context.Context, params *platform.TransferParams) error {
	// 1. 调用balance服务扣款
	g.Log().Info(ctx, "正在连接balance服务: jh_balance_service")

	balanceClient, conn, err := p.getBalanceClient(ctx)
	if err != nil {
		return fmt.Errorf("连接balance服务失败: %v", err)
	}
	defer conn.Close()

	g.Log().Infof(ctx, "balance服务连接成功，准备调用AdjustBalance: siteId=%d, userId=%d, amount=%.2f",
		params.SiteId, params.UserId, params.Amount)

	adjustRes, err := balanceClient.AdjustBalance(ctx, &balanceV1.AdjustBalanceReq{
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
	buildUsername := util.BuildUsername(params.Username, params.SiteCode)
	traceId, _ := p.generateUUID()

	data := map[string]interface{}{
		"username":       buildUsername,
		"traceId":        traceId,
		"transferAmount": params.Amount,
		"currency":       params.Currency,
		"referenceId":    params.OrderNo,
	}

	signature, err := p.getSignature(data, params.SiteGame.ApiKey)
	if err != nil {
		// 失败：补偿余额
		p.compensateBalance(ctx, balanceClient, params.SiteId, params.UserId, adjustRes.OrderNo, params.Amount)
		return err
	}

	headers := map[string]string{
		"Content-Type": "application/json",
		"X-API-Key":    params.SiteGame.ApiId,
		"X-Signature":  signature,
	}

	url := strings.TrimSuffix(params.SiteGame.ApiUrl, "/") + "/cash/deposit"
	res, err := util.CurlPostHeader(url, headers, data)
	if err != nil {
		// 失败：补偿余额
		p.compensateBalance(ctx, balanceClient, params.SiteId, params.UserId, adjustRes.OrderNo, params.Amount)
		return fmt.Errorf("游戏API请求失败: %v", err)
	}

	// 3. 检查响应
	status, _ := res["status"].(string)
	if status != "SC_OK" {
		// 失败：补偿余额
		p.compensateBalance(ctx, balanceClient, params.SiteId, params.UserId, adjustRes.OrderNo, params.Amount)
		return fmt.Errorf("转账失败: %s", status)
	}

	// 4. 成功：更新转账状态
	_, err = balanceClient.UpdateTransferStatus(ctx, &balanceV1.UpdateTransferStatusReq{
		OrderNo: adjustRes.OrderNo,
	})
	if err != nil {
		g.Log().Errorf(ctx, "更新转账状态失败: %v", err)
	}

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
func (p *OnePlatform) TransferOut(ctx context.Context, params *platform.TransferParams) error {
	// 1. 调用balance服务创建pending订单（Type=1，不扣用户余额）
	g.Log().Info(ctx, "正在连接balance服务: jh_balance_service")

	balanceClient, conn, err := p.getBalanceClient(ctx)
	if err != nil {
		return fmt.Errorf("连接balance服务失败: %v", err)
	}
	defer conn.Close()

	g.Log().Infof(ctx, "balance服务连接成功，准备调用AdjustBalance: siteId=%d, userId=%d, amount=%.2f, type=1(转出)",
		params.SiteId, params.UserId, params.Amount)

	adjustRes, err := balanceClient.AdjustBalance(ctx, &balanceV1.AdjustBalanceReq{
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
	buildUsername := util.BuildUsername(params.Username, params.SiteCode)
	traceId, _ := p.generateUUID()

	// 如果没有传入OrderNo，生成一个
	if params.OrderNo == "" {
		params.OrderNo, _ = p.generateUUID()
	}

	data := map[string]interface{}{
		"username":       buildUsername,
		"traceId":        traceId,
		"transferAmount": params.Amount,
		"currency":       params.Currency,
		"referenceId":    params.OrderNo,
	}

	signature, err := p.getSignature(data, params.SiteGame.ApiKey)
	if err != nil {
		// 失败：标记订单失败（转出失败不需要补偿余额）
		p.markTransferFailed(ctx, balanceClient, adjustRes.OrderNo)
		return err
	}

	headers := map[string]string{
		"Content-Type": "application/json",
		"X-API-Key":    params.SiteGame.ApiId,
		"X-Signature":  signature,
	}

	// 注意：转出使用 /cash/withdraw 接口
	url := strings.TrimSuffix(params.SiteGame.ApiUrl, "/") + "/cash/withdraw"
	res, err := util.CurlPostHeader(url, headers, data)
	if err != nil {
		// 失败：标记订单失败
		p.markTransferFailed(ctx, balanceClient, adjustRes.OrderNo)
		return fmt.Errorf("游戏API请求失败: %v", err)
	}

	// 3. 检查响应
	status, _ := res["status"].(string)
	if status != "SC_OK" {
		// 失败：标记订单失败
		p.markTransferFailed(ctx, balanceClient, adjustRes.OrderNo)
		return fmt.Errorf("转出失败: %s", status)
	}

	// 4. 成功：更新转账状态（会增加用户余额）
	_, err = balanceClient.UpdateTransferStatus(ctx, &balanceV1.UpdateTransferStatusReq{
		OrderNo: adjustRes.OrderNo,
	})
	if err != nil {
		g.Log().Errorf(ctx, "更新转账状态失败: %v", err)
	}

	g.Log().Infof(ctx, "转出成功，已增加用户余额")

	return nil
}

// PullBetRecords 拉取注单
func (p *OnePlatform) PullBetRecords(ctx context.Context, params *platform.PullRecordsParams) ([]platform.BetRecord, error) {
	// 1. 构建请求参数
	traceId, _ := p.generateUUID()

	// 时间转换为毫秒时间戳（使用 GMT+8 时区）
	loc, _ := time.LoadLocation("Asia/Shanghai")
	fromTime, err := time.ParseInLocation("2006-01-02 15:04:05", params.StartTime, loc)
	if err != nil {
		return nil, fmt.Errorf("开始时间格式错误: %v", err)
	}
	toTime, err := time.ParseInLocation("2006-01-02 15:04:05", params.EndTime, loc)
	if err != nil {
		return nil, fmt.Errorf("结束时间格式错误: %v", err)
	}

	fromTimeMs := fromTime.UnixMilli()
	toTimeMs := toTime.UnixMilli()

	g.Log().Infof(ctx, "拉单时间范围(GMT+8): %s ~ %s (毫秒: %d ~ %d)",
		params.StartTime, params.EndTime, fromTimeMs, toTimeMs)

	// 2. 第一次请求获取总数
	data := map[string]interface{}{
		"traceId":  traceId,
		"fromTime": fromTimeMs,
		"toTime":   toTimeMs,
		"pageNo":   1,
	}

	signature, err := p.getSignature(data, params.SiteGame.ApiKey)
	if err != nil {
		return nil, fmt.Errorf("生成签名失败: %v", err)
	}

	headers := map[string]string{
		"Content-Type": "application/json",
		"X-API-Key":    params.SiteGame.ApiId,
		"X-Signature":  signature,
	}

	url := strings.TrimSuffix(params.SiteGame.ApiUrl, "/") + "/transaction/list"
	res, err := util.CurlPostHeader(url, headers, data)
	if err != nil {
		return nil, fmt.Errorf("请求API失败: %v", err)
	}

	// 格式化打印响应
	resJson, _ := json.MarshalIndent(res, "", "  ")
	g.Log().Infof(ctx, "ONE平台注单API响应:\n%s", string(resJson))

	// 3. 检查响应
	status, ok := res["status"].(string)
	if !ok || status != "SC_OK" {
		g.Log().Warningf(ctx, "ONE平台拉单失败: status=%v", status)
		return nil, nil
	}

	// 4. 解析响应数据
	dataMap, ok := res["data"].(map[string]interface{})
	if !ok {
		return nil, fmt.Errorf("响应数据格式错误")
	}

	totalItems, _ := dataMap["totalItems"].(float64)
	if totalItems == 0 {
		g.Log().Infof(ctx, "ONE平台暂无注单数据: %s ~ %s", params.StartTime, params.EndTime)
		return nil, nil
	}

	g.Log().Infof(ctx, "ONE平台拉单: 时间范围=%s ~ %s, 总记录数=%.0f", params.StartTime, params.EndTime, totalItems)

	// 5. 解析表头（headers是map，不是数组）
	headers_data, _ := dataMap["headers"].(map[string]interface{})
	headerMap := make(map[string]int)
	for headerName, idx := range headers_data {
		if index, ok := idx.(float64); ok {
			headerMap[headerName] = int(index)
		}
	}

	g.Log().Infof(ctx, "ONE平台headers映射: %+v", headerMap)

	// 6. 分页拉取所有数据
	pageSize := 2000
	pageCount := int((totalItems + float64(pageSize) - 1) / float64(pageSize))
	allRecords := make([]platform.BetRecord, 0)

	for pageNo := 1; pageNo <= pageCount; pageNo++ {
		pageData := map[string]interface{}{
			"traceId":  traceId,
			"fromTime": fromTimeMs,
			"toTime":   toTimeMs,
			"pageNo":   pageNo,
		}

		pageSignature, _ := p.getSignature(pageData, params.SiteGame.ApiKey)
		pageHeaders := map[string]string{
			"Content-Type": "application/json",
			"X-API-Key":    params.SiteGame.ApiId,
			"X-Signature":  pageSignature,
		}

		pageRes, err := util.CurlPostHeader(url, pageHeaders, pageData)
		if err != nil {
			g.Log().Errorf(ctx, "ONE平台拉单第%d页失败: %v", pageNo, err)
			continue
		}

		pageDataMap, ok := pageRes["data"].(map[string]interface{})
		if !ok {
			continue
		}

		transactions, ok := pageDataMap["transactions"].([]interface{})
		if !ok || len(transactions) == 0 {
			g.Log().Infof(ctx, "ONE平台第%d页无数据", pageNo)
			continue
		}

		// 7. 解析注单数据
		for _, trans := range transactions {
			transArray, ok := trans.([]interface{})
			if !ok {
				continue
			}

			record := p.parseTransaction(transArray, headerMap)
			if record != nil {
				allRecords = append(allRecords, *record)
			}
		}

		g.Log().Infof(ctx, "ONE平台拉单第%d/%d页完成，本页记录数=%d", pageNo, pageCount, len(transactions))
	}

	g.Log().Infof(ctx, "ONE平台拉单完成: 总记录数=%d", len(allRecords))

	// 批量入库到MySQL
	if len(allRecords) > 0 {
		if err := p.saveBetRecordsToMySQL(ctx, params, allRecords); err != nil {
			g.Log().Errorf(ctx, "批量入库失败: %v", err)
			return allRecords, err
		}
	}

	return allRecords, nil
}

// parseTransaction 解析注单数据
func (p *OnePlatform) parseTransaction(transArray []interface{}, headerMap map[string]int) *platform.BetRecord {
	// 只打印第一条用于调试
	//if len(transArray) > 0 {
	//	g.Log().Infof(context.Background(), "解析注单: arrayLen=%d, headerMap=%+v, firstData=%v",
	//		len(transArray), headerMap, transArray)
	//}

	getValue := func(key string) string {
		idx, ok := headerMap[key]
		if !ok {
			g.Log().Warningf(context.Background(), "字段 %s 在headerMap中不存在", key)
			return ""
		}
		if idx >= len(transArray) {
			g.Log().Warningf(context.Background(), "字段 %s 索引越界: idx=%d, arrayLen=%d", key, idx, len(transArray))
			return ""
		}
		if transArray[idx] == nil {
			return ""
		}
		return fmt.Sprintf("%v", transArray[idx])
	}

	getFloat := func(key string) float64 {
		val := getValue(key)
		if val == "" {
			return 0
		}
		f, _ := strconv.ParseFloat(val, 64)
		return f
	}

	getInt := func(key string) int {
		val := getValue(key)
		if val == "" {
			return 0
		}
		i, _ := strconv.Atoi(val)
		return i
	}

	getTimeFromMs := func(key string) string {
		idx, ok := headerMap[key]
		if !ok {
			return ""
		}
		if idx >= len(transArray) || transArray[idx] == nil {
			return ""
		}

		// 直接从interface{}转换，避免科学计数法问题
		var ms int64
		switch v := transArray[idx].(type) {
		case float64:
			ms = int64(v)
		case int64:
			ms = v
		case int:
			ms = int64(v)
		default:
			// 尝试字符串解析
			val := fmt.Sprintf("%v", v)
			f, err := strconv.ParseFloat(val, 64)
			if err != nil {
				return ""
			}
			ms = int64(f)
		}

		// 转换为时间（GMT+8）
		loc, _ := time.LoadLocation("Asia/Shanghai")
		t := time.Unix(ms/1000, (ms%1000)*1000000).In(loc)
		return t.Format("2006-01-02 15:04:05")
	}

	// 构建注单记录
	betTime := getTimeFromMs("vendorBetTime")
	settleTime := getTimeFromMs("vendorSettleTime")

	//g.Log().Infof(context.Background(), "解析注单: transactionId=%s, username=%s, betTime=%s, settleTime=%s",
	//	getValue("externalTransactionId"), getValue("username"), betTime, settleTime)

	record := &platform.BetRecord{
		OrderNo:     getValue("externalTransactionId"), // 使用externalTransactionId而不是transactionId
		Username:    getValue("username"),
		GameName:    getValue("gameCode"), // 使用gameCode而不是gameName
		BetAmount:   getFloat("betAmount"),
		WinAmount:   getFloat("winAmount"),
		ValidAmount: getFloat("effectiveTurnover"), // 使用effectiveTurnover而不是validBetAmount
		BetTime:     betTime,
		SettleTime:  settleTime,
		Status:      getInt("status"),
	}

	// 将原始数据转为JSON
	rawData, _ := json.Marshal(transArray)
	record.RawData = string(rawData)

	return record
}

// ========== 辅助方法 ==========

// getPlatformType 获取平台类型
func (p *OnePlatform) getPlatformType(mobile int32) string {
	if mobile > 0 {
		return "H5"
	}
	return "WEB"
}

// generateUUID 生成UUID
func (p *OnePlatform) generateUUID() (string, error) {
	b := make([]byte, 16)
	_, err := rand.Read(b)
	if err != nil {
		return "", err
	}

	b[6] = (b[6] & 0x0f) | 0x40
	b[8] = (b[8] & 0x3f) | 0x80

	return fmt.Sprintf("%08x-%04x-%04x-%04x-%012x",
		b[0:4], b[4:6], b[6:8], b[8:10], b[10:16]), nil
}

// getSignature 生成签名
func (p *OnePlatform) getSignature(data interface{}, secretKey string) (string, error) {
	b, err := json.Marshal(data)
	if err != nil {
		return "", err
	}

	h := hmac.New(sha256.New, []byte(secretKey))
	h.Write(b)

	return hex.EncodeToString(h.Sum(nil)), nil
}

// getBalanceClient 获取balance服务客户端
func (p *OnePlatform) getBalanceClient(ctx context.Context) (balanceV1.BalanceClient, *grpc.ClientConn, error) {
	// 从Consul获取balance服务地址
	addr, err := registry.GetServiceAddr("jh_balance_service")
	if err != nil {
		return nil, nil, fmt.Errorf("从Consul获取balance服务地址失败: %v", err)
	}

	g.Log().Infof(ctx, "balance服务地址: %s", addr)

	// 创建gRPC连接
	conn, err := grpc.NewClient(
		addr,
		grpc.WithTransportCredentials(insecure.NewCredentials()),
	)
	if err != nil {
		return nil, nil, fmt.Errorf("无法连接到balance服务: %v", err)
	}
	return balanceV1.NewBalanceClient(conn), conn, nil
}

// compensateBalance 补偿余额
func (p *OnePlatform) compensateBalance(ctx context.Context, client balanceV1.BalanceClient, siteId int, userId uint, orderNo string, amount float64) {
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
func (p *OnePlatform) markTransferFailed(ctx context.Context, client balanceV1.BalanceClient, orderNo string) {
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

//GetGameSource
/**
 * @desc：获取游戏列表并同步到数据库
 * @param ctx
 * @param params
 * @return error
 * @author : Carson
 */
func (p *OnePlatform) GetGameSource(ctx context.Context, params *platform.GetGameSourceParams) error {
	g.Log().Info(ctx, "开始同步ONE平台游戏列表...")

	// 1. 获取所有运营商（vendors）
	vendors, err := p.getVendors(ctx, params)
	if err != nil {
		return fmt.Errorf("获取运营商列表失败: %v", err)
	}

	g.Log().Infof(ctx, "获取到 %d 个运营商", len(vendors))

	// 2. 遍历每个运营商，获取游戏列表
	allGames := make([]map[string]interface{}, 0)
	for _, vendor := range vendors {
		vendorCode := vendor["code"].(string)
		g.Log().Infof(ctx, "正在获取运营商 %s 的游戏列表...", vendorCode)

		games, err := p.getGamesByVendor(ctx, params, vendorCode)
		if err != nil {
			g.Log().Errorf(ctx, "获取运营商 %s 游戏列表失败: %v", vendorCode, err)
			continue
		}

		allGames = append(allGames, games...)
		g.Log().Infof(ctx, "运营商 %s 获取到 %d 个游戏", vendorCode, len(games))
	}

	g.Log().Infof(ctx, "总共获取到 %d 个游戏", len(allGames))

	// 3. 同步到数据库
	if len(allGames) > 0 {
		err = p.syncGamesToDatabase(ctx, params, allGames)
		if err != nil {
			return fmt.Errorf("同步游戏到数据库失败: %v", err)
		}
	}

	g.Log().Info(ctx, "ONE平台游戏列表同步完成")
	return nil
}

// getVendors 获取所有运营商
func (p *OnePlatform) getVendors(ctx context.Context, params *platform.GetGameSourceParams) ([]map[string]interface{}, error) {
	traceId, _ := p.generateUUID()

	data := map[string]interface{}{
		"traceId": traceId,
	}

	signature, err := p.getSignature(data, params.SiteGame.ApiKey)
	if err != nil {
		return nil, err
	}

	headers := map[string]string{
		"Content-Type": "application/json",
		"X-API-Key":    params.SiteGame.ApiId,
		"X-Signature":  signature,
	}

	url := strings.TrimSuffix(params.SiteGame.ApiUrl, "/") + "/game/vendors"
	res, err := util.CurlPostHeader(url, headers, data)
	if err != nil {
		return nil, fmt.Errorf("请求API失败: %v", err)
	}

	status, _ := res["status"].(string)
	if status != "SC_OK" {
		return nil, fmt.Errorf("获取运营商失败: %s", status)
	}

	dataValue, ok := res["data"].([]interface{})
	if !ok {
		return nil, fmt.Errorf("响应数据格式错误")
	}

	vendors := make([]map[string]interface{}, 0)
	for _, v := range dataValue {
		vendor, ok := v.(map[string]interface{})
		if ok {
			vendors = append(vendors, vendor)
		}
	}

	return vendors, nil
}

// getGamesByVendor 获取指定运营商的所有游戏
func (p *OnePlatform) getGamesByVendor(ctx context.Context, params *platform.GetGameSourceParams, vendorCode string) ([]map[string]interface{}, error) {
	pageSize := 100
	allGames := make([]map[string]interface{}, 0)

	// 第一次请求获取总页数
	traceId, _ := p.generateUUID()
	data := map[string]interface{}{
		"traceId":    traceId,
		"vendorCode": vendorCode,
		"pageNo":     1,
		"pageSize":   pageSize,
	}

	signature, err := p.getSignature(data, params.SiteGame.ApiKey)
	if err != nil {
		return nil, err
	}

	headers := map[string]string{
		"Content-Type": "application/json",
		"X-API-Key":    params.SiteGame.ApiId,
		"X-Signature":  signature,
	}

	url := strings.TrimSuffix(params.SiteGame.ApiUrl, "/") + "/game/list"
	res, err := util.CurlPostHeader(url, headers, data)
	if err != nil {
		return nil, fmt.Errorf("请求API失败: %v", err)
	}

	status, _ := res["status"].(string)
	if status != "SC_OK" {
		return nil, fmt.Errorf("获取游戏列表失败: %s", status)
	}

	dataMap, ok := res["data"].(map[string]interface{})
	if !ok {
		return nil, fmt.Errorf("响应数据格式错误")
	}

	totalPages := 1
	if tp, ok := dataMap["totalPages"].(float64); ok {
		totalPages = int(tp)
	}

	// 遍历所有页
	for pageNo := 1; pageNo <= totalPages; pageNo++ {
		if pageNo > 1 {
			// 重新生成请求
			traceId, _ = p.generateUUID()
			data = map[string]interface{}{
				"traceId":    traceId,
				"vendorCode": vendorCode,
				"pageNo":     pageNo,
				"pageSize":   pageSize,
			}

			signature, _ = p.getSignature(data, params.SiteGame.ApiKey)
			headers["X-Signature"] = signature

			res, err = util.CurlPostHeader(url, headers, data)
			if err != nil {
				g.Log().Errorf(ctx, "获取第 %d 页失败: %v", pageNo, err)
				continue
			}

			dataMap, ok = res["data"].(map[string]interface{})
			if !ok {
				continue
			}
		}

		// 解析游戏列表
		games, ok := dataMap["games"].([]interface{})
		if !ok || len(games) == 0 {
			continue
		}

		for _, game := range games {
			gameArray, ok := game.([]interface{})
			if !ok || len(gameArray) < 8 {
				continue
			}

			// 游戏数据格式：[code, name, style, img, img1, language, device, currency]
			gameMap := map[string]interface{}{
				"type":     consts.GameTypeSlot,
				"code":     gameArray[0],
				"name":     gameArray[1],
				"style":    gameArray[2],
				"img":      gameArray[3],
				"img1":     gameArray[4],
				"language": gameArray[5],
				"device":   gameArray[6],
				"currency": gameArray[7],
				"vendor":   vendorCode,
			}

			allGames = append(allGames, gameMap)
		}
	}

	return allGames, nil
}

// syncGamesToDatabase 同步游戏到数据库
func (p *OnePlatform) syncGamesToDatabase(ctx context.Context, params *platform.GetGameSourceParams, games []map[string]interface{}) error {
	if len(games) == 0 {
		return nil
	}

	// 使用动态数据库配置
	groupName := fmt.Sprintf("site_%s", params.SiteCode)
	db := g.DB(groupName)
	if db == nil {
		return fmt.Errorf("数据库连接不存在: %s", groupName)
	}

	g.Log().Infof(ctx, "开始同步游戏到数据库，共 %d 个游戏", len(games))

	successCount := 0
	errorCount := 0
	now := time.Now()

	// 逐条处理（因为需要区分插入和更新的字段）
	for _, game := range games {
		code, _ := game["code"].(string)
		name, _ := game["name"].(string)
		style, _ := game["style"].(string)
		img, _ := game["img"].(string)
		img1, _ := game["img1"].(string)
		language, _ := game["language"].(string)
		device, _ := game["device"].(string)
		currency, _ := game["currency"].(string)
		vendor, _ := game["vendor"].(string)

		// 检查记录是否存在
		count, err := db.Model("game_list").
			Where("site_id", params.SiteId).
			Where("platform", "ONE").
			Where("code", code).
			Count()

		if err != nil {
			g.Log().Errorf(ctx, "查询游戏失败: code=%s, err=%v", code, err)
			errorCount++
			continue
		}

		if count > 0 {
			// 记录已存在：只更新三方字段，保留后台管理字段
			_, err = db.Model("game_list").
				Where("site_id", params.SiteId).
				Where("platform", "ONE").
				Where("code", code).
				Data(g.Map{
					// 只更新三方同步字段
					"name":       name,
					"style":      style,
					"img":        img,
					"img1":       img1,
					"language":   language,
					"device":     device,
					"currency":   currency,
					"vendor":     vendor,
					"updated_at": now,
					// 不更新这些字段，保持原值：
					// hot, recommend, is_new, status, sort, alias_name, like_num
				}).
				Update()

			if err != nil {
				g.Log().Errorf(ctx, "更新游戏失败: code=%s, err=%v", code, err)
				errorCount++
			} else {
				successCount++
			}
		} else {
			// 记录不存在：插入新记录，设置默认值
			gameType, _ := game["type"].(int)
			_, err = db.Model("game_list").Data(g.Map{
				// 三方字段
				"site_id":  params.SiteId,
				"platform": "ONE",
				"type":     gameType, // 添加type字段
				"code":     code,
				"name":     name,
				"style":    style,
				"img":      img,
				"img1":     img1,
				"language": language,
				"device":   device,
				"currency": currency,
				"vendor":   vendor,
				"game_id":  params.GameId,
				// 后台管理字段（新记录使用默认值）
				"alias_name": name, // 默认使用原始名称
				"hot":        0,
				"recommend":  0,
				"like_num":   0,
				"is_new":     0,
				"status":     1, // 默认启用
				"sort":       0,
				// 时间字段
				"created_at": now,
				"updated_at": now,
			}).Insert()

			if err != nil {
				g.Log().Errorf(ctx, "插入游戏失败: code=%s, err=%v", code, err)
				errorCount++
			} else {
				successCount++
			}
		}
	}

	g.Log().Infof(ctx, "游戏同步完成: 成功=%d, 失败=%d", successCount, errorCount)
	return nil
}
func (p *OnePlatform) saveBetRecordsToMySQL(ctx context.Context, params *platform.PullRecordsParams, records []platform.BetRecord) error {
	if len(records) == 0 {
		return nil
	}

	// 使用动态数据库配置
	groupName := fmt.Sprintf("site_%s", params.SiteCode)
	db := g.DB(groupName)
	if db == nil {
		return fmt.Errorf("数据库连接不存在: %s", groupName)
	}

	// 1. 批量查询 user_mapping 获取 user_id
	usernames := make([]string, 0, len(records))
	for _, record := range records {
		usernames = append(usernames, record.Username)
	}

	// 查询用户映射
	userMappings := make(map[string]uint64) // username -> user_id
	if len(usernames) > 0 {
		var mappings []struct {
			Username string `json:"username"`
			UserId   uint64 `json:"user_id"`
		}
		err := db.Model("user_mapping").
			Where("site_id", params.SiteId).
			WhereIn("username", usernames).
			Fields("username, user_id").
			Scan(&mappings)
		if err != nil {
			g.Log().Warningf(ctx, "查询用户映射失败: %v", err)
		} else {
			for _, m := range mappings {
				userMappings[m.Username] = m.UserId
			}
			g.Log().Infof(ctx, "查询到 %d 个用户映射", len(userMappings))
		}
	}

	// 2. 批量插入注单到 MySQL
	insertData := make([]g.Map, 0, len(records))
	loc, _ := time.LoadLocation("Asia/Shanghai")

	for _, record := range records {
		// 解析时间
		betTime, _ := time.ParseInLocation("2006-01-02 15:04:05", record.BetTime, loc)
		settleTime, _ := time.ParseInLocation("2006-01-02 15:04:05", record.SettleTime, loc)

		// 计算净输赢
		netAmount := record.WinAmount - record.BetAmount

		// 从映射表获取 user_id
		userId := userMappings[record.Username]
		if userId == 0 {
			g.Log().Warningf(ctx, "用户 %s 未找到映射，user_id 设为 0", record.Username)
		}

		// 获取游戏类型（转为字符串）
		gameType := ""
		if params.SiteGame.Type > 0 {
			gameType = fmt.Sprintf("%d", params.SiteGame.Type)
		}

		insertData = append(insertData, g.Map{
			"site_id":        params.SiteId,
			"site_code":      params.SiteCode,
			"user_id":        userId, // 使用映射的 user_id
			"username":       record.Username,
			"platform":       params.SiteGame.Platform,
			"currency":       "PHP",
			"game_id":        params.SiteGame.GameId,
			"game_name":      record.GameName,
			"game_type":      gameType,
			"order_no":       record.OrderNo,
			"round_id":       "", // BetRecord 没有 RoundId 字段
			"bet_amount":     record.BetAmount,
			"win_amount":     record.WinAmount,
			"valid_amount":   record.ValidAmount,
			"net_amount":     netAmount,
			"bet_time":       betTime,
			"settle_time":    settleTime,
			"status":         record.Status,
			"raw_data":       record.RawData,
			"ch_sync_status": 0,
		})
	}

	if len(insertData) == 0 {
		g.Log().Warning(ctx, "没有有效的注单数据可插入")
		return nil
	}

	// 批量插入到 bet_records 表
	_, err := db.Model("bet_records").Data(insertData).InsertIgnore()
	if err != nil {
		return fmt.Errorf("批量插入注单失败: %v", err)
	}

	g.Log().Infof(ctx, "批量入库MySQL成功: 站点=%s, 插入 %d 条", params.SiteCode, len(insertData))

	// 3. 发送 MQ 消息同步到 ClickHouse（只同步已结算的注单）
	successCount := 0
	failCount := 0

	for _, record := range records {
		// 只同步已结算的注单 (status=1)
		if record.Status != 1 {
			continue
		}

		userId := userMappings[record.Username]
		if userId == 0 {
			// 没有映射的用户跳过同步
			continue
		}

		// 获取游戏类型（转为字符串）
		gameType := ""
		if params.SiteGame.Type > 0 {
			gameType = fmt.Sprintf("%d", params.SiteGame.Type)
		}

		// 构建 MQ 消息
		mqMsg := &mq.SyncBetToClickHouseMessage{
			SiteID:      uint32(params.SiteId),
			SiteCode:    params.SiteCode,
			UserID:      userId,
			Username:    record.Username,
			Platform:    params.SiteGame.Platform,
			GameID:      uint32(params.SiteGame.GameId),
			GameName:    record.GameName,
			GameType:    gameType,
			OrderNo:     record.OrderNo,
			RoundID:     "", // BetRecord 没有 RoundId 字段
			BetAmount:   record.BetAmount,
			WinAmount:   record.WinAmount,
			ValidAmount: record.ValidAmount,
			NetAmount:   record.WinAmount - record.BetAmount,
			BetTime:     record.BetTime,
			SettleTime:  record.SettleTime,
			Status:      uint8(record.Status), // 转换为 uint8
		}

		// 发送 MQ 消息
		if err := mq.PublishBetToClickHouse(ctx, mqMsg); err != nil {
			g.Log().Errorf(ctx, "发送注单同步消息失败: order=%s, err=%v", record.OrderNo, err)
			failCount++
		} else {
			successCount++
		}
	}

	if successCount > 0 || failCount > 0 {
		g.Log().Infof(ctx, "注单同步消息发送完成: 成功=%d, 失败=%d", successCount, failCount)
	}

	return nil
}
