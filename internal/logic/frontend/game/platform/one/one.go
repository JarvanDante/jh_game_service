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
	"jh_game_service/internal/logic/frontend/game/platform"
	"jh_game_service/internal/registry"
	"jh_game_service/internal/util"
	"math/rand"
	"strconv"
	"strings"
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
	buildUsername := p.buildUsername(params.Username, params.SiteCode)

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

		// 3.1 查询旧游戏余额
		oldGameBalance, balanceErr := p.GetBalance(ctx, params)
		if balanceErr != nil {
			g.Log().Errorf(ctx, "查询旧游戏余额失败: %v", balanceErr)
		} else {
			g.Log().Infof(ctx, "旧游戏余额: %.2f", oldGameBalance)

			// 3.2 如果旧游戏有余额，先转出到中心钱包
			if oldGameBalance > 0 {
				orderNo, _ := p.generateUUID()
				transferErr := p.TransferOut(ctx, &platform.TransferParams{
					SiteId:   params.SiteId,
					UserId:   params.UserId,
					Username: params.Username,
					SiteCode: params.SiteCode,
					Amount:   oldGameBalance,
					Currency: params.Currency,
					OrderNo:  orderNo,
					SiteGame: params.SiteGame,
				})
				if transferErr != nil {
					return nil, fmt.Errorf("从旧游戏转出失败: %v", transferErr)
				}
				g.Log().Infof(ctx, "从旧游戏转出成功: %.2f", oldGameBalance)

				// 更新用户余额
				userBalance += oldGameBalance
			}
		}

		// 3.3 将中心钱包余额全部转入新游戏
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
	buildUsername := p.buildUsername(params.Username, params.SiteCode)

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
	buildUsername := p.buildUsername(params.Username, params.SiteCode)
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
	buildUsername := p.buildUsername(params.Username, params.SiteCode)
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
	// TODO: 实现拉单逻辑
	g.Log().Warning(ctx, "ONE平台拉单功能待实现")
	return nil, fmt.Errorf("功能待实现")
}

// SupplementBetRecords 补单
func (p *OnePlatform) SupplementBetRecords(ctx context.Context, params *platform.SupplementParams) error {
	// TODO: 实现补单逻辑
	g.Log().Warning(ctx, "ONE平台补单功能待实现")
	return fmt.Errorf("功能待实现")
}

// ========== 辅助方法 ==========

// buildUsername 构建游戏用户名
func (p *OnePlatform) buildUsername(username, siteCode string) string {
	return siteCode + "_" + username
}

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

// markTransferFailed 标记转账失败（用于转出失败，不需要补偿余额）
func (p *OnePlatform) markTransferFailed(ctx context.Context, client balanceV1.BalanceClient, orderNo string) {
	// TODO: 需要在balance服务添加一个标记失败的方法
	// 暂时使用CompensateBalance，传入0金额
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
