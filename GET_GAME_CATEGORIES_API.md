# 获取游戏种类接口

## 接口说明

获取当前站点的游戏种类列表，支持按游戏类型过滤，只返回状态为1（启用）的游戏。

**注意**: 此接口不需要用户登录，站点ID从请求的域名/header中自动识别（通过 middleware.GetSiteIdFromContext）。

## 接口信息

- **服务**: Game Service (Frontend)
- **方法**: GetGameCategories
- **协议**: gRPC
- **请求方式**: GET (通过gRPC)
- **认证**: 不需要（公开接口）

## 请求参数

| 参数名 | 类型  | 必填 | 说明                                               |
| ------ | ----- | ---- | -------------------------------------------------- |
| type   | int32 | 否   | 游戏类型（1=体育，2=彩票，3=真人，4=电子，5=棋牌） |

**注意**: `site_id` 从 context 中自动获取（通过 middleware.GetSiteIdFromContext），无需传递

## 响应参数

| 参数名        | 类型   | 说明         |
| ------------- | ------ | ------------ |
| categories    | array  | 游戏种类列表 |
| ├─ game_id    | int32  | 游戏ID       |
| ├─ name       | string | 游戏名称     |
| ├─ platform   | string | 平台         |
| ├─ type       | int32  | 游戏类型     |
| ├─ image_code | string | 图片代码     |
| ├─ db_name    | string | 数据库名称   |
| ├─ sort       | int32  | 排序（降序） |
| └─ status     | int32  | 状态         |

## 查询逻辑

1. 从 context 中获取当前站点的 `site_id`（通过 middleware.GetSiteIdFromContext）
   - 优先从 gRPC metadata 获取
   - 如果没有，从 context 获取
2. 查询 `site_game` 表
3. 过滤条件：
   - `site_id` = 从 context 获取的站点ID
   - `status` = 1（固定为启用状态）
   - `type` = 请求参数的 type（如果提供）
4. 按 `sort` 字段降序排序

## 使用示例

### 请求示例1：获取所有游戏种类

```json
{
  "type": 0
}
```

### 请求示例2：获取指定类型的游戏

```json
{
  "type": 4
}
```

### 响应示例

```json
{
  "categories": [
    {
      "game_id": 1,
      "name": "ONE游戏",
      "platform": "ONE",
      "type": 4,
      "image_code": "one_game",
      "db_name": "one_db",
      "sort": 100,
      "status": 1
    },
    {
      "game_id": 2,
      "name": "JILI游戏",
      "platform": "IG_JILI",
      "type": 4,
      "image_code": "jili_game",
      "db_name": "jili_db",
      "sort": 90,
      "status": 1
    }
  ]
}
```

## 部署步骤

### 1. 重新生成 protobuf 代码

```bash
cd jh_game_service
make proto
```

### 2. 编译服务

```bash
make build
```

### 3. 重启服务

```bash
# 停止旧服务
pkill -f jh_game_service

# 启动新服务
./main
```

## 文件修改清单

1. **jh_game_service/manifest/protobuf/frontend/game/v1/game.proto**
   - 添加 `GetGameCategories` RPC 方法
   - 添加 `GetGameCategoriesReq` 消息（只包含 type 参数）
   - 添加 `GetGameCategoriesRes` 消息
   - 添加 `GameCategory` 消息

2. **jh_game_service/internal/controller/frontend/game/game.go**
   - 添加 `GetGameCategories` 方法

3. **jh_game_service/internal/service/frontend/game.go**
   - 在 `IGame` 接口添加 `GetGameCategories` 方法签名

4. **jh_game_service/internal/logic/frontend/game/game.go**
   - 实现 `GetGameCategories` 方法
   - 使用 `middleware.GetSiteIdFromContext` 获取 site_id（不需要登录）
   - 查询 site_game 表

## 注意事项

1. **不需要登录**：此接口为公开接口，不需要用户登录即可访问
2. **site_id 自动识别**：从请求的域名或 header 中自动识别站点ID
3. 接口固定查询 `status = 1` 的游戏，不支持查询禁用的游戏
4. 结果按 `sort` 字段降序排序，sort 值越大越靠前
5. `type` 参数可选，不传则返回所有类型的游戏
6. 需要先运行 `make proto` 重新生成 protobuf 代码

## 错误处理

### 获取站点信息失败

```
获取站点信息失败
```

原因：context 中没有站点信息，可能是 Gateway 没有正确传递站点ID

### 查询失败

```
查询游戏种类失败: database error
```

原因：数据库查询出错

## 站点ID获取机制

`middleware.GetSiteIdFromContext` 的获取顺序：

1. 优先从 gRPC metadata 中获取（Gateway 传递）
2. 如果没有，从 context 中获取
3. 如果都没有，返回失败

Gateway 需要在调用时将站点ID放入 gRPC metadata 中。
