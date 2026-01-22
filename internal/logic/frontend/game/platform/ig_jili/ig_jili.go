package ig_jili

import (
	"context"
	"crypto/hmac"
	"crypto/md5"
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
	mrand "math/rand"
	"strconv"
	"strings"
	"time"
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

	//----------
	//4. 判断是否需要转账
	var needTransfer bool
	var amount float64
	needTransfer = false
	amount = 0
	fmt.Println(needTransfer, amount)

	if lastGameId == 0 {
		// 3.1 查询旧游戏余额
		oldGameBalance, balanceErr := p.GetBalance(ctx, params)
		if balanceErr != nil {
			g.Log().Errorf(ctx, "查询旧游戏余额失败: %v", balanceErr)
		} else {
			g.Log().Infof(ctx, "旧游戏余额: %.2f", oldGameBalance)
		}
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

			//// 3.2 如果旧游戏有余额，先转出到中心钱包
			//if oldGameBalance > 0 {
			//	orderNo, _ := p.generateUUID()
			//	transferErr := p.TransferOut(ctx, &platform.TransferParams{
			//		SiteId:   params.SiteId,
			//		UserId:   params.UserId,
			//		Username: params.Username,
			//		SiteCode: params.SiteCode,
			//		Amount:   oldGameBalance,
			//		Currency: params.Currency,
			//		OrderNo:  orderNo,
			//		SiteGame: params.SiteGame,
			//	})
			//	if transferErr != nil {
			//		return nil, fmt.Errorf("从旧游戏转出失败: %v", transferErr)
			//	}
			//	g.Log().Infof(ctx, "从旧游戏转出成功: %.2f", oldGameBalance)
			//
			//	// 更新用户余额
			//	userBalance += oldGameBalance
			//}
		}

		//// 3.3 将中心钱包余额全部转入新游戏
		//if userBalance > 0 {
		//	needTransfer = true
		//	amount = userBalance
		//}
	}

	////5. 执行转账
	//if needTransfer && amount > 0 {
	//	orderNo, _ := p.generateUUID()
	//	err = p.TransferIn(ctx, &platform.TransferParams{
	//		SiteId:   params.SiteId,
	//		UserId:   params.UserId,
	//		Username: params.Username,
	//		SiteCode: params.SiteCode,
	//		Amount:   amount,
	//		Currency: params.Currency,
	//		OrderNo:  orderNo,
	//		SiteGame: params.SiteGame,
	//	})
	//	if err != nil {
	//		return nil, fmt.Errorf("转入游戏失败: %v", err)
	//	}
	//	g.Log().Infof(ctx, "转入游戏成功: %.2f", amount)
	//}

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
	bodyValue, exists := res["body"]
	if !exists {
		g.Log().Errorf(ctx, "游戏API返回格式错误，完整响应: %v", res)
		return nil, fmt.Errorf("游戏API返回格式错误，缺少body字段")
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
		return nil, fmt.Errorf("游戏API返回格式错误，body字段类型不正确: %T", bodyValue)
	}

	// 解析JSON
	var dataMap map[string]interface{}
	if err := json.Unmarshal(bodyBytes, &dataMap); err != nil {
		g.Log().Errorf(ctx, "解析游戏API响应失败: %v, body: %s", err, string(bodyBytes))
		return nil, fmt.Errorf("解析游戏API响应失败: %v", err)
	}

	g.Log().Infof(ctx, "解析后的响应: %v", dataMap)

	// 检查错误码
	errorCode := 0
	if ec, ok := dataMap["ErrorCode"].(float64); ok {
		errorCode = int(ec)
	}

	if errorCode != 0 {
		message := dataMap["Message"]
		return nil, fmt.Errorf("获取游戏链接失败。错误代码: %d, 错误消息: %v", errorCode, message)
	}

	gameUrl, ok := dataMap["Data"].(string)
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
func (p *IGJiliPlatform) GetBalance(ctx context.Context, params *platform.GameParams) (float64, error) {
	buildUsername := p.buildUsername(params.Username, params.SiteCode)
	//要注册-查询余额前
	keyRegister, err := p.generateKeyRegister(buildUsername, params)
	if err != nil {
		return 0, err
	}
	// 3. 构建请求数据
	dataRegister := map[string]string{
		"Account": buildUsername,
		"AgentId": params.SiteGame.ApiId,
		"Key":     keyRegister,
	}

	// 7. 调用API
	urlRegister := strings.TrimSuffix(params.SiteGame.ApiUrl, "/") + "/CreateMember"

	resRegister, err := util.SendRequest(urlRegister, dataRegister)
	if err != nil {
		return 0, fmt.Errorf("游戏API请求失败: %v", err)
	}

	g.Log().Infof(ctx, "游戏API响应: %v", resRegister)

	if resRegister["http_code"] != 200 {
		g.Log().Errorf(ctx, "网络请求失败，请稍后再试。HTTP代码:%s", resRegister["http_code"])
		return 0, fmt.Errorf("网络请求失败，请稍后再试。HTTP代码:%s", resRegister["http_code"])
	}

	// 8. 解析响应
	bodyValue, exists := resRegister["body"]
	if !exists {
		g.Log().Errorf(ctx, "游戏API返回格式错误，缺少body字段")
		return 0, fmt.Errorf("游戏API返回格式错误，缺少body字段")
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
		return 0, fmt.Errorf("游戏API返回格式错误，body字段类型不正确: %T", bodyValue)
	}

	// 解析JSON
	var dataMap map[string]interface{}
	if err := json.Unmarshal(bodyBytes, &dataMap); err != nil {
		g.Log().Errorf(ctx, "解析游戏API响应失败: %v, body: %s", err, string(bodyBytes))
		return 0, fmt.Errorf("解析游戏API响应失败: %v", err)
	}

	if dataMap["ErrorCode"] != 0 {
		g.Log().Errorf(ctx, "创建会员失败。错误代码:%d", dataMap["ErrorCode"])
		return 0, fmt.Errorf("创建会员失败。错误代码:%d", dataMap["ErrorCode"])
	}

	g.Log().Infof(ctx, "创建IG_JILI玩家账号成功: %s", buildUsername)

	//=======获取余额============

	//要注册-查询余额前
	keyBalance, err := p.generateKeyBalance(buildUsername, params)
	if err != nil {
		return 0, err
	}
	// 3. 构建请求数据
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

	g.Log().Infof(ctx, "游戏API响应: %v", resBalance)

	if resBalance["http_code"] != 200 {
		g.Log().Errorf(ctx, "网络请求失败，请稍后再试。HTTP代码:%s", resBalance["http_code"])
		return 0, fmt.Errorf("网络请求失败，请稍后再试。HTTP代码:%s", resBalance["http_code"])
	}

	// 8. 解析响应
	bodyValueBalance, exists := resBalance["body"]
	if !exists {
		g.Log().Errorf(ctx, "游戏API返回格式错误，缺少body字段")
		return 0, fmt.Errorf("游戏API返回格式错误，缺少body字段")
	}

	// body可能是字节数组或字符串，需要解析成JSON
	var bodyBytesBalance []byte
	switch v := bodyValueBalance.(type) {
	case []byte:
		bodyBytesBalance = v
	case string:
		bodyBytesBalance = []byte(v)
	case []interface{}:
		// 字节数组形式 [123, 34, ...]
		bodyBytesBalance = make([]byte, len(v))
		for i, b := range v {
			if num, ok := b.(float64); ok {
				bodyBytesBalance[i] = byte(num)
			}
		}
	default:
		return 0, fmt.Errorf("游戏API返回格式错误，body字段类型不正确: %T", bodyValue)
	}

	// 解析JSON
	var dataMapBalance map[string]interface{}
	if err := json.Unmarshal(bodyBytesBalance, &dataMapBalance); err != nil {
		g.Log().Errorf(ctx, "解析游戏API响应失败: %v, body: %s", err, string(bodyBytesBalance))
		return 0, fmt.Errorf("解析游戏API响应失败: %v", err)
	}

	if dataMapBalance["ErrorCode"] != 0 {
		g.Log().Errorf(ctx, "创建会员失败。错误代码:%d", dataMapBalance["ErrorCode"])
		return 0, fmt.Errorf("创建会员失败。错误代码:%d", dataMapBalance["ErrorCode"])
	}

	g.Log().Infof(ctx, "创建IG_JILI玩家账号成功: %s", buildUsername)

	balance, err := p.GetJsonBalance(dataMapBalance)
	if err != nil {
		return 0, err
	}
	fmt.Println("balance =", balance)

	return balance, nil
	//
	//// 2. 生成TraceID
	//traceId, _ := p.generateUUID()
	//
	//// 3. 构建请求数据
	//data := map[string]interface{}{
	//	"username": buildUsername,
	//	"traceId":  traceId,
	//	"currency": params.Currency,
	//}
	//
	//// 4. 生成签名
	//signature, err := p.getSignature(data, params.SiteGame.ApiKey)
	//if err != nil {
	//	return 0, err
	//}
	//
	//// 5. 构建请求头
	//headers := map[string]string{
	//	"Content-Type": "application/json",
	//	"X-API-Key":    params.SiteGame.ApiId,
	//	"X-Signature":  signature,
	//}
	//
	//// 6. 调用API
	//url := strings.TrimSuffix(params.SiteGame.ApiUrl, "/") + "/cash/balance"
	//
	//g.Log().Infof(ctx, "查询余额请求: url=%s, username=%s", url, buildUsername)
	//res, err := util.CurlPostHeader(url, headers, data)
	//if err != nil {
	//	return 0, fmt.Errorf("游戏API请求失败: %v", err)
	//}
	//
	//g.Log().Infof(ctx, "查询余额响应: %v", res)
	//
	//// 7. 检查响应状态
	//status, _ := res["status"].(string)
	//switch status {
	//case "SC_OK":
	//	// 解析余额
	//	dataValue, exists := res["data"]
	//	if !exists {
	//		return 0, fmt.Errorf("游戏API返回格式错误，缺少data字段")
	//	}
	//
	//	dataMap, ok := dataValue.(map[string]interface{})
	//	if !ok {
	//		return 0, fmt.Errorf("游戏API返回格式错误，data字段类型不正确")
	//	}
	//
	//	amountAny, ok := dataMap["amount"]
	//	if !ok {
	//		return 0, fmt.Errorf("缺少字段 data.amount")
	//	}
	//
	//	// 处理不同的数字类型
	//	var amount float64
	//	switch v := amountAny.(type) {
	//	case float64:
	//		amount = v
	//	case float32:
	//		amount = float64(v)
	//	case int:
	//		amount = float64(v)
	//	case int64:
	//		amount = float64(v)
	//	default:
	//		return 0, fmt.Errorf("data.amount 类型不正确: %T", amountAny)
	//	}
	//
	//	return amount, nil
	//
	//case "SC_INVALID_OPERATOR":
	//	return 0, fmt.Errorf("未经授权的访问")
	//case "SC_INVALID_SIGNATURE":
	//	return 0, fmt.Errorf("验证失败")
	//case "SC_WRONG_CURRENCY":
	//	return 0, fmt.Errorf("不支持该货币")
	//case "SC_INVALID_REQUEST":
	//	return 0, fmt.Errorf("请求无效")
	//case "SC_UNKNOWN_ERROR":
	//	return 0, fmt.Errorf("未知错误")
	//case "SC_DUPLICATE_REQUEST":
	//	return 0, fmt.Errorf("请求重复")
	//case "SC_WALLET_NOT_SUPPORTED":
	//	return 0, fmt.Errorf("钱包类型不支持")
	//default:
	//	return 0, fmt.Errorf("查询余额失败: %s", status)
	//}
	return 0, nil
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
	//// 1. 调用balance服务扣款
	//g.Log().Info(ctx, "正在连接balance服务: jh_balance_service")
	//
	//balanceClient, conn, err := p.getBalanceClient(ctx)
	//if err != nil {
	//	return fmt.Errorf("连接balance服务失败: %v", err)
	//}
	//defer conn.Close()
	//
	//g.Log().Infof(ctx, "balance服务连接成功，准备调用AdjustBalance: siteId=%d, userId=%d, amount=%.2f",
	//	params.SiteId, params.UserId, params.Amount)
	//
	//adjustRes, err := balanceClient.AdjustBalance(ctx, &balanceV1.AdjustBalanceReq{
	//	SiteId:   int32(params.SiteId),
	//	UserId:   int32(params.UserId),
	//	Amount:   float32(params.Amount),
	//	Type:     0, // 0=转出：从中心账户转账到游戏账户
	//	GameName: params.SiteGame.Name,
	//	GameId:   strconv.Itoa(params.SiteGame.GameId),
	//})
	//
	//g.Log().Infof(ctx, "balance服务调用完成: err=%v", err)
	//
	//if err != nil {
	//	return fmt.Errorf("调用balance服务失败: %v", err)
	//}
	//
	//if !adjustRes.Success {
	//	return fmt.Errorf("扣款失败: %s", adjustRes.Message)
	//}
	//
	//g.Log().Infof(ctx, "扣款成功，订单号: %s", adjustRes.OrderNo)
	//
	//// 2. 调用三方游戏API转账
	//buildUsername := p.buildUsername(params.Username, params.SiteCode)
	//traceId, _ := p.generateUUID()
	//
	//data := map[string]interface{}{
	//	"username":       buildUsername,
	//	"traceId":        traceId,
	//	"transferAmount": params.Amount,
	//	"currency":       params.Currency,
	//	"referenceId":    params.OrderNo,
	//}
	//
	//signature, err := p.getSignature(data, params.SiteGame.ApiKey)
	//if err != nil {
	//	// 失败：补偿余额
	//	p.compensateBalance(ctx, balanceClient, params.SiteId, params.UserId, adjustRes.OrderNo, params.Amount)
	//	return err
	//}
	//
	//headers := map[string]string{
	//	"Content-Type": "application/json",
	//	"X-API-Key":    params.SiteGame.ApiId,
	//	"X-Signature":  signature,
	//}
	//
	//url := strings.TrimSuffix(params.SiteGame.ApiUrl, "/") + "/cash/deposit"
	//res, err := util.CurlPostHeader(url, headers, data)
	//if err != nil {
	//	// 失败：补偿余额
	//	p.compensateBalance(ctx, balanceClient, params.SiteId, params.UserId, adjustRes.OrderNo, params.Amount)
	//	return fmt.Errorf("游戏API请求失败: %v", err)
	//}
	//
	//// 3. 检查响应
	//status, _ := res["status"].(string)
	//if status != "SC_OK" {
	//	// 失败：补偿余额
	//	p.compensateBalance(ctx, balanceClient, params.SiteId, params.UserId, adjustRes.OrderNo, params.Amount)
	//	return fmt.Errorf("转账失败: %s", status)
	//}
	//
	//// 4. 成功：更新转账状态
	//_, err = balanceClient.UpdateTransferStatus(ctx, &balanceV1.UpdateTransferStatusReq{
	//	OrderNo: adjustRes.OrderNo,
	//})
	//if err != nil {
	//	g.Log().Errorf(ctx, "更新转账状态失败: %v", err)
	//}
	//
	//return nil
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
	//// 1. 调用balance服务创建pending订单（Type=1，不扣用户余额）
	//g.Log().Info(ctx, "正在连接balance服务: jh_balance_service")
	//
	//balanceClient, conn, err := p.getBalanceClient(ctx)
	//if err != nil {
	//	return fmt.Errorf("连接balance服务失败: %v", err)
	//}
	//defer conn.Close()
	//
	//g.Log().Infof(ctx, "balance服务连接成功，准备调用AdjustBalance: siteId=%d, userId=%d, amount=%.2f, type=1(转出)",
	//	params.SiteId, params.UserId, params.Amount)
	//
	//adjustRes, err := balanceClient.AdjustBalance(ctx, &balanceV1.AdjustBalanceReq{
	//	SiteId:   int32(params.SiteId),
	//	UserId:   int32(params.UserId),
	//	Amount:   float32(params.Amount),
	//	Type:     1, // 1=转入：从游戏账户转账到中心账户
	//	GameName: params.SiteGame.Name,
	//	GameId:   strconv.Itoa(params.SiteGame.GameId),
	//})
	//
	//g.Log().Infof(ctx, "balance服务调用完成: err=%v", err)
	//
	//if err != nil {
	//	return fmt.Errorf("调用balance服务失败: %v", err)
	//}
	//
	//if !adjustRes.Success {
	//	return fmt.Errorf("创建转出订单失败: %s", adjustRes.Message)
	//}
	//
	//g.Log().Infof(ctx, "创建转出订单成功，订单号: %s", adjustRes.OrderNo)
	//
	//// 2. 调用三方游戏API转出
	//buildUsername := p.buildUsername(params.Username, params.SiteCode)
	//traceId, _ := p.generateUUID()
	//
	//data := map[string]interface{}{
	//	"username":       buildUsername,
	//	"traceId":        traceId,
	//	"transferAmount": params.Amount,
	//	"currency":       params.Currency,
	//	"referenceId":    params.OrderNo,
	//}
	//
	//signature, err := p.getSignature(data, params.SiteGame.ApiKey)
	//if err != nil {
	//	// 失败：标记订单失败（转出失败不需要补偿余额）
	//	p.markTransferFailed(ctx, balanceClient, adjustRes.OrderNo)
	//	return err
	//}
	//
	//headers := map[string]string{
	//	"Content-Type": "application/json",
	//	"X-API-Key":    params.SiteGame.ApiId,
	//	"X-Signature":  signature,
	//}
	//
	//// 注意：转出使用 /cash/withdraw 接口
	//url := strings.TrimSuffix(params.SiteGame.ApiUrl, "/") + "/cash/withdraw"
	//res, err := util.CurlPostHeader(url, headers, data)
	//if err != nil {
	//	// 失败：标记订单失败
	//	p.markTransferFailed(ctx, balanceClient, adjustRes.OrderNo)
	//	return fmt.Errorf("游戏API请求失败: %v", err)
	//}
	//
	//// 3. 检查响应
	//status, _ := res["status"].(string)
	//if status != "SC_OK" {
	//	// 失败：标记订单失败
	//	p.markTransferFailed(ctx, balanceClient, adjustRes.OrderNo)
	//	return fmt.Errorf("转出失败: %s", status)
	//}
	//
	//// 4. 成功：更新转账状态（会增加用户余额）
	//_, err = balanceClient.UpdateTransferStatus(ctx, &balanceV1.UpdateTransferStatusReq{
	//	OrderNo: adjustRes.OrderNo,
	//})
	//if err != nil {
	//	g.Log().Errorf(ctx, "更新转账状态失败: %v", err)
	//}
	//
	//g.Log().Infof(ctx, "转出成功，已增加用户余额")
	//
	//return nil
	return nil
}

// PullBetRecords 拉取注单
func (p *IGJiliPlatform) PullBetRecords(ctx context.Context, params *platform.PullRecordsParams) ([]platform.BetRecord, error) {
	// TODO: 实现拉单逻辑
	g.Log().Warning(ctx, "ONE平台拉单功能待实现")
	return nil, fmt.Errorf("功能待实现")
}

// SupplementBetRecords 补单
func (p *IGJiliPlatform) SupplementBetRecords(ctx context.Context, params *platform.SupplementParams) error {
	// TODO: 实现补单逻辑
	g.Log().Warning(ctx, "ONE平台补单功能待实现")
	return fmt.Errorf("功能待实现")
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

// getBalanceClient 获取balance服务客户端
func (p *IGJiliPlatform) getBalanceClient(ctx context.Context) (balanceV1.BalanceClient, *grpc.ClientConn, error) {
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

// markTransferFailed 标记转账失败（用于转出失败，不需要补偿余额）
func (p *IGJiliPlatform) markTransferFailed(ctx context.Context, client balanceV1.BalanceClient, orderNo string) {
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
